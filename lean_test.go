package lean_test

import (
	"os"
	"regexp"
	"testing"

	"github.com/lean-docs/lean"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// C0.3
func TestVersionString(t *testing.T) {
	v := lean.Version()
	require.NotEmpty(t, v)
	// Must be valid semver (with optional -prerelease suffix)
	matched, err := regexp.MatchString(`^\d+\.\d+\.\d+(-[\w.]+)?$`, v)
	require.NoError(t, err)
	assert.True(t, matched, "version %q is not valid semver", v)
}

// C0.6
func TestCIConfigPresent(t *testing.T) {
	info, err := os.Stat(".github/workflows/ci.yml")
	require.NoError(t, err, "CI config file must exist")
	assert.False(t, info.IsDir())

	data, err := os.ReadFile(".github/workflows/ci.yml")
	require.NoError(t, err)

	content := string(data)
	assert.Contains(t, content, "go test", "CI must contain test step")
	assert.Contains(t, content, "go vet", "CI must contain vet step")
	assert.Contains(t, content, "staticcheck", "CI must contain lint step")
}
