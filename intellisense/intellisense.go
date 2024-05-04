package intellisense

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/mcu-art/ergomcutool/config"
	"github.com/mcu-art/ergomcutool/utils"
)

// CCppPropertiesReplacements are JSON entries for 'c_cpp_properties.json'
// that are generated automatically based on the project configuration.
type CCppPropertiesReplacements struct {
	IncludePath  []string `json:"includePath"`
	Defines      []string `json:"defines"`
	CompilerPath string   `json:"compilerPath"`
}

// LaunchReplacements are JSON entries for 'launch.json'
// that are generated automatically based on the project configuration.
type LaunchReplacements struct {
	Executable  string   `json:"executable"`
	ConfigFiles []string `json:"configFiles"`
	SvdFile     string   `json:"svdFile"`
}

// SettingsReplacements are JSON entries for 'settings.json'
// that are generated automatically based on the project configuration.
type SettingsReplacements struct {
	IncludePaths                []string `json:"C_Cpp_Runner.includePaths"`
	CCompilerPath               string   `json:"C_Cpp_Runner.cCompilerPath"`
	CppCompilerPath             string   `json:"C_Cpp_Runner.cppCompilerPath"`
	DebuggerPath                string   `json:"C_Cpp_Runner.debuggerPath"`
	CortexDebugArmToolchainPath string   `json:"cortex-debug.armToolchainPath"`
	CortexDebugOpenocdPath      string   `json:"cortex-debug.openocdPath"`
	CortexDebugGdbPath          string   `json:"cortex-debug.gdbPath"`
}

// ConfigurationEntry is an auxiliary type that stores
// Configuration taken from JSON file as a map,
// and its index in the 'configurations' array.
type ConfigurationEntry struct {
	// C is a configuration entry
	C     map[string]any
	Index int
}

func readJsonFileToMap(filePath string) (map[string]any, error) {
	r := make(map[string]any)
	if utils.FileExists(filePath) {
		currentFileData, err := os.ReadFile(filePath)
		if err != nil {
			return r, fmt.Errorf("failed to read %q: %w", filePath, err)
		}
		err = json.Unmarshal(currentFileData, &r)
		if err != nil {
			return r, fmt.Errorf("failed to unmarshal json %q: %w", filePath, err)
		}
	}
	return r, nil
}

// findConfigurations takes c_cpp_properties.json as an input
// and returns a slice of configurations that have names
// starting with specified prefix along with their index.
func findConfigurations(
	root map[string]any, prefix string) ([]ConfigurationEntry, error) {
	r := make([]ConfigurationEntry, 0, 10)

	configurations, ok := root["configurations"].([]any)
	if !ok {
		return r, fmt.Errorf("'configurations' must be an array")
	}

	for i, configuration := range configurations {
		confmap, ok := configuration.(map[string]any)
		if !ok {
			return r, fmt.Errorf("each 'configuration' must be a map")
		}
		name, ok := confmap["name"].(string)
		if !ok {
			return r, fmt.Errorf("'configuration.name' must be a string")
		}
		if strings.HasPrefix(name, prefix) {
			r = append(r, ConfigurationEntry{
				C:     confmap,
				Index: i,
			})
		}
	}
	return r, nil
}

