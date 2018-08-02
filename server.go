package main

import (
	"log"
	"net/http"

	"encoding/base64"
	"encoding/json"
	"github.com/gorilla/websocket"
	"gocv.io/x/gocv"
	"strings"
)

type Message struct {
	Type string `json:"type"`
	Data string `json:"data"`
}

var upgrader = websocket.Upgrader{} // use default options

func socketHandler(w http.ResponseWriter, r *http.Request) {
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

		log.Printf("recv: %s", message)

		var msg Message
		err = json.Unmarshal(message, &msg)
		if err != nil {
			log.Println("unmarshal:", err)
			break
		}

		if msg.Type == "img" {

			encodedData := strings.Split(msg.Data, ",")[1]

			base64data, err := base64.StdEncoding.DecodeString(encodedData)
			if err != nil {
				log.Println("base64 decode:", err)
				break
			}

			img, err := gocv.IMDecode(base64data, gocv.IMReadColor)
			if err != nil {
				log.Println("decode:", err)
				break
			}

			encoded, err := gocv.IMEncode(".jpg", img)
			if err != nil {
				log.Println("encode:", err)
				break
			}

			base64Encoded := base64.StdEncoding.EncodeToString(encoded)

			msg.Data = "data:image/jpeg;base64," + base64Encoded

			msg.Type = "frame"
			newMsg, err := json.Marshal(msg)
			if err != nil {
				log.Println("marshal:", err)
				break
			}
			err = c.WriteMessage(mt, newMsg)
			if err != nil {
				log.Println("write:", err)
				break
			}
		}
	}
}

func main() {
	upgrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}
	log.SetFlags(0)
	http.HandleFunc("/", socketHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
