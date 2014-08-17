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

	mouse struct {
		coordx, coordy float64
		p              mgl.Vec2
		pos            mgl.Vec2
		down           mgl.Vec2
		pressed        bool
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

	a.FramebufferSized(window, w, h)

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

func (a *Arcball) CursorPositioned(w *glfw.Window, x, y float64) {

	a.mouse.coordx, a.mouse.coordy = w.GetCursorPosition()

	if a.mouse.pressed {
		a.mouse.pos = a.mouseVector(x, y)

		start := SpherePoint(a.mouse.down)
		end := SpherePoint(a.mouse.pos)
		perp := start.Cross(end)
		log.Println("Perp:", perp)
		mag := start.Dot(end)

		a.dragged = mgl.QuatRotate(mag/2000, perp)
	} else {
		a.mouse.p = mgl.Vec2{x, y}
	}

	if a.oldCursorPositioned != nil {
		a.oldCursorPositioned(w, x, y)
	}
}

func (a *Arcball) MouseIn3DSpace() mgl.Vec3 {
	win := a.mouse.p.Vec3(0)
	win[1] = a.h - win[1]

	ident := mgl.Ident4()
	unproj, err := mgl.UnProject(win, ident, a.Projection, 0, 0, int(a.w), int(a.h))
	if err != nil {
		log.Panic("Error in UnProject:", err)
	}

	// The point doesn't show up if you draw it exactly on the camera plane.
	unproj[2] *= 0.9999999

	return unproj
}

var q unsafe.Pointer

func (a *Arcball) Reset() {
	a.rotation = mgl.QuatIdent()
}

func (a *Arcball) Draw() {
	if q == nil {
		q = glu.NewQuadric()
	}

	drawSphere := func() {
		glh.With(glh.Attrib{gl.ENABLE_BIT | gl.POLYGON_BIT}, func() {
			gl.PolygonMode(gl.FRONT_AND_BACK, gl.LINE)
			gl.Enable(gl.BLEND)
			// glBlendFunc(GL_SRC_ALPHA, GL_ONE_MINUS_SRC_ALPHA)
			gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
			// gl.Color4f(0.4, 0.4, 0.4, 0.5)
			gl.LineWidth(2)
			gl.Color4f(0.75, 0.75, 0.75, 0.25)
			glu.Sphere(q, 1, 20, 20)
		})
	}

	p := a.MouseIn3DSpace()
	// p[0] = 0
	// t := glfw.GetTime()
	// p = mgl.Vec3{math.Sin(t), math.Cos(t), 0}
	// p = p.Mul(0.75)

	v2 := p.Vec2()
	if v2.Len() < 1 {
		p = v2.Vec3(-math.Sqrt(1 - math.Pow(v2.Len(), 2)))
	} else {
		v2 = v2.Normalize()
		p = v2.Vec3(0)
	}
	// p[2] = -1

	var up mgl.Vec3

	glfw.GetTime()
	_ = math.Cos

	// v := math.Cos(math.Pi * (1 + p.Y()) / 2)
	yPos := mgl.Clamp(p.Y(), -1, 1)
	//theta := math.Asin(yPos) + math.Pi/2
	//v := math.Cos(theta)

	up = mgl.Vec3{0, -1, -yPos}
	up = p.Cross(up)
	up = p.Cross(up)
	up = up.Normalize()
	// up = p.Cross(mgl.Vec3{0, 0, -1})
	eye := mgl.Vec3{0, 0, 0}
	// p = mgl.Vec3{-p.X(), -p.Y(), p.Z()}

	initDir := up.Cross(mgl.Vec3{1, 0, 0})
	rA := mgl.QuatLookAtV(eye, initDir, up)
	rB := mgl.QuatLookAtV(eye, p, up)
	t, err := glfw.GetTime()
	if err != nil {
		panic(err)
	}
	amt := 0.5 + math.Sin(t*4)/2
	amt = 1

	theRotation := mgl.QuatSlerp(rA, rB, amt).Normalize().Mat4()

	p = mgl.Vec3{p.X(), p.Y(), -p.Z()}
	up = mgl.Vec3{up.X(), up.Y(), -up.Z()}

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
			// _ = theRotation
			glh.DrawAxes()
			drawSphere()
		})
	}

	// Draw the three views
	show(func() {}, func() {})

	show(func() {
		gl.Translated(-3, 0, 0)
	}, func() {
		gl.Rotated(90, 0, 1, 0)
	})

	show(func() {
		gl.Translated(3, 0, 0)
	}, func() {
		gl.Rotated(90, 1, 0, 0)
	})

	// Draw the points in the three views without applying the lookAt rotation
	glh.With(glh.Matrix{gl.MODELVIEW}, func() {
		gl.LoadIdentity()

		gl.PointSize(10)

		pts := func() {
			glh.With(glh.Primitive{gl.POINTS}, func() {
				gl.Color3d(1, 0, 0)
				Vertex(p)
				gl.Color3d(0, 1, 0)
				Vertex(eye)
				gl.Color3d(0, 0, 1)
				Vertex(up)
				newPt := theRotation.Mul4x1(mgl.Vec4{0, 0, ah1, 0})
				gl.Color3d(0, 1, 1)
				Vertex(newPt.Vec3())
			})

			gl.Color3d(1, 1, 0.25)
			// Draw white line to red dot
			glh.With(glh.Primitive{gl.LINES}, func() {
				Vertex(mgl.Vec3{0, 0, 0})
				Vertex(p)
			})
			// Draw "up" from red dot
			glh.With(glh.Primitive{gl.LINES}, func() {
				Vertex(p)
				Vertex(p.Add(up.Mul(0.25)))
			})
		}

		pts()

		gl.LoadIdentity()
		gl.Translated(-3, 0, 0)
		gl.Rotated(90, 0, 1, 0)
		pts()

		gl.LoadIdentity()
		gl.Translated(3, 0, 0)
		gl.Rotated(90, 1, 0, 0)
		pts()

	})

}

func (a *Arcball) Rotation() mgl.Quat {
	return a.dragged.Mul(a.rotation).Normalize()
}
