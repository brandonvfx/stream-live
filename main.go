package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"

	"github.com/gorilla/mux"
)

const font_dir = "./assets/fonts/"
const twitch_streams_api = "https://api.twitch.tv/kraken/streams/"
const twitch_channels_api = "https://api.twitch.tv/kraken/channels/"
const no_profile_image = "./assets/img/missing_profile_image.png"

var text_regular, text_bold ImageText
var missing_profile_image image.Image

func init() {
	// Load fonts
	text_regular = ImageText{
		fontfile: path.Join(font_dir, "Chivo-Regular.ttf"),
	}
	text_regular.Init()
	text_bold = ImageText{
		fontfile: path.Join(font_dir, "Chivo-Black.ttf"),
	}
	text_bold.Init()

	// Load missing preview image
	data, err := ioutil.ReadFile(no_profile_image)
	if err != nil {
		log.Fatalln(err)
	}
	buf := bytes.NewBuffer(data)

	missing_profile_image, _, err = image.Decode(buf)
	if err != nil {
		log.Fatalln(err)
	}
}

// Old test handler
func SmallTextHandler(w http.ResponseWriter, req *http.Request) {
	log.Println("Start")
	vars := mux.Vars(req)
	stream := vars["stream"]
	stream_info, err := GetStream(stream)
	if err != nil {
		fmt.Fprint(w, err)
		return
	}

	if stream_info.Live {
		stream = stream_info.Stream.Channel.DisplayName
	}

	length := 250
	rgba := image.NewRGBA(image.Rect(0, 0, length, 100))
	draw.Draw(rgba, rgba.Bounds(), image.White, image.ZP, draw.Src)

	text_bold.AddText(rgba, stream, 16, image.Point{10, 21}, color.White)
	offset := (len(stream) * 8) + 10
	text_regular.AddText(rgba, "Testing 1234", 16, image.Point{10 + offset, 21}, color.White)
	err = png.Encode(w, rgba)
	if err != nil {
		fmt.Fprint(w, err)
		return
	}
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/small/{stream}.png", SmallTextHandler)
	r.HandleFunc("/large/{stream}.png", LargeImageHandler)
	r.HandleFunc("/debug/{stream}", DebugHandler)

	http.Handle("/", r)
	port := os.Getenv("PORT")
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))

}
