package main

import (
	"log"

	"github.com/go-gl/gl"
	glfw "github.com/go-gl/glfw3"
	"github.com/go-gl/glh"
	// "github.com/go-gl/glu"
	mgl "github.com/go-gl/mathgl/mgl64"
)

func MakeCube() *glh.MeshBuffer {
	cube := glh.NewMeshBuffer(
		glh.RenderBuffered,
		glh.NewPositionAttr(3, gl.FLOAT, gl.STATIC_DRAW),
	)

	cube.Add([]float32{
		-1, -1, -1, -1, -1, 1,
		-1, -1, -1, -1, 1, -1,
		-1, -1, -1, 1, -1, -1,

		-1, 1, 1, -1, -1, 1,
		-1, 1, 1, -1, 1, -1,
		-1, 1, 1, 1, 1, 1,

		1, -1, 1, -1, -1, 1,
		1, -1, 1, 1, -1, -1,
		1, -1, 1, 1, 1, 1,

		1, 1, -1, -1, 1, -1,
		1, 1, -1, 1, -1, -1,
		1, 1, -1, 1, 1, 1,
	})

	return cube
}

func main() {

	const (
		startFullscreen = false
		windowTitle     = "ToggleFullscreen"
		defW            = 1280
		defH            = 768
	)

	window, err := NewWindow(windowTitle, startFullscreen, defW, defH, nil)
	if err != nil {
		log.Fatalln("Unable to create window: ", err)
	}
	defer window.Destroy()

	window.SetFramebufferSizeCallback(func(_ *glfw.Window, w, h int) {
		gl.Viewport(0, 0, w, h)

		gl.MatrixMode(gl.PROJECTION)
		gl.LoadIdentity()

		ratio := float64(h) / float64(w)
		// gl.Frustum(-1, 1, -ratio, ratio, 4, 100)
		const s = 10
		gl.Ortho(-1*s, 1*s, -ratio*s, ratio*s, -100, 100)

		gl.Translated(0, 0, -20)
	})

	arcball := NewArcball(window.Window)

	cube := MakeCube()

	// ident := mgl.QuatRotate(0.2, mgl.Vec3{1, 0, 0})
	ident := mgl.QuatIdent()
	delta := mgl.QuatRotate(0, mgl.Vec3{0, 1, 0})

	for !window.ShouldClose() {

		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		gl.MatrixMode(gl.MODELVIEW)
		gl.LoadIdentity()

		delta = delta
		i := [16]float64(ident.Mul(arcball.Rotation()).Normalize().Mat4())
		gl.MultMatrixd(&i)

		gl.Color4f(1, 1, 1, 1)
		cube.Render(gl.LINES)
		// _ = cube

		arcball.Draw()
		// gl.LoadIdentity()
		// i = [16]float64(arcball.Rotation().Mat4())
		// gl.MultMatrixd(&i)
		// gl.Rotated(90, 1, 0, 0)

		window.SwapBuffers()
		glfw.PollEvents()
	}
}
