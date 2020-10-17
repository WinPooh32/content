package model

import (
	"net/http"

	"github.com/anacrolix/torrent"
)

type Header struct {
	Name string `json:"name"`
	Hash string `json:"hash"`
}

type TorrentFile struct {
	Name string   `json:"name"`
	Path []string `json:"path"`
	Size int64    `json:"size"`
}

type FilesList struct {
	Header  Header        `json:"header"`
	Content []TorrentFile `json:"content"`
}

func (s *FilesList) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func NewFilesList(t *torrent.Torrent) *FilesList {
	var files []TorrentFile

	if t.Info().IsDir() {
		files = make([]TorrentFile, len(t.Files()))

		for i, fi := range t.Files() {
			var name string
			var path []string = fi.FileInfo().Path

			if len(path) != 0 {
				name = path[len(path)-1]
			}

			files[i] = TorrentFile{
				Name: name,
				Path: path,
				Size: fi.Length(),
			}
		}
	} else {
		files = []TorrentFile{
			{
				Name: t.Name(),
				Path: []string{t.Name()},
				Size: t.Length(),
			},
		}
	}

	return &FilesList{
		Header: Header{
			Name: t.Name(),
			Hash: t.InfoHash().String(),
		},
		Content: files,
	}
}
