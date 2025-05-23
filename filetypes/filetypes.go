package filetypes

import (
	"path/filepath"

	"github.com/efarrer/adhoccasts/filesystem"
)

var episodeMediaExtensions map[string]struct{} = map[string]struct{}{
	".mp3":  {},
	".mp4":  {},
	".3gp":  {},
	".3g2":  {},
	".aiff": {},
	".amr":  {},
	".wav":  {},
	".m4a":  {},
	".m4b":  {},
	".m4p":  {},
	".mov":  {},
	".m4v":  {},
	".feed": {}, // This is a special case used for rss feeds only
}

var podcastTitleExtensions map[string]struct{} = map[string]struct{}{
	".png":  {},
	".jpeg": {},
	".jpg":  {},
}

func FilterArtwork(files []filesystem.File) []filesystem.File {
	filtered := make([]filesystem.File, 0, len(files))
	for _, file := range files {
		extension := filepath.Ext(file.Name)
		if _, ok := podcastTitleExtensions[extension]; ok {
			filtered = append(filtered, file)
		}
	}

	return filtered
}

func FilterEpisodes(files []filesystem.File) []filesystem.File {
	filtered := make([]filesystem.File, 0, len(files))
	for _, file := range files {
		extension := filepath.Ext(file.Name)
		if _, ok := episodeMediaExtensions[extension]; ok {
			filtered = append(filtered, file)
		}
	}

	return filtered
}
