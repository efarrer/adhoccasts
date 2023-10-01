package filesystem

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"time"
)

type File struct {
	Name    string
	ModTime time.Time
	Size    int64
}

type FS struct {
	isDir     func(string) (bool, error)
	dirNames  func(string) ([]string, error)
	listFiles func(string) ([]File, error)
}

func Default() FS {
	return New(
		IsDirectory,
		DirectoryNames,
		ListFiles,
	)
}

func New(
	isDir func(string) (bool, error),
	dirNames func(string) ([]string, error),
	listFiles func(string) ([]File, error),
) FS {
	return FS{
		isDir:     isDir,
		dirNames:  dirNames,
		listFiles: listFiles,
	}
}

func (fs FS) ValidatePath(path string) (string, error) {
	cleanPath := filepath.Clean(path)
	cleanPath, err := filepath.Abs(cleanPath)
	if err != nil {
		return "", errors.New(fmt.Sprintf("%s is not a valid path.", path))
	}

	return cleanPath, nil
}

func (fs FS) ValidateDir(dir string) (string, error) {
	cleanDir, err := fs.ValidatePath(dir)
	if err != nil {
		return "", err
	}
	isDir, err := fs.isDir(cleanDir)
	if !isDir || err != nil {
		return "", errors.New(fmt.Sprintf("%s is not a directory.", dir))
	}

	return cleanDir, nil
}

func (fs FS) IsDirectory(dir string) (bool, error) {
	return fs.isDir(dir)
}

func (fs FS) DirectoryNames(dir string) ([]string, error) {
	return fs.dirNames(dir)
}

func (fs FS) ListFiles(dir string) ([]File, error) {
	return fs.listFiles(dir)
}

// IsDirectory returns true if dir is a directory.
func IsDirectory(dir string) (bool, error) {
	stat, err := os.Stat(dir)
	if err != nil {
		return false, err
	}
	if stat.IsDir() {
		return true, nil
	}
	return false, nil
}

// DirectoryNames returns the names of all subdirectories under dir.
func DirectoryNames(dir string) ([]string, error) {
	fileInfos, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("reading directories from %s, %w", dir, err)
	}
	dirs := make([]string, 0, len(fileInfos))
	for _, fileInfo := range fileInfos {
		fullDir := path.Join(dir, fileInfo.Name())
		dirName := fileInfo.Name()
		isDir, err := IsDirectory(fullDir)
		if err != nil {
			return nil, fmt.Errorf("unable to stat dir %s, %w", fullDir, err)
		}
		if isDir {
			dirs = append(dirs, dirName)
		}
	}
	return dirs, nil
}

func ListFiles(dir string) ([]File, error) {
	fileInfos, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("listing files %s, %w", dir, err)
	}

	files := make([]File, 0, len(fileInfos))
	for _, fileInfo := range fileInfos {
		if fileInfo.IsDir() {
			continue
		}
		info, err := fileInfo.Info()
		if err != nil {
			return nil, fmt.Errorf("getting file info %s, %w", dir, err)
		}

		files = append(files, File{
			Name:    fileInfo.Name(),
			ModTime: info.ModTime(),
			Size:    info.Size(),
		})
	}
	return files, nil
}
