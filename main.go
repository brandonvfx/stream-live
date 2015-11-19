package main

import (
	"bytes"
	"fmt"
	"image"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"

	"github.com/gorilla/mux"
	"github.com/codegangsta/cli"
)

const (
	fontDir = "./assets/fonts/"
	twitchStreamsAPI = "https://api.twitch.tv/kraken/streams/"
	twitchChannelsAPI = "https://api.twitch.tv/kraken/channels/"
	noProfileImage = "./assets/img/missing_profile_image.png"
)

var Version string
var textRegular, textBold ImageText
var missingProfileImage image.Image

func init() {
	// Load fonts
	textRegular = ImageText{
		fontfile: path.Join(fontDir, "Chivo-Regular.ttf"),
	}
	textRegular.Init()
	textBold = ImageText{
		fontfile: path.Join(fontDir, "Chivo-Black.ttf"),
	}
	textBold.Init()

	// Load missing preview image
	data, err := ioutil.ReadFile(noProfileImage)
	if err != nil {
		log.Fatalln(err)
	}
	buf := bytes.NewBuffer(data)

	missingProfileImage, _, err = image.Decode(buf)
	if err != nil {
		log.Fatalln(err)
	}
}

func main() {
	app := cli.NewApp()
	app.Name = "stream-live"
	app.Version = Version

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "port, p",
			Value:  "8000",
			Usage:  "Port to listen on",
			EnvVar: "PORT",
		},
		cli.BoolFlag{
			Name:   "debug-endpoint, d",
			Usage:  "Enable debug endpoint",
		},
	}

	app.Action = func(c *cli.Context) {
		log.Printf("stream-live Version: %v", Version)
		r := mux.NewRouter()
		r.HandleFunc("/{stream}.png", largeImageHandler)
		if c.Bool("debug-endpoint") {
			r.HandleFunc("/debug/{stream}", debugHandler)
		}
		
		http.Handle("/", r)
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", c.String("port")), nil))
	}
	app.Run(os.Args)
}
