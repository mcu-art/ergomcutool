package cli

import (
	"fmt"

	"github.com/mcu-art/ergomcutool/config"
	"github.com/spf13/cobra"
)

var appShortDescription = fmt.Sprintf("ergomcutool version %s", config.Version)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "ergomcutool",
	Short:   appShortDescription,
	Version: config.Version,
	Long:    `ergomcutool is a tool for integrating STM32CubeMX projects into VSCode.`,

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

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVarP(&gconf.P.ConfigURI, "config", "", "./config.yaml", "config file")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {

	// xzdummy.SayHello() // Test package replacement with local version

	// Validate preliminary config.
	// fmt.Printf("* config file in preliminary configuration: %q\n", gconf.P.ConfigURI)
	// fmt.Println("This is initConfig!")

}
