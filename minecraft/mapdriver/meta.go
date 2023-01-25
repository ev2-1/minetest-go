package minecraft_map

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"errors"
	"io"
)

type NodeDef struct {
	Image        string
	FriendlyName string
	Name         string
}

type Meta struct {
	Image        string `json:ImgSrc`
	FriendlyName string `json:FName`
	Name         string `json:TName`
	ID           uint32 `json:ID`
	Data         uint8  `json:Data`
}

type Identifier struct {
	ID   uint32
	Data uint8
}

//go:embed metastream.json
var eStream []byte

var IdBlockMap map[Identifier]string
var BlockIdMap map[string]Identifier
var Metas []Meta

func EnsureMap() {
	if IdBlockMap == nil || BlockIdMap == nil {
		IdBlockMap, BlockIdMap, Metas, _ = ParseEmbeddedStream()
	}
}

func ParseEmbeddedStream() (IdBlockMap map[Identifier]string, BlockIdMap map[string]Identifier, s []Meta, err error) {
	return ParseStream(bytes.NewReader(eStream))
}

func ParseStream(r io.Reader) (IdBlockMap map[Identifier]string, BlockIdMap map[string]Identifier, s []Meta, err error) {
	IdBlockMap = make(map[Identifier]string)
	BlockIdMap = make(map[string]Identifier)

	d := json.NewDecoder(r)

	var meta Meta
	for {
		err = d.Decode(&meta)
		if err != nil {
			if errors.Is(io.EOF, err) {
				err = nil
			}
			return
		}

		i := Identifier{meta.ID, meta.Data}

		IdBlockMap[i] = meta.Name
		BlockIdMap[meta.Name] = i

		s = append(s, meta)
	}
}
