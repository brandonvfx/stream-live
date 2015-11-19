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

	"github.com/llgcode/draw2d/draw2dimg"
)

func drawRoundedRect(dst draw.Image,
	width float64,
	height float64,
	fillColor color.Color,
	strokeColor color.Color,
	strokeWidth float64) {
	x, y := 0.0, 0.0
	aspect := 1.0                /* aspect ratio */
	cornerRadius := height / 7.5 /* and corner curvature radius */

	radius := cornerRadius / aspect
	degrees := math.Pi / 180.0

	gcbg := draw2dimg.NewGraphicContext(dst)
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
	gc := draw2dimg.NewGraphicContext(dst)
	x, y := 0.0, 0.0
	aspect := 1.0                /* aspect ratio */
	cornerRadius := height / 7.5 /* and corner curvature radius */

	radius := cornerRadius / aspect
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

func getPreviewImage(imgURL string) (image.Image, error) {
	var tmpImg image.Image
	resp, err := http.Get(imgURL)
	if err != nil {
		return tmpImg, err
	}

	previewImg, _, err := image.Decode(resp.Body)
	if err != nil {
		return tmpImg, err
	}
	return previewImg, nil
}

func largeImageHandler(w http.ResponseWriter, req *http.Request) {
	// Settings
	width, height := 430.0, 64.0
	previewWidth := 150
	previewImageOffset := 24

	var previewImageURL string

	// url vars
	vars := mux.Vars(req)
	stream := vars["stream"]
	log.Printf("LargeImageHandler - %s", stream)
	streamInfo, err := getStream(stream)
	if err != nil {
		log.Println(fmt.Sprintf("[ERROR] - %v", err))
		// fmt.Fprint(w, err)
		// return
	}

	if streamInfo.Live {
		previewImageOffset = 0
		previewImageURL = streamInfo.Stream.Preview["medium"]
	} else {
		previewImageURL = streamInfo.Stream.Channel.Logo
	}

	// Get preview image
	var previewSrcImage image.Image
	if previewImageURL != "" {
		previewSrcImage, err = getPreviewImage(previewImageURL)
		if err != nil {
			log.Println(err)
			fmt.Fprint(w, err)
			return
		}
	} else {
		previewSrcImage = missingProfileImage
	}
	// resize the preview
	previewImage := resize.Resize(0, uint(height-2.0), previewSrcImage, resize.Lanczos3)

	// create the output image and it's background
	outputImg := image.NewRGBA(image.Rect(0, 0, int(width), int(height)))
	drawRoundedRect(outputImg,
		width,
		height,
		color.RGBA{0x64, 0x41, 0xA5, 0xFF},  // Twitch purple
		image.White,
		1.0)

	// Create mask for preview image
	maskImg := image.NewRGBA(image.Rect(0, 0, int(previewWidth), int(height)))
	drawMask(maskImg)

	// Comp the preview image over the bg using the mask.
	draw.DrawMask(outputImg,
		outputImg.Bounds(),
		previewImage,
		image.Point{-previewImageOffset, 0},
		maskImg,
		image.Point{0, 0},
		draw.Over)

	// Redraw the border.
	drawRoundedRect(outputImg,
		width,
		height,
		image.Transparent,
		image.White,
		1.0)

	// Draw divider
	gcLine := draw2dimg.NewGraphicContext(outputImg)
	gcLine.SetStrokeColor(image.White)
	gcLine.SetFillColor(image.Transparent)
	gcLine.MoveTo(110, 0)
	gcLine.LineTo(110, 0)
	gcLine.LineTo(110, height)
	gcLine.SetLineWidth(1.5)
	gcLine.FillStroke()

	leftSide := 115
	if streamInfo.Live {
		red := color.RGBA{0xDF, 0x2D, 0x28, 0xFF}
		textBold.AddText(outputImg, streamInfo.Name, 16 , image.Point{leftSide + 37, 18}, color.White)
		textBold.AddText(outputImg, "[LIVE]", 12 , image.Point{leftSide, 17}, red)
		// max length 52 chars
		textRegular.AddText(
			outputImg, 
			TruncString(streamInfo.Stream.Channel.Status, 52), 
			11 , 
			image.Point{leftSide, 32}, 
			color.White)
		// max length 45 chars
		textRegular.AddText(
			outputImg, 
			fmt.Sprintf("Playing - %s", TruncString(streamInfo.Stream.Game, 45)), 
			11 , 
			image.Point{leftSide, 45}, 
			color.White)
		textRegular.AddText(
			outputImg, 
			humanize.Comma(int64(streamInfo.Stream.Viewers))+" Viewers", 
			11 , 
			image.Point{leftSide, 58}, 
			color.White)
	} else {
		textBold.AddText(outputImg, streamInfo.Name, 16, image.Point{leftSide + 52, 18}, color.White)
		textBold.AddText(outputImg, "[Offline]", 12, image.Point{leftSide, 17}, color.White)
	}

	err = png.Encode(w, outputImg)
	if err != nil {
		log.Println(err)
		fmt.Fprint(w, err)
		return
	}
}
