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
	app                    *application.Application
	materialsByID          map[string]*material.Phong
	attributesIDByMaterial map[*material.Phong]string
	attributesByID         map[string][]string
	materialsByXY          map[float32]map[float32]*material.Phong
	size                   float32
	assignedX              float32
	assignedY              float32

	command string
	Header  []string

	selectedX float32
	selectedY float32
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
	size := float32(10.0)

	c := &CubePlane{
		app:                    app,
		materialsByID:          make(map[string]*material.Phong),
		attributesIDByMaterial: make(map[*material.Phong]string),
		attributesByID:         make(map[string][]string),
		size:                   size,
		assignedX:              -size,
		assignedY:              -size,
		materialsByXY:          make(map[float32]map[float32]*material.Phong),
		command:                cmd,
	}

	app.Subscribe(application.OnAfterRender,
		func(ev string, i interface{}) {
			app.Scene().RotateOnAxis(&math32.Vector3{0, 0, 1}, .003)
		})

	app.Window().SubscribeID(window.OnKeyDown, 1, func(ev string, i interface{}) {
		handleKeyPress(ev, i, c)
	})

	c.initCubePlane(size)

	return c
}

func handleKeyPress(ev string, i interface{}, c *CubePlane) {
	mat := c.materialsByXY[c.selectedX][c.selectedY]
	mat.SetEmissiveColor(&math32.Color{0, 0, 0})

	key := i.(*window.KeyEvent)
	switch key.Keycode {
	case window.KeyUp:
	case window.KeyW:
		c.selectedY++
		break

	case window.KeyDown:
	case window.KeyS:
		c.selectedY--
		break

	case window.KeyLeft:
	case window.KeyA:
		c.selectedX--
		break

	case window.KeyRight:
	case window.KeyD:
		c.selectedX++
		break
	}

	// wrap cursor on plane
	if c.selectedX > 0 && c.selectedX > c.size {
		c.selectedX = -c.size
	} else if c.selectedX < 0 && c.selectedX < -c.size {
		c.selectedX = c.size
	} else if c.selectedY > 0 && c.selectedY > c.size {
		c.selectedY = -c.size
	} else if c.selectedY < 0 && c.selectedY < -c.size {
		c.selectedY = c.size
	}

	// update details text
	mat = c.materialsByXY[c.selectedX][c.selectedY]
	mat.SetEmissiveColor(&math32.Color{0, 100, 0})
	id := c.attributesIDByMaterial[mat]

	c.app.Gui().RemoveAll(false)
	l1 := gui.NewLabel("oq command: " + c.command)
	width, _ := c.app.Gui().Window().Size()
	l1.SetPosition(float32(width)-230, 10)
	l1.SetPaddings(2, 2, 2, 2)
	l1.SetFontSize(12.0)
	c.app.Gui().Add(l1)

	// if we dont' have an ID it's because the cube wasn't
	// assigned anything so we can skip the details update
	if id == "" {
		return
	}

	for i := range c.Header {
		var basename string
		if val := c.attributesByID[id][i]; val != "" {
			basename = path.Base(c.attributesByID[id][i]) // basename everything to fit things on screen better, esp paths
		}
		selected := fmt.Sprintf("%v %v", c.Header[i], basename)
		attrs := gui.NewLabel(selected)
		attrs.SetPosition(float32(width)-230, 50.0+(float32(i)*15.0))
		attrs.SetPaddings(2, 2, 2, 2)
		c.app.Gui().Add(attrs)
	}
}

func (c *CubePlane) Add(id string, attrs []string) {
	mat := c.materialsByXY[c.assignedX][c.assignedY]
	c.materialsByID[id] = mat
	c.attributesByID[id] = attrs
	c.attributesIDByMaterial[mat] = id
	if c.assignedX < c.size {
		c.assignedX++
	} else if c.assignedX >= c.size {
		c.assignedX = -c.size
		c.assignedY++
	}
	if c.assignedY > c.size {
		panic("out of grid space")
	}
}

func (c *CubePlane) Update(id string, attrs []string) {
	if mat, ok := c.materialsByID[id]; ok && mat != nil {
		cpu, _ := strconv.ParseFloat(attrs[2], 64)
		mat.SetEmissiveColor(&math32.Color{float32(cpu), 0, 0})
	}
}

func (c *CubePlane) initCubePlane(size float32) {
	for x := -size; x <= size; x++ {
		for y := -size; y <= size; y++ {
			cube := geometry.NewCube(.5)
			mat := material.NewPhong(math32.NewColorHex(0x002b36))
			mesh := graphic.NewMesh(cube, mat)
			mesh.SetPosition(float32(x), float32(y), 0.0)
			c.app.Scene().Add(mesh)

			if _, ok := c.materialsByXY[x]; !ok {
				c.materialsByXY[x] = make(map[float32]*material.Phong)
			}
			c.materialsByXY[x][y] = mat
		}
	}
}
