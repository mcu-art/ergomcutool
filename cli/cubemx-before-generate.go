package cli

import (
	"log"

	"github.com/mcu-art/ergomcutool/config"
	"github.com/mcu-art/ergomcutool/mkf"
	"github.com/mcu-art/ergomcutool/utils"
	"github.com/spf13/cobra"
)

var cubemxBeforeGenerateCmd = &cobra.Command{
	Use:   "cubemx-before-generate",
	Short: "Do preliminary tasks before STM32CubeMX generates the code",
	Run:   preCubeMXGenerate,
}

var (
	cubemxBeforeGen_Makefile string
)

func init() {
	rootCmd.AddCommand(cubemxBeforeGenerateCmd)
	cubemxBeforeGenerateCmd.PersistentFlags().StringVarP(
		&cubemxBeforeGen_Makefile, "makefile", "m", "Makefile", "Specify custom path to Makefile")
}

func preCubeMXGenerate(cmd *cobra.Command, args []string) {
	if len(args) > 0 {
		log.Fatalf("error: too many CLI argument(s): %+v\n", args)
	}
	// Ensure that ergomcutool is initialized
	config.EnsureUserConfigExists()

	// Check iof makefile exists
	if !utils.FileExists(cubemxBeforeGen_Makefile) {
		// Nothing to do
		log.Printf("'cubemx-before-generate': no makefile found, backup job skipped.\n")
		return
	}

	// Backup the makefile
	err := mkf.BackupMakefile(cubemxBeforeGen_Makefile)
	if err != nil {
		log.Printf("error: failed to backup the makefile:%v\n", err)
	} else {
		log.Printf("Makefile backup created successfully.\n")
	}
}
