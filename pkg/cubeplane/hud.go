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
	"path"
	"strings"

	"github.com/g3n/engine/gui"
	"github.com/g3n/engine/math32"
)

type Hud struct {
	fontSize float64
	color    *math32.Color
	main     *gui.Panel
	headers  *gui.Panel
	values   *gui.Panel
	usage    *gui.Panel
	buttons  []*gui.Button
}

type HudData struct {
	attrIdx int
}

const USAGE_TEXT = `
Keys
F                   Wireframe
R                   Start/stop rotation
Arrows          Move cursor (also vi and awsd)
Q                   Quit
H                   Show usage help

Mouse
Wheel                      Zoom in/out
Right click & drag  Rotate plane
Right click on cube View details of the cube
`

func (cp *CubePlane) initHud() {

	// main panel covers screen so we can display the usage in lower right
	width, height := cp.app.Window().Size()
	cp.hud.main = gui.NewPanel(float32(width), float32(height))
	cp.hud.main.SetPosition(10, 10)

	// header from the table we're displaying
	cp.hud.headers = gui.NewPanel(500, 500)
	cp.hud.headers.SetPosition(10, 10)
	cp.hud.main.Add(cp.hud.headers)
	//cp.hud.headers.SetBorders(1, 1, 1, 1)

	// values/details of the cubes
	cp.hud.values = gui.NewPanel(500, 500)
	//cp.hud.values.SetBorders(1, 1, 1, 1)
	cp.hud.headers.Add(cp.hud.values)

	// usage text panel
	cp.hud.usage = gui.NewPanel(100, 100)
	cp.hud.usage.SetPosition(float32(width)-340, float32(height)-260)
	cp.hud.usage.SetBorders(1, 1, 1, 1)
	cp.hud.usage.SetColor4(&math32.Color4{0.871, 0.494, 0.267, .6})
	cp.hud.usage.SetBordersColor4(&math32.Color4{0.871, 0.494, 0.267, 1.0})
	cp.hud.usage.SetPaddings(5, 5, 5, 5)

	// usage text
	usagetext := gui.NewLabel(USAGE_TEXT)
	usagetext.SetColor4(math32.NewColor4("White", 1.0))
	usagetext.SetPaddings(5, 5, 5, 5)
	cp.hud.usage.SetContentSize(usagetext.Size())
	cp.hud.usage.Add(usagetext)
	cp.hud.main.Add(cp.hud.usage)

	// reposition the usage panel on a screen resize
	cp.app.Gui().Subscribe(gui.OnResize, func(evname string, ev interface{}) {
		width, height := cp.app.Window().Size()
		cp.hud.main.SetSize(float32(width), float32(height))
		cp.hud.usage.SetPosition(float32(width)-340, float32(height)-350)
	})
}

func (cp *CubePlane) updateHud() {

	// add headers
	if cp.hud.headers.Root() == nil {
		for i := range cp.header {
			lineSpace := float32(8.0)
			name := cp.header[i]
			header := gui.NewButton(name)
			header.SetPosition(0, 20.0+(float32(i)*(float32(cp.hud.fontSize)+lineSpace)))
			header.SetStyles(&gui.ButtonStyles{
				Over:   gui.ButtonStyle{FgColor: *math32.NewColor4("Gold", 1.0)},
				Normal: gui.ButtonStyle{FgColor: *math32.NewColor4("White", 1.0)},
			})

			// set an id on the button so we know which one was clicked
			ud := HudData{attrIdx: i}
			header.SetUserData(ud)

			header.Subscribe(gui.OnClick, func(evname string, ev interface{}) {
				if cp.selectedHeaderIdx > -1 {
					unselected := cp.hud.buttons[cp.selectedHeaderIdx]
					unselected.SetStyles(&gui.ButtonStyles{
						Over:   gui.ButtonStyle{FgColor: *math32.NewColor4("Gold", 1.0)},
						Normal: gui.ButtonStyle{FgColor: *math32.NewColor4("White", 1.0)},
					})
				}

				selected := header
				selected.SetStyles(&gui.ButtonStyles{
					Over:   gui.ButtonStyle{FgColor: *math32.NewColor4("Gold", 1.0)},
					Normal: gui.ButtonStyle{FgColor: *math32.NewColor4("Gold", 1.0)},
				})
				ud := selected.UserData().(HudData)
				cp.selectedHeaderIdx = ud.attrIdx

			})

			cp.hud.buttons = append(cp.hud.buttons, header)
			cp.hud.headers.Add(header)
		}
		cp.hud.headers.SetTopChild(cp.hud.values)
		cp.app.Gui().Add(cp.hud.main)
	}

	// add values
	node := cp.plane[cp.cursorX][cp.cursorY]
	ud := node.UserData().(CubeData)
	if ud.attrs == nil {
		return
	}

	// display updated values
	cp.hud.values.DisposeChildren(true)
	for i := range ud.attrs {
		lineSpace := float32(8.0)
		name := cleanCommandPaths(ud.attrs[i])
		value := gui.NewLabel(name)
		value.SetPosition(110, 20.0+(float32(i)*(float32(cp.hud.fontSize)+lineSpace)))
		value.SetColor(math32.NewColor("White"))
		cp.hud.values.Add(value)
	}

}

func cleanCommandPaths(name string) string {
	var clean string
	for _, v := range strings.Split(name, " ") {
		if !strings.HasPrefix(v, "-") {
			clean += " " + v
		}
	}
	clean = path.Base(clean)
	return clean
}
