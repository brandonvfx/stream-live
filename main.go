package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

const font_dir = "./fonts/"
const twitch_streams_api = "https://api.twitch.tv/kraken/streams/"
const twitch_channels_api = "https://api.twitch.tv/kraken/channels/"

var text_arial, text_arial_bold, text_arial_black ImageText

func init() {
	text_arial = ImageText{
		fontfile: font_dir + "Arial.ttf",
	}
	text_arial.Init()
	text_arial_bold = ImageText{
		fontfile: font_dir + "Arial Bold.ttf",
	}
	text_arial_bold.Init()
	text_arial_black = ImageText{
		fontfile: font_dir + "Arial Black.ttf",
	}
	text_arial_black.Init()
}

type Stream struct {
	// Links  map[string]string `json:'_links'`
	SearchName string
	Stream     StreamInfo `json:"stream"`
	Live       bool       `json:"live"`
}

type StreamInfo struct {
	// Links   map[string]string `json:'_links'`
	Preview map[string]string `json:"preview"`
	ID      int               `json:"_id"`
	Game    string            `json:"game"`
	Viewers int               `json:"viewers"`
	Channel Channel           `json:"channel"`
}

type Channel struct {
	DisplayName string            `json:"display_name"`
	Status      string            `json:"status"`
	Name        string            `json:"name"`
	Links       map[string]string `json:"_links"`
	Logo        string            `json:"logo"`
}

func GetStream(stream_name string) (Stream, error) {
	log.Printf("Checking Stream: %s", stream_name)
	stream_req, err := http.NewRequest("GET", twitch_streams_api+stream_name, nil)
	stream := Stream{SearchName: stream_name}
	if err != nil {
		return stream, err
	}
	stream_req.Header = map[string][]string{
		"Accept": {"application/vnd.twitchtv.v3+json"},
	}

	client := http.Client{}
	stream_resp, err := client.Do(stream_req)
	if err != nil {
		return stream, err
	}

	if stream_resp.StatusCode == 404 {
		return stream, errors.New("Invalid Username")
	} else if stream_resp.StatusCode != 200 {
		return stream, errors.New("Twitch Error")
	}

	dec := json.NewDecoder(stream_resp.Body)
	err = dec.Decode(&stream)
	if err != nil {
		return stream, err
	}

	if stream.Stream.ID != 0 {
		stream.Live = true
	} else {
		channel_req, err := http.NewRequest("GET", twitch_channels_api+stream_name, nil)
		if err != nil {
			return stream, err
		}
		stream_req.Header = map[string][]string{
			"Accept": {"application/vnd.twitchtv.v3+json"},
		}

		stream_info_resp, err := client.Do(channel_req)
		if err != nil {
			return stream, err
		}

		if stream_info_resp.StatusCode != 200 {
			return stream, errors.New("Twitch Error")
		}

		dec := json.NewDecoder(stream_info_resp.Body)
		var stream_info StreamInfo
		err = dec.Decode(&stream_info.Channel)
		if err != nil {
			return stream, err
		}
		stream.Stream = stream_info
	}

	return stream, nil
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

	text_arial_bold.AddText(rgba, stream, 16, image.Point{10, 21}, color.White)
	offset := (len(stream) * 8) + 10
	text_arial.AddText(rgba, "Testing 1234", 16, image.Point{10 + offset, 21}, color.White)
	err = png.Encode(w, rgba)
	if err != nil {
		fmt.Fprint(w, err)
		return
	}
}

func DebugHandler(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	stream := vars["stream"]
	stream_info, err := GetStream(stream)
	if err != nil {
		fmt.Fprint(w, err)
		return
	}
	json_encoder := json.NewEncoder(w)
	err = json_encoder.Encode(stream_info)
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
