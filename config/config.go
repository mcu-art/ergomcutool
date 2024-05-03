package config

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/mcu-art/ergomcutool/assets"
	"github.com/mcu-art/ergomcutool/utils"
	"gopkg.in/yaml.v3"
)

var (
	// ergomcutool version
	Version = "1.0.0"

	// Owner, group: r+w, others: read only
	DefaultFilePermissions uint32 = 0664

	// Special script permissions that allow execution
	DefaultScriptPermissions uint32 = 0775

	// Owner and group: full access, others: can't create new files
	DefaultDirPermissions uint32 = 0775

	toolConfigWarningPrefix = "configuration warning:"

	toolConfigWarningSuffix = "\nIt is recommended to fix the configuration issues before you continue."

	// User-global ergomcutool configuration
	ToolConfig = &ToolConfigT{}

	UserConfigFileName = "ergomcutool_config.yaml"

	UserConfigFilePath = filepath.Join(UserConfigDir, UserConfigFileName)

	// Number of makefile backup files
	MakefileBackupsLimit = 5

	LocalErgomcuDir = "ergomcutool"

	// ProjectFilePath is the path to the project file from project root.
	ProjectFilePath   = filepath.Join(LocalErgomcuDir, "ergomcu_project.yaml")
	ProjectScriptsDir = filepath.Join(LocalErgomcuDir, "scripts")
)

// user and local tool configuration file names
var UserConfigDir = func() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".ergomcutool")
}()

type ToolConfig_GeneralT struct {
	ArmToolchainPath *string `yaml:"arm_toolchain_path"`
	CCompilerPath    *string `yaml:"c_compiler_path"`
	CppCompilerPath  *string `yaml:"cpp_compiler_path"`
	DebuggerPath     *string `yaml:"debugger_path"`
}

func (g *ToolConfig_GeneralT) Validate() error {
	if g.ArmToolchainPath == nil {
		return fmt.Errorf("'arm_toolchain_path' parameter is not defined")
	}
	dirExists := utils.DirExists(*g.ArmToolchainPath)
	if !dirExists {
		log.Printf("%s general:'arm_toolchain_path' must specify an existing directory.%s",
			toolConfigWarningPrefix, toolConfigWarningSuffix)
	}
	if g.CCompilerPath == nil {
		return fmt.Errorf("'c_compiler_path' parameter is not defined")
	}
	if g.CppCompilerPath == nil {
		return fmt.Errorf("'cpp_compiler_path' parameter is not defined")
	}
	if g.DebuggerPath == nil {
		return fmt.Errorf("'debugger_path' parameter is not defined")
	}
	return nil
}

type ToolConfig_OpenOcdT struct {
	Interface         *string `yaml:"interface"`
	BinPath           *string `yaml:"bin_path"`
	ScriptsPath       *string `yaml:"scripts_path"`
	SvdFilePath       string  `yaml:"svd_file_path"`
	DisableSvdWarning bool    `yaml:"disable_svd_warning"`
}

func (g *ToolConfig_OpenOcdT) Validate() error {

	if g.Interface == nil {
		return fmt.Errorf("openocd:'interface' parameter is not defined")
	}
	// TODO: check if the openocd interface has corresponding file
	if *g.Interface == "" {
		return fmt.Errorf("openocd:'interface' parameter is not defined")
	}

	if g.BinPath == nil {
		return fmt.Errorf("openocd:'bin_path' parameter is not defined")
	}
	exists := utils.FileExists(*g.BinPath)
	if !exists {
		log.Printf("%s openocd:'bin_path' must specify an existing file.%s\n", toolConfigWarningPrefix, toolConfigWarningSuffix)
	}

	if g.ScriptsPath == nil {
		return fmt.Errorf("'scripts_path' parameter is not defined")
	}
	exists = utils.DirExists(*g.ScriptsPath)
	if !exists {
		log.Printf("%s openocd:'scripts_path' must specify an existing directory.%s\n",
			toolConfigWarningPrefix, toolConfigWarningSuffix)
	}

	return nil
}

type ExternalDependencyT struct {
	Var                 string `yaml:"var"`
	Path                string `yaml:"path"`
	CreateInProjectLink bool   `yaml:"create_in_project_link"`
	LinkName            string `yaml:"link_name"`
}

