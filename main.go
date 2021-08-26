package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/viniciusmazon/lab-go-websocket-pubsub/pubsub"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var connectionList = &pubsub.PubSub{}

func autoID() string {
	return uuid.Must(uuid.NewUUID()).String()
}

func handler(writer http.ResponseWriter, request *http.Request) {
	upgrader.CheckOrigin = func(request *http.Request) bool {
		return true
	}

	connection, err := upgrader.Upgrade(writer, request, nil)
	if err != nil {
		log.Print(err)
	}

	client := pubsub.Client{
		ID:         autoID(),
		Connection: connection,
	}

	connectionList.AddClient(client)

	fmt.Println("== New client is connected with id: ", client.ID, "Total de ", len(connectionList.Clients), " clients")
	for {
		messageType, payload, err := connection.ReadMessage()
		if err != nil {
			log.Println("💀 Oops... something went wrong")
			return
		}

		connectionList.HandleReceiveMessage(client, messageType, payload)
	}
}

func main() {
	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		http.ServeFile(writer, request, "public/static/")
	})

	http.HandleFunc("/ws", handler)

	fmt.Println("🛰️  Server is running on http://localhost:3000")
	http.ListenAndServe(":3000", nil)
}
