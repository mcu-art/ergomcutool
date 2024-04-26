package utils

import (
	"bufio"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	homedir "github.com/mitchellh/go-homedir"
)

type OverwriteActionT uint32

/*
const (

	NOTIFY_NONE OverwriteActionT = iota
	NOTIFY_INFO
	NOTIFY_WARN
	NOTIFY_ERROR
	NOTIFY_FATAL

)
*/
const (
	OverwriteFatal OverwriteActionT = iota
	OverwriteError
	OverwriteWarn
	OverwriteInfo
	OverwriteSilently
)

type CopyFileT struct {
	Src      string
	Dest     string
	FileMode uint32
	// Overwrite         bool
	Overwrite OverwriteActionT

	// PrependPrefix is a prefix to be prepended to
	// each line of the file being copied.
	// Useful to comment out the entire file contents.
	PrependPrefix string
}

func DirExists(path string) bool {
	path, _ = homedir.Expand(path)
	if stat, err := os.Stat(path); err == nil && stat.IsDir() {
		return true
	}
	return false
}

func FileExists(path string) bool {
	path, _ = homedir.Expand(path)
	if stat, err := os.Stat(path); err == nil && !stat.IsDir() {
		return true
	}
	return false
}

func GetSortedFileList(path string, suffix string) ([]string, error) {
	l, err := GetFileList(path)
	if err != nil {
		return l, err
	}
	r := make([]string, 0, len(l))
	for _, f := range l {
		if strings.HasSuffix(f, suffix) {
			r = append(r, f)
		}
	}
	sort.Strings(r)
	return r, nil
}

// func SymlinkExists(path string) bool {
// 	path, _ = homedir.Expand(path)
// 	if stat, err := os.Stat(path); err == nil && (!stat.IsDir() && !stat.Mode().IsRegular()) {
// 		return true
// 	}
// 	return false
// }

// See https://stackoverflow.com/a/58148921/3824328
func CreateOrReplaceSymlink(symlinkPath, target string) error {
	if _, err := os.Lstat(symlinkPath); err == nil {
		if err := os.Remove(symlinkPath); err != nil {
			return fmt.Errorf("failed to unlink: %w", err)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to check symlink: %w", err)
	}
	return os.Symlink(target, symlinkPath)
}

// CopyFile copies the contents of the file named src to the file named
// by dst. The file will be created if it does not already exist. If the
// destination file exists, all it's contents will be replaced by the contents
// of the source file.
func CopyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()
	if _, err = io.Copy(out, in); err != nil {
		return err
	}
	err = out.Sync()
	return err
}

func CopyFileEx(p *CopyFileT) (err error) {
	in, err := os.Open(p.Src)
	if err != nil {
		return
	}
	defer in.Close()

	// Check if destination exists
	if FileExists(p.Dest) {
		switch p.Overwrite {
		case OverwriteFatal:
			log.Fatalf("fatal error: existing file %q must not be overwritten\n", p.Dest)
		case OverwriteError:
			return fmt.Errorf("existing file %q must not be overwritten\n", p.Dest)
		case OverwriteWarn:
			log.Printf("warning: existing file %q will be overwritten\n", p.Dest)
		case OverwriteInfo:
			log.Printf("info: existing file %q will be overwritten\n", p.Dest)
		}
	}
	out, err := os.Create(p.Dest)
	if err != nil {
		return
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
		cerr = os.Chmod(p.Dest, os.FileMode(p.FileMode))
		if err == nil {
			err = cerr
		}
	}()
	if _, err = io.Copy(out, in); err != nil {
		return
	}
	err = out.Sync()
	return
}

// CopyDir copies the content of src to dst. src should be a full path.
func CopyDir(src, dst string, dirPerm, filePerm uint32) error {

	return filepath.Walk(src, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// copy to this path
		outpath := filepath.Join(dst, strings.TrimPrefix(path, src))

		if info.IsDir() {
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
		in, err := os.Open(path)
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

// CopyEmbeddedDir copies directory from the embedded FS into specified directory
func CopyEmbeddedDir(src fs.FS, dest string, dirPerm, filePerm uint32) error {
	return fs.WalkDir(src, ".", func(path string, d fs.DirEntry, err error) error {
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
		in, err := src.Open(path)
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

func ReadTextFile(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

// GetFileList returns a list of regular files in the specified directory.
// Sub-directories and symlinks are not included.
func GetFileList(path string) ([]string, error) {
	files, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	result := make([]string, 0, 100)
	for _, file := range files {
		//fmt.Println(file.Name(), file.IsDir())
		if file.Type().IsRegular() {
			result = append(result, file.Name())
		}
	}
	return result, nil
}

// GetSymlinkList returns a list of symlinks in the specified directory.
// Sub-directories are not included.
func GetSymlinkList(path string) ([]string, error) {
	files, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	result := make([]string, 0, 100)
	for _, file := range files {
		if !file.Type().IsDir() && !file.Type().IsRegular() {
			result = append(result, file.Name())
		}
	}
	return result, nil
}

// ReadTextFileEx reads a text file into lines.
// Trailing (only) whitespace of each line is trimmed.
// Returns Lines, LineEnding, error.
func ReadTextFileEx(path string) ([]string, string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, "", err
	}
	if len(data) == 0 {
		return nil, "", fmt.Errorf("empty file")
	}

	lines := strings.Split(string(data), "\n")
	if len(lines) <= 1 {
		return nil, "", fmt.Errorf("unsupported line endings detected (bad file format?)")
	}
	carriageReturnDetected := false
	r := make([]string, 0, len(lines))
	lineEnding := ""
	for i, line := range lines {
		if strings.HasSuffix(line, "\r") {
			carriageReturnDetected = true
		}
		line = TrimRightSpace(line)
		if i < len(lines)-1 {
			r = append(r, TrimRightSpace(line))
		} else {
			if len(line) > 0 {
				r = append(r, line)
			}
		}
	}
	if carriageReturnDetected {
		lineEnding = "\r\n"
	} else {
		lineEnding = "\n"
	}
	return r, lineEnding, nil
}
