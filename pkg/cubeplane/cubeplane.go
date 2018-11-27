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
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/cove/oq/pkg/fonts"

	"github.com/g3n/engine/text"

	"github.com/g3n/engine/camera/control"

	"github.com/cove/oq/pkg/colors"

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
	app *application.Application

	plane              [][]*core.Node
	size               int64
	secondsPerRotation float32
	ttl                int64

	cubeSize          float32
	cubeInactiveColor *math32.Color
	cubeActiveColor   *math32.Color

	cursorX       int64
	cursorY       int64
	selected      *core.Node
	selectedColor *math32.Color

	backgroundColor *math32.Color

	hudTextSize  float64
	hudTextColor *math32.Color
	hudPanel     *gui.Panel
	hudFont      *text.Font

	rc      *core.Raycaster
	command string
	header  []string
	rotate  bool

	UpdateChan CubeUpdateChan
}

type CubeUpdateChan chan CubeUpdate
type CubeUpdate [][]string
type CubeHeader []string

type CubeData struct {
	attrs  []string
	locX   int64
	locY   int64
	ttl    int64
	active bool
}

func Init(app *application.Application, cmd string) *CubePlane {

	// Add lights to the scene
	ambientLight := light.NewAmbient(&math32.Color{1.0, 1.0, 1.0}, 0.8)
	app.Scene().Add(ambientLight)
	pointLight := light.NewPoint(&math32.Color{1, 1, 1}, 10.0)
	pointLight.SetPosition(1, 0, 10)
	app.Scene().Add(pointLight)

	// Add an axis helper to the scene
	//axis := graphic.NewAxisHelper(1.5)
	//app.Scene().Add(axis)

	// Position the camera to look at center of board at an angle
	app.CameraPersp().SetPosition(0, -20, 10)
	app.CameraPersp().LookAt(&math32.Vector3{0, 0, 0})

	// Init for hud
	gs, err := gls.New()
	if err != nil {
		panic(err)
	}
	gui.NewRoot(gs, app.Window())
	root := gui.NewRoot(gs, app.Window())

	// Creates a renderer and adds default shaders
	rend := renderer.NewRenderer(gs)
	err = rend.AddDefaultShaders()
	if err != nil {
		panic(err)
	}
	app.SetGui(root)

	// Only use the mouse for orbital controls, since the arrow keys are used to cursor to cubes
	app.Orbit().Dispose()
	ob := control.NewOrbitControl(app.Camera(), app.Window())
	ob.EnableKeys = false
	app.SetOrbit(ob)

	// Create cube plane with defaults
	cp := &CubePlane{
		app:                app,
		size:               int64(20),
		secondsPerRotation: float32(30),
		cubeSize:           float32(.5),
		cubeInactiveColor:  colors.Solarized("base1"),
		cubeActiveColor:    colors.Solarized("violet"),
		backgroundColor:    colors.Solarized("base03"),
		selectedColor:      colors.Solarized("cyan"),
		hudTextSize:        float64(12.0),
		hudTextColor:       colors.Solarized("base0"),
		command:            cmd,
		rc:                 core.NewRaycaster(&math32.Vector3{}, &math32.Vector3{}),
		rotate:             true,
		UpdateChan:         make(CubeUpdateChan, 500),
	}

	// Sets window background color
	c := cp.backgroundColor
	gs.ClearColor(c.R, c.G, c.B, 1.0)

	// Init event handling
	app.TimerManager.Initialize()
	app.Window().Subscribe(window.OnKeyDown, func(evname string, ev interface{}) {
		cp.onKey(evname, ev)
	})
	app.Window().Subscribe(window.OnKeyRepeat, func(evname string, ev interface{}) {
		cp.onKey(evname, ev)
	})
	app.Window().Subscribe(window.OnMouseDown, func(evname string, ev interface{}) {
		cp.onMouse(evname, ev)
	})
	cp.app.SubscribeID(application.OnAfterRender, 1, func(evname string, ev interface{}) {
		if cp.rotate {
			cp.app.Scene().RotateOnAxis(&math32.Vector3{0, 0, 1},
				app.FrameDeltaSeconds()*-2*math32.Pi/cp.secondsPerRotation)
		}
	})

	cp.initCubePlane()
	cp.initHud()

	app.SetInterval(time.Duration(5*time.Second), nil,
		func(i interface{}) {
			select {
			case table := <-cp.UpdateChan:
				cp.CullExpiredCubes()
				for i := range table {
					cp.Update(table[i][1], table[i])
				}
			}
		})

	return cp
}

