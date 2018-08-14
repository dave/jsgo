package cmd

import (
	"fmt"
	"os"

	"github.com/dave/jsgo/cmd/cmdconfig"
	"github.com/spf13/cobra"
)

var global = &cmdconfig.Config{}

func init() {
	rootCmd.PersistentFlags().StringVarP(&global.Index, "index", "i", "index.jsgo.html", "Specify the index page. If omitted, use `index.jsgo.html` if it exists.")
	rootCmd.PersistentFlags().BoolVarP(&global.Quiet, "quiet", "q", false, "Suppress status messages.")
	rootCmd.PersistentFlags().BoolVarP(&global.Open, "open", "o", false, "Open the page in a browser.")
	rootCmd.PersistentFlags().StringVarP(&global.Command, "command", "c", "go", "Name of the go command.")
	rootCmd.PersistentFlags().StringVarP(&global.Flags, "flags", "f", "", "Flags to pass to the go build command.")
}

var rootCmd = &cobra.Command{
	Use:   "jsgo",
	Short: "Compile Go to WASM, test locally or deploy to jsgo.io",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
