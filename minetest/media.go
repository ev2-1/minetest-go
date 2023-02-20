package minetest

import (
	"github.com/anon55555/mt"

	"strings"
	"sync"
)

type Alias struct{ Alias, Orig string }

var aliasesMu sync.RWMutex
var aliases map[string]string

var mediaURLs []string
var media []struct{ Name, Base64SHA1 string }
var mediaMu sync.RWMutex

// Add a Alias to the pool
// pls only use while init func
func AddAlias(alias ...Alias) {
	aliasesMu.Lock()
	defer aliasesMu.Unlock()

	for _, a := range alias {
		aliases[a.Alias] = a.Orig
	}
}

// Add a file to the media pool
// pls only use while init func
func AddMedia(m ...struct{ Name, Base64SHA1 string }) {
	mediaMu.Lock()
	defer mediaMu.Unlock()

	media = append(media, m...)
}

// Add a file to the mediaURL
// pls only use while init func
func AddMediaURL(url ...string) {
	mediaMu.Lock()
	defer mediaMu.Unlock()

	mediaURLs = append(mediaURLs, url...)
}

// Send (cached) AnnounceMedia to client
func (c *Client) SendAnnounceMedia() {
	mediaMu.RLock()
	cmd := &mt.ToCltAnnounceMedia{
		Files: media,
		URL:   strings.Join(mediaURLs, ","),
	}
	mediaMu.RUnlock()

	c.SendCmd(cmd)
}
