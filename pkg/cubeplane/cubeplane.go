// Copyright © 2018 Cove Schneider
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cubeplane

import (
	"fmt"
	"path"
	"strconv"

	"github.com/g3n/engine/core"

	"github.com/g3n/engine/renderer"

	"github.com/g3n/engine/gls"

	"github.com/g3n/engine/gui"

	"github.com/g3n/engine/window"

	"github.com/g3n/engine/geometry"
	"github.com/g3n/engine/graphic"
	"github.com/g3n/engine/light"
	"github.com/g3n/engine/material"
	"github.com/g3n/engine/math32"
	"github.com/g3n/engine/util/application"
)

type CubePlane struct {
	app  *application.Application
	size int64

	command string
	Header  []string

	cursorX      int64
	cursorY      int64
	selectedNode *core.Node

	rc *core.Raycaster

	plane [][]*core.Node

	rotate bool
}

type CubeData struct {
	attrs []string
	locX  int64
	locY  int64
}

func Init(app *application.Application, cmd string) *CubePlane {

	// Add lights to the scene
	ambientLight := light.NewAmbient(&math32.Color{1.0, 1.0, 1.0}, 0.8)
	app.Scene().Add(ambientLight)
	pointLight := light.NewPoint(&math32.Color{1, 1, 1}, 5.0)
	pointLight.SetPosition(1, 0, 2)
	app.Scene().Add(pointLight)

	// Add an axis helper to the scene
	//axis := graphic.NewAxisHelper(1.5)
	//app.Scene().Add(axis)

	app.CameraPersp().SetPosition(0, -15, 10)
	app.CameraPersp().LookAt(&math32.Vector3{0, 0, 0})

	// init stuff for hud
	gs, err := gls.New()
	if err != nil {
		panic(err)
	}
	gui.NewRoot(gs, app.Window())
	root := gui.NewRoot(gs, app.Window())

	l1 := gui.NewLabel("oq command: " + cmd)
	width, _ := app.Window().Size()
	l1.SetPosition(float32(width)-230, 10)
	l1.SetPaddings(2, 2, 2, 2)
	l1.SetFontSize(12.0)
	root.Add(l1)

	// why doesn't this resize adjust the text location?
	root.Subscribe(window.OnWindowSize, func(evname string, ev interface{}) {
		width, _ := app.Window().Size()
		l1.SetPosition(float32(width)-130, 10)
	})

	// Creates a renderer and adds default shaders
	rend := renderer.NewRenderer(gs)
	err = rend.AddDefaultShaders()
	if err != nil {
		panic(err)
	}
	app.SetGui(root)

	// Sets window background color
	gs.ClearColor(0.0394, 0.1601, 0.1983, 1.0)

	app.TimerManager.Initialize()

	// default plane size
	size := int64(20)

	c := &CubePlane{
		app:     app,
		size:    size,
		command: cmd,
	}

	app.Window().Subscribe(window.OnKeyDown, func(evname string, ev interface{}) {
		c.onKey(evname, ev)
	})

	app.Window().Subscribe(window.OnKeyRepeat, func(evname string, ev interface{}) {
		c.onKey(evname, ev)
	})

	c.app.SubscribeID(application.OnAfterRender, 1, func(evname string, ev interface{}) {
		if c.rotate {
			c.app.Scene().RotateOnAxis(&math32.Vector3{0, 0, 1}, -.03)
		}
	})

	c.rc = core.NewRaycaster(&math32.Vector3{}, &math32.Vector3{})
	app.Window().Subscribe(window.OnMouseDown, func(evname string, ev interface{}) {
		c.onMouse(evname, ev)
	})

	c.initCubePlane(size)

	return c
}

