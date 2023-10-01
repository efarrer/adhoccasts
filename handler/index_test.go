package handler_test

import (
	"testing"

	"github.com/efarrer/adhoccasts/handler"
	"github.com/stretchr/testify/require"
)

type dirNamer []string

func (dn dirNamer) DirectoryNames(dir string) ([]string, error) {
	return []string(dn), nil
}

func TestRender_HandlesWeirdDirNames(t *testing.T) {
	indexer := handler.NewIndexer(
		"http://localhost/",
		"/some/root/",
		dirNamer([]string{"foo/", "This Directory Has Spaces/"}),
	)
	index, err := indexer.Render()

	expectedIndex := handler.Index{
		NameToPath: map[string]string{
			"foo":                       "http://localhost/foo.xml",
			"This Directory Has Spaces": "http://localhost/This%20Directory%20Has%20Spaces.xml",
		},
	}
	require.NoError(t, err)
	require.Equal(t, expectedIndex, index)
}
