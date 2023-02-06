package minecraft_map

import (
	"github.com/anon55555/mt"
	"github.com/ev2-1/minetest-go/minetest"
	mc "vimagination.zapto.org/minecraft"

	"sync"
)

type MinecraftMapDriver struct {
	sync.RWMutex

	*mc.Level
}

func (drv *MinecraftMapDriver) Open(file string) (err error) {
	drv.Lock()
	defer drv.Unlock()

	path, err := mc.NewFilePath(file)
	if err != nil {
		return
	}

	drv.Level, err = mc.NewLevel(path)
	if err != nil {
		return
	}

	return nil
}

func (drv *MinecraftMapDriver) GetBlk(pos [3]int16) (blk minetest.DriverMapBlk, err error) {
	// convert to normal pos:
	pos = mt.Blkpos2Pos(pos, 0)

	// Fill chunk
	var Chunk = [16][16][16]mt.Node{}

	for x := int16(0); x < 16; x++ {
		for y := int16(0); y < 16; y++ {
			for z := int16(0); z < 16; z++ {
				block, err := drv.GetBlock(int32(pos[0]+x), int32(pos[1]+y), int32(pos[2]+z))
				if err != nil {
					return nil, err
				}

				Chunk[x][y][z] = mt.Node{Param0: MapContent(block.ID, block.Data)}
			}
		}
	}

	return nil, nil
}

func MapContent(id uint16, data uint8) mt.Content {

	return 0
}

func init() {
	minetest.RegisterMapDriver("minecraft", new(MinecraftMapDriver))
}
