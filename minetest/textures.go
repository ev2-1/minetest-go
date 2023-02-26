package minetest

import (
	"github.com/anon55555/mt"

	"image/color"
	"strings"
)

type ItemTextures struct {
	InvImg     Texture
	InvOverlay Texture

	WieldImg     Texture
	WieldOverlay Texture

	Palette Texture // TODO: why? and what does?
	Color   color.NRGBA

	WieldScale [3]float32
}

func mtTexture(t Texture) mt.Texture {
	return mt.Texture(t.Texture())
}

type Texture interface {
	// should return string representation of texture
	Texture() string
}

type StrTexture mt.Texture

func (str StrTexture) Texture() string {
	return string(str)
}

type TextureOverlay struct {
	Textures []Texture
}

func (o TextureOverlay) Texture() string {
	return strings.Join(TextureS2StringS(o.Textures), "^")
}

// Converts a Texture slice to a string slice
func TextureS2StringS(s []Texture) (r []string) {
	r = make([]string, len(s))

	for k := range r {
		r[k] = s[k].Texture()
	}

	return
}

func TextureStr(t Texture) mt.Texture {
	if t == nil {
		return ""
	} else {
		return mt.Texture(t.Texture())
	}
}

// Converts Color to NRGBA
func NRGBA(c color.Color) color.NRGBA {
	r, g, b, a := c.RGBA()

	return color.NRGBA{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), uint8(a >> 8)}
}
