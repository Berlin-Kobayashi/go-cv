package main

import (
	"log"
	"net/http"

	"encoding/json"
	"github.com/gorilla/websocket"
)

type Message struct {
	Type string `json:"type"`
	Data string `json:"data"`
}

var upgrader = websocket.Upgrader{} // use default options

func echo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Accept, Authorization, User-Agent, Connection")
	w.Header().Set("Content-Type", "application/json")

	log.Print("got request:")
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}

		var msg Message
		err = json.Unmarshal(message, &msg)
		if err != nil {
			log.Println("unmarshal:", err)
			break
		}

		msg.Type = "frame"
		newMsg, err := json.Marshal(msg)
		if err != nil {
			log.Println("marshal:", err)
			break
		}
		log.Printf("recv: %s", message)
		err = c.WriteMessage(mt, newMsg)
		if err != nil {
			log.Println("write:", err)
			break
		}
	}
}

func main() {
	upgrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}
	log.SetFlags(0)
	http.HandleFunc("/", echo)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
