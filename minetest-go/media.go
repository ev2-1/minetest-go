package minetest

import (
	"github.com/anon55555/mt"

	"strings"
	"sync"
)

var itemDefsMu sync.RWMutex
var itemDefs []mt.ItemDef

var nodeDefsMu sync.RWMutex
var nodeDefs []mt.NodeDef

var aliasesMu sync.RWMutex
var aliases []struct{ Alias, Orig string }

var mediaURLs []string
var media []struct{ Name, Base64SHA1 string }
var mediaMu sync.RWMutex

// Add more item definitions to pool
// pls only use while init func
func AddItemDef(defs ...mt.ItemDef) {
	itemDefsMu.Lock()
	defer itemDefsMu.Unlock()

	itemDefs = append(itemDefs, defs...)
}

// Add more item definitions to pool
// pls only use while init func
func AddNodeDef(defs ...mt.NodeDef) {
	nodeDefsMu.Lock()
	defer nodeDefsMu.Unlock()

	nodeDefs = append(nodeDefs, defs...)
}

// Add a Alias to the pool
// pls only use while init func
func AddAlias(alias ...struct{ Alias, Orig string }) {
	aliasesMu.Lock()
	defer aliasesMu.Unlock()

	aliases = append(aliases, alias...)
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

// Send (cached) ItemDefinitions to client
func (c *Client) SendItemDefs() {
	itemDefsMu.RLock()
	cmd := &mt.ToCltItemDefs{
		Defs:    itemDefs,
		Aliases: aliases,
	}
	itemDefsMu.RUnlock()

	c.SendCmd(cmd)
}

// Send (cached) NodeDefinitions to client
func (c *Client) SendNodeDefs() {
	nodeDefsMu.RLock()
	cmd := &mt.ToCltNodeDefs{
		Defs: nodeDefs,
	}
	nodeDefsMu.RUnlock()

	c.SendCmd(cmd)
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
