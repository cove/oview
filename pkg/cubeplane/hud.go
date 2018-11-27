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
	"strings"

	"github.com/cove/oq/pkg/colors"
	"github.com/cove/oq/pkg/fonts"
	"github.com/g3n/engine/gui"
	"github.com/g3n/engine/math32"
	"github.com/g3n/engine/text"
)

func (cp *CubePlane) initHud() {

	// panel
	panel := gui.NewPanel(230, 200)
	// XXX keep invisable, for later use
	panel.SetVisible(false)
	panel.SetContentSize(200, 190)
	panel.SetPosition(20, 20)
	style := gui.PanelStyle{
		Margin:      gui.RectBounds{0, 0, 0, 0},
		Border:      gui.RectBounds{1, 1, 1, 1},
		Padding:     gui.RectBounds{8, 8, 8, 8},
		BorderColor: *colors.Solarized4("cyan", .2),
		BgColor:     math32.Color4{0, 0, 0, .2},
	}
	panel.ApplyStyle(&style)
	cp.hudPanel = panel

	// font
	font, err := text.NewFontFromData(fonts.OrbitronRegular())
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
	cp.app.Gui().GetPanel().RemoveAll(true)
	cp.app.Gui().GetPanel().Add(cp.hudPanel)

	node := cp.plane[cp.cursorX][cp.cursorY]
	d := node.UserData().(CubeData)
	if d.attrs != nil {
		for i := range cp.header {

			lineSpace := float32(8.0)

			// heading
			heading := gui.NewButton(strings.ToLower(cp.header[i]))
			heading.SetPosition(20, 20.0+(float32(i)*(float32(cp.hudTextSize)+lineSpace)))
			heading.Label.SetFont(cp.hudFont)
			heading.SetStyles(&gui.ButtonStyles{
				Over:   gui.ButtonStyle{FgColor: *math32.NewColor4("cyan", 1.0)},
				Normal: gui.ButtonStyle{FgColor: *colors.Solarized4("base1", 1.0)}},
			)
			heading.Subscribe(gui.OnClick, func(evname string, ev interface{}) {
				fmt.Printf("button %v OnClick\n", cp.header[i])
			})
			cp.app.Gui().GetPanel().Add(heading)

			// values
			basename := path.Base(d.attrs[i]) // everything gets basenamed
			values := gui.NewButton(basename)
			values.SetPosition(110, 20.0+(float32(i)*(float32(cp.hudTextSize)+lineSpace)))
			values.Label.SetFont(cp.hudFont)
			values.SetStyles(&gui.ButtonStyles{
				Over:   gui.ButtonStyle{FgColor: *math32.NewColor4("cyan", 1.0)},
				Normal: gui.ButtonStyle{FgColor: *colors.Solarized4("base2", 1.0)}},
			)
			values.Subscribe(gui.OnClick, func(evname string, ev interface{}) {
				fmt.Printf("button %v OnClick\n", cp.header[i])
			})
			cp.app.Gui().GetPanel().Add(values)
		}
	}
}
