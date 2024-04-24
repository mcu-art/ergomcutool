package proj

import (
	"github.com/mcu-art/ergomcutool/tpl"
)

type ErgomcuProjectTemplateReplacements struct {
	ErgomcutoolVersion string
	ProjectName        string
	DeviceId           string
	OpenocdTarget      string
}

// InstantiateFileErgomcuProjectYaml instantiates ergomcu_project.yaml
// from the template read from embedded assets.
func InstantiateFileErgomcuProjectYaml(
	destPath string,
	r *ErgomcuProjectTemplateReplacements, filePerm uint32) error {
	err := tpl.InstantiateInitCmdTemplate(
		"ergomcu_project.yaml.tmpl",
		destPath,
		r, filePerm)
	return err
}
