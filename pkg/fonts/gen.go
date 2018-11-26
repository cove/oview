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

// +build ignore
//go:generate go run gen.go

package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

func main() {
	fonts := []string{
		"Orbitron/Orbitron-Regular.ttf",
		"Orbitron/Orbitron-Medium.ttf",
		"Orbitron/Orbitron-Bold.ttf",
		"Orbitron/Orbitron-Black.ttf",
	}

	out, err := os.Create("fonts.go")
	if err != nil {
		panic(err)
	}

	header := `// Code generated fonts.go DO NOT EDIT.
package fonts

`
	fmt.Fprintln(out, header)

	for i := range fonts {
		fd, err := os.Open(fonts[i])
		if err != nil {
			panic(err)
		}

		name := path.Base(fonts[i])
		name = strings.TrimSuffix(name, ".ttf")
		name = strings.Replace(name, "-", "", -1)
		fmt.Fprintf(out, "func %s() []byte {\t\n\tb := []byte{\n\t", name)

		bin, _ := ioutil.ReadAll(fd)
		for j := range bin {
			fmt.Fprintf(out, "%#x, ", bin[j])
			if j%20 == 0 {
				fmt.Fprint(out, "\n\t")
			}
		}

		footer := `}
	return b
}
	`
		_, err = fmt.Fprintln(out, footer)
	}

}
