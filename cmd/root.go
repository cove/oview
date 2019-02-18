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

	"github.com/cove/oview/pkg/cubeplane"
	"github.com/cove/oview/pkg/text2table"
	"github.com/g3n/engine/util/application"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

var (
	profile   bool
	refresh   = 3
	wireframe bool
	size      = int64(30)
	rotation  = 32
	pause     bool
	file      string
	command   string
	usage     bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "oview",
	Version: "unknown",
	Short:   "Displays a text table as a 3D rotating plane of cubes",
	Long: `Takes a text table and displays it as a 3D rotating plane of cubes,
where the size of the cube grows and shrinks based on the values
associated with the cube. This allows one to see a large table that may 
not otherwise fit on the screen and quickly see the changes in it as it
changes over time.
`,
	Run: view,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.oview.yaml)")
	rootCmd.Flags().BoolVar(&profile, "profile", profile, "Profile CPU and memory usage")
	rootCmd.PersistentFlags().Int64VarP(&size, "size", "s", size, "Size of cube plane")
	rootCmd.PersistentFlags().IntVarP(&refresh, "interval", "i", refresh, "Refresh data interval in seconds")
	rootCmd.PersistentFlags().IntVarP(&rotation, "rotations", "r", rotation, "How many seconds each rotation takes")
	rootCmd.PersistentFlags().BoolVarP(&pause, "pause", "p", pause, "Start up with rotation paused to improve performance")
	rootCmd.PersistentFlags().BoolVarP(&wireframe, "wireframe", "w", wireframe, "Render cubes as wireframes to improve performance")
	rootCmd.PersistentFlags().StringVarP(&file, "file", "f", file, "Load data from file or use '-' to read from stdin")
	rootCmd.PersistentFlags().StringVarP(&command, "command", "c", command, "Command to run to get data from")
	rootCmd.PersistentFlags().BoolVarP(&usage, "usage", "u", true, "Show usage text in screen on startup")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".oview" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".oview")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

func view(cmd *cobra.Command, args []string) {

	// validate command line args
	if file == "" && command == "" {
		fmt.Fprintln(os.Stderr, "Please specify either -f or -c to load data")
		cmd.Usage()
		os.Exit(-1)
	}

	app, _ := application.Create(application.Options{
		Title:  "oview",
		Width:  900,
		Height: 900,
	})

	cp := cubeplane.Init(
		app,
		strings.Join(args, " "),
		refresh,
		wireframe,
		size,
		rotation,
		pause,
		usage)

	if command != "" {
		go PollCmd(command, cp)
	} else if file != "" {
		go PollFile(file, cp)
	}

	app.Run()
}

func PollCmd(command string, cp *cubeplane.CubePlane) {

	// split out cmd and args
	tmp := strings.Fields(command)
	cmd := tmp[0]
	args := strings.Join(tmp[1:], " ")

	failed := func(err error) {
		fmt.Fprintf(os.Stderr, "Command failed to run command %s %s: %s\n", cmd, args, err)
		time.Sleep(3 * time.Second)
	}

	// main loop that polls the command
	needsHeader := true
	for {
		run := exec.Command(cmd, args)
		stdout, err := run.StdoutPipe()
		if err != nil {
			failed(err)
			continue
		}

		if err := run.Start(); err != nil {
			failed(err)
			continue
		}

		header, table, err := text2table.NewTable(stdout)
		if err = run.Wait(); err != nil {
			failed(err)
			continue
		}

		if needsHeader {
			cp.SetHeader(header)
			needsHeader = false
		}

		// send the table to the cube plane
		cp.UpdateChan <- table
		time.Sleep(3 * time.Second)
	}
}

func PollFile(file string, cp *cubeplane.CubePlane) {

	failed := func(err error) {
		fmt.Fprintf(os.Stderr, "Failed to load data form file %s: %s\n", file, err)
		time.Sleep(3 * time.Second)
	}

	needsHeader := true
	for {
		var input io.Reader
		var fd *os.File
		if file == "-" {
			input = bufio.NewReader(os.Stdin)
		} else {
			fd, err := os.Open(file)
			if err != nil {
				failed(err)
				fd.Close()
				continue
			}
			input = bufio.NewReader(fd)
		}

		header, table, _ := text2table.NewTable(input)
		if needsHeader {
			cp.SetHeader(header)
			needsHeader = false
		}
		cp.UpdateChan <- table
		time.Sleep(3 * time.Second)
		fd.Close()
	}
}
