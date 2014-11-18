package main

import (
	"image"
	"image/color"
	"image/draw"
	"io/ioutil"

	"code.google.com/p/freetype-go/freetype"
	"code.google.com/p/freetype-go/freetype/truetype"
)

type ImageText struct {
	fontfile string
	font     *truetype.Font
}

func (ti *ImageText) Init() error {
	fontBytes, err := ioutil.ReadFile(ti.fontfile)
	if err != nil {
		return err
	}
	ti.font, err = freetype.ParseFont(fontBytes)
	if err != nil {
		return err
	}
	return nil
}

func (ti *ImageText) AddText(img draw.Image, text string, fontsize int, loc image.Point, textColor color.Color) {
	c := freetype.NewContext()
	c.SetDPI(72)
	c.SetFont(ti.font)
	c.SetFontSize(float64(fontsize))
	c.SetClip(img.Bounds())
	c.SetDst(img)
	c.SetSrc(image.NewUniform(textColor))
	c.SetHinting(freetype.NoHinting)
	pt := freetype.Pt(loc.X, loc.Y)
	c.DrawString(text, pt)
}