// MergeSpecial merges two external dependencies.
// 'projectSetting' is supposed to be from the project configuration,
// and 'configSetting' - from the tool configuration.
// 'Var' fields are supposed to be equal and are not modified.
// 'projectSetting.CreateInProjectLink' is only modified if
// 'configSetting.CreateInProjectLink==true'
// Precedence has: 'configSetting.Path',
// 'projectSetting.LinkName' is only modified if
// 'configSetting.LinkName' is not empty.
func (projectSetting *ExternalDependencyT) MergeSpecial(
	configSetting *ExternalDependencyT) {
	if projectSetting.Var != configSetting.Var {
		return
	}
	if configSetting.CreateInProjectLink {
		projectSetting.CreateInProjectLink = true
	}
	if configSetting.Path != "" {
		projectSetting.Path = configSetting.Path
	}
	if configSetting.LinkName != "" {
		projectSetting.LinkName = configSetting.LinkName
	}
}

// Validate validates the external dependency.
func (g *ExternalDependencyT) Validate() error {

	if g.Var == "" {
		return fmt.Errorf("external_dependencies:'var' parameter is not defined")
	}

	if g.Path == "" {
		return fmt.Errorf("external_dependencies:'path' parameter is not defined for %q",
			g.Var)
	}

	if g.CreateInProjectLink {
		if g.LinkName == "" {
			return fmt.Errorf("external_dependencies:'link_name' parameter must be defined if 'create_in_project_link' is true for variable %q",
				g.Var)
		}
		if g.LinkName == "" {
			return fmt.Errorf("external_dependencies:'link_name' parameter must not be empty if 'create_in_project_link' is true for variable %q",
				g.Var)
		}
	}

	exists := utils.DirExists(g.Path)
	if !exists {
		log.Printf("%s external_dependencies:'path' %q must specify an existing directory.%s\n",
			toolConfigWarningPrefix, g.Path, toolConfigWarningSuffix)
	}
	return nil
}

type BuildOptionsT struct {
	BuildDir          *string `yaml:"build_dir"`
	Debug             *string `yaml:"debug"`
	OptimizationFlags *string `yaml:"optimization_flags"`
}

// Validate validates the build options.
// Currently, nil values are considered legitimate.
func (o *BuildOptionsT) Validate() error {
	if o.Debug != nil {
		if !(*o.Debug == "0" || *o.Debug == "1") {
			return fmt.Errorf("build_options:'debug' must have value '0' or '1'")
		}
	}
	return nil
}

type ToolConfigT struct {
	General              *ToolConfig_GeneralT  `yaml:"general"`
	Openocd              *ToolConfig_OpenOcdT  `yaml:"openocd"`
	ExternalDependencies []ExternalDependencyT `yaml:"external_dependencies"`
	BuildOptions         *BuildOptionsT        `yaml:"build_options"`
}

func (g *ToolConfigT) String() string {
	data, _ := yaml.Marshal(g)
	return string(data)
}

// readConfigFile reads user or local configuration file into **config**
func readConfigFile(file string, config *ToolConfigT) error {
	data, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(data, config)
	if err != nil {
		return err
	}

	return err
}

// copyTextFileEx copies text file from src to dest, prepends optional prefix
// to each line.
func copyTextFileEx(src, dest string, prefix string, filePerm uint32) error {
	data, err := os.ReadFile(src)
	if err != nil {
		log.Fatalf("failed to read file %q: %v\n", src, err)
	}
	if prefix != "" {
		var sb strings.Builder
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			if len(line) > 0 {
				sb.WriteString(prefix)
				sb.WriteString(line)
			}
			sb.WriteString("\n")
		}
		return os.WriteFile(dest, []byte(sb.String()), fs.FileMode(filePerm))
	}
	return os.WriteFile(dest, data, fs.FileMode(filePerm))
}

// copyAssetsIntoUserConfigDir creates user config dir if not exists and
// copies all necessary assets into it.
// It doesn't return error but exits instead.
func copyAssetsIntoUserConfigDir() {
	if err := os.MkdirAll(UserConfigDir,
		fs.FileMode(DefaultDirPermissions)); err != nil {
		log.Fatalf("failed to create user config directory %q: %v\n",
			UserConfigDir, err)
	}
	err := assets.CopyAssets(UserConfigDir,
		DefaultDirPermissions, DefaultFilePermissions)

	if err != nil {
		log.Fatalf("failed to copy embedded assets into user config directory %q: %v\n",
			UserConfigDir, err)
	}

	// Move user config file from assets dir to user config dir
	src := filepath.Join(UserConfigDir, "assets", UserConfigFileName)
	dest := filepath.Join(UserConfigDir, UserConfigFileName)
	_ = os.Rename(src, dest)
}

