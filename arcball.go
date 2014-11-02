package main

import (
	"log"
	"math"
	"unsafe"

	gl "github.com/go-gl/gl"
	glfw "github.com/go-gl/glfw3"
	"github.com/go-gl/glh"
	"github.com/go-gl/glu"
	mgl "github.com/go-gl/mathgl/mgl64"
)

var _ = log.Println

type Arcball struct {
	Projection mgl.Mat4

	w, h                float64
	oldCursorPositioned glfw.CursorPositionCallback
	oldFramebufferSized glfw.FramebufferSizeCallback
	oldMouseClicked     glfw.MouseButtonCallback
	oldKeyPressed       glfw.KeyCallback

	mouse struct {
		p       mgl.Vec2
		down    mgl.Vec2
		pressed bool
	}

	sphereSize float64

	initRotation, currentRotation, rotation, dragged mgl.Quat
}

func NewArcball(window *glfw.Window) *Arcball {

	w, h := window.GetFramebufferSize()

	a := &Arcball{
		w:          float64(w),
		h:          float64(h),
		rotation:   mgl.QuatIdent(),
		dragged:    mgl.QuatIdent(),
		sphereSize: 4,
	}

	a.oldFramebufferSized = window.SetFramebufferSizeCallback(a.FramebufferSized)
	a.oldCursorPositioned = window.SetCursorPositionCallback(a.CursorPositioned)
	a.oldMouseClicked = window.SetMouseButtonCallback(a.mouseClicked)
	a.oldKeyPressed = window.SetKeyCallback(a.KeyPressed)

	a.FramebufferSized(window, w, h)

	return a
}

func (a *Arcball) KeyPressed(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	if a.oldKeyPressed != nil {
		a.oldKeyPressed(w, key, scancode, action, mods)
	}

	switch key {
	case glfw.KeyR:
		if action == glfw.Press {
			a.rotation = mgl.QuatIdent()
		}
	}
}

func (a *Arcball) mouseVector(x, y float64) mgl.Vec2 {
	return mgl.Vec2{(x / a.w * 2) - 1, -((y / a.h * 2) - 1)}
}

