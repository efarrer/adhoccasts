package handler

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/efarrer/adhoccasts/filesystem"
	"github.com/efarrer/adhoccasts/filetypes"
)

const (
	timeFormat = "Mon Jan 2 15:04:05 -0700 MST 2006"
)

type FileLister interface {
	ListFiles(string) ([]filesystem.File, error)
	IsDirectory(string) (bool, error)
}

type Rsser struct {
	rootUrl string
	rootDir string
	fs      FileLister
}

func NewRsser(rootUrl string, rootDir string, fs FileLister) Rsser {
	return Rsser{
		rootUrl: rootUrl,
		rootDir: rootDir,
		fs:      fs,
	}
}

type Rss struct {
	XMLName xml.Name `xml:"rss"`
	Version string   `xml:"version,attr"`
	Channel Channel  `xml:"channel"`
}

type Channel struct {
	Title         string `xml:"title"`
	Link          string `xml:"link"`
	Image         string `xml:"itunes:image"`
	Description   string `xml:"description"`
	LastBuildDate string `xml:"lastBuildDate"`
	Items         []Item `xml:"item"`
}

type Item struct {
	Title       string    `xml:"title"`
	Description string    `xml:"description"`
	PubDate     string    `xml:"pubDate"`
	Enclosure   Enclosure `xml:"enclosure"`
	Guid        Guid      `xml:"guid"`
}

type Enclosure struct {
	Url    string `xml:"url,attr"`
	Length int64  `xml:"length,attr"`
	Type   string `xml:"type,attr"`
}

type Guid struct {
	IsPermaLink bool   `xml:"isPermaLink,attr"`
	Value       string `xml:"guid"`
}

func (rsser Rsser) Render(podcastDir string) (Rss, error) {

	fullDir := path.Join(rsser.rootDir, podcastDir)
	files, err := rsser.fs.ListFiles(fullDir)
	if err != nil {
		return Rss{}, fmt.Errorf("list files for %s, %w", fullDir, err)
	}

	// Find a title image
	image := ""
	artwork := filetypes.FilterArtwork(files)
	if len(artwork) > 0 {
		parsedUrl, err := url.Parse(rsser.rootUrl)
		if err != nil {
			return Rss{}, fmt.Errorf("parsing url for %s, %w", podcastDir, err)
		}
		parsedUrl.Path = "/" + path.Base(podcastDir) + "/" + artwork[0].Name
		image = parsedUrl.String()
	}

	rss := Rss{
		XMLName: xml.Name{"", ""},
		Version: "2.0",
		Channel: Channel{
			Title:         dirToTitle(podcastDir),
			Link:          rsser.rootUrl,
			Image:         image,
			Description:   DirToDescription(podcastDir),
			LastBuildDate: time.Now().Format(timeFormat),
			Items:         nil,
		},
	}

	// Filter out non-supported episode types
	episodes := filetypes.FilterEpisodes(files)

	items := make([]Item, 0, len(episodes))
	for _, file := range episodes {
		title := nameToTitle(file.Name)
		parsedUrl, err := url.Parse(rsser.rootUrl)
		if err != nil {
			return rss, fmt.Errorf("parsing url for %s, %w", podcastDir, err)
		}
		parsedUrl.Path = "/" + path.Base(podcastDir) + "/" + file.Name
		item := Item{
			Title:       title,
			Description: title,
			PubDate:     file.ModTime.Format(time.RFC1123Z),
			Enclosure: Enclosure{
				Url:    parsedUrl.String(),
				Length: file.Size,
				Type:   contentType(file.Name),
			},
			Guid: Guid{
				IsPermaLink: true,
				Value:       parsedUrl.String(),
			},
		}
		items = append(items, item)
	}
	rss.Channel.Items = items

	return rss, nil
}

func (rsser Rsser) Wrap(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		basePath := path.Base(r.URL.Path)
		// Don't generate a rss feed for the root directory
		if basePath == "/" {
			handler.ServeHTTP(w, r)
			return
		}
		extension := filepath.Ext(basePath)
		// Only generate rss feed for .xml paths
		if extension != ".xml" {
			handler.ServeHTTP(w, r)
			return
		}

		dirPath := strings.TrimSuffix(basePath, filepath.Ext(basePath))
		fullDir := path.Join(rsser.rootDir, dirPath)

		isDir, err := rsser.fs.IsDirectory(fullDir)
		if err != nil {
			httpError(w, http.StatusInternalServerError, "Invalid podcast", err)
			return
		}
		if isDir {
			rss, err := rsser.Render(dirPath)
			if err != nil {
				httpError(w, http.StatusInternalServerError, "Unable to render podcast", err)
				return
			}
			encoder := xml.NewEncoder(w)
			encoder.Indent("", "  ")
			err = encoder.Encode(rss)
			if err != nil {
				httpError(w, http.StatusInternalServerError, "Unable to encode podcast", err)
				return
			}
			return
		}

		handler.ServeHTTP(w, r)
	})
}

func DirToDescription(dir string) string {
	title__description := strings.Split(filepath.Base(dir), "__")
	description := ""
	if len(title__description) == 1 {
		return dir
	}
	description = title__description[1]
	return strings.Replace(description, "_", " ", -1)
}

func nameToTitle(name string) string {
	title := strings.TrimSuffix(name, filepath.Ext(name))
	return strings.Replace(title, "_", " ", -1)
}

func contentType(name string) string {
	switch path.Ext(name) {
	case ".mp3":
		return "audio/mpeg"
	case ".mp4":
		return "video/mp4"
	default:
		return "audio/mpeg"
	}
}