func CheckUserConfigDirExists() bool {
	return utils.DirExists(UserConfigDir)
}

// CreateUserConfig creates user config directory and its contents.
// If the directory already exists, it returns error.
func CreateUserConfig() error {
	if utils.DirExists(UserConfigDir) {
		return fmt.Errorf("user configuration directory already exists")
	}
	// Copy assets from the embedded FS
	copyAssetsIntoUserConfigDir()
	return nil
}

// EnsureUserConfigExists checks that ergomcutool user config directory
// exists. If not, it prints error message and exists with an error.
func EnsureUserConfigExists() {
	if !utils.DirExists(UserConfigDir) {
		log.Fatalf(`error: ergomcutool is not initialized yet.
Run 'ergomcutool init' first.
`)
	}
}

// ParseErgomcutoolConfig parses ergomcutool configuration,
// both user and local configurations are taken into account.
// 'createLocalConfigIfNotExists': if true, creates a local configuration
// file in CWD that is a commented-out copy of the current user configuration.
func ParseErgomcutoolConfig(createLocalConfigIfNotExists bool) {
	// Read user config file
	userConfigFilePath := filepath.Join(UserConfigDir, UserConfigFileName)
	err := readConfigFile(userConfigFilePath, ToolConfig)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			log.Fatalf("error: ergomcutool configuration file doesn't exist, please run 'ergomcutool init' first.\n")
		} else {
			log.Fatalf("error: failed to read user configuration file: %+v\n", err)
		}
	}

	localConfigFilePath := filepath.Join("_non_persistent", UserConfigFileName)
	// Override with values taken from local config file
	err = readConfigFile(localConfigFilePath, ToolConfig)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			if createLocalConfigIfNotExists {
				// Write a default configuration file
				// Check if path exists
				dirPath := filepath.Dir(localConfigFilePath)
				if err = os.MkdirAll(dirPath, fs.FileMode(DefaultDirPermissions)); err != nil {
					log.Fatalf("Failed to create local directory %q: %v\n",
						dirPath, err)
				}
				if err = copyTextFileEx(
					userConfigFilePath,
					localConfigFilePath, "", // "# "
					DefaultFilePermissions); err != nil {
					log.Fatalf("Failed to write local configuration to file %q: %v\n",
						localConfigFilePath, err)
				}

				// Re-read file
				if err = readConfigFile(localConfigFilePath, ToolConfig); err != nil {
					log.Fatalf("Failed to read local configuration file: %+v\n", err)
				}
			}
		} else {
			log.Fatalf("Failed to read local configuration file: %+v\n", err)
		}
	}

	// Override with values taken from environment variables
	// Currently, no variables are used

	// Override with values taken from CLI
	// Currently, no CLI variables are used

	// Validate
	msgPrefix := "error: ergomcutool configuration validation failed:"
	msgSuffix := "Fix the configuration errors and try again."
	if ToolConfig.General == nil {
		log.Fatalf("%s 'general' section is missing.\n%s\n", msgPrefix, msgSuffix)
	}
	if err = ToolConfig.General.Validate(); err != nil {
		log.Fatalf("%s %v.\n%s\n",
			msgPrefix, err, msgSuffix)
	}

	if ToolConfig.Openocd == nil {
		log.Fatalf("%s 'openocd' section is missing.\n%s\n", msgPrefix, msgSuffix)
	}
	if err = ToolConfig.Openocd.Validate(); err != nil {
		log.Fatalf("%s %v.\n%s\n",
			msgPrefix, err, msgSuffix)
	}

	if ToolConfig.BuildOptions != nil {
		err = ToolConfig.BuildOptions.Validate()
		if err != nil {
			log.Fatalf("%s 'build_options' validation failed: %v.\n%s\n", msgPrefix, err, msgSuffix)
		}
	}

	// Do not validate external dependencies here,
	// they should be validated in the update-project cmd
}
