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
	"strconv"
	"time"

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
	hudFont      *text.Font
	hudHeaders   *gui.Panel
	hudValues    *gui.Panel

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
			if !isActive(node) {
				ud := node.UserData().(CubeData)
				ud.attrs = attrs
				ud.ttl = cp.ttl
				node.SetUserData(ud)
				node.SetName(id)
				makeActive(node)
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
			if isActive(node) && cp.isExpired(node) {
				makeInactive(node)
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

func isActive(node *core.Node) bool {
	ud := node.UserData().(CubeData)
	return ud.active
}

func makeInactive(node *core.Node) {
	node.SetName("")
	ud := node.UserData().(CubeData)
	ud.active = false
	node.SetUserData(ud)
}

func makeActive(node *core.Node) {
	ud := node.UserData().(CubeData)
	ud.active = true
	node.SetUserData(ud)
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

	if isActive(node) {
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
			makeInactive(node)
			node.Add(mesh)
			cp.app.Scene().Add(node)
			cp.plane[x][y] = node
		}
	}
}
