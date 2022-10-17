package must

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMustRequire(t *testing.T) {
	_, err := os.OpenFile("mock.txt", os.O_APPEND, 0666)
	require.Equal(t, "open mock.txt: no such file or directory", err.Error())

}

func TestMustAssert(t *testing.T) {
	_, err := os.OpenFile("./test.txt", os.O_CREATE, 0666)
	assert.Equal(t, err, (nil))

}
