package main

/*
	TODO:
		- FLAG:agoOffset
			- (option) if agoOffset is specifed, subtract this amount of time to derive API:sinceTime ; allows overlap to ensure all messages consumed
			- should ??? be capped,
		- FUNC: deduplication
			- keeping this self contained, we should do the following
				- send all Rx'd messages to a channel in a go routine
				- have worker read from channel and check seen status in a db
				- IF !seen, send to shipChannel for processing
					- should do this in a mutex; the writes will error at sqlite side anyway
					but we should ensure that only one routine tries to write to db at a time
					- db should only be written to after successful write to NDJSON
						- guarantees 'shipped' ; i.e. if we wrote to db but had err on output
						we would miss events
		- read API key as an environment variable ??
		- libeat file outputter
			- review these
				- https://github.com/elastic/beats/blob/main/libbeat/outputs/fileout/file.go
			- juice not really worth the squeeze on this one... especially since generating
			custom beats was deprecated in 8.0 just dump poll and dump
		-
*/

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"strconv"

	//"net/url"
	"os"
	"time"

	coredata "github.com/dukeofdisaster/proofpoint-on-demand-consumer-go/core/data"
	"github.com/elastic/go-ucfg"
	"github.com/elastic/go-ucfg/yaml"
	"github.com/gorilla/websocket"
	_ "github.com/mattn/go-sqlite3"
)

var (
	agoOffset          *string
	checkpointPath     *string
	configPath         *string
	dbPath             *string
	endPointUrl        *string
	lastCheckpointTime int64
	logPath            *string
	seen               string
	seenTime           string
)

const (
	QUERY_SELECT_ALL = "select * from proofpoint_events"
	QUERY_SELECT_ID  = "select * from proofpoint_events where id = ?"
	QUERY_INSERT_ID  = "insert into proofpoint_events values(?,?)"
)

func init() {
	agoOffset = flag.String("a", "15m", "an ago time in <integer><units> format e.g. 15m")
	checkpointPath = flag.String("checkpoint", "none", "path to write checkpoint file")
	configPath = flag.String("c", "none", "path to config")
	dbPath = flag.String("db", "none", "path to sqlite3 db")
	endPointUrl = flag.String("u", "none", "url of the websockets endpoint")
	logPath = flag.String("l", "none", "path to log file")
}

func convertAgoString(a string) (*coredata.AgoType, error) {
	if len(a) < 2 {
		return nil, fmt.Errorf("convertAgoTime(): a valid ago time has at least 2 characters")
	}
	suffix := a[len(a)-1:]
	prefix := a[:len(a)-1]
	if suffix == "m" || suffix == "h" {
		val, err := strconv.Atoi(prefix)
		if err != nil {
			return nil, err
		}
		agoObj := coredata.AgoType{
			Units: suffix,
			Value: val,
		}
		return &agoObj, nil
	}
	return nil, fmt.Errorf("convertAgoTime(): only m|h are accepted unit")
}

// checks wither the given string is of the format <integer><unit>"
// onyl minutes and hours are valid
//func isValidAgoTime(a string) bool {
//}