func (cp *CubePlane) onMouse(evname string, ev interface{}) {

	// Convert mouse coordinates to normalized device coordinates
	mev := ev.(*window.MouseEvent)
	width, height := cp.app.Window().Size()
	var x, y float32
	if runtime.GOOS == "darwin" {
		// OSX, not sure why it's different than Windows and Linux
		x = 1*(mev.Xpos/float32(width)) - 1
		y = -1*(mev.Ypos/float32(height)) + 1
	} else {
		// Linux and Windows
		x = 2*(mev.Xpos/float32(width)) - 1
		y = -2*(mev.Ypos/float32(height)) + 1
	}

	// Set the raycaster from the current camera and mouse coordinates
	cp.app.Camera().SetRaycaster(cp.rc, x, y)
	//fmt.Printf("rc:%+v\n", cp.rc.Ray)

	// Checks intersection with all objects in the scene
	intersects := cp.rc.IntersectObjects(cp.app.Scene().Children(), true)
	//fmt.Printf("intersects:%+v\n", intersects)
	if len(intersects) == 0 {
		return
	}

	// Get the first object we intersected with
	obj := intersects[0].Object
	ig, ok := obj.(graphic.IGraphic)
	if !ok {
		return
	}

	// Save the x y coordinate on the node so we can identify it later after a raycast
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
		fallthrough
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

	case window.KeyRight:
		fallthrough
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

	case window.KeyUp:
		fallthrough
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

	case window.KeyDown:
		fallthrough
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

	case window.KeyR:
		cp.rotate = !cp.rotate

	case window.KeyQ:
		cp.app.Quit()
	}
}

func (cp *CubePlane) updateSelected() {

	type matI interface {
		EmissiveColor() math32.Color
		SetEmissiveColor(*math32.Color)
	}

	// Un-highlight previous selection
	if cp.selected != nil {
		ig, _ := cp.selected.Children()[0].(graphic.IGraphic)
		gr := ig.GetGraphic()
		imat := gr.GetMaterial(0)
		cubemat := imat.(matI)
		cubemat.SetEmissiveColor(&math32.Color{0, 0, 0})
	}

	// Highlight new selection
	cp.selected = cp.plane[cp.cursorX][cp.cursorY]
	ig, _ := cp.selected.Children()[0].(graphic.IGraphic)
	gr := ig.GetGraphic()
	imat := gr.GetMaterial(0)

	cubemat := imat.(matI)
	cubemat.SetEmissiveColor(cp.selectedColor)

	cp.updateHud()
}

func (cp *CubePlane) initHud() {

	// panel
	panel := gui.NewPanel(230, 200)
	// XXX keep invisable, for later use
	panel.SetVisible(false)
	panel.SetContentSize(200, 190)
	panel.SetPosition(20, 20)
	style := gui.PanelStyle{
		Margin:      gui.RectBounds{0, 0, 0, 0},
		Border:      gui.RectBounds{1, 1, 1, 1},
		Padding:     gui.RectBounds{8, 8, 8, 8},
		BorderColor: *colors.Solarized4("cyan", .2),
		BgColor:     math32.Color4{0, 0, 0, .2},
	}
	panel.ApplyStyle(&style)
	cp.hudPanel = panel

	// font
	font, err := text.NewFontFromData(fonts.OrbitronRegular())
	if err != nil {
		panic(err.Error())
	}
	font.SetLineSpacing(1.0)
	font.SetPointSize(cp.hudTextSize)
	font.SetDPI(72)
	font.SetFgColor(&math32.Color4{0, 0, 1, 1})
	font.SetBgColor(&math32.Color4{1, 1, 0, 0.8})
	cp.hudFont = font
}

func (cp *CubePlane) updateHud() {
	cp.app.Gui().GetPanel().RemoveAll(true)
	cp.app.Gui().GetPanel().Add(cp.hudPanel)

	node := cp.plane[cp.cursorX][cp.cursorY]
	d := node.UserData().(CubeData)
	if d.attrs != nil {
		for i := range cp.header {

			lineSpace := float32(8.0)

			// heading
			heading := gui.NewButton(strings.ToLower(cp.header[i]))
			heading.SetPosition(20, 20.0+(float32(i)*(float32(cp.hudTextSize)+lineSpace)))
			heading.Label.SetFont(cp.hudFont)
			heading.SetStyles(&gui.ButtonStyles{
				Over:   gui.ButtonStyle{FgColor: *math32.NewColor4("cyan", 1.0)},
				Normal: gui.ButtonStyle{FgColor: *colors.Solarized4("base1", 1.0)}},
			)
			heading.Subscribe(gui.OnClick, func(evname string, ev interface{}) {
				fmt.Printf("button %v OnClick\n", cp.header[i])
			})
			cp.app.Gui().GetPanel().Add(heading)

			// values
			basename := path.Base(d.attrs[i]) // everything gets basenamed
			values := gui.NewButton(basename)
			values.SetPosition(110, 20.0+(float32(i)*(float32(cp.hudTextSize)+lineSpace)))
			values.Label.SetFont(cp.hudFont)
			values.SetStyles(&gui.ButtonStyles{
				Over:   gui.ButtonStyle{FgColor: *math32.NewColor4("cyan", 1.0)},
				Normal: gui.ButtonStyle{FgColor: *colors.Solarized4("base2", 1.0)}},
			)
			values.Subscribe(gui.OnClick, func(evname string, ev interface{}) {
				fmt.Printf("button %v OnClick\n", cp.header[i])
			})
			cp.app.Gui().GetPanel().Add(values)
		}
	}
}

