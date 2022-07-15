// this file contains serialization and deserializing methods
// for minetest mabbulkdata blobs
// this will probably be moved to its own library once finished

package main

import (
	"github.com/anon55555/mt"

	"encoding/binary"
	"io"
	"log"
	"time"
)

type MapBlock struct {
	Underground, NightDayDiff, Generated bool

	Param0 [4096]mt.Content
	Param1 [4096]uint8
	Param2 [4096]uint8
}

type Writer struct {
	io.Writer
}

func (w Writer) writeU8(d uint8) {
	w.Write([]byte{byte(d)})
}

func (w Writer) writeByte(d byte) {
	w.Write([]byte{d})
}

func (w Writer) writeU16(d uint16) {
	buff := make([]byte, 2)
	binary.LittleEndian.PutUint16(buff, d)

	w.Write(buff)
}

func (w Writer) writeU32(d uint32) {
	buff := make([]byte, 4)
	binary.LittleEndian.PutUint32(buff, d)

	w.Write(buff)
}

func (w Writer) writeString(str string) {
	w.Write([]byte(str))
}

func (blk MapBlock) Serialize(ver uint8, compression int, ndef func(id mt.Content) *mt.NodeDef, w Writer) {
	// flags:
	var flags byte
	if blk.Underground {
		flags |= 0x01
	}
	if blk.NightDayDiff {
		flags |= 0x02
	}
	if blk.Generated {
		flags |= 0x03
	}

	w.writeByte(flags)

	if ver >= 27 { // write dummy m_lighting_complete
		w.writeU16(0x0000)
	}

	// Blk node data
	tmp_nodes := blk

	mapping, nimap := blk.getMapping(ndef, &blk)

	buf := blk.serializeBulk(ver, tmp_nodes)

	w.writeU8(2)
	w.writeU8(2)

	if ver >= 29 {
		w.writeU32(uint32(time.Now().Unix()))

		nimap.serialize(w)
	} else {
		compress(buf[:], w, ver, compression)
	}

	// Node Meta // TODO:
	if ver >= 29 {
		m_node_metadata.serialize(os, version, disk)
	} else {
		// use os_raw from above to avoid allocating another stream object
		m_node_metadata.serialize(os_raw, version, disk)
		// prior to 29 node data was compressed individually
		compress(os_raw.str(), os, version, compression_level)
	}

	if ver <= 24 {
		// Node timers
		m_node_timers.serialize(os, version)
	}

	// Static objects
	m_static_objects.serialize(os)

	if ver < 29 {
		// Timestamp
		writeU32(os, getTimestamp())

		// Write block-specific node definition id mapping
		nimap.serialize(os)
	}

	if ver >= 25 {
		// Node timers
		m_node_timers.serialize(os, version)
	}

	if ver >= 29 {
		// now compress the whole thing
		compress(os_raw.str(), os_compressed, version, compression_level)
	}
}

func (blk MapBlock) serializeBulk(ver uint8, nodes MapBlock) (buff [4096 * (2 + 2)]byte) {
	// Can't do this anymore; we have 16-bit dynamically allocated node IDs
	// in memory; conversion just won't work in this direction.
	if ver < 24 {
		log.Println("cant serialize Bulk to ver < 24")
	}

	start1 := 2 * 4096
	start2 := (2 + 1) * 4096

	// Serialize content
	for i := 0; i < 4096; i++ {
		// writeU16(&databuf[i * 2], nodes[i].param0)
		buff[i*2] = uint8(nodes.Param0[i] >> 8)
		buff[i*2+1] = uint8(nodes.Param0[i])

		// writeU8(&databuf[start1+i], nodes[i].param1)
		buff[start1+i] = nodes.Param1[i]

		// writeU8(&databuf[start2+i], nodes[i].param2)
		buff[start2+i] = nodes.Param2[i]
	}

	return buff
}

type mapping map[mt.Content]uint16
type nimap map[uint16]string

func (nm nimap) serialize(w Writer) {
	w.writeU8(0)
	w.writeU16(uint16(len(nm)))

	for k, v := range nm {
		w.writeU16(k)
		w.writeString(v)
	}
}

func (blk MapBlock) getMapping(ndef func(id mt.Content) *mt.NodeDef, tmp_nodes *MapBlock) (m mapping, nm nimap) {
	var id_counter uint16 = 0

	for i := uint16(0); i < 16*16*16; i++ {
		global_id := blk.Param0[i]
		id := uint16(mt.Ignore)

		// Try to find an existing mapping
		if _, ok := m[global_id]; ok {
			id = m[global_id]
		} else {
			// We have to assign a new mapping
			id = id_counter
			id_counter++

			m[global_id] = id

			def := ndef(global_id)
			if def.Name != "" {
				nm[id] = def.Name
			}
		}

		// Update the MapNode
		tmp_nodes.Param0[i] = mt.Content(id)
	}

	return m, nm
}
