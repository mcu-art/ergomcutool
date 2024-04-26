package cli

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"

	"github.com/mcu-art/ergomcutool/config"
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
	up_Verbose  bool
	up_Makefile string
)

func init() {
	rootCmd.AddCommand(updateProjectCmd)
	updateProjectCmd.PersistentFlags().BoolVarP(
		&up_Verbose, "verbose", "", false, "Verbose mode")
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
	projectFile := filepath.Join(cwd, "ergomcu_project.yaml")
	pc, err := proj.ReadAndValidate(projectFile)
	if err != nil {
		log.Fatalf("error: failed to read project file %q: %v\n", projectFile, err)
	}

	log.Printf("Updating project %q...\n", *pc.ProjectName)
	if up_Verbose {
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
					log.Fatalf("failed to create directory %q: %v\n", externalDir, err)
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
				log.Fatalf("failed to create or replace symlink %q to %q: %v\n",
					newLink, dep.Path, err)
			}
		}
		log.Printf("Done.\n")
	}

	// Read the Makefile
	makefile, err := mkf.FromFile(up_Makefile)
	if err != nil {
		log.Fatalf("failed to read and parse the makefile %q: %v\n",
			up_Makefile, err)
	}
	preEditedMakefilePath := filepath.Join(cwd, "_non_persistent", "Makefile.pre-edit")
	moveMakefileToPreEdited := false
	if makefile.IsAutoEdited() {
		// Read the original version
		makefile, err = mkf.FromFile(preEditedMakefilePath)
		if err != nil {
			log.Fatalf("failed to read and parse %q: %v\n",
				preEditedMakefilePath, err)
		}
	} else { // Makefile is not edited, move it later to _non_persistent/Makefile.pre-edit
		moveMakefileToPreEdited = true
	}

	if up_Verbose {
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
		log.Fatalf("failed to read C_SOURCES from the makefile: %v\n", err)
	}
	c_src = append(c_src, pc.CSrc...)

	// Resolve CSrcDirs
	for _, d := range pc.CSrcDirs {
		l, err := utils.GetSortedFileList(d, ".c")
		if err != nil {
			log.Fatalf("failed to read source directory %q: %v\n", d, err)
		}
		for _, filename := range l {
			fullName := filepath.Join(d, filename)
			c_src = append(c_src, fullName)
		}
	}
	// Expand external dependencies in each line
	c_src, err = expandExternalDependencies(c_src, externalDepExpansionMap)
	if err != nil {
		log.Fatalf("failed to expand external dependencies in C_SOURCES: %v\n", err)
	}
	err = makefile.ReplaceValue("C_SOURCES", c_src)
	if err != nil {
		log.Fatalf("failed to replace C_SOURCES in the makefile: %v\n", err)
	}

	// c_includes
	c_includes, err := makefile.ReadValue("C_INCLUDES")
	if err != nil {
		log.Fatalf("failed to read C_INCLUDES from the makefile: %v\n", err)
	}
	c_includes = append(c_includes, pc.CIncludeDirs...)
	err = makefile.ReplaceValue("C_INCLUDES", c_includes)
	if err != nil {
		log.Fatalf("failed to replace C_INCLUDES in the makefile: %v\n", err)
	}

	// C_DEFS
	c_defs, err := makefile.ReadValue("C_DEFS")
	if err != nil {
		log.Fatalf("failed to read C_DEFS from the makefile: %v\n", err)
	}

	c_defs = append(c_defs, pc.CDefs...)

	err = makefile.ReplaceValue("C_DEFS", c_defs)
	if err != nil {
		log.Fatalf("failed to replace C_DEFS in the makefile: %v\n", err)
	}

	// Instantiate and append the 'prog' target
	progSnippetUserDir := filepath.Join(config.UserConfigDir, "assets", "snippets")
	progSnippetLocalDir := filepath.Join(cwd, ".ergomcutool", "assets", "snippets")
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
		log.Fatalf("failed to instantiate snippet template %q: %v\n", up_Makefile, err)
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
				log.Fatalf("failed to remove old %q: %v\n",
					preEditedMakefilePath, err)
			}
		}

		err = os.Rename(up_Makefile, preEditedMakefilePath)
		if err != nil {
			log.Fatalf("failed to move %q to %q: %v\n",
				up_Makefile, preEditedMakefilePath, err)
		}
	}

	if up_Verbose {
		log.Printf("* updated makefile contains %d lines.\n", len(makefile.Lines))
	}

	err = os.WriteFile(up_Makefile, makefile.Bytes(), fs.FileMode(config.DefaultFilePermissions))
	if err != nil {
		log.Fatalf("failed to write the updated %q: %v\n", up_Makefile, err)
	}
	log.Printf("The project was successfully updated.")
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
