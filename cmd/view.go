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

package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/cove/oq/pkg/text2table"
	profile2 "github.com/pkg/profile"

	"github.com/cove/oq/pkg/cubeplane"

	"github.com/g3n/engine/util/application"

	"github.com/spf13/cobra"
)

// viewCmd represents the view command
var viewCmd = &cobra.Command{
	Use:   "view",
	Short: "View text table",
	Run:   cmdView,
}

var (
	profile   bool
	refresh   = 5
	wireframe bool
	size      = int64(20)
	rotation  = 32
	pause     bool
	file      string
	command   string
)

func init() {
	rootCmd.AddCommand(viewCmd)
	viewCmd.Flags().BoolVar(&profile, "profile", profile, "Profile CPU and memory usage")
	viewCmd.PersistentFlags().Int64VarP(&size, "size", "s", size, "Size of cube plane")
	viewCmd.PersistentFlags().IntVarP(&refresh, "interval", "i", refresh, "Refresh data interval in seconds")
	viewCmd.PersistentFlags().IntVarP(&rotation, "rotations", "r", rotation, "How many seconds each rotation takes")
	viewCmd.PersistentFlags().BoolVarP(&pause, "pause", "p", pause, "Start up with rotation paused to improve performance")
	viewCmd.PersistentFlags().BoolVarP(&wireframe, "wireframe", "w", wireframe, "Render cubes as wireframes to improve performance")
	viewCmd.PersistentFlags().StringVarP(&file, "file", "f", file, "Load data from file or use '-' to read from stdin")
	viewCmd.PersistentFlags().StringVarP(&command, "command", "c", command, "Command to run to get data from")

}

func cmdView(cmd *cobra.Command, args []string) {

	if profile {
		fmt.Println("PROFILING")
		defer profile2.Start(profile2.MemProfile).Stop()
	}

	// validate command line args
	if file == "" && command == "" {
		fmt.Fprintln(os.Stderr, "Please specify either -f or -c to load data")
		cmd.Usage()
		os.Exit(-1)
	}

	app, _ := application.Create(application.Options{
		Title:  "oq",
		Width:  800,
		Height: 600,
	})

	cp := cubeplane.Init(
		app,
		strings.Join(args, " "),
		refresh,
		wireframe,
		size,
		rotation,
		pause)

	if command != "" {
		go PollCmd(command, strings.Join(args[1:], " "), cp)
	} else if file != "" {
		go PollFile(file, cp)
	}

	app.Run()
}

func PollCmd(cmd, args string, cp *cubeplane.CubePlane) {

	needsHeader := true
	for {
		run := exec.Command(cmd, args)
		stdout, err := run.StdoutPipe()
		if err != nil {
			panic(err)
		}

		if err := run.Start(); err != nil {
			panic(err)
		}

		header, table, err := text2table.NewTable(stdout)
		if needsHeader {
			cp.SetHeader(header)
			needsHeader = false
		}
		cp.UpdateChan <- table

		if err := run.Wait(); err != nil {
			panic(err)
		}
		time.Sleep(5 * time.Second)
	}
}

func PollFile(file string, cp *cubeplane.CubePlane) {
	needsHeader := true
	for {
		var input io.Reader
		var fd *os.File
		if file == "-" {
			input = bufio.NewReader(os.Stdin)
		} else {
			fd, err := os.Open(file)
			if err != nil {
				panic(err)
			}
			input = bufio.NewReader(fd)
		}

		header, table, _ := text2table.NewTable(input)
		if needsHeader {
			cp.SetHeader(header)
			needsHeader = false
		}
		cp.UpdateChan <- table

		fd.Close()
	}
}
