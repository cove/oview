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
	"log"
	"os"
	"os/exec"
	"runtime"
	"runtime/pprof"
	"strings"
	"time"

	"github.com/cove/oq/pkg/txt"

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
	profile bool
)

func init() {
	rootCmd.AddCommand(viewCmd)
	viewCmd.Flags().BoolVar(&profile, "profile", false, "profile CPU and memory usage")

}

func cmdView(cmd *cobra.Command, args []string) {
	app, _ := application.Create(application.Options{
		Title:  "oq",
		Width:  800,
		Height: 600,
	})

	if len(args) < 1 {
		cmd.Usage()
		os.Exit(1)
	}

	cp := cubeplane.Init(app, strings.Join(args, " "))
	go ReadInTable(args[0], strings.Join(args[1:], " "), cp)

	if profile {
		fmt.Println("PROFILING")
		f, err := os.Create("profilecpu.prof")
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
		defer pprof.StopCPUProfile()
	}

	app.Run()

	if profile {
		f, err := os.Create("profilemem.prof")
		if err != nil {
			log.Fatal("could not create memory profile: ", err)
		}
		runtime.GC() // get up-to-date statistics
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Fatal("could not write memory profile: ", err)
		}
		f.Close()
	}
}

func ReadInTable(cmd, args string, cp *cubeplane.CubePlane) {

	for {
		run := exec.Command(cmd, args)
		stdout, err := run.StdoutPipe()
		if err != nil {
			panic(err)
		}

		if err := run.Start(); err != nil {
			panic(err)
		}

		var header cubeplane.CubeHeader
		table, header := txt.NewTable(stdout)

		if err := run.Wait(); err != nil {
			panic(err)
		}

		if header == nil {
			cp.SetHeader(header)
		}
		cp.UpdateChan <- table
		time.Sleep(5 * time.Second)
	}
}
