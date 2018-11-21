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

	"github.com/g3n/engine/window"

	"github.com/g3n/engine/geometry"
	"github.com/g3n/engine/graphic"
	"github.com/g3n/engine/light"
	"github.com/g3n/engine/material"
	"github.com/g3n/engine/math32"
	"github.com/g3n/engine/util/application"
)

type CubePlane struct {
	app            *application.Application
	materialsById  map[string]*material.Phong
	attributesById map[string][]string
	materialsByXY  map[float32]map[float32]*material.Phong
	size           float32
	assignedX      float32
	assignedY      float32

	selectedX float32
	selectedY float32
}

func Init(app *application.Application) *CubePlane {

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
	app.TimerManager.Initialize()
	size := float32(10.0)

	c := &CubePlane{
		app:            app,
		materialsById:  make(map[string]*material.Phong),
		attributesById: make(map[string][]string),
		size:           size,
		assignedX:      -size,
		assignedY:      -size,
		materialsByXY:  make(map[float32]map[float32]*material.Phong),
	}

	app.Subscribe(application.OnAfterRender,
		func(evname string, ev interface{}) {
			app.Scene().RotateOnAxis(&math32.Vector3{0, 0, 1}, .003)
		})

	app.Window().SubscribeID(window.OnKeyDown, 1, func(ev string, i interface{}) {
		c.materialsByXY[c.selectedX][c.selectedY].SetEmissiveColor(&math32.Color{0, 0, 0})
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

		if c.selectedX > 0 && c.selectedX > c.size {
			c.selectedX = -c.size
		} else if c.selectedX < 0 && c.selectedX < -c.size {
			c.selectedX = c.size
		} else if c.selectedY > 0 && c.selectedY > c.size {
			c.selectedY = -c.size
		} else if c.selectedY < 0 && c.selectedY < -c.size {
			c.selectedY = c.size
		}

		c.materialsByXY[c.selectedX][c.selectedY].SetEmissiveColor(&math32.Color{0, 1, 1})
		fmt.Printf("interface i = %v", i)
	})

	c.initCubePlane(size)

	return c
}

func (c *CubePlane) Add(id string, attrs []string) {
	c.materialsById[id] = c.materialsByXY[c.assignedX][c.assignedY]
	c.attributesById[id] = attrs
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
	if mat, ok := c.materialsById[id]; ok && mat != nil {
		cpu, _ := strconv.ParseFloat(attrs[2], 64)
		mat.SetColor(&math32.Color{float32(cpu), 0, 0})
	}
}

func (c *CubePlane) initCubePlane(size float32) {
	for x := -size; x <= size; x++ {
		for y := -size; y <= size; y++ {
			cube := geometry.NewCube(.5)
			mat := material.NewPhong(math32.NewColor("DarkBlue"))
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
