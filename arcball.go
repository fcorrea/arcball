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
	println("Framebuffer sized")
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
	gl.PolygonMode(gl.FRONT_AND_BACK, gl.LINE)
	gl.Color4f(0.4, 0.4, 0.4, 1)
	glu.Sphere(q, 1, 10, 10)
	gl.PolygonMode(gl.FRONT_AND_BACK, gl.FILL)

	p := a.MouseIn3DSpace()

	v2 := p.Vec2()
	if v2.Len() < 1 {
		p = v2.Vec3(-math.Sqrt(1 - math.Pow(v2.Len(), 2)))
	} else {
		v2 = v2.Normalize()
		p = v2.Vec3(-math.Sqrt(1 - math.Pow(v2.Len(), 2)))
	}
	log.Println(p[2])
	// log.Println(p, p.Normalize())
	var up mgl.Vec3
	// up := p.Vec2().Vec3(0).Normalize().Cross(mgl.Vec3{0, 0, -1})
	// p[1] *= -1
	gl.PushMatrix()
	gl.LoadIdentity()

	dir := mgl.Vec2{0, 1}
	dir = mgl.Rotate2D(glfw.GetTime()).Mul2x1(dir)
	// log.Println(unitY)

	glfw.GetTime()
	_ = math.Cos

	up = mgl.Vec3{0, 2, 0}
	// p = mgl.Vec3{dir.X(), dir.Y(), math.Sin(glfw.GetTime())}
	up = p.Cross(mgl.Vec3{0, 0, -1})
	eye := mgl.Vec3{0, 0, 0}
	// up = p.Cross(mgl.Vec3{0, 0, -1})
	// log.Println("Dir =", p, "Up =", up)

	// lookat := mgl.QuatLookAtV(mgl.Vec3{0, 0, 0}, p, up)

	lookat := mgl.Ident4()
	lookat = mgl.LookAtV(eye, p, up).Mul4(lookat)
	lookat = mgl.HomogRotate3DY(mgl.DegToRad(90)).Mul4(lookat)
	lookat = mgl.HomogRotate3DZ(mgl.DegToRad(180)).Mul4(lookat)
	lookat = lookat.Mul4(a.Rotation().Normalize().Mat4())
	// lookat = mgl.HomogRotate3DX(mgl.DegToRad(90)).Mul4(lookat)
	// LoadQuat(mgl.QuatSlerp(mgl.QuatIdent(), lookat, 10))
	// LoadQuat(lookat)
	LoadMatrix(lookat)
	// log.Println("mgl =", lookat)
	_ = lookat

	// glu.LookAt(0, 0, 0, up.X(), up.Y(), up.Z(), p.X(), p.Y(), p.Z())

	// m := GetMatrix(gl.MODELVIEW_MATRIX)
	// log.Println("gl =", m)
	glh.DrawAxes()

	gl.PopMatrix()

	// p[1] *= -1
	gl.PushMatrix()
	gl.LoadIdentity()
	// gl.Rotated(90, 1, 0, 0)

	gl.PointSize(10)
	gl.Begin(gl.POINTS)

	gl.Color3d(1, 0, 0)
	Vertex(p)

	gl.End()

	gl.PopMatrix()

}

func (a *Arcball) Rotation() mgl.Quat {
	return a.dragged.Mul(a.rotation).Normalize()
}