func (cp *CubePlane) onMouse(evname string, ev interface{}) {
	// Convert mouse coordinates to normalized device coordinates
	mev := ev.(*window.MouseEvent)
	width, height := cp.app.Window().Size()
	// Linux and Windows
	//x := 2*(mev.Xpos/float32(width)) - 1
	//y := -2*(mev.Ypos/float32(height)) + 1

	// OSX
	x := 1*(mev.Xpos/float32(width)) - 1
	y := -1*(mev.Ypos/float32(height)) + 1

	// Set the raycaster from the current camera and mouse coordinates
	_ = cp.app.Camera().SetRaycaster(cp.rc, x, y)
	//fmt.Printf("rc:%+v\n", cp.rc.Ray)

	// Checks intersection with all objects in the scene
	intersects := cp.rc.IntersectObjects(cp.app.Scene().Children(), true)
	//fmt.Printf("intersects:%+v\n", intersects)
	if len(intersects) == 0 {
		return
	}

	obj := intersects[0].Object
	ig, ok := obj.(graphic.IGraphic)
	if !ok {
		return
	}

	node := ig.GetNode().Parent().GetNode()
	ud, ok := node.UserData().(CubeData)
	if ok {
		cp.cursorX = ud.locX
		cp.cursorY = ud.locY
	}

	cp.updateSelected()
}

func (cp *CubePlane) onKey(evname string, ev interface{}) {

	key := ev.(*window.KeyEvent)
	switch key.Keycode {
	case window.KeyLeft:
	case window.KeyA:
		z := cp.app.Scene().Rotation().Z
		cp.cursorX -= int64(math32.Round(math32.Cos(z)))
		if cp.cursorX > cp.size-1 {
			cp.cursorX = cp.size - 1
		}
		if cp.cursorX < 0 {
			cp.cursorX = 0
		}

		cp.cursorY += int64(math32.Round(math32.Sin(z)))
		if cp.cursorY < 0 {
			cp.cursorY = 0
		}
		if cp.cursorY > cp.size-1 {
			cp.cursorY = cp.size - 1
		}
		cp.updateSelected()
		break

	case window.KeyRight:
	case window.KeyD:
		z := cp.app.Scene().Rotation().Z
		cp.cursorX += int64(math32.Round(math32.Cos(z)))
		if cp.cursorX > cp.size-1 {
			cp.cursorX = cp.size - 1
		}
		if cp.cursorX < 0 {
			cp.cursorX = 0
		}

		cp.cursorY -= int64(math32.Round(math32.Sin(z)))
		if cp.cursorY < 0 {
			cp.cursorY = 0
		}
		if cp.cursorY > cp.size-1 {
			cp.cursorY = cp.size - 1
		}
		cp.updateSelected()
		break

	case window.KeyUp:
	case window.KeyW:
		z := cp.app.Scene().Rotation().Z
		cp.cursorX += int64(math32.Round(math32.Sin(z)))
		if cp.cursorX > cp.size-1 {
			cp.cursorX = cp.size - 1
		}
		if cp.cursorX < 0 {
			cp.cursorX = 0
		}

		cp.cursorY += int64(math32.Round(math32.Cos(z)))
		if cp.cursorY < 0 {
			cp.cursorY = 0
		}
		if cp.cursorY > cp.size-1 {
			cp.cursorY = cp.size - 1
		}

		cp.updateSelected()
		break

	case window.KeyDown:
	case window.KeyS:
		z := cp.app.Scene().Rotation().Z
		cp.cursorX -= int64(math32.Round(math32.Sin(z)))
		if cp.cursorX > cp.size-1 {
			cp.cursorX = cp.size - 1
		}
		if cp.cursorX < 0 {
			cp.cursorX = 0
		}

		cp.cursorY -= int64(math32.Round(math32.Cos(z)))
		if cp.cursorY < 0 {
			cp.cursorY = 0
		}
		if cp.cursorY > cp.size-1 {
			cp.cursorY = cp.size - 1
		}
		cp.updateSelected()
		break

	case window.KeyR:
		if cp.rotate {
			cp.rotate = false
		} else {
			cp.rotate = true
		}

	case window.KeyQ:
		cp.app.Quit()
	}
}

