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

package colors

import (
	"strings"

	"github.com/g3n/engine/math32"
)

var solarized = map[string]math32.Color{
	"base03":  {0.0159, 0.1265, 0.1597},
	"base02":  {0.0394, 0.1601, 0.1983},
	"base01":  {0.2767, 0.3567, 0.3830},
	"base00":  {0.3244, 0.4072, 0.4385},
	"base0":   {0.4406, 0.5096, 0.5168},
	"base1":   {0.5060, 0.5649, 0.5637},
	"base2":   {0.9161, 0.8900, 0.7978},
	"base3":   {0.9895, 0.9579, 0.8641},
	"yellow":  {0.6475, 0.4676, 0.0235},
	"orange":  {0.7418, 0.2132, 0.0735},
	"red":     {0.8192, 0.1084, 0.1414},
	"magenta": {0.7774, 0.1081, 0.4352},
	"violet":  {0.3479, 0.3513, 0.7179},
	"blue":    {0.1275, 0.4627, 0.7823},
	"cyan":    {0.1468, 0.5709, 0.5250},
	"green":   {0.4498, 0.5412, 0.0202},
}

func Solarized(name string) *math32.Color {
	c := solarized[strings.ToLower(name)]
	return &c
}
func Solarized4(name string, alpha float32) *math32.Color4 {
	c := solarized[strings.ToLower(name)]
	c4 := &math32.Color4{c.R, c.B, c.G, alpha}
	return c4
}
