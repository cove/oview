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

package text2table

import (
	"bufio"
	"io"
	"strings"
)

func NewTable(fd io.Reader) ([]string, [][]string, error) {

	var table [][]string
	var header []string
	sep := ""

	scanner := bufio.NewScanner(fd)
	for scanner.Scan() {
		line := scanner.Text()

		// Assume first line is header
		if header == nil {
			if strings.ContainsAny(line, ",") {
				sep = ","
			} else if strings.ContainsAny(line, "\t") {
				sep = "\t"
			} else if strings.ContainsAny(line, ":") {
				sep = ":"
			}

			if sep != "" {
				for _, v := range strings.Split(line, sep) {
					header = append(header, strings.TrimSpace(v))
				}
			} else {
				for _, v := range strings.Fields(line) {
					header = append(header, strings.TrimSpace(v))
				}
			}

			continue
		}

		var fields []string
		if sep != "" {
			fields = strings.Split(line, sep)
		} else {
			fields = strings.Fields(line)
		}

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
