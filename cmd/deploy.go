package cmd

import (
	"fmt"
	"os"

	"github.com/dave/jsgo/cmd/deployer"
	"github.com/spf13/cobra"
)

func init() {
	deployCmd.PersistentFlags().StringVarP(&global.Template, "template", "t", "{{ .Page }}", "Template defining the output returned by the deploy command. Variables: Page (string), Loader (string).")
	deployCmd.PersistentFlags().BoolVarP(&global.Json, "json", "j", false, "Return all template variables as a json blob from the deploy command.")
	rootCmd.AddCommand(deployCmd)
}

var deployCmd = &cobra.Command{
	Use:   "deploy [package]",
	Short: "Compile and deploy",
	Long:  "Compiles Go to WASM and deploys to the jsgo.io CDN.",
	Args:  cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			global.Path = args[0]
		}
		if err := deployer.Start(global); err != nil {
			fmt.Fprintf(os.Stderr, err.Error())
			os.Exit(1)
		}
	},
}
