package main

import (
	"fmt"
	"log"
	"runtime"

	"github.com/go-gl/gl"
	glfw "github.com/go-gl/glfw3"
)

type Window struct {
	*glfw.Window
	title      string
	fullscreen bool
	w, h       int
}

var firstWindow = true

func NewWindow(title string, fullscreen bool, w, h int, oldWindow *Window) (*Window, error) {
	var err error
	var window, oldW *glfw.Window

	if oldWindow != nil {
		oldW = oldWindow.Window
	}

	if firstWindow {
		// We're the first window to be constructed.
		// LockOSThread.
		runtime.LockOSThread()

		err := glfw.Init()
		if err != nil {
			return nil, fmt.Errorf("glfw.Init() failed: %q", err)
		}
		// defer glfw.Terminate()

		glfw.SwapInterval(1)
	}

	if fullscreen {
		window, err = FullscreenWindow(title, nil, nil, oldW)
	} else {
		window, err = glfw.CreateWindow(w, h, title, nil, oldW)
	}

	if firstWindow {
		window.MakeContextCurrent()

		status := gl.Init()
		if status != 0 {
			return nil, fmt.Errorf("gl.Init() failed, status =", status)
		}
		firstWindow = false
	}

	ret := &Window{window, title, fullscreen, w, h}
	ret.SetKeyCallback(nil) // default key handler
	return ret, err
}

func (w *Window) Destroy() {
	// Needed so that `defer w.Destroy()` doesn't close over the original
	// w.Window (before we made it a enew one, for example).
	w.Window.Destroy()
}

// Replace `window` with `newWindow`, copying across all of the callbacks.
// Also makes `newWindow` the current OpenGL context and calls the
// framebufferSizeCallback.
func ReplaceWindow(window **glfw.Window, newWindow *glfw.Window) {
	newWindow.SetKeyCallback(window.SetKeyCallback(nil))
	newWindow.SetCharacterCallback(window.SetCharacterCallback(nil))
	newWindow.SetMouseButtonCallback(window.SetMouseButtonCallback(nil))
	newWindow.SetCursorPositionCallback(window.SetCursorPositionCallback(nil))
	newWindow.SetCursorEnterCallback(window.SetCursorEnterCallback(nil))
	newWindow.SetScrollCallback(window.SetScrollCallback(nil))
	newWindow.SetPositionCallback(window.SetPositionCallback(nil))
	newWindow.SetCloseCallback(window.SetCloseCallback(nil))
	newWindow.SetRefreshCallback(window.SetRefreshCallback(nil))
	newWindow.SetFocusCallback(window.SetFocusCallback(nil))
	newWindow.SetIconifyCallback(window.SetIconifyCallback(nil))

	sizeCallback := window.SetSizeCallback(nil)
	newWindow.SetSizeCallback(sizeCallback)
	framebufferSizeCallback := window.SetFramebufferSizeCallback(nil)
	newWindow.SetFramebufferSizeCallback(framebufferSizeCallback)

	window.Destroy()

	*window = newWindow
	window.MakeContextCurrent()

	if framebufferSizeCallback != nil {
		w, h := window.GetFramebufferSize()
		framebufferSizeCallback(*window, w, h)
	}
	if sizeCallback != nil {
		w, h := window.GetSize()
		sizeCallback(*window, w, h)
	}
}

func FullscreenWindow(title string, monitor *glfw.Monitor, mode *glfw.VideoMode, oldWindow *glfw.Window) (*glfw.Window, error) {
	var err error

	if monitor == nil {
		monitor, err = glfw.GetPrimaryMonitor()
		if err != nil {
			return nil, fmt.Errorf("failed to obtain primary monitor: ", err)
		}
	}

	if mode == nil {
		mode, err = monitor.GetVideoMode()
		if err != nil {
			return nil, fmt.Errorf("failed to obtain video mode:", err)
		}
	}

	glfw.WindowHint(glfw.RefreshRate, mode.RefreshRate)
	glfw.WindowHint(glfw.Samples, 4)

	window, err := glfw.CreateWindow(mode.Width, mode.Height, title, monitor, oldWindow)
	if err != nil {
		return nil, fmt.Errorf("failed to create window:", err)
	}

	return window, nil
}

func (w *Window) ToggleFullscreen() {
	w.fullscreen = !w.fullscreen

	newWindow, err := NewWindow(w.title, w.fullscreen, w.w, w.h, w)
	if err != nil {
		log.Fatalln("Unable to create window in ToggleFullscreen(): ", err)
	}

	ReplaceWindow(&w.Window, newWindow.Window)
}

func (win *Window) SetFramebufferSizeCallback(cb func(_ *glfw.Window, w, h int)) {
	w, h := win.GetFramebufferSize()
	if !win.fullscreen {
		// Fullscren mode automatically calls this in my testing
		cb(win.Window, w, h)
	}
	win.Window.SetFramebufferSizeCallback(cb)
}

func (win *Window) SetKeyCallback(cb func(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey)) {
	win.Window.SetKeyCallback(func(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
		if cb != nil {
			defer cb(w, key, scancode, action, mods)
		}

		switch key {
		case glfw.KeyEscape, glfw.KeyQ:
			w.SetShouldClose(true)

		case glfw.KeyF11, glfw.KeyF:
			if action == glfw.Press {
				win.ToggleFullscreen()
			}
		}
	})
}
