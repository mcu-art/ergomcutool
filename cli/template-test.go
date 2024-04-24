// THIS IS A DUMMY COMMAND TO BE REMOVED
package cli

import (
	"log"

	"github.com/spf13/cobra"
)

var templateTestCmd = &cobra.Command{
	Use:   "template-test",
	Short: "Test template instantiation",
	Run:   OnTemplateTestCmd,
}

func init() {
	rootCmd.AddCommand(templateTestCmd)
	templateTestCmd.PersistentFlags().StringP("name", "n", "", "template name")
}

func OnTemplateTestCmd(cmd *cobra.Command, args []string) {
	if len(args) > 0 {
		log.Fatalf("error: too many CLI argument(s): %+v\n", args)
	}
	name := cmd.Flag("name").Value.String()
	if name == "" {
		log.Fatalf("error: template name must be specified\n")

	}

	/*
		r := &proj.ErgomcuProjectTemplateReplacements{
			ErgomcutoolVersion: config.Version,
			ProjectName:        "Dummy Project",
			DeviceId:           "STM32F1x",
			OpenocdTarget:      "STM32F103x",
		}
		err := proj.InstantiateFileErgomcuProjectYaml(
			"./ergomcu_project.yaml", r, config.DefaultFilePermissions)

		if err != nil {
			log.Fatalf("error: template instantiation failed: %v\n", err)
		}
	*/

}
