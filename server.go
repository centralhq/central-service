package main

import (
	"fmt"
	"log"
	"net/http"
	"github.com/gorilla/websocket"
	"github.com/google/uuid"
)

type EngineServer struct {
	config *EngineConfig
	hub *Hub
	handler *OperationManager
}

type CentralObject struct {
	OperationType 	string		`json:"operationType"`
	CentralUid		string		`json:"uuId"`
	ObjectId		string		`json:"objectId"` //constructed in server
	MutationCounter uint64 		`json:"mutationCounter"`
	ConflictId 		string  	`json:"conflictId"`
	IsDeleted 		bool		`json:"isDeleted"`
	Payload 		interface{}	`json:"payload"`
}

type AckOperation struct {
	Status 			string 		`json:"status"`
	OperationType	string		`json:"operationType"`
	CentralUid		string		`json:"uuId"`
	ObjectId		string		`json:"objectId"`
	MutationCounter uint64 		`json:"mutationCounter"`
	ConflictId 		string		`json:"conflictId"`
	Payload 		interface{} `json:"payload"`
}

var WsMessageType = 1

func NewEngineServer(handler *OperationManager, config *EngineConfig, hub *Hub) *EngineServer {
    return &EngineServer{
		config: config,
		hub: hub,
		handler: handler,
	}
}

func (s *EngineServer) Run() {
	fmt.Println("Hello WebSocket")
	s.Handler()
	log.Fatal(http.ListenAndServe(":8080", nil))
}

var upgrader = websocket.Upgrader{
	ReadBufferSize: 1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool { return true },
}

func (s *EngineServer) Handler() {
	http.HandleFunc("/", s.wsEndpoint)
}

func (s *EngineServer) wsEndpoint(w http.ResponseWriter, r *http.Request) {
	// upgrade this connection to a WebSocket connection

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	// There will be many instances of Client, so it shouldn't be in the DI container
	uuid := uuid.NewString()
	client := NewClient(s.hub, conn, s.handler, uuid) // possibly store sessionId here
	client.hub.register <- client
	// problem: the uuid is sent to the user the first time, but not properly stored.
	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	go client.readPump()

	log.Println("Client Connected")
}

