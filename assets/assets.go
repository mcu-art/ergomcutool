package assets

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"os"
	"text/template"

	"github.com/mcu-art/ergomcutool/utils"
)

//go:embed all:assets
var EmbeddedAssets embed.FS

var templatesLoaded bool
var templateSet *template.Template

/*
func _copyAsset(src, dest string, filePerm uint32) error {
	srcData, err := EmbeddedAssets.ReadFile(src)
	if err != nil {
		log.Fatalf("failed to read asset file %q: %v\n",
			src, err)
	}
	if err := os.WriteFile(dest, srcData, fs.FileMode(filePerm)); err != nil {
		log.Fatalf("failed to copy asset file %q into %q: %v\n",
			src, dest, err)
	}
	return nil
}

// copyAssetFile copies a file from assets/files into dest.
func copyAssetFile(src, dest string, filePerm uint32) error {
	src = filepath.Join("assets", "files", src)
	return _copyAsset(src, dest, filePerm)
}

// copyAssetTemplate copies a file from assets/templates into dest.
func copyAssetTemplate(src, dest string, filePerm uint32) error {
	src = filepath.Join("assets", "templates", src)
	return _copyAsset(src, dest, filePerm)
}
*/

func CopyAssets(dest string, dirPerm, filePerm uint32) error {
	return utils.CopyEmbeddedDir(EmbeddedAssets, dest, dirPerm, filePerm)
}

/*
// CopyAssets copies all embedded assets into specified directory
func CopyAssets(dest string, dirPerm, filePerm uint32) error {
	return fs.WalkDir(EmbeddedAssets, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// copy to this path
		outpath := filepath.Join(dest, path)
		info, err := d.Info()
		if err != nil {
			return fmt.Errorf("failed to get DirEntry info: %w", err)
		}
		if d.IsDir() {
			return os.MkdirAll(outpath, fs.FileMode(dirPerm)) // means recursive
		}

		// handle irregular files
		if !info.Mode().IsRegular() {
			switch info.Mode().Type() & os.ModeType {
			case os.ModeSymlink:
				link, err := os.Readlink(path)
				if err != nil {
					return err
				}
				return os.Symlink(link, outpath)
			}
			return nil
		}

		// copy contents of regular file efficiently

		// open input
		in, err := EmbeddedAssets.Open(path)
		if err != nil {
			return err
		}
		defer in.Close()

		// create output
		fh, err := os.Create(outpath)
		if err != nil {
			return err
		}
		defer fh.Close()

		// make it the same
		err = fh.Chmod(fs.FileMode(filePerm))
		if err != nil {
			return err
		}

		// copy content
		_, err = io.Copy(fh, in)
		return err
	})
}
*/

// loadAssetTemplates loads and pareses templates from the embedded FS
// into global variable `templateSet`.
func loadAssetTemplates() {
	templatesFs, err := fs.Sub(EmbeddedAssets, "assets/templates")
	if err != nil {
		log.Fatalf("failed to create embedded sub-filesystem from EmbeddedAssets: %v", err)
	}

	templateSet, err = template.New("template-set").ParseFS(templatesFs, "*.tmpl")
	if err != nil {
		log.Fatalf("failed to load templates from embedded assets: %v",
			err)
	}
	templatesLoaded = true
}

func InstantiateAssetTemplate(templateFileName, dest string,
	replacements map[string]any, filePerm uint32) error {

	errMsgPrefix := "ergomcutool.assets.InstantiateAssetTemplate: "
	if !templatesLoaded {
		loadAssetTemplates()
	}

	file, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY, os.FileMode(filePerm))
	if err != nil {
		return fmt.Errorf(errMsgPrefix+"failed to open file %q for writing: %v", dest, err)
	}
	defer file.Close()

	err = templateSet.ExecuteTemplate(file, templateFileName, replacements)
	if err != nil {
		log.Fatalf(errMsgPrefix+"template execution failed: %s", err)
	}

	return nil
}

// DELETE IT

// // CopyUserFiles copies all necessary for user settings assets into destDir.
// func CopyUserFiles(destDir string, filePerm, dirPerm uint32) error {
// 	// Create destination if not exists
// 	if err := os.MkdirAll(destDir, fs.FileMode(dirPerm)); err != nil {
// 		log.Fatalf("failed to create user directory %q: %v\n",
// 			destDir, err)
// 	}
// 	srcData, _ := EmbeddedAssets.ReadFile("assets/default_ergomcutool_config.yaml")
// 	os.WriteFile(filepath.Join(destDir, ""))
// 	// utils.CopyFile(srcDir)
// 	return nil
// }
