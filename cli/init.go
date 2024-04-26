package cli

import (
	"log"
	"os"

	"github.com/mcu-art/ergomcutool/config"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize ergomcutool globally",
	Run:   InitErgomcutool,
}

var initCmdForce bool

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.PersistentFlags().BoolVarP(&initCmdForce, "force", "f", false, "Replace current user config directory if exists")
}

func InitErgomcutool(cmd *cobra.Command, args []string) {
	if len(args) > 0 {
		log.Fatalf("error: too many CLI argument(s): %+v\n", args)
	}

	userConfigDirExists := config.CheckUserConfigDirExists()
	if userConfigDirExists {
		if !initCmdForce {
			log.Fatalf(`error: ergomcutool is already initialized.
If you want to delete current user settings and re-initialize the tool, use --force flag.
`)
		}
	}
	if userConfigDirExists && initCmdForce {
		err := os.RemoveAll(config.UserConfigDir)
		if err != nil {
			log.Fatalf(`error: failed to remove %q.
`, config.UserConfigDir)
		}
	}
	err := config.CreateUserConfig()
	if err != nil {
		log.Fatalf("error: failed to create ergomcutool config: %v\n", err)
	}

	if userConfigDirExists && initCmdForce {
		log.Printf("ergomcutool was successfully re-initialized.")
	} else {
		log.Printf("ergomcutool was successfully initialized.")
	}
}