// ProcessCCppPropertiesJson processes 'c_cpp_properties.json'
// so that it contains values required for VSCode intellisense.
// If the file doesn't exist, it will be created
// from the 'c_cpp_properties.persistent.json'.
func ProcessCCppPropertiesJson(r CCppPropertiesReplacements) error {
	var err error
	// Read c_cpp_properties.json if exists
	currentFile := filepath.Join(".vscode", "c_cpp_properties.json")
	persistentFile := filepath.Join(".vscode", "c_cpp_properties.persistent.json")
	currentFileExists := utils.FileExists(currentFile)

	var currentFileMap map[string]any
	var persistentFileMap map[string]any

	// If c_cpp_properties.json doesn't exist, create it as a copy
	// of c_cpp_properties.persistent.json
	if !currentFileExists {
		err = utils.CopyFile(persistentFile, currentFile)
		if err != nil {
			return fmt.Errorf("failed to copy %q to %q: %w", persistentFile, currentFile, err)
		}
	}

	currentFileMap, err = readJsonFileToMap(currentFile)
	if err != nil {
		return fmt.Errorf("failed to read %q: %w", currentFile, err)
	}

	// Read c_cpp_properties.persistent.json if exists
	if !utils.FileExists(persistentFile) {
		return fmt.Errorf("%q does not exist", persistentFile)
	}

	persistentFileMap, err = readJsonFileToMap(persistentFile)
	if err != nil {
		return fmt.Errorf("failed to read %q: %w", persistentFile, err)
	}

	prefix := "linux-gcc-arm"
	destConfigs, err := findConfigurations(currentFileMap, prefix)
	if err != nil {
		return fmt.Errorf("failed to read %q: %w", currentFile, err)
	}

	srcConfigs, err := findConfigurations(persistentFileMap, prefix)
	if err != nil {
		return fmt.Errorf("failed to read %q: %w", persistentFile, err)
	}
	if len(srcConfigs) < 1 {
		return fmt.Errorf(
			"file %q doesn't contain any configuration that has name starting with %q", persistentFile, prefix)
	}

	if len(destConfigs) < 1 {
		return fmt.Errorf(
			"file %q doesn't contain any configuration that has name starting with %q", currentFile, prefix)
	}

	// Overwrite all top-level values in dest with the values from src[0]
	for _, dest := range destConfigs {
		for k, v := range srcConfigs[0].C {
			dest.C[k] = v
		}
		// Overwrite auto-generated values with the values from the replacements
		dest.C["includePath"] = r.IncludePath
		dest.C["defines"] = r.Defines
		dest.C["compilerPath"] = r.CompilerPath

		// Save changes in the currentFileMap
		configs := currentFileMap["configurations"].([]any)
		configs[dest.Index] = dest.C
	}

	data, err := json.MarshalIndent(currentFileMap, "", "  ")
	if err != nil {
		return fmt.Errorf("ProcessCCppPropertiesJson: failed to marshal json: %w", err)
	}

	// Save the updated file
	return os.WriteFile(currentFile, data, fs.FileMode(config.DefaultFilePermissions))
}

// ProcessLaunchJson processes 'launch.json'
// so that it contains values required for VSCode intellisense.
// If the file doesn't exist, it will be created
// from 'launch.persistent.json'.
func ProcessLaunchJson(r LaunchReplacements) error {
	var err error
	// Read c_cpp_properties.json if exists
	currentFile := filepath.Join(".vscode", "launch.json")
	persistentFile := filepath.Join(".vscode", "launch.persistent.json")
	currentFileExists := utils.FileExists(currentFile)

	var currentFileMap map[string]any
	var persistentFileMap map[string]any

	// If launch.json doesn't exist, create it as a copy
	// of launch.persistent.json
	if !currentFileExists {
		err = utils.CopyFile(persistentFile, currentFile)
		if err != nil {
			return fmt.Errorf("failed to copy %q to %q: %w", persistentFile, currentFile, err)
		}
	}

	currentFileMap, err = readJsonFileToMap(currentFile)
	if err != nil {
		return fmt.Errorf("failed to read %q: %w", currentFile, err)
	}

	// Read launch.persistent.json if exists
	if !utils.FileExists(persistentFile) {
		return fmt.Errorf("%q does not exist", persistentFile)
	}

	persistentFileMap, err = readJsonFileToMap(persistentFile)
	if err != nil {
		return fmt.Errorf("failed to read %q: %w", persistentFile, err)
	}

	prefix := "STM32_Debug"
	destConfigs, err := findConfigurations(currentFileMap, prefix)
	if err != nil {
		return fmt.Errorf("failed to read %q: %w", currentFile, err)
	}

	srcConfigs, err := findConfigurations(persistentFileMap, prefix)
	if err != nil {
		return fmt.Errorf("failed to read %q: %w", persistentFile, err)
	}
	if len(srcConfigs) < 1 {
		return fmt.Errorf(
			"file %q doesn't contain any configuration that has name starting with %q", persistentFile, prefix)
	}

	if len(destConfigs) < 1 {
		return fmt.Errorf(
			"file %q doesn't contain any configuration that has name starting with %q", currentFile, prefix)
	}

	// Overwrite all top-level values in dest with the values from src[0]
	for _, dest := range destConfigs {
		for k, v := range srcConfigs[0].C {
			dest.C[k] = v
		}
		// Overwrite auto-generated values with the values from the replacements
		dest.C["configFiles"] = r.ConfigFiles
		dest.C["executable"] = r.Executable
		dest.C["svdFile"] = r.SvdFile

		// Save changes in the currentFileMap
		configs := currentFileMap["configurations"].([]any)
		configs[dest.Index] = dest.C
	}

	data, err := json.MarshalIndent(currentFileMap, "", "  ")
	if err != nil {
		return fmt.Errorf("ProcessLaunchJson: failed to marshal json: %w", err)
	}

	// Save the updated file
	return os.WriteFile(currentFile, data, fs.FileMode(config.DefaultFilePermissions))
}

