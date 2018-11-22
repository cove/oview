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

package txt

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode"
)

func NewTable(fd io.Reader) (map[string][]string, []string, error) {
	tmpTable := make([][]string, 500)
	scanner := bufio.NewScanner(fd)
	for i := range tmpTable {
		if ok := scanner.Scan(); !ok {
			break
		}

		line := scanner.Text()
		fields := strings.Fields(line)
		tmpTable[i] = make([]string, len(fields))
		tmpTable[i] = fields

		if err := scanner.Err(); err != nil {
			fmt.Fprintln(os.Stderr, "reading standard input:", err)
		}
	}

	key, header := findFirstUniqueValuedColumn(tmpTable)
	if key < 0 {
		panic("no key column")
	}

	newTable := make(map[string][]string, len(tmpTable))
	for y := range tmpTable {
		if tmpTable[y] == nil {
			break
		}
		newTable[tmpTable[y][key]] = tmpTable[y]
	}

	return newTable, header, nil
}

func findFirstUniqueValuedColumn(table [][]string) (int, []string) {

	var header []string
	uniqueColumn := -1
	for x := range table[0] {
		previous := ""
		for y := range table {
			if isTableHeader(strings.Join(table[y], " ")) {
				header = table[y]
				continue
			}

			if table[y] == nil {
				break
			}

			if previous == "" {
				previous = table[y][x]
			} else if previous != table[y][x] {
				uniqueColumn = x
			} else {
				uniqueColumn = -1
				break
			}
		}

		if uniqueColumn != -1 {
			break
		}
	}

	return uniqueColumn, header
}

func isTableHeader(s string) bool {
	score := 0.0
	for _, c := range s {
		if unicode.IsUpper(c) || unicode.IsSpace(c) || c == '-' || c == '=' {
			score += 1.0
		}
		if unicode.IsDigit(c) || c == '/' || c == '\\' {
			score -= 1.0
		}
	}
	confidence := score/float64(len(s)) > .8

	return confidence
}