func newTrue() *bool {
	t := true
	return &t
}
func newFalse() *bool {
	f := false
	return &f
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

func insertEvent(db *sql.DB, id string, seentime string) (int, error) {
	_, e := db.Exec(QUERY_INSERT_ID, id, seentime)
	if e != nil {
		return 0, e
	}
	return 1, nil
}

func isSeenEvent(db *sql.DB, id string) (*bool, error) {
	if err := db.QueryRow(QUERY_SELECT_ID, id).Scan(&seen, &seenTime); err != nil {
		if err == sql.ErrNoRows {
			return newFalse(), nil
		} else {
			return nil, fmt.Errorf("isSeenEvent(): %v", err)
		}
	}
	return newTrue(), nil
}

func getCheckpoint(p string) (int64, error) {
	f, err := os.Open(p)
	if err != nil {
		log.Println(err)
		return int64(0), err
	}
	b, err := io.ReadAll(f)
	if err != nil {
		log.Println(err)
		return int64(0), err
	}
	var cp = coredata.Checkpoint{}
	err = json.Unmarshal(b, &cp)
	if err != nil {
		log.Println(err)
		return int64(0), err
	}
	return cp.Timestamp, nil
}
func writeCheckpoint(path string, t int64) error {
	var current = coredata.Checkpoint{Timestamp: t}
	b, err := json.Marshal(current)
	if err != nil {
		log.Println(err)
		return err
	}
	err = os.WriteFile(path, b, 0644)
	if err != nil {
		log.Println(err)
		return err
	}
	log.Println("writeCheckpoint(): ", t)
	return nil
}

func getConfigFromPath(p string) (*coredata.Config, error) {
	var newConf coredata.Config
	cf, err := yaml.NewConfigWithFile(p, ucfg.PathSep("."))
	if err != nil {
		return nil, err
	}
	err = cf.Unpack(&newConf)
	if err != nil {
		return nil, err
	}
	return &newConf, nil
}

func dumpConfigJson(c *coredata.Config) error {
	if c != nil {
		b, err := json.Marshal(c)
		if err != nil {
			return err
		}
		log.Println(string(b))
	}
	return fmt.Errorf("dumpConfigJson(): got a nil config...")
}

// given the start time of a polling period in epoch, the interval minutes, and the checkpoint path, determine if checkpoint needs
// to be written
func handleCheckpointStatus(start int64, m int64, p string) error {
	var now = time.Now().Unix()
	var diffSecs = now - start
	var intervalSecs = m * 60
	var secondsSinceLastCheckpoint = now - lastCheckpointTime
	var diffSecsModInterval = diffSecs % intervalSecs
	if (diffSecsModInterval >= 0) && (diffSecsModInterval <= 5) && (secondsSinceLastCheckpoint > intervalSecs) {
		if p != "none" {
			err := writeCheckpoint(p, now)
			if err != nil {
				return err
			}
			lastCheckpointTime = now
			return nil
		} else {
			log.Println("no checkpoint path specified: checkpoint:", now)
			return nil
		}
	}
	return nil
}

// add helper for manipulating endponit from string
func main() {
	// this is kinda jank... need refactor but leave
	flag.Parse()
	if *configPath != "none" {
		setLogFromPath(*logPath)
		newConfig, err := getConfigFromPath(*configPath)
		if err != nil {
			log.Println(err)
			return
		}
		dumpConfigJson(newConfig)
		log.Printf("connecting to %s", newConfig.Endpoint)
		conn, resp, err := websocket.DefaultDialer.Dial(newConfig.Endpoint, nil)
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
				log.Println("read:", err)
				return
			}
			log.Printf("message type: %d", messageType)
			log.Printf("recv: %s", message)
			time.Sleep(1 * time.Second)

		}
	}
	// if configPath is none than we need at least endpoint url, possibly environemnt variable
	if *endPointUrl != "none" {
		fmt.Println(*endPointUrl)
		if *dbPath != "none" {
			fmt.Println(*dbPath)
			// open connection to db
			dbconn, err := sql.Open("sqlite3", *dbPath)
			if err != nil {
				log.Println("err on db open:", err)
				//return err
				return
			}
			conn, _, err := websocket.DefaultDialer.Dial(*endPointUrl, nil)
			if err != nil {
				log.Fatal("dial: ", err)
				//return err
				return
			}
			defer conn.Close()
			// holder struct for reading the message
			var start = time.Now().Unix()
			var interval = int64(2)
			lastCheckpointTime = start
			var currentEvent = coredata.PODEvent{}
			for {
				if err := conn.ReadJSON(&currentEvent); err != nil {
					log.Println("READJSON: ", err)
					if websocket.IsUnexpectedCloseError(err) {
						log.Println("unexpected 1006 ok")
					}
					return
				} else {
					fmt.Println("GOT THIS GUID: ", currentEvent.Guid)
					if seen, err := isSeenEvent(dbconn, currentEvent.Guid); err != nil {
						log.Println("err determing seen status: ", err)
					} else {
						if *seen {
							fmt.Println("seen this guid: ", currentEvent.Guid)
						} else {
							fmt.Println("NOT seen this guid: ", currentEvent.Guid)
							_, err := insertEvent(dbconn, currentEvent.Guid, fmt.Sprintf("%d", time.Now().Unix()))
							if err != nil {
								fmt.Println("err on insert: ", err)
							}
						}
					}
				}
				err := handleCheckpointStatus(start, interval, *checkpointPath)
				if err != nil {
					log.Println(err)
				}
			}
		} else {
			// write initial checkpoint path
			log.Println("[WARNING] - No db path specified, poller will not attempt deduplication")
			log.Printf("connecting to %s", *endPointUrl)
			conn, resp, err := websocket.DefaultDialer.Dial(*endPointUrl, nil)
			if err != nil {
				log.Fatal("dial:", err)
			}
			log.Println("status code: ", resp.StatusCode)
			defer conn.Close()
			var start = time.Now().Unix()
			var interval = int64(2)
			lastCheckpointTime = start
			for {
				messageType, message, err := conn.ReadMessage()
				if err != nil {
					log.Println("read:", err)
					return
				}
				log.Printf("message type: %d", messageType)
				log.Printf("recv: %s", message)
				time.Sleep(1 * time.Second)
				err = handleCheckpointStatus(start, interval, *checkpointPath)
				if err != nil {
					log.Println(err)
				}
			}
		}
	}
	flag.PrintDefaults()
}
