package pubsub

import (
	"encoding/json"
	"fmt"

	"github.com/gorilla/websocket"
)

var (
	PUBLISH     = "publish"
	SUBSCRIBE   = "subscribe"
	UNSUBSCRIBE = "unsubscribe"
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

func (connectionList *PubSub) GetSubscriptions(topic string, client *Client) []Subscription {
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

func (connectionList *PubSub) Publish(topic string, message []byte, excludeClient *Client) {
	subscriptions := connectionList.GetSubscriptions(topic, nil)

	for _, sub := range subscriptions {

		fmt.Printf("✅ Publishing to client id %s message is %s\n", sub.Client.ID, message)
		sub.Client.Send(message)
	}
}

func (connectionList *PubSub) Subscribe(client *Client, topic string) *PubSub {

	clientSubs := connectionList.GetSubscriptions(topic, client)
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

func (client *Client) Send(message []byte) error {
	err := client.Connection.WriteMessage(1, message)

	return err
}

func (connectionList *PubSub) Unsubscribe(client *Client, topic string) *PubSub {
	for index, sub := range connectionList.Subscriptions {
		if sub.Client.ID == client.ID && sub.Topic == topic {
			// found client subscription

			connectionList.Subscriptions = append(connectionList.Subscriptions[:index], connectionList.Subscriptions[index+1:]...)
		}
	}

	return connectionList
}

func (connectionList *PubSub) RemoveClient(client Client) *PubSub {
	// remove all client subscriptions
	for index, sub := range connectionList.Subscriptions {
		if client.ID == sub.Client.ID {
			connectionList.Subscriptions = append(connectionList.Subscriptions[:index], connectionList.Subscriptions[index+1:]...)
		}
	}

	// remove client
	for index, iclient := range connectionList.Clients {
		if iclient.ID == client.ID {
			connectionList.Clients = append(connectionList.Clients[:index], connectionList.Clients[index+1:]...)
		}
	}

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
		connectionList.Publish(msg.Topic, msg.Message, nil)
		fmt.Println("✅ New PUBLISH to topic: ", msg.Topic)
		fmt.Println("== Total of subscriptions: ", len(connectionList.Subscriptions))
		break

	case SUBSCRIBE:
		connectionList.Subscribe(&client, msg.Topic)
		fmt.Println("✅ New SUBSCRIBE to topic: ", msg.Topic)
		fmt.Println("== Total of subscriptions: ", len(connectionList.Subscriptions))
		break

	case UNSUBSCRIBE:
		connectionList.Unsubscribe(&client, msg.Topic)
		fmt.Printf("✅ Client %s UNSUBSCRIBED of topic %s \n", client.ID, msg.Topic)
		break

	default:
		break
	}

	return connectionList
}
