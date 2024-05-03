package proj

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReadAndValidate(t *testing.T) {
	sample1Path := "./test_data/ergomcu_project_sample1.yaml"
	m, err := ReadAndValidate(sample1Path)
	require.Nil(t, err)
	require.NotNil(t, m)
}
