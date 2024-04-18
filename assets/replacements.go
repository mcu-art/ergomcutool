package assets

import "github.com/mcu-art/ergomcutool/utils"

type ErgomcuProjectYamlReplacements struct {
	ErgomcutoolVersion string
	ProjectName        string
	DeviceId           string
	OpenocdTarget      string
}

// InstantiateFileErgomcuProjectYaml instantiates ergomcu_project.yaml
// from the template read from embedded assets.
func InstantiateFileErgomcuProjectYaml(
	destPath string,
	r *ErgomcuProjectYamlReplacements, filePerm uint32) error {
	replacementMap, err := utils.StructToMap(r)
	if err != nil {
		return err
	}
	err = InstantiateAssetTemplate(
		"default_ergomcu_project.yaml.tmpl",
		destPath,
		replacementMap, filePerm)
	return err
}
