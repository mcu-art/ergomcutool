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
