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
	"io"
	"strings"
	"unicode"
)

func NewTable(fd io.Reader) ([]string, [][]string, error) {

	var table [][]string
	var header []string

	scanner := bufio.NewScanner(fd)
	for scanner.Scan() {
		line := scanner.Text()

		// skip to header
		if header == nil {
			if IsTableHeader(line) {
				for _, v := range strings.Fields(line) {
					header = append(header, strings.TrimSpace(v))
				}
			}
			continue
		}

		fields := strings.Fields(line)

		if len(fields) > len(header) {
			// concat trailing fields to last column
			// (e.g. a process name with spaces in it)
			cat1 := fields[len(header)-1:]
			cat2 := strings.Join(cat1, " ")
			catFields := append(fields[0:len(header)-1], cat2)
			table = append(table, catFields)
		} else if len(fields) == len(header) {
			// normal line matches up with header
			table = append(table, fields)
		} else {
			// incomplete line
			continue
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, nil, err
	}

	return header, table, nil
}

func findFirstUniqueValuedColumn(table [][]string) int {

	uniqueColumn := -1
	for x := range table[0] {
		previous := ""
		for y := range table {
			if IsTableHeader(strings.Join(table[y], " ")) {
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

	return uniqueColumn
}

func IsTableHeader(s string) bool {
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
