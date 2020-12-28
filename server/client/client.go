package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 16384
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

type ConnectionState int32

const(
	INITIAL ConnectionState = 0
	CONNECTED = 1
	DISCONNECTED = 2
)

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	hub *Hub

	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte

	state ConnectionState

	room string
}

type RoomJoin struct {
	// The Client which wants to join
	client *Client

	// Join this Room
	roomId string
}

func NewClient(hub *Hub, conn *websocket.Conn) Client {
	return Client{hub: hub, conn: conn, send: make(chan []byte, 16384), state: INITIAL}
}

// readPump pumps messages from the websocket connection to the hub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
/*func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		var sb strings.Builder
		sb.Write(bytes.TrimSpace(bytes.Replace(message, newline, space, -1)))
		sb.WriteByte('\n')
		message = []byte(sb.String())
		c.hub.broadcast <- message
	}
}*/

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)
			fmt.Printf("Sent Message: %s\n", message)

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) ReadPump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		c.conn.Close()
		c.hub.unregister <- c
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		/**/
		switch c.state {
		case INITIAL:
			c.handleInitialMessage(message)
			break
		case CONNECTED:
			var sb strings.Builder
			sb.Write(bytes.TrimSpace(bytes.Replace(message, newline, space, -1)))
			sb.WriteByte('\n')
			msg := []byte(sb.String())
			c.hub.broadcast[c.room] <-msg
			if c.readMessages(ticker) {
				return
			}
			break
		case DISCONNECTED:

		}
	}
}

func (c *Client) readMessages(ticker *time.Ticker) bool {
	select {
	case message, ok := <-c.send:
		c.conn.SetWriteDeadline(time.Now().Add(writeWait))
		if !ok {
			// The hub closed the channel.
			c.conn.WriteMessage(websocket.CloseMessage, []byte{})
			return true
		}

		w, err := c.conn.NextWriter(websocket.TextMessage)
		if err != nil {
			return true
		}
		w.Write(message)
		fmt.Printf("Sent Message: %s\n", message)

		// Add queued chat messages to the current websocket message.
		n := len(c.send)
		for i := 0; i < n; i++ {
			w.Write(newline)
			w.Write(<-c.send)
		}

		if err := w.Close(); err != nil {
			return true
		}
	case <-ticker.C:
		c.conn.SetWriteDeadline(time.Now().Add(writeWait))
		if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
			return true
		}
	}
	return false
}

func (c *Client) handleInitialMessage(message []byte) {
	var m MessageWrapper
	err := json.Unmarshal(message, &m)
	var rj JoinRoomMessage
	json.Unmarshal(m.Message, &rj)
	if err != nil || rj.RoomId == "" {
		errorMessage := ErrorMessage{Type: InvalidMessage, Description: "failed to register before sending messages!"}
		errorMessage.writeError(c.conn)
	} else {
		c.hub.register <- &RoomJoin{client: c, roomId: rj.RoomId}
		c.state = CONNECTED
		sendMessageWrapper(c.conn, MessageWrapper{Type: JoinRoomSuccessType})
		c.room = rj.RoomId
		fmt.Println("joined!")
	}
}