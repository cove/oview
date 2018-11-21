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
	"strconv"

	"github.com/g3n/engine/geometry"
	"github.com/g3n/engine/graphic"
	"github.com/g3n/engine/light"
	"github.com/g3n/engine/material"
	"github.com/g3n/engine/math32"
	"github.com/g3n/engine/util/application"
)

type CubePlane struct {
	app          *application.Application
	Cubemat      map[string]*material.Phong
	CubematIndex []*material.Phong

	Cubemesh  map[string]*graphic.Mesh
	Cubeattrs map[string][]string
}

func Init(app *application.Application) *CubePlane {

	// Add lights to the scene
	ambientLight := light.NewAmbient(&math32.Color{1.0, 1.0, 1.0}, 0.8)
	app.Scene().Add(ambientLight)
	pointLight := light.NewPoint(&math32.Color{1, 1, 1}, 15.0)
	pointLight.SetPosition(1, 0, 2)
	app.Scene().Add(pointLight)

	// Add an axis helper to the scene
	//axis := graphic.NewAxisHelper(1.5)
	//app.Scene().Add(axis)

	app.CameraPersp().SetPosition(0, -15, 10)
	app.CameraPersp().LookAt(&math32.Vector3{0, 0, 0})
	app.TimerManager.Initialize()

	app.Subscribe(application.OnAfterRender,
		func(evname string, ev interface{}) {
			app.Scene().RotateOnAxis(&math32.Vector3{0, 0, 1},
				.003)
		})

	c := &CubePlane{
		app:       app,
		Cubemat:   make(map[string]*material.Phong),
		Cubemesh:  make(map[string]*graphic.Mesh),
		Cubeattrs: make(map[string][]string),
	}

	c.initCubes(10)

	return c
}

func (c *CubePlane) Add(id string, attrs []string) {
	cube := geometry.NewCube(1.0)
	mat := material.NewPhong(math32.NewColor("DarkBlue"))
	mesh := graphic.NewMesh(cube, mat)
	c.Cubemat[id] = mat
	c.Cubemesh[id] = mesh
	c.Cubeattrs[id] = attrs

	//pos := c.getNextGridPosition()
	//mesh.SetPositionVec(pos)

	c.app.Scene().Add(mesh)
}

func (c *CubePlane) Update(id string, attrs []string) {
	if mat, ok := c.Cubemat[id]; ok {
		cpu, _ := strconv.ParseFloat(attrs[2], 64)
		redpercent := float32(cpu / 255)
		mat.SetColor(&math32.Color{redpercent, 0, 0})
	}
}

func (c *CubePlane) initCubes(size float32) {
	for x := -size; x < size; x++ {
		cube := geometry.NewCube(1.0)
		mat := material.NewPhong(math32.NewColor("DarkBlue"))
		mesh := graphic.NewMesh(cube, mat)
		mesh.SetPosition(float32(x), 0, 0.0)
		c.app.Scene().Add(mesh)

		for y := -size; y < size; y++ {
			cube := geometry.NewCube(1.0)
			mat := material.NewPhong(math32.NewColor("DarkBlue"))
			mesh := graphic.NewMesh(cube, mat)
			mesh.SetPosition(float32(x), float32(y), 0.0)
			c.app.Scene().Add(mesh)
		}
	}
}