func (cp *CubePlane) Update(id string, attrs []string) {

	for j := range cp.plane {
		for i := range cp.plane {
			node := cp.plane[i][j]
			if node.Name() == id {
				ud := node.UserData().(CubeData)
				ud.attrs = attrs
				ud.ttl++
				node.SetUserData(ud)
				cp.updateCubeStatus(node)

				if cp.selected == node {
					cp.updateSelected()
				}
				return
			}
		}
	}

	for j := range cp.plane {
		for i := range cp.plane {
			node := cp.plane[i][j]
			if !cp.isActive(node) {
				ud := node.UserData().(CubeData)
				ud.attrs = attrs
				ud.ttl = cp.ttl
				node.SetUserData(ud)
				node.SetName(id)
				cp.makeActive(node)
				cp.updateCubeStatus(node)

				if cp.selected == node {
					cp.updateSelected()
				}
				return
			}
		}
	}

}

func (cp *CubePlane) CullExpiredCubes() {
	for x := range cp.plane {
		for y := range cp.plane {
			node := cp.plane[x][y]
			if cp.isActive(node) && cp.isExpired(node) {
				cp.makeInactive(node)
				cp.updateCubeStatus(node)
			}
		}
	}
	cp.ttl++
}

func (cp *CubePlane) isExpired(node *core.Node) bool {
	ud := node.UserData().(CubeData)
	return ud.ttl < (cp.ttl - 1)
}

func (cp *CubePlane) isActive(node *core.Node) bool {
	ud := node.UserData().(CubeData)
	return ud.active
}

func (cp *CubePlane) makeInactive(node *core.Node) {
	node.SetName("")
	ud := node.UserData().(CubeData)
	ud.active = false
	ud.attrs = nil
	ud.ttl = 0
	node.SetUserData(ud)
}

func (cp *CubePlane) makeActive(node *core.Node) {
	ud := node.UserData().(CubeData)
	ud.active = true
	node.SetUserData(ud)
	ud.ttl = cp.ttl
}

func (cp *CubePlane) SetHeader(header CubeHeader) {
	cp.header = header
}

func (cp *CubePlane) dumpPlane() {

	fmt.Println("Dumping CubePlane:")
	for y := range cp.plane {
		for x := range cp.plane {
			name := cp.plane[x][y].Name()
			fmt.Printf("%5s\t", name)
		}
		fmt.Println("")
	}

}

func (cp *CubePlane) updateCubeStatus(node *core.Node) {

	type meshI interface {
		EmissiveColor() math32.Color
		SetEmissiveColor(*math32.Color)
		SetColor(*math32.Color)
	}
	ig := node.Children()[0].(graphic.IGraphic)
	gr := ig.GetGraphic()
	imesh := gr.GetMaterial(0).(meshI)

	if cp.isActive(node) {
		ud := node.UserData().(CubeData)
		imesh.SetColor(cp.cubeActiveColor)

		cpu, _ := strconv.ParseFloat(ud.attrs[2], 64)
		cpu /= 10

		if float32(cpu) >= .5 {
			gr.SetMatrix(math32.NewMatrix4().MakeTranslation(0, 0, float32(cpu)/4))
			gr.SetScaleZ(float32(cpu))
		} else {
			gr.SetMatrix(math32.NewMatrix4().MakeTranslation(0, 0, float32(cp.cubeSize)/4))
			gr.SetScaleZ(float32(cp.cubeSize))
		}

	} else {
		imesh.SetColor(cp.cubeInactiveColor)
		gr.SetMatrix(math32.NewMatrix4().MakeTranslation(0, 0, cp.cubeSize/4))
		gr.SetScaleZ(cp.cubeSize)
	}
}

func (cp *CubePlane) initCubePlane() {

	// allocate matrix
	cp.plane = make([][]*core.Node, cp.size)
	for x := int64(0); x < cp.size; x++ {
		cp.plane[x] = make([]*core.Node, cp.size)
	}

	// Create nodes
	for y := int64(0); y < cp.size; y++ {
		for x := int64(0); x < cp.size; x++ {
			node := core.NewNode()
			cube := geometry.NewCube(cp.cubeSize)
			mat := material.NewPhong(cp.cubeInactiveColor)
			mesh := graphic.NewMesh(cube, mat)

			// XXX: pre-scale cubes so when they're scaled they all line up
			mesh.SetMatrix(math32.NewMatrix4().MakeTranslation(0, 0, cp.cubeSize/4))
			mesh.SetScaleZ(cp.cubeSize)

			// Shift cube positions so that rotational axis is in the center,
			// while keeping simpler zero based grid coordinates
			posX := float32(x) - (float32(cp.size) / 2)
			posY := float32(y) - (float32(cp.size) / 2)
			node.SetPosition(posX, posY, 0.0)
			d := CubeData{locX: x, locY: y}
			node.SetUserData(d)
			cp.makeInactive(node)
			node.Add(mesh)
			cp.app.Scene().Add(node)
			cp.plane[x][y] = node
		}
	}
}
