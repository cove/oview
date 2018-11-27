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
	"runtime"

	"github.com/g3n/engine/graphic"
	"github.com/g3n/engine/window"
)

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
