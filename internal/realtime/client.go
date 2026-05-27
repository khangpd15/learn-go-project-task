package realtime

import (
	"log"

	"github.com/gorilla/websocket"
)

type Client struct {
	hub  *Hub
	conn *websocket.Conn
	send chan []byte
}

func (c *Client) ReadPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			log.Println("[WS] read error:", err)
			break
		}
	}
}

func (c *Client) WritePump() {
	defer c.conn.Close()

	for message := range c.send {
		err := c.conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			log.Println("[WS] write error:", err)
			break
		}
	}
}