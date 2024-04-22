package cli

import (
	"log"
	"os"

	"github.com/mcu-art/ergomcutool/config"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update project and regenerate makefile",
	Run:   OnUpdateCmd,
}

func init() {
	rootCmd.AddCommand(updateCmd)
	updateCmd.PersistentFlags().StringP("path", "p", "", "Path to project root directory")
}

func OnUpdateCmd(cmd *cobra.Command, args []string) {
	if len(args) > 0 {
		log.Fatalf("error: too many CLI argument(s): %+v\n", args)
	}
	path := cmd.Flag("path").Value.String()
	if path != "" {
		if err := os.Chdir(path); err != nil {
			log.Fatalf("error: can't cd into directory %q: %v\n", path, err)
		}
	}
	// cwd, _ := os.Getwd()
	log.Printf("Updating the project...\n")
	config.ParseErgomcutoolConfig()

	log.Printf("The project was successfully updated.")

}
