package client

import (
	"../rtc"
	"encoding/json"
	"fmt"
	"log"
	"sync"
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

const (
	INITIAL            ConnectionState = 0
	WAITING_FOR_STREAM                 = 1
	CONNECTED                          = 2
	DISCONNECTED                       = 3
)

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	hub *Hub

	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte

	recv chan []byte

	state ConnectionState

	room string

	mu sync.Mutex // todo

	webRTC *rtc.WebRTC
}

type RoomJoin struct {
	// The Client which wants to join
	client *Client

	// Join this Room
	roomId string
}

func NewClient(hub *Hub, conn *websocket.Conn) Client {
	return Client{hub: hub, conn: conn, send: make(chan []byte, 16384), recv: make(chan []byte, 16384), state: INITIAL}
}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *Client) WritePump() {
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

		switch c.state {
		case INITIAL:
			c.handleInitialMessage(message)
			break
		case WAITING_FOR_STREAM:
			c.handleStreamerMessage(message)
		case CONNECTED:
			c.recv <- message
			break
		case DISCONNECTED:

		}
	}
}

func (c *Client) readMessages(ticker *time.Ticker) bool {
	for {
		select {
		case message, ok := <-c.recv:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return true
			}
			c.webRTC.Recv <- message
		case message := <-c.webRTC.Send:
			sendMessageWrapper(c.conn, MessageWrapper{Type: AnswerType, Message: message})
			break
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return true
			}
		}
	}
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
		if c.hub.rooms[rj.RoomId] != nil && c.hub.rooms[rj.RoomId].streamer != nil {
			c.setStateToConnected()
		} else {
			c.state = WAITING_FOR_STREAM
		}

		c.room = rj.RoomId
		fmt.Println("joined!")
	}
}

func (c *Client) setStateToConnected() {
	if c.state != CONNECTED {
		c.state = CONNECTED
		go c.readMessages(time.NewTicker(pingPeriod))
	}
}

func (c *Client) handleStreamerMessage(message []byte) {
	var m MessageWrapper
	err := json.Unmarshal(message, &m)
	if err != nil {
		fmt.Println("There was an error while trying to decode message")
		return
	}
	switch m.Type {
	case StartStreamType:
		defer c.hub.rooms[c.room].mu.Unlock()
		c.hub.rooms[c.room].mu.Lock()

		if c.hub.rooms[c.room].streamer == nil {
			c.hub.rooms[c.room].streamer = c
			m, _ := json.Marshal(StartStreamInfoMessage{StreamStartSuccess: true})
			sendMessageWrapper(c.conn, MessageWrapper{Type: StartStreamType, Message: m})
			c.webRTC = rtc.NewWebRTC()
			go c.webRTC.Start()
			for client := range c.hub.rooms[c.room].clients {
				m, _ := json.Marshal(StreamerMessage{RoomHasStreamer: true})
				sendMessageWrapper(client.conn, MessageWrapper{Type: JoinRoomSuccessType, Message: m})
				c.setStateToConnected()
			}
		} else {
			m, _ := json.Marshal(StartStreamInfoMessage{StreamStartSuccess: false})
			sendMessageWrapper(c.conn, MessageWrapper{Type: StartStreamType, Message: m})
		}
		break
	}
}
