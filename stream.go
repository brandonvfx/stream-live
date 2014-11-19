package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
)

type Stream struct {
	// Links  map[string]string `json:'_links'`
	Name   string
	Stream StreamInfo `json:"stream"`
	Live   bool       `json:"live"`
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
	stream := Stream{Name: stream_name, Live: false}
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
		log.Printf("[ERROR] %s - %s", stream_name, stream_resp.Status)
		return stream, errors.New(fmt.Sprintf("Twitch Error - %s", stream_resp.Status))
	}

	dec := json.NewDecoder(stream_resp.Body)
	err = dec.Decode(&stream)
	if err != nil {
		return stream, err
	}

	if stream.Stream.ID != 0 {
		stream.Live = true
		stream.Name = stream.Stream.Channel.DisplayName
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
