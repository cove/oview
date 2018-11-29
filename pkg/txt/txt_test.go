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
	"io"
	"os"
	"testing"
)

func TestNewTable(t *testing.T) {
	psdata, _ := os.Open("testdata/ps-aux-osx.txt")

	type args struct {
		fd io.Reader
	}
	tests := []struct {
		name string
		args args
		want error
	}{
		{
			name: "ps aux header",
			args: args{fd: psdata},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := NewTable(tt.args.fd)
			if err != tt.want {
				t.Errorf("isTableHeader() = %v, want %v", err, tt.want)
			}
		})
	}
}

func Test_isTableHeader(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "ps aux header",
			args: args{s: "USER               PID  %CPU %MEM      VSZ    RSS   TT  STAT STARTED      TIME COMMAND"},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsTableHeader(tt.args.s); got != tt.want {
				t.Errorf("IsTableHeader() = %v, want %v", got, tt.want)
			}
		})
	}
}
