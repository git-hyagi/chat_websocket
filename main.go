package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/git-hyagi/chat_websocket/db"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

//const corsServer = "http://chatserver:8000"
const corsServer = "http://localhost:8081"

// when client connects bring only the last `lastNMsg` messages
const lastNMsg = 5

// message struct
type msg struct {
	Name    string
	Message string
	When    time.Time
	Doctor  string
	Patient string
	SentBy  string
}

// websocket struct
type wsStruct struct {
	redisStruct
	esStruct
	context.Context
	listConn map[*websocket.Conn]bool
	newConn  chan *websocket.Conn
}

type user struct {
	db.User
}

type register struct {
	db.User
}

type doctor struct {
	db.User
	Username string
	Name     string
	Password string
	Subtitle string
	Avatar   string
}

type patients struct {
	db.User
	Username string
	Name     string
	Password string
	Avatar   string
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

	database, err := db.Connect("telemedicine", "root", "password", "localhost:3306")
	if err != nil {
		log.Fatalln("[FATAL] Error on database conncetion", err)
	}

	user := &user{}
	user.DB = database

	doctor := &doctor{}
	doctor.DB = database

	patients := &patients{}
	patients.DB = database

	register := &register{}
	register.DB = database

	// websocket struct
	webSkt := &wsStruct{redisStruct: *rdis, Context: ctx, esStruct: *es, listConn: map[*websocket.Conn]bool{}, newConn: make(chan *websocket.Conn)}

	log.Println("[INFO] Starting server on port", port)

	r := mux.NewRouter()
	r.Handle("/ws/{doctor}/{patient}", webSkt)
	r.Handle("/login", user)
	r.Handle("/doctors", doctor)
	r.Handle("/patients/{doctor}", patients)
	r.Handle("/register", register)

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

	//webSkt.esIndex = mux.Vars(r)["patient"] + "-" + mux.Vars(r)["doctor"]
	webSkt.esIndex = "telemedicine"
	webSkt.doctor = mux.Vars(r)["doctor"]
	webSkt.patient = mux.Vars(r)["patient"]

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

// Login
func (user *user) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	var userLogin db.User
	//Allow CORS here By * or specific origin
	w.Header().Set("Access-Control-Allow-Origin", corsServer)
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	if err := json.NewDecoder(r.Body).Decode(&userLogin); err != nil {
		log.Println(err)
	}

	password, err := user.GetAttribute(userLogin.Name, "password")
	if err != nil {
		log.Println("User", userLogin.Name, "not found!")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	userType, _ := user.GetAttribute(userLogin.Name, "type")
	name, _ := user.GetAttribute(userLogin.Name, "name")
	avatar, _ := user.GetAttribute(userLogin.Name, "avatar")

	if userLogin.Password == password {
		http.SetCookie(w, &http.Cookie{
			Name: "user",
			//Value:    url.QueryEscape(name), //cookie v0 should not contain spaces, to avoid that we are using the "url encode/queryescape"
			Value:    (&url.URL{Path: name}).String(), //encode ' ' as %20 instead of +
			SameSite: http.SameSiteNoneMode,
			//Path:     "/",
		})
		http.SetCookie(w, &http.Cookie{
			Name:     "username",
			Value:    userLogin.Name,
			SameSite: http.SameSiteNoneMode,
			//Path:     "/",
		})
		http.SetCookie(w, &http.Cookie{
			Name:     "password",
			Value:    password,
			SameSite: http.SameSiteNoneMode,
			//Path:     "/",
		})
		http.SetCookie(w, &http.Cookie{
			Name:     "type",
			Value:    userType,
			SameSite: http.SameSiteNoneMode,
			//Path:     "/",
		})
		http.SetCookie(w, &http.Cookie{
			Name:     "avatar",
			Value:    avatar,
			SameSite: http.SameSiteNoneMode,
			//Path:     "/",
		})

		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusUnauthorized)
	}
}

// newConnections creates a goroutine for every new websocket connection to handle the messages from each client
func (webSkt *wsStruct) newConnections() {
	for {
		// received a new connection from channel
		ws := <-webSkt.newConn

		//func retrieveMessages(lastNMsg int, patient, doctor string, es *esStruct) ([]msg, error) {
		esMsgs, _ := retrieveMessages(lastNMsg, webSkt.patient, webSkt.doctor, &webSkt.esStruct)

		// iterate over the messages from slice
		for i := range esMsgs {
			var jsonMsg msg

			// the decoding is made in the reverse order (from the last element to the first)
			json.Unmarshal([]byte(`{"When": "`+esMsgs[len(esMsgs)-1-i].When.Format("2006-01-02T15:04:05Z")+
				`","Name": "`+esMsgs[len(esMsgs)-1-i].Name+
				`", "Message": "`+esMsgs[len(esMsgs)-1-i].Message+
				`", "Patient": "`+esMsgs[len(esMsgs)-1-i].Patient+
				`", "Doctor": "`+esMsgs[len(esMsgs)-1-i].Doctor+
				`"}`), &jsonMsg)
			if err := ws.WriteJSON(jsonMsg); err != nil {
				log.Println("[ERROR]", err)
				delete(webSkt.listConn, ws)
				return
			}
		}
		//}

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
		jsonMsg.SentBy = strings.Replace(jsonMsg.SentBy, "+", " ", -1)

		// publish the message on redis pubsub channel
		err = webSkt.redisStruct.client.Publish(webSkt.Context, webSkt.channel, `{"Name": "`+jsonMsg.SentBy+`","Message": "`+jsonMsg.Message+`","When": "`+date+`"}`).Err()
		if err != nil {
			log.Println("[ERROR]", err)
		}

		// define the elasticsearch fields
		webSkt.esStruct.patient = jsonMsg.Patient
		webSkt.esStruct.doctor = jsonMsg.Doctor
		webSkt.esStruct.sentBy = jsonMsg.SentBy
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

// Get the list of doctors
func (d *doctor) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	//Allow CORS here By * or specific origin
	w.Header().Set("Access-Control-Allow-Origin", corsServer)
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Content-Type", "application/json")

	doc, err := d.GetDoctors()
	if err != nil {
		log.Println("no doc found!", err)
		w.WriteHeader(http.StatusNoContent)
	}

	jsonMsg, _ := json.Marshal(doc)
	w.Write(jsonMsg)
	w.WriteHeader(http.StatusOK)
}

// Get the list of patients from a specific doctor
func (p *patients) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	//Allow CORS here By * or specific origin
	w.Header().Set("Access-Control-Allow-Origin", corsServer)
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Content-Type", "application/json")

	unescapeName, _ := url.QueryUnescape(mux.Vars(r)["doctor"])
	doc, err := p.GetPatients(unescapeName)
	if err != nil {
		log.Println("no doc found!", err)
		w.WriteHeader(http.StatusNoContent)
	}

	jsonMsg, _ := json.Marshal(doc)
	w.Write(jsonMsg)
	w.WriteHeader(http.StatusOK)
}

// Register a new user on database
func (u *register) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	var newUser db.User

	//Allow CORS here By * or specific origin
	w.Header().Set("Access-Control-Allow-Origin", corsServer)
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	if err := json.NewDecoder(r.Body).Decode(&newUser); err != nil {
		log.Println(err)
	}

	// [TO DO] need to check if the user exists already
	err := u.CreateUser(newUser.Username, newUser.Name, newUser.Password, newUser.Type, newUser.Subtitle, "")
	if err != nil {
		w.Write([]byte("Failed to create the user!"))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
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
		es.hosts = "http://localhost:9201"
	}

	return port, rdis, es
}
