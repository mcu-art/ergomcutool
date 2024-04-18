package tpl

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"text/template"

	"github.com/mcu-art/ergomcutool/config"
)

var initCmdTemplatesLoaded bool
var initCmdTemplates *template.Template

// loadInitCmdTemplates loads and pareses templates from the user config directory
// into global variable `initCmdTemplates`.
func loadInitCmdTemplates() {
	path := filepath.Join(config.UserConfigDir, "assets", "init_cmd", "templates")
	templatesFs := os.DirFS(path)

	var err error
	initCmdTemplates, err = template.New("initCmdTemplates").ParseFS(templatesFs, "*.tmpl")
	if err != nil {
		log.Fatalf("failed to load templates from embedded assets: %v",
			err)
	}
	initCmdTemplatesLoaded = true
}

func InstantiateInitCmdTemplate(templateFileName, dest string,
	replacements any, filePerm uint32) error {

	errMsgPrefix := "ergomcutool.InstantiateInitCmdTemplate: "
	if !initCmdTemplatesLoaded {
		loadInitCmdTemplates()
	}

	file, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY, os.FileMode(filePerm))
	if err != nil {
		return fmt.Errorf(errMsgPrefix+"failed to open file %q for writing: %v", dest, err)
	}
	defer file.Close()

	err = initCmdTemplates.ExecuteTemplate(file, templateFileName, replacements)
	if err != nil {
		log.Fatalf(errMsgPrefix+"template execution failed: %s", err)
	}

	return nil
}
