package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsInclude(t *testing.T) {

	tableTests := []struct {
		name     string
		data     []string
		target   string
		expected bool
	}{
		{
			name:     "IsInclude matched",
			data:     []string{"a", "b", "c"},
			target:   "b",
			expected: true,
		},
		{
			name:     "IsInclude not matched",
			data:     []string{"a", "b", "c"},
			target:   "d",
			expected: false,
		},
	}

	for _, test := range tableTests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, IsInclude(test.data, test.target))
		})

	}

}
