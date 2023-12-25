package gpio

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNumber(t *testing.T) {
	t.Parallel()

	p := &pin{
		number: 123,
	}

	require.Equal(t, p.number, p.Number())
}

// TODO improve testing here. This specific package is hard to test without adding testing code to the actual implementation.
