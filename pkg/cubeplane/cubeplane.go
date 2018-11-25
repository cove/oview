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

	rc      *core.Raycaster
	command string
	Header  []string
	rotate  bool
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
		cubeInactiveColor:  colors.Solaried("base1"),
		cubeActiveColor:    colors.Solaried("violet"),
		backgroundColor:    colors.Solaried("base03"),
		selectedColor:      colors.Solaried("cyan"),
		hudTextSize:        float64(12.0),
		hudTextColor:       colors.Solaried("base0"),
		command:            cmd,
		rc:                 core.NewRaycaster(&math32.Vector3{}, &math32.Vector3{}),
		rotate:             true,
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
	cp.app.SubscribeID(application.OnAfterRender, 1, func(evname string, ev interface{}) {
		if cp.rotate {
			cp.app.Scene().RotateOnAxis(&math32.Vector3{0, 0, 1},
				app.FrameDeltaSeconds()*-2*math32.Pi/cp.secondsPerRotation)
		}
	})
	app.Window().Subscribe(window.OnMouseDown, func(evname string, ev interface{}) {
		cp.onMouse(evname, ev)
	})

	cp.initCubePlane()

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
		break

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
		break

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

func (cp *CubePlane) updateHud() {

	cp.app.Gui().RemoveAll(false)
	//	width, _ := cp.app.Gui().Window().Size()

	font, err := text.NewFont("/Users/cove/go/src/github.com/cove/oq/fonts/Orbitron/Orbitron-Regular.ttf")
	//font, err := text.NewFont("/Users/cove/go/src/github.com/cove/oq/fonts/spacemono/SpaceMono-Regular.ttf")
	if err != nil {
		panic(err.Error())
	}
	font.SetLineSpacing(1.0)
	font.SetPointSize(cp.hudTextSize)
	font.SetDPI(72)
	font.SetFgColor(&math32.Color4{0, 0, 1, 1})
	font.SetBgColor(&math32.Color4{1, 1, 0, 0.8})

	node := cp.plane[cp.cursorX][cp.cursorY]
	d := node.UserData().(CubeData)
	if d.attrs != nil {
		for i := range cp.Header {
			basename := path.Base(d.attrs[i]) // everything gets basenamed

			lineSpace := float32(10.0)
			b2 := gui.NewButton(cp.Header[i])
			b2.SetPosition(20, 20.0+(float32(i)*(float32(cp.hudTextSize)+lineSpace)))
			b2.Label.SetFont(font)
			b2.SetBordersColor4(&math32.Color4{0, 0, 0, 0})
			b2.SetColor4(&math32.Color4{0, 0, 0, 0})

			fg := colors.Solaried("cyan")
			s := gui.ButtonStyle{FgColor: math32.Color4{fg.R, fg.G, fg.B, 1}}
			fo := colors.Solaried("green")
			n := gui.ButtonStyle{FgColor: math32.Color4{fo.R, fo.G, fo.B, 1}}

			b2.SetStyles(&gui.ButtonStyles{Over: s, Normal: n})

			b2.Subscribe(gui.OnClick, func(evname string, ev interface{}) {
				fmt.Printf("button %v OnClick\n", cp.Header[i])
			})
			cp.app.Gui().Add(b2)

			attrs := gui.NewLabel(basename)
			attrs.SetPosition(120, 20.0+(float32(i)*(float32(cp.hudTextSize)+lineSpace)))
			attrs.SetPaddings(5, 2, 5, 2)
			attrs.SetColor(cp.hudTextColor)
			attrs.SetFont(font)
			cp.app.Gui().Add(attrs)

		}
	}
}

func (cp *CubePlane) Update(id string, attrs []string) {

	// Try to position ID's on plane in a predictable order
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
				cp.updateCubeStatus(node)
				return
			}
		}
	}
}

func (cp *CubePlane) updateCubeStatus(node *core.Node) {

	type meshI interface {
		EmissiveColor() math32.Color
		SetEmissiveColor(*math32.Color)
		SetColor(color *math32.Color)
	}

	ig := node.Children()[0].(graphic.IGraphic)
	gr := ig.GetGraphic()
	imesh := gr.GetMaterial(0).(meshI)
	ud := node.UserData().(CubeData)

	imesh.SetColor(cp.cubeActiveColor)

	cpu, _ := strconv.ParseFloat(ud.attrs[2], 64)
	cpu /= 10

	if float32(cpu) >= .5 {
		gr.SetMatrix(math32.NewMatrix4().MakeTranslation(0, 0, float32(cpu)/4))
		gr.SetScaleZ(float32(cpu))
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
			node.Add(mesh)
			cp.app.Scene().Add(node)
			cp.plane[x][y] = node
		}
	}
}
