package minetest

import (
	"github.com/anon55555/mt"

	"fmt"
	"strings"
	"sync"
)

var itemDefsMu sync.RWMutex
var itemDefs []mt.ItemDef

type nodeDefNotFoundErr struct {
	name string
}

func (ndnfe nodeDefNotFoundErr) Error() string {
	return fmt.Sprintf("Node definition not Found: '%s'", ndnfe.name)
}

var (
	nodeDefsMu sync.RWMutex
	nodeDefs   []*mt.NodeDef

	nodeDefsMapMu sync.RWMutex
	nodeDefsMap   = make(map[string]*mt.NodeDef)
	nodeDefID     mt.Content
)

var aliasesMu sync.RWMutex
var aliases []struct{ Alias, Orig string }

var mediaURLs []string
var media []struct{ Name, Base64SHA1 string }
var mediaMu sync.RWMutex

var NodeIdMap map[mt.Content]string
var IdNodeMap map[string]mt.Content

// Add more item definitions to pool
// pls only use while init func
func AddItemDef(defs ...mt.ItemDef) {
	itemDefsMu.Lock()
	defer itemDefsMu.Unlock()

	itemDefs = append(itemDefs, defs...)
}

// Add more item definitions to pool
// pls only use while init func
// Param0 field is overwritten
func AddNodeDef(defs ...*mt.NodeDef) {
	nodeDefsMu.Lock()
	nodeDefsMapMu.Lock()
	defer nodeDefsMu.Unlock()
	defer nodeDefsMapMu.Unlock()

	// add id
	for k := range defs {
		defs[k].Param0 = getNodeDefID()
		nodeDefsMap[defs[k].Name] = defs[k]
	}

	nodeDefs = append(nodeDefs, defs...)
}

func getNodeDefID() mt.Content {
	if nodeDefID == 125 {
		nodeDefID += 3

		// insert meta defs:
		// unknown air and ignore
		nodeDefs = append(nodeDefs,
			&mt.NodeDef{Param0: mt.Unknown, Name: "unknown"},
			&mt.NodeDef{Param0: mt.Air, Name: "air"},
			&mt.NodeDef{Param0: mt.Ignore, Name: "ignore"},
		)
	}

	id := nodeDefID
	nodeDefID++

	return id
}

// GetNodeDef returns pointer to node def if registerd
// otherwise nil
func GetNodeDef(name string) (def *mt.NodeDef) {
	nodeDefsMapMu.Lock()
	defer nodeDefsMapMu.Unlock()

	def, found := nodeDefsMap[name]
	if !found {
		return nil
	}

	return def
}

// GetNodeID returns the Param0 of a node
// panics if not found
func GetNodeID(name string) mt.Content {
	def := GetNodeDef(name)

	if def == nil {
		panic(nodeDefNotFoundErr{name})
	}

	return def.Param0
}

func FillNameIdMap() {
	NodeIdMap = make(map[mt.Content]string)
	IdNodeMap = make(map[string]mt.Content)

	nodeDefsMapMu.RLock()
	defer nodeDefsMapMu.RUnlock()

	for k, v := range nodeDefsMap {
		IdNodeMap[k] = v.Param0
		NodeIdMap[v.Param0] = k
	}
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
	cmd := &mt.ToCltNodeDefs{
		Defs: nodeDefReferenced(),
	}

	c.SendCmd(cmd)
}

func nodeDefReferenced() (s []mt.NodeDef) {
	nodeDefsMu.RLock()
	defer nodeDefsMu.RUnlock()

	for _, v := range nodeDefs {
		s = append(s, *v)
	}

	return
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
