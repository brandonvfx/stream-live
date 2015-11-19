package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
)

type stream struct {
	Name   string
	Stream streamData `json:"stream"`
	Live   bool       `json:"live"`
}

type streamData struct {
	Preview map[string]string `json:"preview"`
	ID      int               `json:"_id"`
	Game    string            `json:"game"`
	Viewers int               `json:"viewers"`
	Channel channel           `json:"channel"`
}

type channel struct {
	DisplayName string            `json:"display_name"`
	Status      string            `json:"status"`
	Name        string            `json:"name"`
	Links       map[string]string `json:"_links"`
	Logo        string            `json:"logo"`
}

func getStream(streamName string) (stream, error) {
	log.Printf("Checking Stream: %s", streamName)
	streamReq, err := http.NewRequest("GET", twitchStreamsAPI+streamName, nil)
	streamInst := stream{Name: streamName, Live: false}
	if err != nil {
		return streamInst, err
	}
	streamReq.Header = map[string][]string{
		"Accept": {"application/vnd.twitchtv.v3+json"},
	}

	client := http.Client{}
	streamResp, err := client.Do(streamReq)
	if err != nil {
		return streamInst, err
	}

	if streamResp.StatusCode == 404 {
		return streamInst, errors.New("Invalid Username")
	} else if streamResp.StatusCode != 200 {
		log.Printf("[ERROR] %s - %s", streamName, streamResp.Status)
		return streamInst, fmt.Errorf("Twitch Error - %s", streamResp.Status)
	}

	dec := json.NewDecoder(streamResp.Body)
	err = dec.Decode(&streamInst)
	if err != nil {
		return streamInst, err
	}

	if streamInst.Stream.ID != 0 {
		streamInst.Live = true
		streamInst.Name = streamInst.Stream.Channel.DisplayName
	} else {
		channelReq, err := http.NewRequest("GET", twitchChannelsAPI+streamName, nil)
		if err != nil {
			return streamInst, err
		}
		streamReq.Header = map[string][]string{
			"Accept": {"application/vnd.twitchtv.v3+json"},
		}

		streamInfoResp, err := client.Do(channelReq)
		if err != nil {
			return streamInst, err
		}

		if streamInfoResp.StatusCode != 200 {
			return streamInst, errors.New("Twitch Error")
		}

		dec := json.NewDecoder(streamInfoResp.Body)
		var streamInfo streamData
		err = dec.Decode(&streamInfo.Channel)
		if err != nil {
			return streamInst, err
		}
		streamInst.Stream = streamInfo
	}

	return streamInst, nil
}
