package handler_test

import (
	"encoding/xml"
	"testing"
	"time"

	"github.com/efarrer/adhoccasts/filesystem"
	"github.com/efarrer/adhoccasts/handler"
	"github.com/stretchr/testify/require"
)

type fileLister []filesystem.File

func (fl fileLister) ListFiles(string) ([]filesystem.File, error) {
	return []filesystem.File(fl), nil
}

func (fl fileLister) IsDirectory(string) (bool, error) {
	return true, nil
}

func TestRsser_HandlesFunkyNames(t *testing.T) {
	rsser := handler.NewRsser(
		"http://localhost/",
		"/some/root/",
		fileLister([]filesystem.File{
			{
				Name:    "foo.mp3",
				ModTime: time.Time{},
				Size:    1234,
			}, {
				Name:    "this file has spaces.mp3",
				ModTime: time.Time{},
				Size:    5678,
			},
		}),
	)

	rss, err := rsser.Render("this dir has spaces")
	require.NoError(t, err)

	// Clear this out so we aren't comparing time.Now()
	rss.Channel.LastBuildDate = ""

	expectedRss := handler.Rss{
		XMLName: xml.Name{"", ""},
		Version: "2.0",
		Channel: handler.Channel{
			Title: "this dir has spaces",
			Link:  "http://localhost/",
			Items: []handler.Item{
				{
					Title:       "foo",
					Description: "foo",
					PubDate:     "Mon, 01 Jan 0001 00:00:00 +0000",
					Enclosure: handler.Enclosure{
						Url:    "http://localhost/this%20dir%20has%20spaces/foo.mp3",
						Length: 1234,
						Type:   "audio/mpeg",
					},
					Guid: handler.Guid{
						IsPermaLink: true,
						Value:       "http://localhost/this%20dir%20has%20spaces/foo.mp3",
					},
				},
				{
					Title:       "this file has spaces",
					Description: "this file has spaces",
					PubDate:     "Mon, 01 Jan 0001 00:00:00 +0000",
					Enclosure: handler.Enclosure{
						Url:    "http://localhost/this%20dir%20has%20spaces/this%20file%20has%20spaces.mp3",
						Length: 5678,
						Type:   "audio/mpeg",
					},
					Guid: handler.Guid{
						IsPermaLink: true,
						Value:       "http://localhost/this%20dir%20has%20spaces/this%20file%20has%20spaces.mp3",
					},
				},
			},
		},
	}

	require.Equal(t, expectedRss, rss)
}
