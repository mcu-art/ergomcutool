package iocfile

import (
	"errors"
	"strings"

	"github.com/mcu-art/ergomcutool/utils"
)

var (
	ErrEntryNotFound = errors.New("entry not found")
)

type Ioc struct {
	Lines      []string
	LineEnding string
}

// ParsedIoc contains some fields of interest from the .ioc file
// generated by CubeMX.
type ParsedIoc struct {
	ProjectName        string
	DeviceId           string
	UAScriptAfterPath  string // ProjectManager.UAScriptAfterPath
	UAScriptBeforePath string // ProjectManager.UAScriptBeforePath
}

func FromFile(path string) (*Ioc, error) {
	ioc := &Ioc{}
	var err error
	ioc.Lines, ioc.LineEnding, err = utils.ReadTextFileEx(path)
	return ioc, err
}

func (ioc *Ioc) Parse() (*ParsedIoc, error) {
	result := &ParsedIoc{}
	for _, line := range ioc.Lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "ProjectManager.ProjectName") {
			substrings := strings.Split(line, "=")
			if len(substrings) > 1 {
				result.ProjectName = substrings[1]
			}
		}
		if strings.HasPrefix(line, "ProjectManager.DeviceId") {
			substrings := strings.Split(line, "=")
			if len(substrings) > 1 {
				result.DeviceId = substrings[1]
			}
		}
		if strings.HasPrefix(line, "ProjectManager.UAScriptAfterPath") {
			substrings := strings.Split(line, "=")
			if len(substrings) > 1 {
				result.UAScriptAfterPath = substrings[1]
			}
		}
		if strings.HasPrefix(line, "ProjectManager.UAScriptBeforePath") {
			substrings := strings.Split(line, "=")
			if len(substrings) > 1 {
				result.UAScriptBeforePath = substrings[1]
			}
		}
	}
	return result, nil
}

// ReadValue returns the value of the specified key or error if key doesn't exist.
func (ioc *Ioc) ReadValue(key string) (string, error) {
	for _, line := range ioc.Lines {
		if strings.HasPrefix(line, key) {
			split := strings.Split(line, "=")
			if len(split) > 1 {
				return split[1], nil
			}
		}
	}
	return "", ErrEntryNotFound
}

// ReplaceValue replaces the value of the specified key and
// returns the old value and the error if any.
func (ioc *Ioc) ReplaceValue(key string, value string) (string, error) {
	for i, line := range ioc.Lines {
		if strings.HasPrefix(line, key) {
			split := strings.Split(line, "=")
			result := split[0] + "=" + value
			ioc.Lines[i] = result
			if len(split) > 1 {
				return split[1], nil
			}
			return "", nil
		}
	}
	return "", ErrEntryNotFound
}

func (ioc *Ioc) Bytes() []byte {
	return []byte(ioc.String())
}

func (ioc *Ioc) String() string {
	b := strings.Builder{}
	for _, line := range ioc.Lines {
		b.WriteString(line)
		b.WriteString(ioc.LineEnding)
	}
	return b.String()
}
