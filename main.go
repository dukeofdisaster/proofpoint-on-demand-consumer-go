package main

import (
	"fmt"
	"flag"
	"log"
	"net/url"
	"os"
	"time"

        "github.com/elastic/go-ucfg"
        "github.com/elastic/go-ucfg/yaml"
	coredata "github.com/dukeofdisaster/proofpoint-on-demand-consumer-go/core/data"
	"github.com/gorilla/websocket"
)
var (
	configPath *string
	logPath *string
	
)
func init() {
	configPath = flag.String("c", "none", "path to config")
	logPath = flag.String("l", "none", "path to log file")
}

func setLogFromPath(p string) error {
	if p != "none" {
		log_file, err := os.OpenFile(p, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			log.Fatal(err)
			return err
		}
		defer log_file.Close()
		log.SetOutput(log_file)
		return nil
	}
	return nil
}
// add helper for manipulating endponit from string
func main() {
	// this is kinda jank... need refactor but leave
	flag.Parse()
	setLogFromPath(*logPath)
	newConfig := coredata.Config{}
	configFromFile, err := yaml.NewConfigWithFile(*configPath, ucfg.PathSep("."))
	if err != nil {
		log.Println(err)
		return
	}
	err = configFromFile.Unpack(&newConfig)
	if err != nil {
		log.Println(err)
		return
	}
	fmt.Println(newConfig.Endpoint)
	//u := url.URL{Scheme: "ws", Host: "localhost:8080", Path: "/"}
	u := url.URL{Scheme: "wss", Host: "localhost:8080", Path: "/"}
	log.Printf("connecting to %s", u.String())
	conn, resp, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		//log.Println("handshake fail with status: %d", resp.StatusCode)
		log.Fatal("dial:", err)
	} else {
		log.Println("resp:", resp.StatusCode)
	}
	defer conn.Close()
	for {
		// receive messages from the connection
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("read:",err)
			return
		}
		log.Printf("message type: %d",messageType)
		log.Printf("recv: %s",message)
		time.Sleep(1 * time.Second)
	}
}
