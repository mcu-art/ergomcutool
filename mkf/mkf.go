// package mkf is for working with a Makefile.
package mkf

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/mcu-art/ergomcutool/config"
	"github.com/mcu-art/ergomcutool/utils"
)

type ParsedMkf struct {
	// IsAutoEdited is true if the Makefile has been patched by ergomcutool
	IsAutoEdited bool
	// ErgomcutoolVersion is the version of ergomcutool that patched the Makefile
	ErgomcutoolVersion string
	// Debug is 1 if the Debug mode is selected, or 0 in Release mode.
	Debug string
	// Opt contains optimization level arguments, default is -Og
	Opt string

	// BuildDir is the build directory, default is 'build'
	BuildDir string

	// CSources is a list of .c files from the Makefile,
	// the order is preserved.
	CSources []string

	// CDefs is a list of definitions from the Makefile,
	// the order is preserved.
	CDefs []string

	// CIncludes is a list of include paths from the Makefile,
	// the order is preserved.
	CIncludes []string
}

type Mkf struct {
	// Lines are the lines read from a Makefile.
	Lines []string
	// LineEnding specifies the Makefile line endings ("\n" and "\r\n" are supported).
	LineEnding string
}

var (
	AutoEditedMarkComment = `# This file was patched by ergomcutool:`
	AutoEditedMarkPrefix  = "# ERGOMCUTOOL_VERSION ="
	ErrValueNotFound      = errors.New("value not found")
	ErrEntryNotFound      = errors.New("entry not found")
)

// FromFile reads a Makefile from the specified file.
// Line endings are detected automatically.
func FromFile(path string) (m *Mkf, err error) {
	m = &Mkf{}
	data, err := os.ReadFile(path)
	if err != nil {
		return m, err
	}
	if len(data) == 0 {
		return m, fmt.Errorf("empty file")
	}
	dataString := strings.Split(string(data), "\r\n")
	if len(dataString) == 1 {
		dataString = strings.Split(string(data), "\n")
		if len(dataString) == 1 {
			return m, fmt.Errorf("unsupported line endings detected (bad file format?)")
		}
		m.LineEnding = "\n"
	} else {
		m.LineEnding = "\r\n"
	}

	// Trim trailing whitespace and remove last element if empty
	lastElementIndex := len(dataString) - 1
	for i := 0; i < lastElementIndex; i++ {
		// Trim only trailing whitespace!
		m.Lines = append(m.Lines, utils.TrimRightSpace(dataString[i]))
	}
	lastElement := utils.TrimRightSpace(dataString[lastElementIndex])
	if lastElement != "" {
		m.Lines = append(m.Lines, lastElement)
	}
	return m, err
}

// RemoveValue removes the value of the specified entry.
// It doesn't remove the entry itself.
func (m *Mkf) RemoveValue(entryName string) error {
	entryFound := false
	var startIndex, endIndex int
	for i, line := range m.Lines {
		if strings.HasPrefix(line, entryName) {
			startIndex = i
			endIndex = i
			entryFound = true

		}
		if entryFound {
			if strings.HasSuffix(line, "\\") {
				continue
			}
			endIndex = i
			break
		}
	}
	if !entryFound {
		return ErrEntryNotFound
	}
	m.Lines[startIndex] = entryName + " = "
	if startIndex != endIndex {
		tmp := make([]string, 0, len(m.Lines))
		tmp = append(tmp, m.Lines[:startIndex+1]...)
		tmp = append(tmp, m.Lines[endIndex+1:]...)
		m.Lines = tmp
	}
	return nil
}

// InsertValue writes the value of the specified entry.
// It doesn't create the entry itself.
func (m *Mkf) InsertValue(entryName string, values []string) error {
	entryFound := false
	index := 0
	for i, line := range m.Lines {
		if strings.HasPrefix(line, entryName) {
			entryFound = true
			index = i
			if len(values) == 1 {
				m.Lines[i] = line + values[0]
				return nil
			}
			break
		}
	}
	if !entryFound {
		return ErrEntryNotFound
	}
	// Insert multiline value
	m.Lines[index] += " \\"
	tmp := make([]string, 0, len(m.Lines)+len(values))
	tmp = append(tmp, m.Lines[:index+1]...)

	for j := 0; j < len(values)-1; j++ {
		tmp = append(tmp, values[j]+" \\")
	}
	// last value must not contain trailing backslash
	tmp = append(tmp, values[len(values)-1])
	tmp = append(tmp, m.Lines[index+1:]...)
	m.Lines = tmp
	return nil
}

// ReplaceValue removes the original entry value and inserts the new one.
func (m *Mkf) ReplaceValue(entryName string, values []string) error {
	err := m.RemoveValue(entryName)
	if err != nil {
		return err
	}
	return m.InsertValue(entryName, values)
}

// AppendTextLines appends a text block at the end of the makefile,
// but before `# *** EOF ***` line if such line exists.
// The text block must not contain line endings.
func (m *Mkf) AppendTextLines(textLines []string, appendEmptyLine bool) error {
	index := len(m.Lines) - 1
	for i := index; i >= 0; i-- {
		line := m.Lines[i]
		if strings.Contains(line, "*** EOF ***") {
			index = i
			break
		}
	}
	extraLines := 0
	if appendEmptyLine {
		extraLines = 1
	}
	tmp := make([]string, 0, len(m.Lines)+len(textLines)+extraLines)
	tmp = append(tmp, m.Lines[:index]...)
	tmp = append(tmp, textLines...)
	if appendEmptyLine {
		tmp = append(tmp, "")
	}
	tmp = append(tmp, m.Lines[index:]...)
	m.Lines = tmp
	return nil
}

