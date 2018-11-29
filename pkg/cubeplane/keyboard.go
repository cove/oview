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
	"github.com/g3n/engine/math32"
	"github.com/g3n/engine/window"
)

func (cp *CubePlane) onKey(evname string, ev interface{}) {

	key := ev.(*window.KeyEvent)
	switch key.Keycode {
	case window.KeyLeft:
		fallthrough
	case window.KeyA:
		z := cp.app.Scene().Rotation().Z
		cp.cursorX -= int64(math32.Round(math32.Cos(z)))
		if cp.cursorX > cp.size-1 {
			cp.cursorX = cp.size - 1
		}
		if cp.cursorX < 0 {
			cp.cursorX = 0
		}

		cp.cursorY += int64(math32.Round(math32.Sin(z)))
		if cp.cursorY < 0 {
			cp.cursorY = 0
		}
		if cp.cursorY > cp.size-1 {
			cp.cursorY = cp.size - 1
		}
		cp.updateSelected()

	case window.KeyRight:
		fallthrough
	case window.KeyD:
		z := cp.app.Scene().Rotation().Z
		cp.cursorX += int64(math32.Round(math32.Cos(z)))
		if cp.cursorX > cp.size-1 {
			cp.cursorX = cp.size - 1
		}
		if cp.cursorX < 0 {
			cp.cursorX = 0
		}

		cp.cursorY -= int64(math32.Round(math32.Sin(z)))
		if cp.cursorY < 0 {
			cp.cursorY = 0
		}
		if cp.cursorY > cp.size-1 {
			cp.cursorY = cp.size - 1
		}
		cp.updateSelected()

	case window.KeyUp:
		fallthrough
	case window.KeyW:
		z := cp.app.Scene().Rotation().Z
		cp.cursorX += int64(math32.Round(math32.Sin(z)))
		if cp.cursorX > cp.size-1 {
			cp.cursorX = cp.size - 1
		}
		if cp.cursorX < 0 {
			cp.cursorX = 0
		}

		cp.cursorY += int64(math32.Round(math32.Cos(z)))
		if cp.cursorY < 0 {
			cp.cursorY = 0
		}
		if cp.cursorY > cp.size-1 {
			cp.cursorY = cp.size - 1
		}
		cp.updateSelected()

	case window.KeyDown:
		fallthrough
	case window.KeyS:
		z := cp.app.Scene().Rotation().Z
		cp.cursorX -= int64(math32.Round(math32.Sin(z)))
		if cp.cursorX > cp.size-1 {
			cp.cursorX = cp.size - 1
		}
		if cp.cursorX < 0 {
			cp.cursorX = 0
		}

		cp.cursorY -= int64(math32.Round(math32.Cos(z)))
		if cp.cursorY < 0 {
			cp.cursorY = 0
		}
		if cp.cursorY > cp.size-1 {
			cp.cursorY = cp.size - 1
		}
		cp.updateSelected()

	case window.KeyF:
		cp.cubeWireframe = !cp.cubeWireframe

	case window.KeyP:
		// pause
		fallthrough
	case window.KeyR:
		cp.rotate = !cp.rotate

	case window.KeyQ:
		cp.app.Quit()
	}
}
