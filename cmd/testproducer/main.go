package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	coredata "github.com/dukeofdisaster/proofpoint-on-demand-consumer-go/core/data"
    	"github.com/gorilla/websocket"

)

var upgrader = websocket.Upgrader{}
var dude = "HELLOWORLD"
func main() {
	var sampleEvent = coredata.PODEvent{}
	allowStrings := strings.Split(dude, "O")
	sampleEventBytes, err := json.Marshal(sampleEvent)
	if err != nil {
		log.Println(err)
		return
	}
	fmt.Println(string(sampleEventBytes))
	fmt.Println(upgrader)
	fmt.Println(allowStrings)

	// handle websocket connections
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Print("upgrade fail: ", err)
			return
		}
		defer conn.Close()
		// server forever
		for {
			
			// message type consts defined in the lib
			if err  := conn.WriteMessage(websocket.TextMessage, sampleEventBytes) ; err != nil {
				log.Println(err)
				return
			}
			// simulate throttle
			time.Sleep(2 * time.Second)
		}
	})
	http.ListenAndServe(":8080", nil)
}