// ProcessSettingsJson processes 'settings.json'
// so that it contains values required for VSCode intellisense.
// If the file doesn't exist, it will be created
// from 'settings.persistent.json'.
func ProcessSettingsJson(r SettingsReplacements) error {
	var err error
	// Read c_cpp_properties.json if exists
	currentFile := filepath.Join(".vscode", "settings.json")
	persistentFile := filepath.Join(".vscode", "settings.persistent.json")
	currentFileExists := utils.FileExists(currentFile)

	var currentFileMap map[string]any
	var persistentFileMap map[string]any

	// If settings.json doesn't exist, create it as a copy
	// of settings.persistent.json
	if !currentFileExists {
		err = utils.CopyFile(persistentFile, currentFile)
		if err != nil {
			return fmt.Errorf("failed to copy %q to %q: %w", persistentFile, currentFile, err)
		}
	}

	currentFileMap, err = readJsonFileToMap(currentFile)
	if err != nil {
		return fmt.Errorf("failed to read %q: %w", currentFile, err)
	}

	// Read settings.persistent.json if exists
	if !utils.FileExists(persistentFile) {
		return fmt.Errorf("%q does not exist", persistentFile)
	}

	persistentFileMap, err = readJsonFileToMap(persistentFile)
	if err != nil {
		return fmt.Errorf("failed to read %q: %w", persistentFile, err)
	}

	// Overwrite all top-level values in dest with the values from src
	dest := currentFileMap
	src := persistentFileMap
	for k, v := range src {
		dest[k] = v
	}
	// Overwrite auto-generated values with the values from the replacements
	dest["C_Cpp_Runner.cCompilerPath"] = r.CCompilerPath
	dest["C_Cpp_Runner.cppCompilerPath"] = r.CppCompilerPath
	dest["C_Cpp_Runner.debuggerPath"] = r.DebuggerPath
	dest["C_Cpp_Runner.includePaths"] = r.IncludePaths
	dest["cortex-debug.armToolchainPath"] = r.CortexDebugArmToolchainPath
	dest["cortex-debug.openocdPath"] = r.CortexDebugOpenocdPath
	dest["cortex-debug.gdbPath"] = r.CortexDebugGdbPath

	data, err := json.MarshalIndent(currentFileMap, "", "  ")
	if err != nil {
		return fmt.Errorf("ProcessSettingsJson: failed to marshal json: %w", err)
	}

	// Save the updated file
	return os.WriteFile(currentFile, data, fs.FileMode(config.DefaultFilePermissions))
}

// ProcessTasksJson processes 'tasks.json'
// so that it contains values required for VSCode intellisense.
// If the file doesn't exist, it will be created
// from 'tasks.persistent.json'.
func ProcessTasksJson() error {
	var err error
	// Read c_cpp_properties.json if exists
	currentFile := filepath.Join(".vscode", "tasks.json")
	persistentFile := filepath.Join(".vscode", "tasks.persistent.json")
	currentFileExists := utils.FileExists(currentFile)

	// If tasks.json doesn't exist, create it as a copy
	// of tasks.persistent.json
	if !currentFileExists {
		err = utils.CopyFile(persistentFile, currentFile)
		if err != nil {
			return fmt.Errorf("failed to copy %q to %q: %w", persistentFile, currentFile, err)
		}
	}

	// No additional processing required for tasks.json
	return nil
}
