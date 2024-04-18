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

var Version = "1.0.0"

// Owner, group: r+w, others: read only
var DefaultFilePermissions uint32 = 0664

// Owner and group: full access, others: can't create new files
var DefaultDirPermissions uint32 = 0775

var toolConfigWarningPrefix = "configuration warning:"
var toolConfigWarningSuffix = "\nIt is recommended to fix the ergomcutool configuration."

var ToolConfig = &ToolConfigT{}

// user and local tool configuration file names
var UserConfigDir = func() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".ergomcutool")
}()

var UserConfigFileName = "ergomcutool_config.yaml"
var UserConfigFilePath = filepath.Join(UserConfigDir, UserConfigFileName)

// var UserLocalConfigFilePath = filepath.Join("_non_persistent", "ergomcutool_config.yaml")

type ToolConfig_GeneralT struct {
	ArmToolchainPath *string `yaml:"arm_toolchain_path"`
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
	return nil
}

type ToolConfig_OpenOcdT struct {
	Interface   *string `yaml:"interface"`
	BinPath     *string `yaml:"bin_path"`
	ScriptsPath *string `yaml:"scripts_path"`
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

type ToolConfig_LibraryPathT struct {
	Var  *string
	Path *string
}

func (g *ToolConfig_LibraryPathT) Validate() error {

	if g.Var == nil {
		return fmt.Errorf("library_paths:'var' parameter is not defined")
	}

	if g.Path == nil {
		return fmt.Errorf("library_paths:'path' parameter is not defined for %q",
			*g.Var)
	}

	exists := utils.DirExists(*g.Path)
	if !exists {
		log.Printf("%s library_paths:'path' %q must specify an existing directory.%s\n",
			toolConfigWarningPrefix, *g.Path, toolConfigWarningSuffix)
	}
	return nil
}

type ToolConfigT struct {
	General      *ToolConfig_GeneralT      `yaml:"general"`
	Openocd      *ToolConfig_OpenOcdT      `yaml:"openocd"`
	LibraryPaths []ToolConfig_LibraryPathT `yaml:"library_paths"`
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
// go each line.
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
	err = utils.CopyFile(src, dest)
	if err != nil {
		log.Fatalf("failed to copy %q into %q: %v\n",
			src, dest, err)
	}
	_ = os.Remove(src)
}

func ParseErgomcutoolConfig() {
	// homeDir, err := os.UserHomeDir()
	// if err != nil {
	// 	log.Fatalf("failed to retrieve user home directory: %v\n", err)
	// }

	// Read user config file
	userConfigFilePath := filepath.Join(UserConfigDir, UserConfigFileName)
	err := readConfigFile(userConfigFilePath, ToolConfig)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// Copy assets from the embedded FS
			copyAssetsIntoUserConfigDir()
			// Re-read file
			if err = readConfigFile(userConfigFilePath, ToolConfig); err != nil {
				log.Fatalf("failed to read user configuration file: %+v\n", err)
			}
			// Print message about first run and exit
			firstRunMsg := fmt.Sprintf(`IMPORTANT!
You've just run ergomcutool for the first time on your machine.
Your command hasn't been executed,
but new user settings have been generated.
In case you want to review them,
check %q.
Now you may proceed and re-run your command.`, userConfigFilePath)
			log.Println(firstRunMsg)
			os.Exit(0)
		} else {
			log.Fatalf("Failed to read user configuration file: %+v\n", err)
		}
	}

	localConfigFilePath := filepath.Join("_non_persistent", UserConfigFileName)
	// Override with values taken from local config file
	err = readConfigFile(localConfigFilePath, ToolConfig)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// Write a default configuration file
			// Check if path exists
			dirPath := filepath.Dir(localConfigFilePath)
			if err = os.MkdirAll(dirPath, fs.FileMode(DefaultDirPermissions)); err != nil {
				log.Fatalf("Failed to create local directory %q: %v\n",
					dirPath, err)
			}
			if err = copyTextFileEx(
				userConfigFilePath,
				localConfigFilePath, "# ",
				DefaultFilePermissions); err != nil {
				log.Fatalf("Failed to write local configuration to file %q: %v\n",
					localConfigFilePath, err)
			}

			// Re-read file
			if err = readConfigFile(localConfigFilePath, ToolConfig); err != nil {
				log.Fatalf("Failed to read local configuration file: %+v\n", err)
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
	msgPostfix := "Fix the configuration errors and try again."
	if ToolConfig.General == nil {
		log.Fatalf("%s 'general' section is missing.\n%s\n", msgPrefix, msgPostfix)
	}
	if err = ToolConfig.General.Validate(); err != nil {
		log.Fatalf("%s %v.\n%s\n",
			msgPrefix, err, msgPostfix)
	}

	if ToolConfig.Openocd == nil {
		log.Fatalf("%s 'openocd' section is missing.\n%s\n", msgPrefix, msgPostfix)
	}
	if err = ToolConfig.Openocd.Validate(); err != nil {
		log.Fatalf("%s %v.\n%s\n",
			msgPrefix, err, msgPostfix)
	}

	if ToolConfig.LibraryPaths != nil {
		// log.Fatalf("%s 'openocd' section is missing.\n%s\n", msgPrefix, msgPostfix)
		for _, lp := range ToolConfig.LibraryPaths {
			if err = lp.Validate(); err != nil {
				log.Fatalf("%s %v.\n%s\n",
					msgPrefix, err, msgPostfix)
			}
		}
	}
}
