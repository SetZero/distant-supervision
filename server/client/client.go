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
	Initial          ConnectionState = 0
	WaitingForStream                 = 1
	Streamer                         = 2
	Viewer                           = 4
	Disconnected                     = 5
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

	webRTCStreamer *rtc.WebRTCStreamer
	webRTCViewer   *rtc.WebRTCViewer
}

type RoomJoin struct {
	// The Client which wants to join
	client *Client

	// Join this Room
	roomId string
}

func NewClient(hub *Hub, conn *websocket.Conn) Client {
	return Client{hub: hub, conn: conn, send: make(chan []byte, 16384), recv: make(chan []byte, 16384), state: Initial}
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
			if c.handleMessages(ok, message) {
				return
			}
		case <-ticker.C:
			if c.tickWebsocketPing() {
				return
			}
		}
	}
}

func (c *Client) handleMessages(ok bool, message []byte) bool {
	defer c.mu.Unlock()
	c.mu.Lock()

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
	return false
}

func (c *Client) tickWebsocketPing() bool {
	defer c.mu.Unlock()
	c.mu.Lock()

	c.conn.SetWriteDeadline(time.Now().Add(writeWait))
	if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
		return true
	}
	return false
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
		case Initial:
			c.handleInitialMessage(message)
		case WaitingForStream:
			c.handleStreamerMessage(message)
		case Streamer:
			c.recv <- message
		case Viewer:
			c.recv <- message
		case Disconnected:

		}
	}
}

func (c *Client) getStateObject() rtc.WebRtcClient {
	if c.state == Viewer {
		return c.webRTCViewer
	} else {
		return c.webRTCStreamer
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

			c.getStateObject().Recv() <- message
		case message := <-c.getStateObject().Send():
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
			c.setStateToViewer()
		} else {
			c.state = WaitingForStream
		}

		c.room = rj.RoomId
		fmt.Println("joined!")
	}
}

func (c *Client) setStateToViewer() {
	if c.state != Viewer && c.state != Streamer {
		c.state = Viewer
		c.webRTCViewer = rtc.NewWebRTCViewer()
		go c.readMessages(time.NewTicker(pingPeriod))
	}
}

func (c *Client) setStateToStreamer() {
	if c.state != Streamer {
		c.state = Streamer
		c.webRTCStreamer = rtc.NewWebRTCStreamer()
		go c.readMessages(time.NewTicker(pingPeriod))
		go c.pumpToViewer()
	}
}

func (c *Client) handleStreamerMessage(message []byte) {
	defer c.mu.Unlock()
	c.mu.Lock()
	
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
			c.setStateToStreamer()
			go c.webRTCStreamer.Start()
			for client := range c.hub.rooms[c.room].clients {
				m, _ := json.Marshal(StreamerMessage{RoomHasStreamer: true})
				sendMessageWrapper(client.conn, MessageWrapper{Type: JoinRoomSuccessType, Message: m})
				if client != c {
					client.setStateToViewer()
					go client.webRTCViewer.Start()
				}
			}
		} else {
			m, _ := json.Marshal(StartStreamInfoMessage{StreamStartSuccess: false})
			sendMessageWrapper(c.conn, MessageWrapper{Type: StartStreamType, Message: m})
		}
		break
	}
}

func (c *Client) pumpToViewer() {
	for {
		if c.webRTCStreamer != nil && c.webRTCStreamer.WebRtcStream != nil {
			pkg := <-c.webRTCStreamer.WebRtcStream
			for client := range c.hub.rooms[c.room].clients {
				if client != c && client.state == Viewer {
					client.webRTCViewer.WebRtcStream <- pkg
				}
			}
		}
	}
}
