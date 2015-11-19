package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func debugHandler(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	stream := vars["stream"]
	streamInfo, err := getStream(stream)
	if err != nil {
		fmt.Fprint(w, err)
		return
	}
	jsonEncoder := json.NewEncoder(w)
	err = jsonEncoder.Encode(streamInfo)
	if err != nil {
		fmt.Fprint(w, err)
		return
	}
}
