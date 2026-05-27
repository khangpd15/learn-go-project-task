package realtime

import (

	"encoding/json"
	"log"
	"task_api/internal/events"

)

type Hub struct {
	clients    map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan events.Event
}
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan events.Event),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			log.Println("[WS] client connected")

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				log.Println("[WS] client disconnected")
			}

		case event := <-h.broadcast:
			data, err := json.Marshal(event)
			if err != nil {
				log.Println("[WS] failed to marshal event:", err)
				continue
			}

			for client := range h.clients {
				select {
				case client.send <- data:
				default:
					delete(h.clients, client)
					close(client.send)
				}
			}
		}
	}
}

func (h *Hub) Broadcast(event events.Event) {
	h.broadcast <- event
}