// This code was generated with assistance from Claude AI by Anthropic.
// It is provided under the MIT License, which allows for free use, modification,
// and distribution with proper attribution.
//
// MIT License
//
// Copyright (c) [2025] [Michael Rubin]
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package cmd

import (
	"fmt"
	"net"
	"os"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	gotris "github.com/matchstick/gotris/lib"
)

const serverVersion string = "1.0.0"

func versionCmd() *cobra.Command {
	const numVersionCmdArgs = 0

	return &cobra.Command{
		Use:   "version",
		Short: "Output release version.",
		Long:  `Output release version. `,
		Args:  cobra.MaximumNArgs(numVersionCmdArgs),
		Run: func(cmd *cobra.Command, args []string) {
			version := serverVersion
			fmt.Printf("gotris version: %s\n", version)
		},
	}
}

func isPortLegit(port int) error {
	minPortNum := 1024
	maxPortNum := 49151
	if port < minPortNum || port > maxPortNum {
		return fmt.Errorf("Port must be between %d and %d not %d\n", minPortNum, maxPortNum, port)
	}

	address := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", address)

	if err != nil {
		// Port is in use or otherwise unavailable
		return fmt.Errorf("Port %d is in use\n", port)
	}

	// Don't forget to close the listener when done
	defer listener.Close()

	return nil

}

func newStartCmd() *cobra.Command {
	var port int
	var numberOfPlayers int

	startCmd := &cobra.Command{
		Use:   "start",
		Short: "Start up the server for gotris.",
		Long:  `Starts up the server and requires the arguments <port> and <# sessions>`,
		Run: func(cmd *cobra.Command, args []string) {
			err := isPortLegit(port)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			if numberOfPlayers < 1 || numberOfPlayers > 1 {
				fmt.Fprintf(os.Stderr, "Only one player supported now not %d\n", numberOfPlayers)
				os.Exit(1)
			}

			err = gotris.NewServer(port, numberOfPlayers)
			if err != nil {
				fmt.Printf("NewServer failed. Error %s\n", err)
				os.Exit(1)
			}
		},
	}

	startCmd.Flags().IntVarP(&port, "port", "p", 8080, "Port to Listen on")
	startCmd.Flags().IntVarP(&numberOfPlayers, "players", "n", 1, "Number of Players")

	return startCmd
}

func rootCmd() *cobra.Command {

	// rootCmd represents the base command when called without any subcommands.
	return &cobra.Command{
		Use:   "gotris",
		Short: "Web based tetris game in golang",
		Long: `Fun little project to exercise go skills and write tetris
Check out http://github.com/matchstick/gotris for more details.`,
	}
}

// Execute is the single function to set up the complete set of nested cobra
// based commands that provide CLI functionality to the exifsort library.
func Execute() {
	cobra.OnInitialize(initConfig)

	rootCmd := rootCmd()

	rootCmd.AddCommand(newStartCmd())
	rootCmd.AddCommand(versionCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	const cfgFilename string = ".gotris"
	var cfgFile string
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

		// Search config in home directory with name ".gotris"
		// (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(cfgFilename)
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
