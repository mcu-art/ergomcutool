package proj

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/mcu-art/ergomcutool/config"
	"github.com/mcu-art/ergomcutool/utils"
	"gopkg.in/yaml.v2"
)

type ErgomcuProjectTemplateReplacements struct {
	ErgomcutoolVersion string
	ProjectName        string
	DeviceId           string
	OpenocdTarget      string
}

/*
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
*/

type ErgomcuProjectT struct {
	ErgomcutoolVersion   *string                      `yaml:"ergomcutool_version"`
	ProjectName          *string                      `yaml:"project_name"`
	DeviceId             *string                      `yaml:"device_id"`
	Openocd              *OpenocdDescriptor           `yaml:"openocd"`
	ExternalDependencies []config.ExternalDependencyT `yaml:"external_dependencies"`
	CSrc                 []string                     `yaml:"c_src"`
	CSrcDirs             []string                     `yaml:"c_src_dirs"`
	CIncludeDirs         []string                     `yaml:"c_include_dirs"`
	JoinedExternalDeps   []string                     `yaml:"-"`
}

func (p *ErgomcuProjectT) String() string {
	data, _ := yaml.Marshal(p)
	return string(data)
}

type OpenocdDescriptor struct {
	Disabled bool    `yaml:"disabled"`
	Target   *string `yaml:"target"`
}

func (g *OpenocdDescriptor) Validate() error {
	if g.Disabled {
		return nil
	}
	warningPrefix := "warning: in 'ergomcu_project.yaml' openocd section:"

	if g.Target == nil {
		return fmt.Errorf("openocd target parameter is missing in project configuration file")
	}

	if config.ToolConfig.Openocd == nil {
		log.Printf("warning: config.ToolConfig.Openocd is nil (configuration not read?)")
		return nil
	}

	targetFile := filepath.Join(*config.ToolConfig.Openocd.ScriptsPath, "target",
		*g.Target)

	exists := utils.FileExists(targetFile)
	if !exists {
		log.Printf("%s openocd:'target' %q doesn't exist.\n",
			warningPrefix, targetFile)
	}
	return nil
}

func ReadAndValidate(path string) (*ErgomcuProjectT, error) {
	r := &ErgomcuProjectT{}
	data, err := os.ReadFile(path)
	if err != nil {
		return r, err
	}

	err = yaml.Unmarshal(data, r)
	if err != nil {
		return r, err
	}

	// Validate
	msgPrefix := "error: ergomcutool project validation failed:"
	msgSuffix := fmt.Sprintf("Fix errors in %q and try again.", path)
	if r.ErgomcutoolVersion == nil || *r.ErgomcutoolVersion == "" {
		log.Fatalf("%s 'ergomcutool_version' is missing.\n%s\n", msgPrefix, msgSuffix)
	}

	if r.ProjectName == nil || *r.ProjectName == "" {
		log.Fatalf("%s 'project_name' is missing.\n%s\n", msgPrefix, msgSuffix)
	}

	if r.DeviceId == nil || *r.DeviceId == "" {
		log.Fatalf("%s 'device_id' is missing.\n%s\n", msgPrefix, msgSuffix)
	}

	if r.Openocd == nil {
		log.Fatalf("%s 'openocd' section is missing.\n%s\n", msgPrefix, msgSuffix)
	}
	if err = r.Openocd.Validate(); err != nil {
		log.Fatalf("%s %v.\n%s\n",
			msgPrefix, err, msgSuffix)
	}

	if r.ExternalDependencies != nil {
		for _, d := range r.ExternalDependencies {
			if err = d.Validate(); err != nil {
				log.Fatalf("%s %v.\n%s\n",
					msgPrefix, err, msgSuffix)
			}
		}
	}

	// Create JoinedExternalDeps:
	// in case of conflict, local versions take precedence.

	return r, nil
}
