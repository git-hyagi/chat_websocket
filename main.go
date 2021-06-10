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
	"golang.org/x/crypto/bcrypt"
)

const corsServer = "http://192.168.15.114"

// when client connects bring only the last `lastNMsg` messages
const lastNMsg = 5

// temporarily secret key for sign jwt
const secretKey = "1234567890"

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
	listConn map[string][]*websocket.Conn
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
	port, rdis, es, dbCred := getEnvVars()

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

	log.Println("[INFO] Connecting to mariadb ...")
	database, err := db.Connect(dbCred.Database, dbCred.User, dbCred.Password, dbCred.Host+":"+dbCred.Port)
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
	webSkt := &wsStruct{redisStruct: *rdis, Context: ctx, esStruct: *es, listConn: map[string][]*websocket.Conn{}, newConn: make(chan *websocket.Conn)}

	log.Println("[INFO] Starting server on port", port)

	r := mux.NewRouter()
	r.Handle("/ws/{doctor}/{patient}", webSkt)
	r.Handle("/login", user)
	r.Handle("/doctors", MustAuth(doctor))
	r.Handle("/patients/{doctor}", MustAuth(patients))
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
			//log.Println("[INFO] New msg from redis. channel:", redisMsg.Channel, "payload:", redisMsg.Payload)
			msg <- redisMsg.Payload
		}
	}
}

// ServeHTTP creates the websocket
func (webSkt *wsStruct) ServeHTTP(w http.ResponseWriter, r *http.Request) {

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
	webSkt.listConn[mux.Vars(r)["doctor"]+"-"+mux.Vars(r)["patient"]] = append(webSkt.listConn[mux.Vars(r)["doctor"]+"-"+mux.Vars(r)["patient"]], ws)

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

	// test jwt without expiration
	jwt := &JWTManager{secretKey, 8760 * time.Hour}
	token, _ := jwt.Generate(userLogin.Name, userType, avatar)

	if err = bcrypt.CompareHashAndPassword([]byte(password), []byte(userLogin.Password)); err != nil {
		w.WriteHeader(http.StatusUnauthorized)
	} else {
		http.SetCookie(w, &http.Cookie{
			Name:     "user",
			Value:    (&url.URL{Path: name}).String(), //encode ' ' as %20 instead of +
			SameSite: http.SameSiteNoneMode,
		})
		http.SetCookie(w, &http.Cookie{
			Name:     "username",
			Value:    userLogin.Name,
			SameSite: http.SameSiteNoneMode,
		})
		http.SetCookie(w, &http.Cookie{
			Name:     "type",
			Value:    userType,
			SameSite: http.SameSiteNoneMode,
		})
		http.SetCookie(w, &http.Cookie{
			Name:     "avatar",
			Value:    avatar,
			SameSite: http.SameSiteNoneMode,
		})
		http.SetCookie(w, &http.Cookie{
			Name:     "token",
			Value:    token,
			SameSite: http.SameSiteNoneMode,
		})
		w.WriteHeader(http.StatusOK)
	}
}

// newConnections creates a goroutine for every new websocket connection to handle the messages from each client
func (webSkt *wsStruct) newConnections() {
	for {
		// received a new connection from channel
		ws := <-webSkt.newConn
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
				//delete(webSkt.listConn, ws)
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
			//delete(webSkt.listConn, ws)
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
		err = webSkt.redisStruct.client.Publish(webSkt.Context, webSkt.channel, `{
			"Name": "`+jsonMsg.SentBy+`",
			"Message": "`+jsonMsg.Message+`",
			"Doctor": "`+jsonMsg.Doctor+`",
			"Patient": "`+jsonMsg.Patient+`",
			"SentBy": "`+jsonMsg.SentBy+`",
			"When": "`+date+`"
			}`).Err()
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
		if err := json.Unmarshal([]byte(newMsg), &jsonMsg); err != nil {
			log.Println("err", err)
		}

		// broadcast the message to all connections from the same websocket
		for i, conn := range webSkt.listConn[jsonMsg.Doctor+"-"+jsonMsg.Patient] {
			if err := conn.WriteJSON(jsonMsg); err != nil {
				log.Println("[ERROR]", err)
				webSkt.listConn[jsonMsg.Doctor+"-"+jsonMsg.Patient] = removeConnection(webSkt.listConn[jsonMsg.Doctor+"-"+jsonMsg.Patient], i)
				continue
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
		http.Error(w, err.Error(), http.StatusNoContent)
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
		http.Error(w, err.Error(), http.StatusNoContent)
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
		w.Write([]byte("Failed to create the user!"))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err := u.CreateUser(newUser.Username, newUser.Name, newUser.Password, newUser.Type, newUser.Subtitle, "")
	if err != nil {
		log.Println("[ERROR] Create User err: ", err)
		w.Write([]byte("Failed to create the user!"))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write([]byte("created"))
	w.WriteHeader(http.StatusCreated)
}

// check if some environment variables were declared and if they did define
// the vars with their contents
func getEnvVars() (port string, rdis *redisStruct, es *esStruct, dbServer *db.DbServer) {

	rdis = &redisStruct{}
	es = &esStruct{}
	dbServer = &db.DbServer{}

	if os.Getenv("CHAT_PORT") != "" {
		port = os.Getenv("CHAT_PORT")
	} else {
		port = ":8080"
	}

	if os.Getenv("REDIS_ADDR") != "" {
		(*rdis).addr = os.Getenv("REDIS_ADDR")
	} else {
		(*rdis).addr = "chatserver:6379"
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
		es.hosts = "http://chatserver:9200"
	}

	if os.Getenv("DB_HOST") != "" {
		dbServer.Host = os.Getenv("DB_HOST")
	} else {
		dbServer.Host = "chatserver"
	}
	if os.Getenv("DB_PORT") != "" {
		dbServer.Port = os.Getenv("DB_PORT")
	} else {
		dbServer.Port = "3306"
	}
	if os.Getenv("DB_USER") != "" {
		dbServer.User = os.Getenv("DB_USER")
	} else {
		dbServer.User = "root"
	}
	if os.Getenv("DB_PASSWORD") != "" {
		dbServer.Password = os.Getenv("DB_PASSWORD")
	} else {
		dbServer.Password = "password"
	}
	if os.Getenv("DB_DATABASE") != "" {
		dbServer.Database = os.Getenv("DB_DATABASE")
	} else {
		dbServer.Database = "telemedicine"
	}

	return port, rdis, es, dbServer
}

// remove a connection from list
func removeConnection(conn []*websocket.Conn, i int) []*websocket.Conn {
	if i < len(conn) {
		conn[i] = conn[len(conn)-1]
		return conn[:len(conn)-1]
	}
	return conn
}
