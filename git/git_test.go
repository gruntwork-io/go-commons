package git

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gruntwork-io/go-commons/files"
	"github.com/gruntwork-io/go-commons/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGitClone(t *testing.T) {
	t.Parallel()

	tmpDir, err := os.MkdirTemp("", "git-test")
	require.NoError(t, err)
	require.NoError(t, Clone(logging.GetLogger(t.Name(), ""), "https://github.com/gruntwork-io/go-commons.git", tmpDir))
	assert.True(t, files.FileExists(filepath.Join(tmpDir, "LICENSE.txt")))
}

func TestGitCheckout(t *testing.T) {
	t.Parallel()

	tmpDir, err := os.MkdirTemp("", "git-test")
	require.NoError(t, err)
	require.NoError(t, Clone(logging.GetLogger(t.Name(), ""), "https://github.com/gruntwork-io/go-commons.git", tmpDir))
	require.NoError(t, Checkout(logging.GetLogger(t.Name(), ""), "v0.10.0", tmpDir))
	assert.False(t, files.FileExists(filepath.Join(tmpDir, "git", "git_test.go")))
}
