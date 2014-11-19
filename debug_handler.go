package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

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
