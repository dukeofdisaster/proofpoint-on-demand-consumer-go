package main

/*
	TODO:
		- simulate duplicate events (same guid)
		- use modulus or other method to send a duplicate at every N interval
*/

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
var duplicateGuid = "SOMEGUIDHERE"

func main() {
	var sampleEvent = coredata.PODEvent{}
	var duplicateEvent = coredata.PODEvent{}
	duplicateEvent.Guid = duplicateGuid
	duplicateEvent.Timestamp = fmt.Sprintf("%d", time.Now().Unix())
	allowStrings := strings.Split(dude, "O")
	sampleEventBytes, err := json.Marshal(sampleEvent)
	if err != nil {
		log.Println(err)
		return
	}
	duplicateEventBytes, err := json.Marshal(duplicateEvent)
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
			//for i := 1; i < 1000; i++ {
			nowTime := time.Now().Unix()
			fmt.Println(nowTime)
			if nowTime%5 == 0 {
				if err := conn.WriteMessage(websocket.TextMessage, duplicateEventBytes); err != nil {
					log.Println(err)
					return
				}
			} else {
				// simulate inique by writing <blah>-<timestamp>-<blah>
				semiUniqueGuid := fmt.Sprintf("GUID%dGUID", nowTime)
				sampleEvent.Guid = semiUniqueGuid
				sampleEvent.Timestamp = fmt.Sprintf("%d", time.Now().Unix())
				sampleEventBytes, err := json.Marshal(sampleEvent)
				if err != nil {
					log.Println(err)
					return
				}
				if err := conn.WriteMessage(websocket.TextMessage, sampleEventBytes); err != nil {
					log.Println(err)
					return
				}
			}
			// simulate throttle
			time.Sleep(2 * time.Second)
		}
	})
	//http.ListenAndServe(":8080", nil)
	fmt.Println("listening on localhost:8088")
	http.ListenAndServe(":8088", nil)
}
