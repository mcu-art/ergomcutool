package mkf

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOnCorrectSample(t *testing.T) {
	sample1Path := "./test_data/sample1.txt"
	m, err := FromFile(sample1Path)
	require.Nil(t, err)
	require.NotNil(t, m)
	require.Equal(t, 205, len(m.Lines))
	require.Equal(t, "\r\n", m.LineEnding)
	require.False(t, m.IsAutoEdited())

}

func TestReadValue(t *testing.T) {
	sample1Path := "./test_data/sample1.txt"
	m, _ := FromFile(sample1Path)
	vals, err := m.ReadValue("TARGET")
	require.Nil(t, err)
	require.Equal(t, 1, len(vals), "TARGET entry must contain exactly one value")
	require.Equal(t, "sample1", vals[0])

	vals, err = m.ReadValue("DEBUG")
	require.Nil(t, err)
	require.Equal(t, 1, len(vals), "DEBUG entry must contain exactly one value")
	require.Equal(t, "1", vals[0])

	vals, err = m.ReadValue("OPT")
	require.Nil(t, err)
	require.Equal(t, 1, len(vals), "OPT entry must contain exactly one value")
	require.Equal(t, "-Og", vals[0])

	vals, err = m.ReadValue("C_SOURCES")
	require.Nil(t, err)

	require.Equal(t, 27, len(vals), "C_SOURCES entry must contain 27 values")

	vals, err = m.ReadValue("C_DEFS")
	require.Nil(t, err)
	require.Equal(t, []string{"-DUSE_HAL_DRIVER", "-DSTM32G431xx"}, vals,
		"C_DEFS entry must contain 2 values")
	// for i, v := range vals {
	// 	fmt.Printf("%d. %v\n", i+1, v)
	// }
	// require.True(t, false)

	_, err = m.ReadValue("NON_EXISTING_VALUE")
	require.ErrorIs(t, err, ErrEntryNotFound)

	vals, err = m.ReadValue("AS_DEFS")
	require.Nil(t, err)
	require.Equal(t, 0, len(vals))
}

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
