package minetest

import (
	"github.com/anon55555/mt"

	"fmt"
	"log"
	"strings"
	"sync"
)

var (
	itemDefsMu sync.RWMutex
	itemDefs   = map[string]mt.ItemDef{}
)

type nodeDefNotFoundErr struct {
	name string
}

func (ndnfe nodeDefNotFoundErr) Error() string {
	return fmt.Sprintf("Node definition not Found: '%s'", ndnfe.name)
}

var (
	nodeDefsMu sync.RWMutex
	nodeDefs   = map[string]*mt.NodeDef{}
	nodeDefID  mt.Content // counts up for each node
)

type Alias struct{ Alias, Orig string }

var aliasesMu sync.RWMutex
var aliases map[string]string

var mediaURLs []string
var media []struct{ Name, Base64SHA1 string }
var mediaMu sync.RWMutex

var NodeIdMap map[mt.Content]string
var IdNodeMap map[string]mt.Content

// Add more item definitions to pool
// will panic if serverstate is `StateRunning`
func AddItemDef(defs ...mt.ItemDef) {
	itemDefsMu.Lock()
	defer itemDefsMu.Unlock()

	for _, def := range defs {
		itemDefs[def.Name] = def
	}
}

// Add more item definitions to pool
// pls only use while init func
// Param0 field is overwritten
func AddNodeDef(defs ...*mt.NodeDef) {
	nodeDefsMu.Lock()
	defer nodeDefsMu.Unlock()

	// add id
	for k := range defs {
		defs[k].Param0 = getNodeDefID()
		nodeDefs[defs[k].Name] = defs[k]
	}
}

func getNodeDefID() mt.Content {
	if nodeDefID == 125 {
		nodeDefID += 3

		// insert meta defs:
		// unknown air and ignore
		nodeDefs["unknown"] = &mt.NodeDef{Param0: mt.Unknown, Name: "unknown"}
		nodeDefs["air"] = &mt.NodeDef{Param0: mt.Air, Name: "air"}
		nodeDefs["ignore"] = &mt.NodeDef{Param0: mt.Ignore, Name: "ignore"}
	}

	id := nodeDefID
	nodeDefID++

	return id
}

// RealNodeName returns the underlying name of a given node
// checks aliases map if !ok returns name
func RealNodeName(name string) string {
	aliasesMu.RLock()
	defer aliasesMu.RUnlock()

	rname, ok := aliases[name]
	if !ok {
		return name
	}

	return rname
}

// GetNodeDef returns pointer to node def if registerd
// otherwise nil
func GetNodeDef(name string) (def *mt.NodeDef) {
	// check alias
	name = RealNodeName(name)

	nodeDefsMu.Lock()
	defer nodeDefsMu.Unlock()

	def, found := nodeDefs[name]
	if !found {
		return nil
	}

	return def
}

// GetItemDef returns pointer to ItemDef if registerd
func GetItemDef(name string) (def *mt.ItemDef) {
	itemDefsMu.Lock()
	defer itemDefsMu.Unlock()

	d, found := itemDefs[name]
	if !found {
		return nil
	}

	return &d
}

// GetNodeID returns the Param0 of a node
// panics if not found
func GetNodeID(name string) mt.Content {
	def := GetNodeDef(name)

	if def == nil {
		log.Printf("WARN: %s\n", nodeDefNotFoundErr{name})
		return mt.Unknown
	}

	return def.Param0
}

func FillNameIdMap() {
	NodeIdMap = make(map[mt.Content]string)
	IdNodeMap = make(map[string]mt.Content)

	nodeDefsMu.RLock()
	defer nodeDefsMu.RUnlock()

	for k, v := range nodeDefs {
		IdNodeMap[k] = v.Param0
		NodeIdMap[v.Param0] = k
	}
}

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

// Send (cached) ItemDefinitions to client
func (c *Client) SendItemDefs() (<-chan struct{}, error) {
	itemDefsMu.RLock()

	// add to slice
	defs := make([]mt.ItemDef, len(itemDefs))
	var i int
	for _, v := range itemDefs {
		defs[i] = v
		i++
	}

	itemDefsMu.RUnlock()

	// add aliases to slice
	aliasesMu.RLock()
	alias := make([]struct{ Alias, Orig string }, len(aliases))

	i = 0
	for k, v := range aliases {
		alias[i] = Alias{k, v}
		i++
	}

	aliasesMu.RUnlock()

	cmd := &mt.ToCltItemDefs{
		Defs:    defs,
		Aliases: alias,
	}

	return c.SendCmd(cmd)
}

// Send (cached) NodeDefinitions to client
func (c *Client) SendNodeDefs() (<-chan struct{}, error) {
	cmd := &mt.ToCltNodeDefs{
		Defs: nodeDefReferenced(),
	}

	return c.SendCmd(cmd)
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
