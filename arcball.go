package main

import (
	"log"
	"unsafe"

	gl "github.com/go-gl/gl"
	glfw "github.com/go-gl/glfw3"
	"github.com/go-gl/glh"
	"github.com/go-gl/glu"
	mgl "github.com/go-gl/mathgl/mgl64"
)

var _ = log.Println

type Arcball struct {
	w, h                float64
	oldCursorPositioned glfw.CursorPositionCallback
	oldFramebufferSized glfw.FramebufferSizeCallback
	oldMouseClicked     glfw.MouseButtonCallback

	mouse struct {
		pos     mgl.Vec2
		down    mgl.Vec2
		pressed bool
	}

	rotation, dragged mgl.Quat
}

func NewArcball(window *glfw.Window) *Arcball {

	w, h := window.GetFramebufferSize()

	a := &Arcball{
		w:        float64(w),
		h:        float64(h),
		rotation: mgl.QuatIdent(),
		dragged:  mgl.QuatIdent(),
	}

	a.oldFramebufferSized = window.SetFramebufferSizeCallback(a.FramebufferSized)
	a.oldCursorPositioned = window.SetCursorPositionCallback(a.CursorPositioned)
	a.oldMouseClicked = window.SetMouseButtonCallback(a.mouseClicked)

	return a
}

func (a *Arcball) mouseVector(x, y float64) mgl.Vec2 {
	return mgl.Vec2{(x / a.w * 2) - 1, -((y / a.h * 2) - 1)}
}

func (a *Arcball) mouseClicked(w *glfw.Window, button glfw.MouseButton, action glfw.Action, mod glfw.ModifierKey) {

	switch button {
	case glfw.MouseButtonLeft:
		switch action {
		case glfw.Press:
			a.mouse.down = a.mouseVector(w.GetCursorPosition())
			a.mouse.pressed = true

		case glfw.Release:
			a.mouse.pressed = false
			a.rotation = a.Rotation()
			a.dragged = mgl.QuatIdent()
		}
	}

	if a.oldMouseClicked != nil {
		a.oldMouseClicked(w, button, action, mod)
	}
}

func (a *Arcball) FramebufferSized(win *glfw.Window, w, h int) {
	a.w, a.h = float64(w), float64(h)

	if a.oldFramebufferSized != nil {
		a.oldFramebufferSized(win, w, h)
	}
}

func SpherePoint(in mgl.Vec2) mgl.Vec3 {
	const s = 0.5
	p := mgl.Vec3{s * in.X(), s * in.Y(), 0}
	if p.Len() > 1 {
		p = p.Normalize()
	} else {

		p[2] = 1 - p.Len()
	}
	return p
}

func (a *Arcball) CursorPositioned(w *glfw.Window, x, y float64) {

	log.Printf("%+v", a.mouse)
	if a.mouse.pressed {
		a.mouse.pos = a.mouseVector(x, y)

		start := SpherePoint(a.mouse.down)
		end := SpherePoint(a.mouse.pos)
		perp := start.Cross(end)
		log.Println("Perp:", perp)
		mag := start.Dot(end)

		a.dragged = mgl.QuatRotate(mag*2, perp)
	}

	if a.oldCursorPositioned != nil {
		a.oldCursorPositioned(w, x, y)
	}
}

var q unsafe.Pointer

func (a *Arcball) Draw() {
	if q == nil {
		q = glu.NewQuadric()
	}
	gl.PolygonMode(gl.FRONT_AND_BACK, gl.LINE)
	gl.Color4f(0.4, 0.4, 0.4, 1)
	// glu.Sphere(q, 7.5, 20, 20)
	gl.PolygonMode(gl.FRONT_AND_BACK, gl.FILL)

	glh.DrawAxes()

	gl.Begin(gl.POINTS)
	gl.Vertex3f()
	// glh.DebugLines()

}

func (a *Arcball) Rotation() mgl.Quat {
	return a.dragged.Mul(a.rotation).Normalize()
}
