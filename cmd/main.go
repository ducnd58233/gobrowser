package main

import (
	"fmt"
	"os"

	"github.com/ducnd58233/gobrowser/internal/ui"
	"github.com/spf13/cobra"
)

var (
	debugFlag   bool
	verboseFlag bool
)

const (
	appVersion = "1.0.0"
	appName    = "GoBrowser"
)

var rootCmd = &cobra.Command{
	Use:   "gobrowser",
	Short: "A modern web browser built with Go and Gio UI",
	Long: `GoBrowser is a production-ready web browser engine built in Go with Gio UI.
It features complete HTML/CSS rendering, tab management, and modern layout engine.`,
	Version: appVersion,
	Run:     runBrowser,
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of GoBrowser",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("%s v%s\n", appName, appVersion)
	},
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&debugFlag, "debug", "d", false, "Enable debug mode with detailed logging")
	rootCmd.PersistentFlags().BoolVarP(&verboseFlag, "verbose", "v", false, "Enable verbose output")

	rootCmd.AddCommand(versionCmd)
}

func runBrowser(cmd *cobra.Command, args []string) {
	if debugFlag {
		if err := os.Setenv("GOBROWSER_DEBUG", "true"); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to set debug environment variable: %v\n", err)
		}
	}

	if verboseFlag {
		fmt.Printf("Starting %s v%s\n", appName, appVersion)
		if debugFlag {
			fmt.Println("Debug mode enabled")
		}
	}

	window := ui.NewMainWindow(debugFlag)
	window.Run()
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
