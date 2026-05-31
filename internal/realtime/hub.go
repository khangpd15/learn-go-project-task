package realtime

import (

	"encoding/json"
	"log"
	"task_api/internal/events"

)

type Hub struct {
	clients     map[*Client]bool
	userClients map[int]map[*Client]bool
	register    chan *Client
	unregister  chan *Client
	broadcast   chan events.Event
}

func NewHub() *Hub {
	return &Hub{
		clients:     make(map[*Client]bool),
		userClients: make(map[int]map[*Client]bool),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		broadcast:   make(chan events.Event),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true

			if h.userClients[client.UserID] == nil {
				h.userClients[client.UserID] = make(map[*Client]bool)
			}
			h.userClients[client.UserID][client] = true
           log.Println("[WS] registered user:", client.UserID)
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)

				if _, ok := h.userClients[client.UserID]; ok {
					delete(h.userClients[client.UserID], client)

					if len(h.userClients[client.UserID]) == 0 {
						delete(h.userClients, client.UserID)
					}
				}

				close(client.send)
			}
		case event := <-h.broadcast:
			data, err := json.Marshal(event)
			if err != nil {
				log.Println("[WS] failed to marshal event:", err)
				continue
			}

			targets := map[int]bool{}

			if event.UserID != 0 {
				targets[event.UserID] = true
			}

			for _, userID := range event.UserIDs {
				if userID != 0 {
					targets[userID] = true
				}
			}

			for userID := range targets {
				h.sendToUser(userID, data)
			}
		}
	}
}
func (h *Hub) sendToUser(userID int, data []byte) {
	clients := h.userClients[userID]

	for client := range clients {
		select {
		case client.send <- data:
		default:
			delete(h.clients, client)
			delete(h.userClients[userID], client)
			close(client.send)
		}
	}
	log.Println("[WS] sendToUser:", userID, "clients:", len(clients))
}
func (h *Hub) Broadcast(event events.Event) {
	h.broadcast <- event
	log.Println("[WS] hub received event:", event.Type, "userID:", event.UserID, "userIDs:", event.UserIDs)
}