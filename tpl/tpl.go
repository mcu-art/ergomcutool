package tpl

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"text/template"

	"github.com/mcu-art/ergomcutool/config"
)

var templatesLoaded bool
var assetTemplates *template.Template

// loadAssetTemplates loads and pareses templates from the user config directory
// into global variable `assetTemplates`.
func loadAssetTemplates() {
	path := filepath.Join(config.UserConfigDir, "assets", "templates")
	templatesFs := os.DirFS(path)

	var err error
	assetTemplates, err = template.New("assetTemplates").ParseFS(templatesFs, "*.tmpl")
	if err != nil {
		log.Fatalf("failed to load templates from embedded assets: %v",
			err)
	}
	templatesLoaded = true
}

// loadDirTemplates loads and pareses templates from the specified directory.
func loadDirTemplates(path, templateName string) (*template.Template, error) {
	templatesFs := os.DirFS(path)
	return template.New(templateName).ParseFS(templatesFs, "*.tmpl")
}

// InstantiateAssetTemplate instantiates the specified template file from
// UserConfigDir assets/templates.
func InstantiateAssetTemplate(templateFileName, dest string,
	replacements any, filePerm uint32) error {

	errMsgPrefix := "ergomcutool.InstantiateTemplate: "
	if !templatesLoaded {
		loadAssetTemplates()
	}

	file, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY, os.FileMode(filePerm))
	if err != nil {
		return fmt.Errorf(errMsgPrefix+"failed to open file %q for writing: %v", dest, err)
	}
	defer file.Close()

	err = assetTemplates.ExecuteTemplate(file, templateFileName, replacements)
	if err != nil {
		log.Fatalf(errMsgPrefix+"template execution failed: %s", err)
	}

	return nil
}

func InstantiateToString(
	dirWithTemplates, templateFileName string, replacements any) (string, error) {
	templates, err := loadDirTemplates(dirWithTemplates, templateFileName)
	if err != nil {
		return "", err
	}
	var buff bytes.Buffer
	err = templates.ExecuteTemplate(&buff, templateFileName, replacements)
	joinedPath := filepath.Join(dirWithTemplates, templateFileName)
	if err != nil {
		return "", fmt.Errorf("template execution failed for %q: %v", joinedPath, err)
	}
	return buff.String(), nil
}

// InstantiateFromString instantiates a template passed as a string.
func InstantiateFromString(ts string, replacements any) (string, error) {
	t1 := template.New("string_template").Option("missingkey=error")
	t1, err := t1.Parse(ts)
	if err != nil {
		return "", err
	}
	var buff bytes.Buffer
	err = t1.Execute(&buff, replacements)
	if err != nil {
		return "", err
	}
	return buff.String(), nil
}
