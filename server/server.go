package server

import (
	"fmt"
	"net/http"

	"github.com/efarrer/adhoccasts/filesystem"
	"github.com/efarrer/adhoccasts/handler"
)

type Filesystemer interface {
	ValidateDir(dir string) (string, error)
	DirectoryNames(dir string) ([]string, error)
	ListFiles(string) ([]filesystem.File, error)
	IsDirectory(string) (bool, error)
}

type ServeMux struct {
	rootUrl string
	rootDir string
	fs      Filesystemer

	HttpServeMux *http.ServeMux
}

func NewServeMux(rootUrl, rootDir string, fs Filesystemer) (*ServeMux, error) {
	cleanDir, err := fs.ValidateDir(rootDir)
	if err != nil {
		return nil, fmt.Errorf("create ServeMux for %s, %w", rootDir, err)
	}
	return &ServeMux{
		rootUrl:      rootUrl,
		rootDir:      cleanDir,
		fs:           fs,
		HttpServeMux: http.NewServeMux(),
	}, nil
}

func (mux *ServeMux) ListenAndServe(addr string) error {
	indexer := handler.NewIndexer(mux.rootUrl, mux.rootDir, mux.fs)
	rsser := handler.NewRsser(mux.rootUrl, mux.rootDir, mux.fs)

	h := http.FileServer(http.Dir(mux.rootDir))
	h = indexer.Wrap("/", h)
	h = rsser.Wrap(h)
	mux.HttpServeMux.Handle("/", h)

	return http.ListenAndServe(addr, mux.HttpServeMux)
}
