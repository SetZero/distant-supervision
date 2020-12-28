package client

/*type RoomMessage struct {
	room              string
	broadcastMessages chan []byte
}*/

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered clients.
	rooms map[string]map[*Client]bool

	// Inbound messages from the clients to rooms.
	broadcast map[string]chan []byte

	// Register requests from the clients.
	register chan *RoomJoin

	// Unregister requests from clients.
	unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		broadcast:  make(map[string]chan []byte),
		register:   make(chan *RoomJoin),
		unregister: make(chan *Client),
		rooms:      make(map[string]map[*Client]bool),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case roomJoin := <-h.register:
			if h.rooms[roomJoin.roomId] == nil {
				h.rooms[roomJoin.roomId] = make(map[*Client]bool)
			}

			h.rooms[roomJoin.roomId][roomJoin.client] = true
			/*case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}*/
		}

		for broadcastRoom, broadcastChannel := range h.broadcast {
			select {
			case roomMessages := <-broadcastChannel:
				for client := range h.rooms[broadcastRoom] {
					select {
					case client.send <- roomMessages:
					default:
						close(client.send)
						delete(h.rooms[broadcastRoom], client)
					}
				}
			}
		}
	}
}