func (a *Arcball) mouseClicked(w *glfw.Window, button glfw.MouseButton, action glfw.Action, mod glfw.ModifierKey) {

	switch button {
	case glfw.MouseButtonLeft:
		switch action {
		case glfw.Press:
			// log.Println("Press")
			x, y := w.GetCursorPosition()
			a.mouse.p = mgl.Vec2{x, y}

			a.mouse.pressed = true

			a.initRotation = a.RotateToMouse(a.mouse.p)
			a.currentRotation = a.initRotation

		case glfw.Release:
			// log.Println("Release")
			a.mouse.pressed = false
			a.rotation = a.Rotation()

			a.dragged = mgl.QuatIdent()
			a.initRotation = mgl.QuatIdent()
			a.currentRotation = mgl.QuatIdent()
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

	a.Projection = GetMatrix(gl.PROJECTION_MATRIX)
}

func SpherePoint(in mgl.Vec2) mgl.Vec3 {
	// const s = 0.5
	p := mgl.Vec3{in.X(), in.Y(), 0}
	if p.Len() > 1 {
		p = p.Normalize()
	} else {

		p[2] = 1 - p.Len()
	}
	return p.Mul(10)
}

func (a *Arcball) RotateToMouse(mousePosition mgl.Vec2) mgl.Quat {
	p := a.MouseIn3DSpace(mousePosition)
	p = MouseOnSphere(p.Vec2(), a.sphereSize)

	// I don't understand why I need to flip this coordinate :(
	pPrime := p.Vec2().Vec3(-p.Z())
	return QuatLookAtV(mgl.Vec3{0, 0, 0}, pPrime)
}

func (a *Arcball) CursorPositioned(w *glfw.Window, x, y float64) {

	a.mouse.p = mgl.Vec2{x, y}

	if a.mouse.pressed {
		// log.Println("Position")
		a.currentRotation = a.RotateToMouse(a.mouse.p)
		a.dragged = a.initRotation.Inverse().Mul(a.currentRotation).Normalize()
	}

	if a.oldCursorPositioned != nil {
		a.oldCursorPositioned(w, x, y)
	}
}

func (a *Arcball) MouseIn3DSpace(coord mgl.Vec2) mgl.Vec3 {
	win := coord.Vec3(0)
	win[1] = a.h - win[1]

	ident := mgl.Ident4()
	unproj, err := mgl.UnProject(win, ident, a.Projection, 0, 0, int(a.w), int(a.h))
	if err != nil {
		log.Panic("Error in UnProject:", err)
	}

	// The point doesn't show up if you draw it exactly on the camera plane.
	unproj[2] *= 0.9999999
	log.Println(unproj)

	return unproj
}

var q unsafe.Pointer

func (a *Arcball) Reset() {
	a.rotation = mgl.QuatIdent()
}

func MouseOnSphere(v mgl.Vec2, r float64) mgl.Vec3 {
	// Put the Z coordinate of the mouse on the sphere
	if v.Len() < r {
		// Inside sphere
		return v.Vec3(math.Sqrt(math.Pow(r, 2) - math.Pow(v.Len(), 2)))
	} else {
		// Outside sphere, clamp it there.
		return v.Normalize().Vec3(0).Mul(r)
	}
}

func (a *Arcball) Draw() {
	if q == nil {
		q = glu.NewQuadric()
	}
	const r = 4

	drawSphere := func() {
		glh.With(glh.Attrib{gl.ENABLE_BIT | gl.POLYGON_BIT}, func() {
			gl.PolygonMode(gl.FRONT_AND_BACK, gl.LINE)
			gl.Enable(gl.BLEND)
			// glBlendFunc(GL_SRC_ALPHA, GL_ONE_MINUS_SRC_ALPHA)
			gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
			// gl.Color4f(0.4, 0.4, 0.4, 0.5)

			gl.LineWidth(2)
			gl.Color4f(0.75, 0.75, 0.75, 0.05)
			glu.Sphere(q, r, 8*4, 8*4)
		})
	}

	p := a.MouseIn3DSpace(a.mouse.p)
	p = MouseOnSphere(p.Vec2(), r)

	eye := mgl.Vec3{0, 0, 0}

	// I don't understand why I need to flip this coordinate :(
	pPrime := p.Vec2().Vec3(-p.Z())
	theRotation := QuatLookAtV(eye, pPrime).Mat4()

	// t, _ := glfw.GetTime()
	// x := (1 + math.Sin(t)) / 2
	// x = 1
	// a.dragged = mgl.QuatNlerp(a.initRotation, a.currentRotation, 0*x)
	// a.dragged = a.currentRotation.Sub(a.initRotation).Normalize()
	// log.Println(a.initRotation, a.currentRotation, a.dragged)
	theRotation = a.Rotation().Mat4()

	// Twist
	// t, _ := glfw.GetTime()
	// theRotation = theRotation.Mul4(mgl.Rotate3DZ(t).Mat4())

	showAxes := func(rot func()) {

		glh.With(glh.Matrix{gl.MODELVIEW}, func() {
			const s = 0.5
			gl.Rotated(10, -1, 1, 0)
			gl.Translated(-1, -1, 0)
			gl.Scaled(s, s, s)
			rot()
			glh.DrawAxes()
		})

		rot()
	}

	show := func(mov, rot func()) {
		glh.With(glh.Matrix{gl.MODELVIEW}, func() {
			gl.LoadIdentity()
			mov()
			showAxes(rot)

			MulMatrix(theRotation)
			glh.DrawAxes()
			drawSphere()
		})
	}
	glh.With(glh.Matrix{gl.PROJECTION}, func() {

		// Look at the point pointed at by the mouse
		// eyePos := mgl.Translate3D(0, 0, 20)
		// rot := QuatLookAtV(eyePos.Col(3).Vec3(), p).Mat4()
		// MulMatrix(eyePos.Mul4(rot).Mul4(eyePos.Inv()))

		// Draw the three views
		show(func() {}, func() {})

		// show(func() {
		// 	gl.Translated(-3, 0, 0)
		// }, func() {
		// 	gl.Rotated(90, 0, 1, 0)
		// })

		// show(func() {
		// 	gl.Translated(3, 0, 0)
		// }, func() {
		// 	gl.Rotated(90, 1, 0, 0)
		// })
	})

}

func (a *Arcball) Rotation() mgl.Quat {
	return a.dragged.Mul(a.rotation).Normalize()
}

func QuatLookAtV(eye, center mgl.Vec3) mgl.Quat {
	forward := center.Sub(eye).Normalize()

	initialForwardDirection := mgl.Vec3{0, 0, -1}
	dot := initialForwardDirection.Dot(forward)

	angle := float64(math.Acos(float64(dot)))
	rotationAxis := initialForwardDirection.Cross(forward)
	return mgl.QuatRotate(-angle, rotationAxis.Normalize()).Normalize()
}
