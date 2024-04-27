package iocfile

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOnCorrectSample(t *testing.T) {
	sample1Path := "./test_data/sample1.ioc"
	m, err := FromFile(sample1Path)
	require.Nil(t, err)
	require.NotNil(t, m)
	require.Equal(t, 285, len(m.Lines))
	require.Equal(t, "\n", m.LineEnding)
}

func TestReadValue(t *testing.T) {
	sample1Path := "./test_data/sample1.ioc"
	m, _ := FromFile(sample1Path)
	val, err := m.ReadValue("CAD.formats")
	require.Nil(t, err)
	require.Equal(t, "", val, "CAD.formats mustn't have value")

	val, err = m.ReadValue("File.Version")
	require.Nil(t, err)
	require.Equal(t, "6", val)

	val, err = m.ReadValue("ProjectManager.UAScriptBeforePath")
	require.Nil(t, err)
	require.Equal(t, "_non_persistent/auto_generated/cubemx_actions/before_generate.sh", val)

	_, err = m.ReadValue("NOT_EXISTS")
	require.NotNil(t, err)
}

func TestReplaceValue(t *testing.T) {
	sample1Path := "./test_data/sample1.ioc"
	m, _ := FromFile(sample1Path)
	val, err := m.ReplaceValue("CAD.formats", "dummy_CAD_formats")
	require.Nil(t, err)
	require.Equal(t, "", val, "CAD.formats old value must be empty")

	val, err = m.ReadValue("CAD.formats")
	require.Nil(t, err)
	require.Equal(t, "dummy_CAD_formats", val)

	val, err = m.ReplaceValue("File.Version", "7")
	require.Nil(t, err)
	require.Equal(t, "6", val, "CAD.formats old value must be empty")
	val, err = m.ReadValue("File.Version")
	require.Nil(t, err)
	require.Equal(t, "7", val)

	_, err = m.ReplaceValue("NOT_EXISTS", "dummy")
	require.NotNil(t, err)
}

/*

func TestParse(t *testing.T) {
	sample1Path := "./test_data/sample1.txt"
	m, _ := FromFile(sample1Path)
	parsed, err := m.Parse()
	require.Nil(t, err)
	require.Equal(t, 2, len(parsed.CDefs))
	require.Equal(t, false, m.IsAutoEdited())
}

func TestRemoveValue(t *testing.T) {
	sample1Path := "./test_data/sample1.txt"
	m, _ := FromFile(sample1Path)
	err := m.RemoveValue("OPT")
	require.Nil(t, err)
	err = m.RemoveValue("DEBUG")
	require.Nil(t, err)
	err = m.RemoveValue("C_SOURCES")
	require.Nil(t, err)
	err = m.RemoveValue("NON_EXISTING_ENTRY")
	require.ErrorIs(t, err, ErrEntryNotFound)
}

func TestReplaceValue(t *testing.T) {
	sample1Path := "./test_data/sample1.txt"
	m, _ := FromFile(sample1Path)
	err := m.ReplaceValue("OPT", []string{"dummy-opt-option"})
	require.Nil(t, err)

	err = m.ReplaceValue("DEBUG", []string{"0"})
	require.Nil(t, err)

	values, err := m.ReadValue("C_SOURCES")
	require.Nil(t, err)
	values = append(values, "/dummy/file1.c")
	values = append(values, "file2.c")
	err = m.ReplaceValue("C_SOURCES", values)
	require.Nil(t, err)

	err = m.ReplaceValue("NON_EXISTING_ENTRY", []string{"dummy"})
	require.ErrorIs(t, err, ErrEntryNotFound)
}

func TestInsertAutoEditedMark(t *testing.T) {
	sample1Path := "./test_data/sample1.txt"
	m, _ := FromFile(sample1Path)
	err := m.InsertAutoEditedMark()
	require.Nil(t, err)
	require.True(t, m.IsAutoEdited())
}

func TestAppendTextLines(t *testing.T) {
	sample1Path := "./test_data/sample1.txt"
	m, _ := FromFile(sample1Path)
	lines := []string{"# This is a test entry:", "TEST_ENTRY: \\", "\tTEST_VALUE"}
	err := m.AppendTextLines(lines, true)
	require.Nil(t, err)
}

func TestAppendString(t *testing.T) {
	sample1Path := "./test_data/sample1.txt"
	m, _ := FromFile(sample1Path)
	s := `# This is a test entry:
TEST_ENTRY_2: \
	TEST_VALUE_2`

	err := m.AppendString(s, true)
	require.Nil(t, err)

	// fmt.Println(m.GetString())
	// require.False(t, true)
}
*/