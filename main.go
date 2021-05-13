package main

import (
	"bufio"
	"context"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
)

type connStruct struct {
	closeConn  chan bool
	connection net.Conn
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type msg struct {
	Name    string
	Message string
	When    time.Time
}

type wsTest struct {
	redisStruct
	context.Context
}

var listConn = map[*websocket.Conn]bool{}

func main() {
	//newConn := make(chan string)
	var port string
	var err error
	rdis := &redisStruct{}
	es := &esStruct{}

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

	if os.Getenv("ES_INDEX") != "" {
		es.esIndex = os.Getenv("ES_INDEX")
	} else {
		es.esIndex = rdis.channel
	}

	poolConn := make(map[string]connStruct)
	msg := make(chan string)
	var ctx = context.Background()
	pubsub := make(<-chan *redis.Message)

	log.SetFlags(log.Ltime | log.Lshortfile)

	log.Println("Connecting to redis ...")
	connectRedis(ctx, rdis)
	es.client, err = connectES(es)
	if err != nil {
		log.Println("Error on ES connection", err)
	}

	log.Println("Starting server ...")

	wst := &wsTest{*rdis, ctx}

	http.Handle("/ws", wst)
	go http.ListenAndServe(port, nil)

	// // let a socket in listening state on port "port"
	// ln, err := net.Listen("tcp", port)
	// if err != nil {
	// 	log.Fatalln("Error to bind port ", port, err)
	// 	return
	// }
	// defer ln.Close()

	// go func that handles new connections
	//go newConnections(ctx, &ln, poolConn, newConn)

	// go func that handles new messages
	//go recvMessage(ctx, poolConn, newConn, rdis, es)

	// broadcast msg to all connected clients
	go printMsg(msg, poolConn)

	// according to the go-redis documentation:
	// Message receiving is NOT safe for concurrent use by multiple goroutines.
	// https://pkg.go.dev/github.com/go-redis/redis/v8#PubSub
	for {
		pubsub = rdis.client.Subscribe(ctx, rdis.channel).Channel()
		for redisMsg := range pubsub {
			log.Println("msg from redis:", redisMsg.Channel, redisMsg.Payload)
			msg <- redisMsg.Payload
		}
	}

}

func (wt *wsTest) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}
	listConn[ws] = true
	log.Println("Client Connected", listConn)

	for {
		_, p, err := ws.ReadMessage()
		if err != nil {
			delete(listConn, ws)
			log.Println(err)
			return
		}

		log.Println("redis channel: ", wt.channel)

		var aux msg
		json.Unmarshal(p, &aux)

		err = wt.client.Publish(wt.Context, wt.channel, `{"Name": "`+ws.RemoteAddr().String()+`","Message": "`+aux.Message+`","When": "`+time.Now().Format("2006-01-02T15:04:05Z")+`"}`).Err()
		if err != nil {
			log.Println("[ERROR] ", err)
		}
	}
}

// newConnections waits for the next connection and returns a net.Conn
func newConnections(ctx context.Context, ln *net.Listener, c map[string]connStruct, newConn chan string) {
	for {
		log.Println("Waiting for new clients ...")
		sockConn, err := (*ln).Accept()
		if err != nil {
			log.Println("Error connecting ", err)
			return
		}
		log.Println("connected!")
		log.Println(sockConn.RemoteAddr().String())

		conn := connStruct{
			closeConn:  make(chan bool),
			connection: sockConn,
		}

		c[sockConn.RemoteAddr().String()] = conn
		newConn <- sockConn.RemoteAddr().String()

		log.Println("connection added do the channel!")
	}
}

// closeConnection removes a client from pool after disconnection
func closeConnection(c map[string]connStruct, connName string) {
	c[connName].connection.Close()
	delete(c, connName)
}

// take the client's message and add it to the msg channel
func recvMessage(ctx context.Context, con map[string]connStruct, newConn chan string, rdb *redisStruct, es *esStruct) {

	// let it running forever to handle every new connection
	for {
		connName := <-newConn

		// each new connection will be handled by a different goroutine
		// each goroutine will wait for user input (the message)
		go func() {
			for {
				log.Println("Waiting for new message from ", connName, "...")
				connOutput, err := bufio.NewReader(con[connName].connection).ReadBytes('\n')
				if err != nil {
					log.Println("Connection from ", connName, " closed!")
					closeConnection(con, connName)
					return
				}

				// publish message on redis channel
				err = rdb.client.Publish(ctx, rdb.channel, connName+": "+string(connOutput)).Err()
				if err != nil {
					log.Println("[ERROR] ", err)
				}

				es.chatClients = connName
				es.msg = strings.TrimRight(string(connOutput), "\r\n")

				if err := index(es); err != nil {
					log.Println("indexing error!", err)
				}

				log.Print("Msg from ", connName, ": ", string(connOutput))
			}
		}()
	}

}

// printMsg consumes the message from channel and outputs it to
// all clients still connected
func printMsg(msgChan chan string, connections map[string]connStruct) {

	for {
		msg2 := <-msgChan
		log.Println("Message arrived!")

		// for i, j := range connections {
		// 	log.Println("Delivering msg [", strings.TrimRight(msg2, "\r\n"), "] to client ", i)
		// 	j.connection.Write([]byte(msg2))
		// }

		//log.Println("new message: ", aux.Message)

		aux := msg{}
		json.Unmarshal([]byte(msg2), &aux)
		for i := range listConn {
			if err := i.WriteJSON(aux); err != nil {
				log.Println(err)
				return
			}
		}

	}
}