// AppendString is similar to AppendTextLines
// but it splits 's' into lines first, and then appends the lines.
// No whitespace characters will be removed, except '\r' and '\n'.
func (m *Mkf) AppendString(s string, appendEmptyLine bool) error {
	lines := strings.Split(s, "\n")
	// remove \r symbols at the end of each line if any;
	// do not remove any other whitespace characters.
	for i, line := range lines {
		lines[i] = strings.TrimRight(line, "\r")
	}
	return m.AppendTextLines(lines, appendEmptyLine)
}

// InsertAutoEditedMark inserts a mark
// that helps to identify if the Makefile was edited by ergomcutool.
func (m *Mkf) InsertAutoEditedMark() error {
	index := 0
	// Find first empty line
	for i, line := range m.Lines {
		if line == "" {
			index = i
			break
		}
	}
	tmp := make([]string, 0, len(m.Lines)+3)
	tmp = append(tmp, m.Lines[:index]...)
	tmp = append(tmp, "")
	tmp = append(tmp, AutoEditedMarkComment)
	tmp = append(tmp, fmt.Sprintf("%s %s", AutoEditedMarkPrefix, config.Version))
	tmp = append(tmp, m.Lines[index:]...)
	m.Lines = tmp
	return nil
}

func (m *Mkf) IsAutoEdited() bool {
	for _, line := range m.Lines {
		if strings.HasPrefix(line, AutoEditedMarkPrefix) {
			return true
		}
	}
	return false
}

func (m *Mkf) readAutoEditedVersion() string {
	for _, line := range m.Lines {
		if strings.HasPrefix(line, AutoEditedMarkPrefix) {
			pos := strings.Index(line, "=")
			if pos == -1 {
				return ""
			}
			line = strings.TrimSpace(line[pos+1:])
			return line
		}
	}
	return ""
}

func (m *Mkf) ReadValue(entryName string) ([]string, error) {
	r := make([]string, 0, 50)
	checkNextLine := false
	for _, line := range m.Lines {
		if checkNextLine {
			line = strings.TrimSpace(line)
			if len(line) == 0 {
				return r, nil
			}
			// Check if last character is '\'
			if line[len(line)-1] == '\\' {
				line = strings.TrimSpace(line[:len(line)-1])
				if line != "" {
					r = append(r, line)
				}
			} else { // line doesn't contain '\' at the end, it's the last line
				if line != "" {
					r = append(r, line)
				}
				return r, nil
			}
			continue
		}
		if strings.HasPrefix(line, entryName) {
			pos := strings.Index(line, "=")
			if pos == -1 {
				return r, ErrValueNotFound
			}
			value := line[pos+1:]
			value = strings.TrimSpace(value)

			valueLength := len(value)
			if valueLength == 0 { // empty value
				return r, nil
			}
			// Check if last character is '\'
			if value[valueLength-1] == '\\' {
				value = strings.TrimSpace(value[:valueLength-1])
				checkNextLine = true
				if value != "" {
					r = append(r, value)
				}
			} else { // line doesn't contain '\' at the end
				r = append(r, value)
			}
			if checkNextLine {
				continue
			}
			return r, nil
		}
	}
	return r, ErrEntryNotFound
}

// ParseMkf parses the internal Lines field into ParsedMkf.
func (m *Mkf) Parse() (*ParsedMkf, error) {
	var err error
	var vals []string
	r := &ParsedMkf{}
	r.IsAutoEdited = m.IsAutoEdited()

	if vals, err = m.ReadValue("BUILD_DIR"); err != nil {
		return r, fmt.Errorf("failed to read makefile 'BUILD_DIR' entry: %w", err)
	}
	if len(vals) > 0 {
		r.BuildDir = vals[0]
	}

	if r.CDefs, err = m.ReadValue("C_DEFS"); err != nil {
		return r, fmt.Errorf("failed to read makefile 'C_DEFS' entry: %w", err)
	}

	if r.CIncludes, err = m.ReadValue("C_INCLUDES"); err != nil {
		return r, fmt.Errorf("failed to read makefile 'C_INCLUDES' entry: %w", err)
	}

	if r.CSources, err = m.ReadValue("C_SOURCES"); err != nil {
		return r, fmt.Errorf("failed to read makefile 'C_SOURCES' entry: %w", err)
	}

	if vals, err = m.ReadValue("DEBUG"); err != nil {
		return r, fmt.Errorf("failed to read makefile 'DEBUG' entry: %w", err)
	}
	if len(vals) > 0 {
		r.Debug = vals[0]
	}

	if vals, err = m.ReadValue("OPT"); err != nil {
		return r, fmt.Errorf("failed to read makefile 'OPT' entry: %w", err)
	}
	if len(vals) > 0 {
		r.Opt = vals[0]
	}
	r.ErgomcutoolVersion = m.readAutoEditedVersion()

	return r, nil
}

func (m *Mkf) Bytes() []byte {
	return []byte(m.String())
}

func (m *Mkf) String() string {
	b := strings.Builder{}
	for _, line := range m.Lines {
		b.WriteString(line)
		b.WriteString(m.LineEnding)
	}
	return b.String()
}
