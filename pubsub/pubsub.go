package pubsub

import (
	"encoding/json"
	"fmt"

	"github.com/gorilla/websocket"
)

var (
	PUBLISH   = "publish"
	SUBSCRIBE = "subscribe"
)

type PubSub struct {
	Clients       []Client
	Subscriptions []Subscription
}

type Client struct {
	ID         string
	Connection *websocket.Conn
}

type Message struct {
	Action  string          `json:"action"`
	Topic   string          `json:"topic"`
	Message json.RawMessage `json:"message"`
}

type Subscription struct {
	Topic  string
	Client *Client
}

func (connectionList *PubSub) GetSubscription(topic string, client *Client) []Subscription {
	var subscriptionList []Subscription

	for _, subscription := range connectionList.Subscriptions {
		if client != nil {
			if subscription.Client.ID == client.ID && subscription.Topic == topic {
				subscriptionList = append(subscriptionList, subscription)
			}
		} else {
			if subscription.Topic == topic {
				subscriptionList = append(subscriptionList, subscription)
			}
		}
	}

	return subscriptionList
}

func (connectionList *PubSub) Subscribe(client *Client, topic string) *PubSub {

	clientSubs := connectionList.GetSubscription(topic, client)
	if len(clientSubs) > 0 {
		// The client has already been registered in this topic
		return connectionList
	}

	newSubscription := Subscription{
		Topic:  topic,
		Client: client,
	}

	connectionList.Subscriptions = append(connectionList.Subscriptions, newSubscription)

	return connectionList
}

func (connectionList *PubSub) AddClient(client Client) *PubSub {
	connectionList.Clients = append(connectionList.Clients, client)

	fmt.Println("\u200d Adding a new client to list of clients")

	payload := []byte("Hello Client ID " + client.ID)
	client.Connection.WriteMessage(1, payload)
	return connectionList
}

func (connectionList *PubSub) HandleReceiveMessage(client Client, messageType int, payload []byte) *PubSub {
	msg := Message{}

	err := json.Unmarshal(payload, &msg)
	if err != nil {
		fmt.Println("💀 oops... there's something wrong with the payload")
		return connectionList
	}

	switch msg.Action {
	case PUBLISH:
		fmt.Println("This is publish new message")
		break

	case SUBSCRIBE:
		connectionList.Subscribe(&client, msg.Topic)
		fmt.Println("✅ New Subscribe to topic: ", msg.Topic)
		fmt.Println("== Total of subscriptions: ", len(connectionList.Subscriptions))
		break

	default:
		break
	}

	return connectionList
}
