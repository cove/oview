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
	profile bool
)

func init() {
	rootCmd.AddCommand(viewCmd)
	viewCmd.Flags().BoolVar(&profile, "profile", false, "profile CPU and memory usage")

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

	cp := cubeplane.Init(app, strings.Join(args, " "))
	go PollCmd(args[0], strings.Join(args[1:], " "), cp)

	app.Run()
}

func PollCmd(cmd, args string, cp *cubeplane.CubePlane) {

	var header cubeplane.CubeHeader
	for {
		run := exec.Command(cmd, args)
		stdout, err := run.StdoutPipe()
		if err != nil {
			panic(err)
		}

		if err := run.Start(); err != nil {
			panic(err)
		}

		scanner := bufio.NewScanner(stdout)
		if ok := scanner.Scan(); !ok {
			return
		}

		// XXX assumes first line is header always
		if header == nil {
			line := scanner.Text()
			if txt.IsTableHeader(line) {
				header = strings.Fields(line)
				cp.SetHeader(header)
			}
		}

		table := cubeplane.CubeUpdate{}
		for i := 0; scanner.Scan(); i++ {
			line := scanner.Text()
			fields := strings.Fields(line)
			table = append(table, fields)

			if err := scanner.Err(); err != nil {
				fmt.Fprintln(os.Stderr, "reading standard input:", err)
				return
			}
		}
		cp.UpdateChan <- table

		if err := run.Wait(); err != nil {
			panic(err)
		}
		time.Sleep(5 * time.Second)
	}
}
