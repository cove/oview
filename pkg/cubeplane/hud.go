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

	"github.com/cove/oq/pkg/colors"
	"github.com/cove/oq/pkg/fonts"
	"github.com/g3n/engine/gui"
	"github.com/g3n/engine/math32"
	"github.com/g3n/engine/text"
)

func (cp *CubePlane) initHud() {

	cp.hudHeaders = gui.NewPanel(600, 900)
	cp.hudHeaders.SetPosition(10, 10)
	//cp.hudHeaders.SetBorders(1, 1, 1, 1)

	cp.hudValues = gui.NewPanel(600, 900)
	//cp.hudValues.SetBorders(1, 1, 1, 1)
	cp.hudHeaders.Add(cp.hudValues)

	// load font
	font, err := text.NewFontFromData(fonts.Gallant12x22())
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

	// add headers
	if cp.hudHeaders.Root() == nil {
		for i := range cp.header {
			lineSpace := float32(8.0)
			name := cp.header[i]
			header := gui.NewButton(name)
			header.SetPosition(20, 20.0+(float32(i)*(float32(cp.hudTextSize)+lineSpace)))
			header.Label.SetFont(cp.hudFont)
			header.SetStyles(&gui.ButtonStyles{
				Over:   gui.ButtonStyle{FgColor: *math32.NewColor4("cyan", 1.0)},
				Normal: gui.ButtonStyle{FgColor: *colors.Solarized4("base1", 1.0)},
			})

			// set an id on the button so we know which one was clicked
			ud := HudData{attrIdx: i}
			header.SetUserData(ud)

			header.Subscribe(gui.OnClick, func(evname string, ev interface{}) {
				if cp.selectedHeaderIdx > -1 {
					unselected := cp.hubHeaderButtons[cp.selectedHeaderIdx]
					unselected.SetStyles(&gui.ButtonStyles{
						Over:   gui.ButtonStyle{FgColor: *math32.NewColor4("cyan", 1.0)},
						Normal: gui.ButtonStyle{FgColor: *colors.Solarized4("base1", 1.0)},
					})
				}

				selected := header
				selected.SetStyles(&gui.ButtonStyles{
					Over:   gui.ButtonStyle{FgColor: *math32.NewColor4("yellow", 1.0)},
					Normal: gui.ButtonStyle{FgColor: *math32.NewColor4("white", 1.0)},
				})
				ud := selected.UserData().(HudData)
				cp.selectedHeaderIdx = ud.attrIdx

			})

			cp.hubHeaderButtons = append(cp.hubHeaderButtons, header)
			cp.hudHeaders.Add(header)
		}
		cp.hudHeaders.SetTopChild(cp.hudValues)
		cp.app.Gui().Add(cp.hudHeaders)
	}

	// add values
	node := cp.plane[cp.cursorX][cp.cursorY]
	ud := node.UserData().(CubeData)
	if ud.attrs == nil {
		return
	}

	cp.hudValues.DisposeChildren(true)
	for i := range ud.attrs {
		lineSpace := float32(8.0)
		name := cleanCommandPaths(ud.attrs[i])
		value := gui.NewLabel(name)
		value.SetPosition(110, 20.0+(float32(i)*(float32(cp.hudTextSize)+lineSpace)))
		value.SetColor(colors.Solarized("base1"))
		value.SetFont(cp.hudFont)
		cp.hudValues.Add(value)
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
