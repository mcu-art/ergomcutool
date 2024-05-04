package cli

import (
	"fmt"

	"github.com/mcu-art/ergomcutool/config"
	"github.com/spf13/cobra"
)

var appShortDescription = fmt.Sprintf("ergomcutool version %s", config.Version)

// verbose is a persistent flag that can be used by all CLI commands.
var verbose bool

var rootCmd = &cobra.Command{
	Use:     "ergomcutool",
	Short:   appShortDescription,
	Version: config.Version,
	Long: `ergomcutool is a tiny project manager that helps
to integrate STM32CubeMX projects into VSCode.
More information at https://github.com/mcu-art/ergomcutool`,

	// Comment the following out if there is no need for the root cmd execution.
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(appShortDescription + ".\nType 'ergomcutool -h' for more information.")
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	cobra.OnInitialize(initConfig)

	// Hide the 'completion' command
	rootCmd.CompletionOptions.HiddenDefaultCmd = true
	// Add persistent flags
	rootCmd.PersistentFlags().BoolVarP(
		&verbose, "verbose", "", false, "Verbose mode")
}

func initConfig() {
	// nothing to do
}
