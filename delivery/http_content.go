package delivery

import (
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"time"

	"github.com/anacrolix/torrent"
)

func serveTorrentFile(w http.ResponseWriter, r *http.Request, t *torrent.Torrent, path string) error {
	var name string
	var file *torrent.File

	var basename = filepath.Base(path)
	var torname = t.Info().Name

	path = torname + "/" + path

	if !t.Info().IsDir() && basename == torname {
		file = t.Files()[0]
		name = t.Name()
	} else {
		// Search for file.
		for _, f := range t.Files() {
			var p = f.Path()
			if p == path {
				file = f
				break
			}
		}

		if file == nil {
			return fmt.Errorf("torrent content file not found")
		}

		var fp = file.FileInfo().Path
		name = fp[len(fp)-1]
	}

	var reader = file.NewReader()
	defer reader.Close()

	return serveContent(w, r, reader, name, (file.Length()*10)/100)
}

func serveContent(w http.ResponseWriter, r *http.Request, reader torrent.Reader, name string, readahead int64) error {
	var err error

	// Don't wait for pieces to complete and be verified.
	//reader.SetResponsive()

	if readahead > 0 {
		// Read ahead 10% of file.
		reader.SetReadahead(readahead)
	}

	w.Header().Set("Content-Disposition", `filename="`+url.PathEscape(name)+`"`)

	_, err = reader.Seek(0, 0)
	if err != nil {
		return err
	}

	http.ServeContent(w, r, "", time.Unix(0, 0), reader)
	return nil
}
