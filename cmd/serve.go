package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	serveCmd.PersistentFlags().IntVarP(&global.Port, "port", "p", 8080, "Server port. If this is in use, an unused port is chosen.")
	rootCmd.AddCommand(serveCmd)
}

var serveCmd = &cobra.Command{
	Use:   "serve [package]",
	Short: "Serve locally",
	Long:  "Starts a webserver locally, and recompiles the WASM on every page refresh, for testing and development.",
	Args:  cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("serve!")
	},
}
