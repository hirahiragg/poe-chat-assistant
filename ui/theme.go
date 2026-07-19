package ui

import (
	"image/color"

	"gioui.org/widget/material"
)

var (
	colorBg         = color.NRGBA{R: 0x1a, G: 0x1a, B: 0x20, A: 0xff}
	colorSurface    = color.NRGBA{R: 0x22, G: 0x22, B: 0x2a, A: 0xff}
	colorCard       = color.NRGBA{R: 0x2a, G: 0x2a, B: 0x34, A: 0xff}
	colorBorder     = color.NRGBA{R: 0x3a, G: 0x3a, B: 0x44, A: 0xff}
	colorSelected   = color.NRGBA{R: 0x30, G: 0x40, B: 0x50, A: 0xff}
	colorText       = color.NRGBA{R: 0xcc, G: 0xc8, B: 0xc0, A: 0xff}
	colorTextDim    = color.NRGBA{R: 0x88, G: 0x85, B: 0x80, A: 0xff}
	colorAccent     = color.NRGBA{R: 0x6e, G: 0xb8, B: 0xd4, A: 0xff}
	colorWhisper    = color.NRGBA{R: 0xc9, G: 0xa8, B: 0x4c, A: 0xff}
	colorTrade      = color.NRGBA{R: 0x5a, G: 0x9e, B: 0x6f, A: 0xff}
	colorParty      = color.NRGBA{R: 0x70, G: 0x70, B: 0xff, A: 0xff}
	colorGlobal     = color.NRGBA{R: 0xcc, G: 0xc8, B: 0xc0, A: 0xff}
	colorBtnBg      = color.NRGBA{R: 0x3a, G: 0x5a, B: 0x6a, A: 0xff}
	colorBtnText    = color.NRGBA{R: 0xe0, G: 0xe0, B: 0xe0, A: 0xff}
	colorTranslated = color.NRGBA{R: 0xa0, G: 0xd0, B: 0xa0, A: 0xff}
)

func newTheme() *material.Theme {
	th := material.NewTheme()
	th.Palette.Bg = colorBg
	th.Palette.Fg = colorText
	th.Palette.ContrastBg = colorAccent
	th.Palette.ContrastFg = colorBtnText
	return th
}
