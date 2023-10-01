package filesystem_test

import (
	"testing"

	"github.com/efarrer/adhoccasts/filesystem"
	"github.com/stretchr/testify/require"
)

func newIsDir(b bool, e error) func(string) (bool, error) {
	return func(d string) (bool, error) {
		return b, e
	}
}
func directoryNames(dirs []string, e error) func(string) ([]string, error) {
	return func(d string) ([]string, error) {
		return dirs, e
	}
}

func TestValidatePath_ReturnsAbsPath(t *testing.T) {
	fs := filesystem.New(
		newIsDir(true, nil),
		nil,
		nil,
	)
	path, err := fs.ValidatePath("/home/adhoccasts/../")
	require.NoError(t, err)
	require.Equal(t, "/home", path)
}

func TestValidateDir_ReturnsAbsPathForDir(t *testing.T) {
	fs := filesystem.New(
		newIsDir(true, nil),
		nil,
		nil,
	)
	path, err := fs.ValidateDir("/home/adhoccasts")
	require.NoError(t, err)
	require.Equal(t, "/home/adhoccasts", path)
}

func TestValidateDir_ReturnsErrorForNonDirs(t *testing.T) {
	fs := filesystem.New(
		newIsDir(false, nil),
		nil,
		nil,
	)
	_, err := fs.ValidateDir("/home/adhoccasts/foo.txt")
	require.Error(t, err)
}

func TestDirectoryNames_ReturnsDirectoryNames(t *testing.T) {
	expectedDirs := []string{"/a", "/b"}
	fs := filesystem.New(
		nil,
		directoryNames(expectedDirs, nil),
		nil,
	)
	dirs, err := fs.DirectoryNames("/")
	require.NoError(t, err)
	require.Equal(t, expectedDirs, dirs)
}
