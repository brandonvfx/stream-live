package main

import (
	"image"
	"image/color"
	"image/draw"
	"io/ioutil"

	"golang.org/x/image/font"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
)

// ImageText stores a truetype.Font and allows you 
// to add text to an image.
type ImageText struct {
	fontfile string
	font     *truetype.Font
}

// Init loads the font set as fontfile
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

// AddText to an image at a point 
func (ti *ImageText) AddText(img draw.Image, text string, fontsize float64, loc image.Point, textColor color.Color) {
	c := freetype.NewContext()
	c.SetDPI(72)
	c.SetFont(ti.font)
	c.SetFontSize(fontsize)
	c.SetClip(img.Bounds())
	c.SetDst(img)
	c.SetSrc(image.NewUniform(textColor))
	c.SetHinting(font.HintingNone)
	pt := freetype.Pt(loc.X, loc.Y)
	c.DrawString(text, pt)
}
