package cli

import (
	"log"
	"os"

	"github.com/mcu-art/ergomcutool/config"
	"github.com/spf13/cobra"
)

var updateProjectCmd = &cobra.Command{
	Use:   "update-project",
	Short: "Update project and patch makefile",
	Run:   updateProject,
}

func init() {
	rootCmd.AddCommand(updateProjectCmd)
	updateProjectCmd.PersistentFlags().StringP("path", "p", "", "Path to project root directory")
}

func updateProject(cmd *cobra.Command, args []string) {
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
	config.ParseErgomcutoolConfig(false)

	// Read ergomcu_project.yaml

	log.Printf("The project was successfully updated.")

}
