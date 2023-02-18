package hud

import (
	"github.com/anon55555/mt"
	"github.com/ev2-1/minetest-go/chat"
	"github.com/ev2-1/minetest-go/minetest"

	"bufio"
	_ "embed"
	"fmt"
	"strings"
	"time"
)

//go:embed characters.txt
var chars string

var charmap map[byte]string = func() map[byte]string {
	m := make(map[byte]string)

	r := strings.NewReader(chars)
	s := bufio.NewScanner(r)

	var k byte
	var step uint8
	for s.Scan() {
		switch step {
		case 0:
			k = s.Text()[0]
		case 1:
			_, ok := m[k]
			if ok {
				continue
			}

			m[k] = s.Text()
		case 2:
		}
		s.Text()

		step++
		if step == 3 {
			step = 0
		}
	}

	return m
}()

const LINE_HEIGHT = 14
const CHAR_WIDTH = 5 + 1

func Text(text string) (str string) {
	str = fmt.Sprintf("[combine:%dx%d:", len(text)*CHAR_WIDTH, LINE_HEIGHT)

	for i := 0; i < len(text); i++ {
		char := text[i]

		m, ok := charmap[char]
		if ok {
			str += fmt.Sprintf("%d,%d=%s.png:", i*CHAR_WIDTH, 0, m)
		}
	}

	return str[:len(str)-1]
}

// results in a simmilar filter to what minecraft uses
const darken = "^[invert:rgb^[brighten^[brighten^[invert:rgb"

func CombineMatrix(maxx, maxy int, name string) (str string) {
	str = fmt.Sprintf("[combine:%dx%d:", maxx*16, maxy*16)

	for x := 0; x < maxx; x++ {
		for y := 0; y < maxy; y++ {
			str += fmt.Sprintf("%d,%d=%s:", x*16, y*16, name)
		}
	}

	return str[:len(str)-1] + darken
}

func init() {
	w := 16
	h := 10

	text := make([]string, 4)
	for k := range text {
		text[k] = Text("now loading"+strings.Repeat(".", k)) + "^[brighten"
	}

	matrix := CombineMatrix(w, h, "mc_dirt.png")

	//matrix := Text("I.wrote.Text")
	fmt.Printf("matrix: %s\n", matrix)

	chat.RegisterChatCmd("addhud", func(c *minetest.Client, _ []string) {
		stopCh := make(chan struct{})

		cd, ok := c.GetData("loading_ch")
		if ok && cd != nil {
			chat.SendMsg(c, "Already loading!", mt.NormalMsg)
			return
		}

		c.SetData("loading_ch", stopCh)

		c.SendCmd(&mt.ToCltAddHUD{
			ID: 0,
			HUD: mt.HUD{
				Type: mt.ImgHUD,

				Pos:   [2]float32{0.5, 0.5},
				Scale: [2]float32{10, 10},

				Text: matrix, // is also resource name lol
			},
		})
		c.SendCmd(&mt.ToCltAddHUD{
			ID: 1,
			HUD: mt.HUD{
				Type: mt.ImgHUD,

				Pos:   [2]float32{0.5, 0.25},
				Scale: [2]float32{5, 5},

				Text: text[0], // is also resource name lol
			},
		})

		go func() {
			var state = 0
			var ticker = time.NewTicker(time.Second)

			for {
				select {
				case <-stopCh:
					ticker.Stop()
					c.SendCmd(&mt.ToCltRmHUD{
						ID: 1,
					})

					return
				case <-ticker.C:
					c.SendCmd(&mt.ToCltChangeHUD{
						ID:    1,
						Field: mt.HUDText,
						Text:  text[state], // is also resource name lol
					})
				}

				state++
				if state >= len(text) {
					state = 0
				}
			}
		}()
	})

	chat.RegisterChatCmd("rmhud", func(c *minetest.Client, _ []string) {
		cd, ok := c.GetData("loading_ch")
		if !ok || cd == nil {
			chat.SendMsg(c, "Not loading!", mt.NormalMsg)
			return
		}

		close(cd.(chan struct{}))
		c.SetData("loading_ch", nil)

		c.SendCmd(&mt.ToCltRmHUD{
			ID: 0,
		})
	})

}