func (cp *CubePlane) updateSelected() {

	// unhighlight previous selection
	if cp.selectedNode != nil {
		ig, _ := cp.selectedNode.Children()[0].(graphic.IGraphic)
		gr := ig.GetGraphic()
		imat := gr.GetMaterial(0)

		type matI interface {
			EmissiveColor() math32.Color
			SetEmissiveColor(*math32.Color)
		}
		cubemat := imat.(matI)
		cubemat.SetEmissiveColor(&math32.Color{0, 0, 0})
	}

	// highlight new selection
	cp.selectedNode = cp.plane[cp.cursorX][cp.cursorY]
	ig, _ := cp.selectedNode.Children()[0].(graphic.IGraphic)
	gr := ig.GetGraphic()
	imat := gr.GetMaterial(0)

	type matI interface {
		EmissiveColor() math32.Color
		SetEmissiveColor(*math32.Color)
	}
	cubemat := imat.(matI)
	cubemat.SetEmissiveColor(&math32.Color{0, 1, 0})

	// draw hud text
	cp.app.Gui().RemoveAll(false)
	l1 := gui.NewLabel("oq command: " + cp.command)
	width, _ := cp.app.Gui().Window().Size()
	l1.SetPosition(float32(width)-230, 10)
	l1.SetPaddings(2, 2, 2, 2)
	l1.SetFontSize(12.0)
	cp.app.Gui().Add(l1)

	node := cp.plane[cp.cursorX][cp.cursorY]
	d := node.UserData().(CubeData)
	if d.attrs != nil {
		for i := range cp.Header {
			basename := path.Base(d.attrs[i]) // everything gets basenamed
			selected := fmt.Sprintf("%v %v", cp.Header[i], basename)
			attrs := gui.NewLabel(selected)
			attrs.SetPosition(float32(width)-230, 50.0+(float32(i)*15.0))
			attrs.SetPaddings(2, 2, 2, 2)
			cp.app.Gui().Add(attrs)
		}
	}
}

func (cp *CubePlane) Update(id string, attrs []string) {

	// try to position id's on plane in a predictable order
	x, _ := strconv.ParseInt(id, 10, 64)
	y := x

	x %= cp.size - 1
	y %= cp.size - 1

	for j := range cp.plane[0][y:] {
		for i := range cp.plane[x:][j] {
			node := cp.plane[i][j]
			if node.Name() == "" || node.Name() == id {
				d := CubeData{attrs: attrs, locX: int64(i), locY: int64(j)}
				node.SetUserData(d)
				node.SetName(id)
				cp.updateNodeGfx(node)
				return
			}
		}
	}
}

func (cp CubePlane) updateNodeGfx(node *core.Node) {

	type meshI interface {
		EmissiveColor() math32.Color
		SetEmissiveColor(*math32.Color)
	}
	ig := node.Children()[0].(graphic.IGraphic)
	gr := ig.GetGraphic()
	imesh := gr.GetMaterial(0).(meshI)
	ud := node.UserData().(CubeData)

	cpu, _ := strconv.ParseFloat(ud.attrs[2], 64)
	imesh.SetEmissiveColor(&math32.Color{float32(cpu), 0, 0})
}

func (cp *CubePlane) initCubePlane(size int64) {

	// allocate matrix
	cp.plane = make([][]*core.Node, size)
	for x := int64(0); x < size; x++ {
		cp.plane[x] = make([]*core.Node, size)
	}

	// create nodes
	for y := int64(0); y < size; y++ {
		for x := int64(0); x < size; x++ {
			node := core.NewNode()
			cube := geometry.NewCube(.5)
			mat := material.NewPhong(math32.NewColorHex(0x002b36))
			mesh := graphic.NewMesh(cube, mat)

			// shift cube positions so that rotational axis is in the center,
			// while keeping simpler zero based grid coordinates
			posX := float32(x) - (float32(size) / 2)
			posY := float32(y) - (float32(size) / 2)
			node.SetPosition(posX, posY, 0.0)
			d := CubeData{locX: x, locY: y}
			node.SetUserData(d)
			node.Add(mesh)
			cp.app.Scene().Add(node)
			cp.plane[x][y] = node
		}
	}
}
