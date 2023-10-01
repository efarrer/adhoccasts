package handler

import (
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"text/template"
)

const (
	listing = `<html>
        <head><title>Adhoc Podcasts</title></head>
        <body>
            {{range $name, $url := .NameToPath}}
                <a href="{{ $url }}">{{ $name }}</a><br>
            {{ end }}
        </body>
    </html>
    `
)

type DirectoryNamer interface {
	DirectoryNames(dir string) ([]string, error)
}

type Indexer struct {
	rootUrl string
	rootDir string
	fs      DirectoryNamer
}

func NewIndexer(rootUrl string, rootDir string, fs DirectoryNamer) Indexer {
	return Indexer{
		rootUrl: rootUrl,
		rootDir: rootDir,
		fs:      fs,
	}
}

type Index struct {
	NameToPath map[string]string
}

func (i Indexer) Render() (Index, error) {
	index := Index{
		NameToPath: make(map[string]string),
	}
	dirs, err := i.fs.DirectoryNames(i.rootDir)
	if err != nil {
		return index, fmt.Errorf("render index for %s, %w", i.rootDir, err)
	}
	for _, dir := range dirs {
		parsedUrl, err := url.Parse(i.rootUrl)
		if err != nil {
			return index, fmt.Errorf("parsing generated %s, %s, %w", i.rootDir, dir, err)
		}
		parsedUrl.Path = "/" + filepath.Clean(dir) + ".xml"
		index.NameToPath[dirToTitle(dir)] = parsedUrl.String()
	}

	return index, nil
}

func (i Indexer) Wrap(path string, handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == path {
			index, err := i.Render()
			if err != nil {
				httpError(w, http.StatusInternalServerError, "Unable to render", err)
				return
			}
			templ, err := template.New("Listing").Parse(listing)
			if err != nil {
				httpError(w, http.StatusInternalServerError, "Unable to parse template", err)
				return
			}
			err = templ.Execute(w, &index)
			if err != nil {
				httpError(w, http.StatusInternalServerError, "Unable to execute template", err)
				return
			}
			return
		}
		handler.ServeHTTP(w, r)
	})
}

func dirToTitle(dir string) string {
	title__description := strings.Split(filepath.Base(dir), "__")
	title := title__description[0]
	return strings.Replace(title, "_", " ", -1)
}
