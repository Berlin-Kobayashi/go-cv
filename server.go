package main

import (
	"log"
	"net/http"

	"encoding/base64"
	"encoding/json"
	"github.com/gorilla/websocket"
	"gocv.io/x/gocv"
	"image"
	"strings"
)

type Message struct {
	Type string `json:"type"`
	Data string `json:"data"`
}

var upgrader = websocket.Upgrader{}

func socketHandler(w http.ResponseWriter, r *http.Request) {
	log.Print("got request")
	setResponseHeaders(w.Header(), r.Header.Get("Origin"))

	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	consumeMessages(c)
}

func setResponseHeaders(header http.Header, origin string) {
	header.Set("Access-Control-Allow-Origin", origin)
	header.Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	header.Set("Access-Control-Allow-Credentials", "true")
	header.Set("Access-Control-Allow-Headers", "Content-Type, Accept, Authorization, User-Agent, Connection")
	header.Set("Content-Type", "application/json")
}

func consumeMessages(c *websocket.Conn) {
	for {
		msg, err := receiveMessage(c)
		if err != nil {
			log.Println("receive:", err)
			break
		}

		if msg.Type == "img" {
			img, err := decodeImage(msg.Data)
			if err != nil {
				log.Println("decode:", err)
				break
			}

			sketch := sketchifyImage(img)

			err = sendImage(sketch, c)
			if err != nil {
				log.Println("send:", err)
				break
			}
		}
	}
}

func receiveMessage(c *websocket.Conn) (Message, error) {
	_, message, err := c.ReadMessage()
	if err != nil {
		return Message{}, err
	}

	var msg Message
	err = json.Unmarshal(message, &msg)
	if err != nil {
		return Message{}, err
	}

	return msg, nil
}

func decodeImage(data string) (gocv.Mat, error) {
	encodedData := strings.Split(data, ",")[1]

	base64data, err := base64.StdEncoding.DecodeString(encodedData)
	if err != nil {
		return gocv.Mat{}, err
	}

	return gocv.IMDecode(base64data, gocv.IMReadColor)
}

func sendImage(src gocv.Mat, c *websocket.Conn) error {
	encoded, err := encodeImage(src)
	if err != nil {
		return err
	}

	msg, err := json.Marshal(Message{Type: "frame", Data: encoded})
	if err != nil {
		return err
	}
	err = c.WriteMessage(websocket.TextMessage, msg)
	if err != nil {
		return err
	}

	return nil
}

func encodeImage(src gocv.Mat) (string, error) {
	encoded, err := gocv.IMEncode(".jpg", src)
	if err != nil {
		return "", err
	}

	base64Encoded := base64.StdEncoding.EncodeToString(encoded)

	return "data:image/jpeg;base64," + base64Encoded, nil
}

func sketchifyImage(src gocv.Mat) gocv.Mat {
	gray := src.Clone()
	gocv.CvtColor(src, &gray, gocv.ColorBGRToGray)

	blur := gray.Clone()
	gocv.GaussianBlur(gray, &blur, image.Point{5, 5}, 0, 0, gocv.BorderConstant)

	canny := blur.Clone()
	gocv.Canny(blur, &canny, 10, 70)

	bin := canny.Clone()
	gocv.Threshold(canny, &bin, 70, 255, gocv.ThresholdBinaryInv)

	return bin
}

func main() {
	upgrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}
	log.SetFlags(0)
	http.HandleFunc("/", socketHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
