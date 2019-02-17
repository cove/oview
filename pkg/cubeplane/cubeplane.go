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
	"math"
	"strconv"
	"time"

	"golang.org/x/sync/semaphore"

	"github.com/g3n/engine/camera/control"

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
	app                *application.Application
	plane              [][]*core.Node
	size               int64
	secondsPerRotation float32
	ttl                int64
	cubeSize           float32
	cubeInactiveColor  *math32.Color
	cubeActiveColor    *math32.Color
	cubeWireframe      bool
	cursorX            int64
	cursorY            int64
	selected           *core.Node
	selectedColor      *math32.Color
	selectedHeaderIdx  int
	backgroundColor    *math32.Color
	hud                *Hud
	rc                 *core.Raycaster
	command            string
	header             []string
	rotate             bool
	UpdateChan         CubeUpdateChan
	incomingInProgres  *semaphore.Weighted
	timeout            chan bool
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

func Init(app *application.Application, cmd string, refresh int,
	wireframe bool, size int64, rotations int, pause bool, usage bool) *CubePlane {

	// Add lights to the scene
	ambientLight := light.NewAmbient(&math32.Color{1.0, 1.0, 1.0}, 0.8)
	app.Scene().Add(ambientLight)
	pointLight := light.NewPoint(&math32.Color{1, 1, 1}, 20.0)
	pointLight.SetPosition(1, 0, 10)
	app.Scene().Add(pointLight)

	// Add an axis helper to the scene
	//axis := graphic.NewAxisHelper(1.5)
	//app.Scene().Add(axis)

	// Position the camera to look at center of board at an angle
	app.CameraPersp().SetPosition(0, -20, 30)
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
		size:               int64(size),
		secondsPerRotation: float32(rotations),
		cubeSize:           float32(.5),
		cubeWireframe:      wireframe,
		cubeInactiveColor:  math32.NewColorHex(0x50596C),
		cubeActiveColor:    math32.NewColorHex(0x608E93),
		backgroundColor:    math32.NewColorHex(0x2F2D3E),
		selectedColor:      math32.NewColorHex(0x99BAA4),
		hud: &Hud{
			color:    math32.NewColorHex(0xFCF2C6),
			fontSize: float64(12.0),
		},
		command:           cmd,
		rc:                core.NewRaycaster(&math32.Vector3{}, &math32.Vector3{}),
		rotate:            !pause,
		UpdateChan:        make(CubeUpdateChan, 1024),
		incomingInProgres: semaphore.NewWeighted(1),
		timeout:           make(chan bool, 1),
		selectedHeaderIdx: -1,
	}

	// Sets window background color
	c := cp.backgroundColor
	gs.ClearColor(c.R, c.G, c.B, 1.0)

	// Init event handling
	app.TimerManager.Initialize()
	app.Window().Subscribe(window.OnKeyDown, cp.onKey)
	app.Window().Subscribe(window.OnKeyRepeat, cp.onKey)
	app.Window().Subscribe(window.OnMouseDown, cp.onMouse)
	cp.app.SubscribeID(application.OnAfterRender, 1, func(evname string, ev interface{}) {
		if cp.rotate {
			cp.app.Scene().RotateOnAxis(&math32.Vector3{0, 0, 1},
				app.FrameDeltaSeconds()*-2*math32.Pi/cp.secondsPerRotation)
		}
	})

	cp.initCubePlane()

	// select first cube to start
	cp.selected = cp.plane[0][0]
	cp.initHud()

	// Display or hide usage text in window after startup
	cp.hud.usage.SetVisible(usage)

	app.SetInterval(time.Duration(refresh)*time.Second, nil, cp.processIncoming)
	go cp.processTimeout()

	return cp
}

func (cp *CubePlane) processTimeout() {
	for {
		time.Sleep(1 * time.Second)
		cp.timeout <- true
	}
}

func (cp *CubePlane) processIncoming(i interface{}) {

	// skip this update if we're already running
	if !cp.incomingInProgres.TryAcquire(1) {
		return
	}

	select {
	case table := <-cp.UpdateChan:

		// use first value that's a number for scaling cubes
		if cp.selectedHeaderIdx < 0 {
			for i, v := range table[0] {
				if _, err := strconv.ParseFloat(v, 64); err == nil {
					cp.selectedHeaderIdx = i
					break
				}
			}
			if cp.selectedHeaderIdx < 0 {
				panic("no numbers found to scale cubes with")
			}
		}
		cp.updateHud()
		cp.cullExpiredCubes()
		for i := range table {
			cp.updateCube(table[i][1], table[i])
		}

	case <-cp.timeout:
		break // timeout to prevent blocking
	}
	cp.incomingInProgres.Release(1)
}

func (cp *CubePlane) updateSelectedCube() {

	type matI interface {
		EmissiveColor() math32.Color
		SetEmissiveColor(*math32.Color)
		SetColor(*math32.Color)
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

func (cp *CubePlane) updateCube(id string, attrs []string) {

	// update exiting cube
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
					cp.updateSelectedCube()
				}
				return
			}
		}
	}

	// fine a spot for a new cube
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
					cp.updateSelectedCube()
				}
				return
			}
		}
	}

}

func (cp *CubePlane) cullExpiredCubes() {
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

func (cp *CubePlane) updateCubeStatus(node *core.Node) {

	type meshI interface {
		EmissiveColor() math32.Color
		SetEmissiveColor(*math32.Color)
		SetColor(*math32.Color)
		SetWireframe(bool)
	}
	ig := node.Children()[0].(graphic.IGraphic)
	gr := ig.GetGraphic()
	imesh := gr.GetMaterial(0).(meshI)

	if isActive(node) {
		ud := node.UserData().(CubeData)
		imesh.SetColor(cp.cubeActiveColor)

		value, err := strconv.ParseFloat(ud.attrs[cp.selectedHeaderIdx], 64)
		if err != nil {
			return
		}

		// scale to log if the value is too much larger than our cube
		if value > 2*float64(cp.size) {
			value = math.Log10(value)
		}

		imesh.SetWireframe(cp.cubeWireframe)

		if float32(value) >= .5 {
			gr.SetMatrix(math32.NewMatrix4().MakeTranslation(0, 0, float32(value)/4))
			gr.SetScaleZ(float32(value))
		} else {
			gr.SetMatrix(math32.NewMatrix4().MakeTranslation(0, 0, float32(cp.cubeSize)/4))
			gr.SetScaleZ(float32(cp.cubeSize))
		}

	} else {
		imesh.SetWireframe(cp.cubeWireframe)
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
			mat.SetWireframe(cp.cubeWireframe)
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
