package cli

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/mcu-art/ergomcutool/config"
	"github.com/mcu-art/ergomcutool/intellisense"
	"github.com/mcu-art/ergomcutool/mkf"
	"github.com/mcu-art/ergomcutool/proj"
	"github.com/mcu-art/ergomcutool/tpl"
	"github.com/mcu-art/ergomcutool/utils"
	"github.com/spf13/cobra"
)

var updateProjectCmd = &cobra.Command{
	Use:   "update-project",
	Short: "Update project and patch makefile",
	Run:   updateProject,
}

var (
	up_Makefile string
)

func init() {
	rootCmd.AddCommand(updateProjectCmd)
	updateProjectCmd.PersistentFlags().StringVarP(
		&up_Makefile, "makefile", "m", "", "Specify custom path to Makefile")
}

func updateProject(cmd *cobra.Command, args []string) {
	if len(args) > 0 {
		log.Fatalf("error: too many CLI argument(s): %+v\n", args)
	}
	cwd, _ := os.Getwd()
	config.ParseErgomcutoolConfig(false)

	// Read ergomcu_project.yaml
	pc, err := proj.ReadAndValidate(config.ProjectFilePath)
	if err != nil {
		log.Fatalf("error: failed to read project file %q: %v\n",
			config.ProjectFilePath, err)
	}

	log.Printf("Updating project %q...\n", *pc.ProjectName)
	if verbose {
		fmt.Println("* Using the following project configuration:")
		fmt.Println(pc.String())
	}

	// Check if makefile exists.
	if up_Makefile == "" {
		up_Makefile = filepath.Join(cwd, "Makefile")
	}
	if !utils.FileExists(up_Makefile) {
		log.Fatalf(`error: makefile doesn't exist: %q.
Generate the Makefile first using STM32CubeMX.
`, up_Makefile)
	}

	// Create the _external directory and links if needed
	if len(pc.ExternalDependencies) > 0 {
		createExternalDir := false
		for _, d := range pc.ExternalDependencies {
			if d.CreateInProjectLink {
				createExternalDir = true
				break
			}
		}
		externalDir := filepath.Join(cwd, "_external")

		if createExternalDir {
			if !utils.DirExists(externalDir) {
				err := os.MkdirAll(externalDir, fs.FileMode(config.DefaultDirPermissions))
				if err != nil {
					log.Fatalf("error: failed to create directory %q: %v\n", externalDir, err)
				}
			}
		}

		// Update required symlinks
		log.Printf("Updating symlinks to external dependencies...\n")
		for _, dep := range pc.ExternalDependencies {
			if !dep.CreateInProjectLink {
				continue
			}
			newLink := filepath.Join(externalDir, dep.LinkName)
			err := utils.CreateOrReplaceSymlink(newLink, dep.Path)
			if err != nil {
				log.Fatalf("error: failed to create or replace symlink %q to %q: %v\n",
					newLink, dep.Path, err)
			}
		}
		log.Printf("Done.\n")
	}

	// Read the Makefile
	makefile, err := mkf.FromFile(up_Makefile)
	if err != nil {
		log.Fatalf("error: failed to read and parse the makefile %q: %v\n",
			up_Makefile, err)
	}
	preEditedMakefilePath := filepath.Join(cwd, "_non_persistent", "Makefile.pre-edit")
	moveMakefileToPreEdited := false
	if makefile.IsAutoEdited() {
		// Read the original version
		makefile, err = mkf.FromFile(preEditedMakefilePath)
		if err != nil {
			log.Fatalf("error: failed to read and parse %q: %v\n",
				preEditedMakefilePath, err)
		}
	} else { // Makefile is not edited, move it later to _non_persistent/Makefile.pre-edit
		moveMakefileToPreEdited = true
	}

	if verbose {
		log.Printf("* original makefile contains %d lines.\n", len(makefile.Lines))
	}

	// Create external dependencies expansion map
	externalDepExpansionMap := make(map[string]string, 0)
	for _, d := range pc.ExternalDependencies {
		externalDepExpansionMap[d.Var] = d.Path
	}

	// Merge values from the Makefile with project values
	// c_src
	c_src, err := makefile.ReadValue("C_SOURCES")
	if err != nil {
		log.Fatalf("error: failed to read C_SOURCES from the makefile: %v\n", err)
	}
	c_src = append(c_src, pc.CSrc...)

	// Resolve CSrcDirs
	for _, d := range pc.CSrcDirs {
		l, err := utils.GetSortedFileList(d, ".c")
		if err != nil {
			log.Fatalf("error: failed to read source directory %q: %v\n", d, err)
		}
		for _, filename := range l {
			fullName := filepath.Join(d, filename)
			c_src = append(c_src, fullName)
		}
	}
	// Expand external dependencies in each line
	c_src, err = expandExternalDependencies(c_src, externalDepExpansionMap)
	if err != nil {
		log.Fatalf(`error: failed to expand external dependencies in C_SOURCES: %v
Check your project configuration.`, err)
	}
	err = makefile.ReplaceValue("C_SOURCES", c_src)
	if err != nil {
		log.Fatalf("error: failed to replace C_SOURCES in the makefile: %v\n", err)
	}

	// c_includes
	c_includes, err := makefile.ReadValue("C_INCLUDES")
	if err != nil {
		log.Fatalf("error: failed to read C_INCLUDES from the makefile: %v\n", err)
	}
	// Remove -I prefix for each line
	for i, includeFile := range c_includes {
		c_includes[i] = strings.TrimLeft(includeFile, "-I")
	}
	c_includes = append(c_includes, pc.CIncludeDirs...)
	// Expand external dependencies in each line
	c_includes, err = expandExternalDependencies(c_includes, externalDepExpansionMap)
	if err != nil {
		log.Fatalf("error: failed to expand external dependencies in C_INCLUDES: %v\n", err)
	}
	// Add -I prefix to each line
	prefixedCIncludes := make([]string, 0, len(c_includes))
	for _, includeFile := range c_includes {
		prefixedCIncludes = append(prefixedCIncludes, "-I"+includeFile)
	}
	err = makefile.ReplaceValue("C_INCLUDES", prefixedCIncludes)
	if err != nil {
		log.Fatalf("error: failed to replace C_INCLUDES in the makefile: %v\n", err)
	}

	// C_DEFS
	c_defs, err := makefile.ReadValue("C_DEFS")
	if err != nil {
		log.Fatalf("error: failed to read C_DEFS from the makefile: %v\n", err)
	}
	// Remove -D prefix for each line
	for i, defFile := range c_defs {
		c_defs[i] = strings.TrimLeft(defFile, "-D")
	}
	c_defs = append(c_defs, pc.CDefs...)

	// Expand external dependencies in each line
	c_defs, err = expandExternalDependencies(c_defs, externalDepExpansionMap)
	if err != nil {
		log.Fatalf("error: failed to expand external dependencies in C_DEFS: %v\n", err)
	}

	// Add -D prefix to each line
	prefixedCDefs := make([]string, 0, len(c_defs))
	for _, defFile := range c_defs {
		prefixedCDefs = append(prefixedCDefs, "-D"+defFile)
	}
	err = makefile.ReplaceValue("C_DEFS", prefixedCDefs)
	if err != nil {
		log.Fatalf("error: failed to replace C_DEFS in the makefile: %v\n", err)
	}

	// Instantiate and append the 'prog' target
	progSnippetUserDir := filepath.Join(config.UserConfigDir, "assets", "snippets")
	progSnippetLocalDir := filepath.Join(cwd, config.LocalErgomcuDir, "assets", "snippets")
	progSnippetFileName := "prog_task.txt.tmpl"
	instantiatedProgSnippet := ""
	replacements := map[string]string{
		"OpenocdInterface": *config.ToolConfig.Openocd.Interface,
		"OpenocdTarget":    *pc.Openocd.Target,
	}
	// Check if local snippet exists
	if utils.FileExists(filepath.Join(progSnippetLocalDir, progSnippetFileName)) {
		instantiatedProgSnippet, err = tpl.InstantiateToString(
			progSnippetLocalDir, progSnippetFileName, replacements)
	} else {
		instantiatedProgSnippet, err = tpl.InstantiateToString(
			progSnippetUserDir, progSnippetFileName, replacements)
	}
	if err != nil {
		log.Fatalf("error: failed to instantiate snippet template %q: %v\n", up_Makefile, err)
	}
	if instantiatedProgSnippet != "" {
		_ = makefile.AppendString(instantiatedProgSnippet, false)
	}

	// Update build options
	buildOptions := config.ToolConfig.BuildOptions
	if buildOptions != nil {
		if buildOptions.BuildDir != nil && *buildOptions.BuildDir != "" {
			values := []string{*buildOptions.BuildDir}
			_ = makefile.ReplaceValue("BUILD_DIR", values)
		}
		if buildOptions.Debug != nil {
			values := []string{*buildOptions.Debug}
			_ = makefile.ReplaceValue("DEBUG", values)
		}
		if buildOptions.OptimizationFlags != nil && *buildOptions.OptimizationFlags != "" {
			values := []string{*buildOptions.OptimizationFlags}
			_ = makefile.ReplaceValue("OPT", values)
		}
	}

	// Insert auto-edited mark
	_ = makefile.InsertAutoEditedMark()

	if moveMakefileToPreEdited {
		if utils.FileExists(preEditedMakefilePath) {
			err = os.Remove(preEditedMakefilePath)
			if err != nil {
				log.Fatalf("error: failed to remove old %q: %v\n",
					preEditedMakefilePath, err)
			}
		}

		err = os.Rename(up_Makefile, preEditedMakefilePath)
		if err != nil {
			log.Fatalf("error: failed to move %q to %q: %v\n",
				up_Makefile, preEditedMakefilePath, err)
		}
	}

	if verbose {
		log.Printf("* updated makefile contains %d lines.\n", len(makefile.Lines))
	}

	err = os.WriteFile(up_Makefile, makefile.Bytes(), fs.FileMode(config.DefaultFilePermissions))
	if err != nil {
		log.Fatalf("error: failed to write the updated %q: %v\n", up_Makefile, err)
	}

	// Update intellisense
	// c_cpp_properties.json
	ccppPropertiesReplacements := intellisense.CCppPropertiesReplacements{
		IncludePath:  c_includes,
		Defines:      c_defs,
		CompilerPath: *config.ToolConfig.General.CCompilerPath,
	}
	err = intellisense.ProcessCCppPropertiesJson(ccppPropertiesReplacements)
	if err != nil {
		log.Printf(`warning: failed to update .vscode/c_cpp_properties.json: %v.
The intellisense may not work properly.`, err)
	}

	// launch.json
	buildDir, _ := makefile.ReadValue("BUILD_DIR")
	launchExecutable := filepath.Join(buildDir[0], *pc.ProjectName+".elf")
	launchReplacements := intellisense.LaunchReplacements{
		Executable: launchExecutable,
		ConfigFiles: []string{*config.ToolConfig.Openocd.Interface,
			*pc.Openocd.Target},
		SvdFile: *config.ToolConfig.Openocd.SvdFilePath,
	}
	err = intellisense.ProcessLaunchJson(launchReplacements)
	if err != nil {
		log.Printf(`warning: failed to update .vscode/launch.json: %v.
The intellisense may not work properly.`, err)
	}

	// settings.json
	settingsReplacements := intellisense.SettingsReplacements{
		IncludePaths:    c_includes,
		CCompilerPath:   *config.ToolConfig.General.CCompilerPath,
		CppCompilerPath: *config.ToolConfig.General.CppCompilerPath,
		DebuggerPath:    *config.ToolConfig.General.DebuggerPath,
	}
	err = intellisense.ProcessSettingsJson(settingsReplacements)
	if err != nil {
		log.Printf(`warning: failed to update .vscode/settings.json: %v.
The intellisense may not work properly.`, err)
	}

	// tasks.json
	err = intellisense.ProcessTasksJson()
	if err != nil {
		log.Printf(`warning: failed to update .vscode/tasks.json: %v.
The intellisense may not work properly.`, err)
	}

	log.Printf("The project was successfully updated by ergomcutool.")
}

func expandExternalDependencies(s []string, replacements any) ([]string, error) {
	r := make([]string, 0, len(s))
	for _, l := range s {
		expanded, err := tpl.InstantiateFromString(l, replacements)
		if err != nil {
			return r, err
		}
		r = append(r, expanded)
	}
	return r, nil
}
