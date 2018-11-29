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
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/cove/oq/pkg/txt"
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
)

func init() {
	rootCmd.AddCommand(viewCmd)
	viewCmd.Flags().BoolVar(&profile, "profile", profile, "Profile CPU and memory usage")
	viewCmd.PersistentFlags().Int64Var(&size, "size", size, "Size of cube plane")
	viewCmd.PersistentFlags().IntVar(&refresh, "refresh", refresh, "Refresh data in seconds")
	viewCmd.PersistentFlags().IntVar(&rotation, "rotations", rotation, "How many seconds each rotation takes")
	viewCmd.PersistentFlags().BoolVar(&pause, "pause", pause, "Start up with rotation paused to improve performance")
	viewCmd.PersistentFlags().BoolVar(&wireframe, "wireframe", wireframe, "Render cubes as wireframes to improve performance")
}

func cmdView(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		cmd.Usage()
		os.Exit(1)
	}

	if profile {
		fmt.Println("PROFILING")
		defer profile2.Start(profile2.MemProfile).Stop()
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
	go PollCmd(args[0], strings.Join(args[1:], " "), cp)

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

		header, table, err := txt.NewTable(stdout)
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
