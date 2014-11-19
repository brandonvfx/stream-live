package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	_ "image/jpeg"
	"image/png"
	"log"
	"math"
	"net/http"

	"github.com/dustin/go-humanize"
	"github.com/gorilla/mux"
	"github.com/nfnt/resize"

	"code.google.com/p/draw2d/draw2d"
)

func drawRoundedRect(dst draw.Image,
	width float64,
	height float64,
	fillColor color.Color,
	strokeColor color.Color,
	strokeWidth float64) {
	x, y := 0.0, 0.0
	aspect := 1.0                 /* aspect ratio */
	corner_radius := height / 7.5 /* and corner curvature radius */

	radius := corner_radius / aspect
	degrees := math.Pi / 180.0

	gcbg := draw2d.NewGraphicContext(dst)
	gcbg.SetStrokeColor(image.Black)
	gcbg.SetFillColor(image.White)
	gcbg.ArcTo(x+width-radius, y+radius, radius, radius, -90*degrees, 90*degrees)
	gcbg.ArcTo(x+width-radius, y+height-radius-1, radius, radius, 0*degrees, 90*degrees)
	gcbg.ArcTo(x+radius, y+height-radius-1, radius, radius, 90*degrees, 90*degrees)
	gcbg.ArcTo(x+radius, y+radius, radius, radius, 180*degrees, 90*degrees)
	gcbg.Close()
	gcbg.SetFillColor(fillColor)
	gcbg.SetStrokeColor(image.White)
	gcbg.SetLineWidth(1.0)
	gcbg.FillStroke()
}

func drawMask(dst draw.Image) {
	width := float64(dst.Bounds().Dx())
	height := float64(dst.Bounds().Dy())
	gc := draw2d.NewGraphicContext(dst)
	x, y := 0.0, 0.0
	aspect := 1.0                 /* aspect ratio */
	corner_radius := height / 7.5 /* and corner curvature radius */

	radius := corner_radius / aspect
	degrees := math.Pi / 180.0

	// Mask
	gc.SetStrokeColor(image.Black)
	gc.SetFillColor(image.White)
	gc.MoveTo(x+width, y)
	gc.LineTo(x+width, y)
	gc.LineTo(x+width, y+height)
	gc.ArcTo(x+radius, y+height-radius, radius, radius, 90*degrees, 90*degrees)
	gc.ArcTo(x+radius, y+radius, radius, radius, 180*degrees, 90*degrees)
	gc.Close()

	gc.SetFillColor(image.White)
	gc.SetStrokeColor(image.White)
	gc.SetLineWidth(0.0)
	gc.FillStroke()
}

func getPreviewImage(img_url string) (image.Image, error) {
	var tmp_img image.Image
	resp, err := http.Get(img_url)
	if err != nil {
		return tmp_img, err
	}

	preview_img, _, err := image.Decode(resp.Body)
	if err != nil {
		return tmp_img, err
	}
	return preview_img, nil
}

func LargeImageHandler(w http.ResponseWriter, req *http.Request) {
	// Fonts
	// text_arial := ImageText{
	// 	fontfile: "/Library/Fonts/Arial.ttf",
	// }
	// text_arial.Init()
	// text_arial_bold := ImageText{
	// 	fontfile: "/Library/Fonts/Arial Bold.ttf",
	// }
	// text_arial_bold.Init()
	// text_arial_black := ImageText{
	// 	fontfile: "/Library/Fonts/Arial Black.ttf",
	// }
	// text_arial_black.Init()

	// Settings
	width, height := 430.0, 64.0
	preview_width := 150
	preview_image_offset := 24

	var preview_image_url string

	// url vars
	vars := mux.Vars(req)
	stream := vars["stream"]
	log.Printf("LargeImageHandler - %s", stream)
	stream_info, err := GetStream(stream)
	if err != nil {
		log.Println(fmt.Sprintf("[ERROR] - %v", err))
		// fmt.Fprint(w, err)
		// return
	}

	if stream_info.Live {
		preview_image_offset = 0
		preview_image_url = stream_info.Stream.Preview["medium"]
	} else {
		preview_image_url = stream_info.Stream.Channel.Logo
	}

	// Get preview image
	var preview_src_image image.Image
	if preview_image_url != "" {
		preview_src_image, err = getPreviewImage(preview_image_url)
		if err != nil {
			log.Println(err)
			fmt.Fprint(w, err)
			return
		}
	} else {
		preview_src_image = missing_profile_image
	}
	// resize the preview
	preview_image := resize.Resize(0, uint(height-2.0), preview_src_image, resize.Lanczos3)

	// create the output image and it's background
	output_img := image.NewRGBA(image.Rect(0, 0, int(width), int(height)))
	drawRoundedRect(output_img,
		width,
		height,
		color.RGBA{0x64, 0x41, 0xA5, 0xFF},
		image.White,
		1.0)

	// Create mask for preview image
	mask_img := image.NewRGBA(image.Rect(0, 0, int(preview_width), int(height)))
	drawMask(mask_img)

	// Comp the preview image over the bg using the mask.
	draw.DrawMask(output_img,
		output_img.Bounds(),
		preview_image,
		image.Point{-preview_image_offset, 0},
		mask_img,
		image.Point{0, 0},
		draw.Over)

	// Redraw the border.
	drawRoundedRect(output_img,
		width,
		height,
		image.Transparent,
		image.White,
		1.0)

	// Draw divider
	gc_line := draw2d.NewGraphicContext(output_img)
	gc_line.SetStrokeColor(image.White)
	gc_line.SetFillColor(image.Transparent)
	gc_line.MoveTo(110, 0)
	gc_line.LineTo(110, 0)
	gc_line.LineTo(110, height)
	gc_line.SetLineWidth(1.5)
	gc_line.FillStroke()

	left_side := 115
	if stream_info.Live {
		red := color.RGBA{0xDF, 0x2D, 0x28, 0xFF}
		text_bold.AddText(output_img, stream_info.Name, 16, image.Point{left_side + 37, 18}, color.White)
		text_bold.AddText(output_img, "[LIVE]", 12, image.Point{left_side, 16}, red)
		// max length 52 chars
		text_regular.AddText(output_img, TruncString(stream_info.Stream.Channel.Status, 52), 11, image.Point{left_side, 32}, color.White)
		// max length 45 chars
		text_regular.AddText(output_img, "Playing - "+TruncString(stream_info.Stream.Game, 45), 11, image.Point{left_side, 45}, color.White)
		text_regular.AddText(output_img, humanize.Comma(int64(stream_info.Stream.Viewers))+" Viewers", 11, image.Point{left_side, 58}, color.White)
	} else {
		text_bold.AddText(output_img, stream_info.Name, 16, image.Point{left_side + 52, 18}, color.White)
		text_bold.AddText(output_img, "[Offline]", 12, image.Point{left_side, 16}, color.White)
	}

	err = png.Encode(w, output_img)
	if err != nil {
		log.Println(err)
		fmt.Fprint(w, err)
		return
	}
}
