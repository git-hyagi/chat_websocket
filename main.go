package main

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

// when client connects bring only the last `lastNMsg` messages
const lastNMsg = 5

// message struct
type msg struct {
	Name    string
	Message string
	When    time.Time
}

// websocket struct
type wsStruct struct {
	redisStruct
	esStruct
	context.Context
	listConn map[*websocket.Conn]bool
	newConn  chan *websocket.Conn
}

func main() {
	var err error
	log.SetFlags(log.Ltime | log.Lshortfile)

	// websocket port,redis struct, es struct
	port, rdis, es := getEnvVars()

	// context used by redis
	var ctx = context.Background()

	// channel for sharing the message between goroutines
	msg := make(chan string)

	// channel for concurrently receive the pubsub messages from redis
	pubSubMsg := make(<-chan *redis.Message)

	log.Println("[INFO] Connecting to redis ...")
	err = connectRedis(ctx, rdis)
	if err != nil {
		log.Fatalln("[FATAL] Could not connect to Redis instance!", err)
	}

	log.Println("[INFO] Connecting to elasticsearch ...")
	es.client, err = connectES(es)
	if err != nil {
		log.Fatalln("[FATAL] Error on ES connection", err)
	}

	// websocket struct
	webSkt := &wsStruct{redisStruct: *rdis, Context: ctx, esStruct: *es, listConn: map[*websocket.Conn]bool{}, newConn: make(chan *websocket.Conn)}

	log.Println("[INFO] Starting server on port", port)

	r := mux.NewRouter()
	r.Handle("/ws/{room}", webSkt)

	//http.Handle("/ws", webSkt)
	//go http.ListenAndServe(port, nil)
	go http.ListenAndServe(port, r)

	// creates a goroutine for each client connection
	go webSkt.newConnections()

	// broadcast msg to all connected clients
	go webSkt.printMsg(msg)

	// according to the go-redis documentation:
	// Message receiving is NOT safe for concurrent use by multiple goroutines.
	// https://pkg.go.dev/github.com/go-redis/redis/v8#PubSub
	for {
		pubSubMsg = rdis.client.Subscribe(ctx, rdis.channel).Channel()
		for redisMsg := range pubSubMsg {
			log.Println("[INFO] New msg from redis. channel:", redisMsg.Channel, "payload:", redisMsg.Payload)
			msg <- redisMsg.Payload
		}
	}
}

// ServeHTTP creates the websocket
func (webSkt *wsStruct) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	webSkt.esIndex = "chat-" + mux.Vars(r)["room"]

	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}

	// upgrade the http connection to a websocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}

	// store this websocket connection in the list of connections
	webSkt.listConn[ws] = true
	log.Println("[INFO] New Client Connected! List of connections:", webSkt.listConn)

	// send this new connection to newConn channel, to let rcvMsg goroutine handle it
	webSkt.newConn <- ws
}

// newConnections creates a goroutine for every new websocket connection to handle the messages from each client
func (webSkt *wsStruct) newConnections() {
	for {
		// received a new connection from channel
		ws := <-webSkt.newConn

		var r map[string]interface{}
		var buf bytes.Buffer
		var esMsgs []msg

		// prepare the query
		// - bring only the last N messages (const lastNMsg)
		// - ordered by date (only the earliest messages)
		query := map[string]interface{}{
			"size": lastNMsg,
			"sort": map[string]interface{}{
				"date": map[string]interface{}{
					"order": "desc",
				},
			},
			"query": map[string]interface{}{
				"match_all": map[string]interface{}{},
			},
		}

		if err := json.NewEncoder(&buf).Encode(query); err != nil {
			log.Fatalf("[ERROR] Error encoding query: %s", err)
		}
		res, err := search(&webSkt.esStruct, buf)
		if err != nil {
			log.Fatalf("[ERROR] Error getting response: %s", err)
		}
		defer res.Body.Close()

		if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
			log.Fatalf("[ERROR] Error parsing the response body: %s", err)
		}

		// return the history of messages only if they could be found on elasticsearch
		if res.StatusCode != 404 {
			for _, hit := range r["hits"].(map[string]interface{})["hits"].([]interface{}) {

				// convert it to time.Time because of msg.When field type
				msgDate, _ := time.Parse(hit.(map[string]interface{})["_source"].(map[string]interface{})["date"].(string), "2006-01-02T15:04:05Z")
				aux := msg{
					Message: hit.(map[string]interface{})["_source"].(map[string]interface{})["msg"].(string),
					When:    msgDate,
					Name:    hit.(map[string]interface{})["_source"].(map[string]interface{})["client"].(string),
				}

				// store the elasticsearch results in a slice because the dates are in the wrong order
				esMsgs = append(esMsgs, aux)
			}

			// iterate over the messages from slice
			for i := range esMsgs {
				var jsonMsg msg

				// the decoding is made in the reverse order (from the last element to the first)
				json.Unmarshal([]byte(`{"When": "`+esMsgs[len(esMsgs)-1-i].When.Format("2006-01-02T15:04:05Z")+`","Name": "`+esMsgs[len(esMsgs)-1-i].Name+`", "Message": "`+esMsgs[len(esMsgs)-1-i].Message+`"}`), &jsonMsg)
				if err := ws.WriteJSON(jsonMsg); err != nil {
					log.Println("[ERROR]", err)
					return
				}
			}

		}

		go webSkt.rcvMsg(ws)
	}
}

