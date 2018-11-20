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

package cmd

import (
	"bytes"
	"fmt"
	"os/exec"
	"time"

	"github.com/g3n/engine/geometry"
	"github.com/g3n/engine/graphic"
	"github.com/g3n/engine/light"
	"github.com/g3n/engine/material"
	"github.com/g3n/engine/math32"
	"github.com/g3n/engine/util/application"

	"github.com/spf13/cobra"
)

// viewCmd represents the view command
var viewCmd = &cobra.Command{
	Use:   "view",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: cmdView,
}

func init() {
	rootCmd.AddCommand(viewCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// viewCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// viewCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func cmdView(cmd *cobra.Command, args []string) {
	app, _ := application.Create(application.Options{
		Title:  "oq",
		Width:  800,
		Height: 600,
	})

	phongs := [100][100]*material.Phong{}
	for x := -10; x < 10; x++ {
		cube := geometry.NewCube(1.0)
		mat := material.NewPhong(math32.NewColor("DarkBlue"))
		cube1Mesh := graphic.NewMesh(cube, mat)
		cube1Mesh.SetPosition(float32(x), 0, 0.0)
		app.Scene().Add(cube1Mesh)

		phongs[x+10][0] = mat

		//cube2 := geometry.NewCube(.5)
		//mat2 := material.NewPhong(math32.NewColor("DarkRed"))
		//cube2Mesh := graphic.NewMesh(cube2, mat2)
		//cube2Mesh.SetPosition(float32(x), 0, 1.0)
		//app.Scene().Add(cube2Mesh)

		for y := -10; y < 10; y++ {
			cube := geometry.NewCube(1.0)
			mat := material.NewPhong(math32.NewColor("DarkBlue"))
			cube1Mesh := graphic.NewMesh(cube, mat)
			cube1Mesh.SetPosition(float32(x), float32(y), 0.0)
			app.Scene().Add(cube1Mesh)
			phongs[x+10][y+10] = mat
			y++
		}
		x++
	}

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

	app.Subscribe(application.OnAfterRender,
		func(evname string, ev interface{}) {
			app.Scene().RotateOnAxis(&math32.Vector3{0, 0, 1},
				.003)
		})

	app.Subscribe(application.OnAfterRender,
		func(evname string, ev interface{}) {
			j := 0
			for i := range keys {
				phongs[0][j].SetColor(&math32.Color{0, float32(table[keys[i]]), 0})
				j++
				j++
			}
		})

	app.TimerManager.Initialize()
	app.SetInterval(time.Duration(5*time.Second), nil, func(i interface{}) {
		ReadInTable()
	})

	app.Run()
}

var table = make(map[string]int, 10)
var keys = make([]string, 2)

func ReadInTable() {

	cmd := exec.Command("/bin/ps", "aux")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		panic(err)
	}

	table["init"] = 1
	keys[0] = "init"
	table["run"] = 100
	keys[1] = "run"

	fmt.Printf("in all caps: %q\n", out.String())
}