// rcvMsg waits for new messages from clients
// when a message arrives publish it on redis channel
// and store on elasticsearch
func (webSkt *wsStruct) rcvMsg(ws *websocket.Conn) {
	for {

		// wait for new messages on websocket
		_, p, err := ws.ReadMessage()
		if err != nil {
			// remove the connection from the list in case of error
			delete(webSkt.listConn, ws)
			log.Println("[ERROR]", err)
			return
		}

		date := time.Now().Format("2006-01-02T15:04:05Z")

		// decode the message from []byte to json
		var jsonMsg msg
		json.Unmarshal(p, &jsonMsg)
		jsonMsg.Message = strings.Replace(jsonMsg.Message, "\n", "\\n", -1)

		// publish the message on redis pubsub channel
		err = webSkt.redisStruct.client.Publish(webSkt.Context, webSkt.channel, `{"Name": "`+ws.RemoteAddr().String()+`","Message": "`+jsonMsg.Message+`","When": "`+date+`"}`).Err()
		if err != nil {
			log.Println("[ERROR]", err)
		}

		// define the elasticsearch fields
		webSkt.esStruct.chatClients = ws.RemoteAddr().String()
		webSkt.esStruct.msg = strings.TrimRight(string(jsonMsg.Message), "\r\n")
		webSkt.esStruct.date = date

		// store the message on elasticsearch
		if err := index(&webSkt.esStruct); err != nil {
			log.Println("[ERROR] indexing error!", err)
		}
	}
}

// printMsg consumes the message from channel and outputs it to
// all clients still connected on the websocket
func (webSkt *wsStruct) printMsg(msgChan chan string) {

	for {
		// consume the subscription message received from channel
		newMsg := <-msgChan

		// decode the message from []byte to json
		var jsonMsg msg
		json.Unmarshal([]byte(newMsg), &jsonMsg)

		// broadcast the message to all clients connected
		for i := range webSkt.listConn {
			if err := i.WriteJSON(jsonMsg); err != nil {
				log.Println("[ERROR]", err)
				return
			}
		}
	}
}

// check if some environment variables were declared and if they did define
// the vars with their contents
func getEnvVars() (port string, rdis *redisStruct, es *esStruct) {

	rdis = &redisStruct{}
	es = &esStruct{}

	if os.Getenv("CHAT_PORT") != "" {
		port = os.Getenv("CHAT_PORT")
	} else {
		port = ":8080"
	}

	if os.Getenv("REDIS_ADDR") != "" {
		(*rdis).addr = os.Getenv("REDIS_ADDR")
	} else {
		(*rdis).addr = "localhost:6379"
	}

	if os.Getenv("REDIS_PASS") != "" {
		(*rdis).pass = os.Getenv("REDIS_PASS")
	} else {
		(*rdis).pass = ""
	}

	if os.Getenv("REDIS_CHANNEL") != "" {
		(*rdis).channel = os.Getenv("REDIS_CHANNEL")
	} else {
		(*rdis).channel = "chat"
	}

	if os.Getenv("REDIS_DB") != "" {
		(*rdis).db, _ = strconv.Atoi(os.Getenv("REDIS_DB"))
	} else {
		(*rdis).db = 0
	}

	if os.Getenv("ES_HOST") != "" {
		es.hosts = os.Getenv("ES_HOST")
	} else {
		es.hosts = "http://localhost:9200"
	}

	/*
		if os.Getenv("ES_INDEX") != "" {
			es.esIndex = os.Getenv("ES_INDEX")
		} else {
			es.esIndex = rdis.channel
		}
	*/

	return port, rdis, es
}
