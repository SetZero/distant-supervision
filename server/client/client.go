package client

import (
	"../messages"
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
	writeWait = 20 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 262144
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

	mu *sync.Mutex // todo

	webRTCStreamer *rtc.WebRTCStreamer
	webRTCViewer   *rtc.WebRTCViewer
}

type RoomJoin struct {
	// The Client which wants to join
	client *Client

	// Join this Room
	roomId string
}

func NewClient(hub *Hub, conn *SafeConnection) Client {
	return Client{hub: hub, conn: conn.Conn, send: make(chan []byte, 8192), recv: make(chan []byte, 8192), state: Initial, mu: &conn.Mu}
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
		c.sendViewerUpdate(c.hub.rooms[c.room])
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
			sendMessageWrapper(c.conn, MessageWrapper{Type: message.Type, Message: message.Data})
			break
		case <-ticker.C:
			if func() bool {
				defer c.mu.Unlock()
				c.mu.Lock()
				c.conn.SetWriteDeadline(time.Now().Add(writeWait))
				if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					return true
				}
				return false
			}() {
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
		c.room = rj.RoomId
		fmt.Println("joined!")
		c.hub.register <- &RoomJoin{client: c, roomId: rj.RoomId}
		if c.hub.rooms[rj.RoomId] != nil && c.hub.rooms[rj.RoomId].streamer != nil {
			c.lateStartClient()
		} else {
			c.state = WaitingForStream
		}
	}
}

func (c *Client) setStateToViewer(roomInfo *RoomInfo) {
	if c.state != Viewer && c.state != Streamer {
		c.state = Viewer
		c.webRTCViewer = rtc.NewWebRTCViewer()
		go c.readMessages(time.NewTicker(pingPeriod))
		c.sendViewerUpdate(roomInfo)
	}
}

func (c *Client) sendViewerUpdate(roomInfo *RoomInfo) {
	defer c.mu.Unlock()
	c.mu.Lock()

	var viewer int
	if roomInfo != nil {
		viewer = len(roomInfo.clients)
	} else {
		return
	}

	m, _ := json.Marshal(ViewerMessage{Viewers: uint32(viewer)})
	for client := range roomInfo.clients {
		sendMessageWrapper(client.conn, MessageWrapper{Type: messages.CurrentViewerUpdate, Message: m})
	}
}

func (c *Client) setStateToStreamer() {
	if c.state != Streamer {
		c.state = Streamer
		c.webRTCStreamer = rtc.NewWebRTCStreamer()
		go c.readMessages(time.NewTicker(pingPeriod))
		c.pumpToViewer()
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
	case messages.StartStreamType:
		defer c.hub.rooms[c.room].mu.Unlock()
		c.hub.rooms[c.room].mu.Lock()

		if c.hub.rooms[c.room].streamer == nil {
			c.hub.rooms[c.room].streamer = c
			m, _ := json.Marshal(StartStreamInfoMessage{StreamStartSuccess: true})
			sendMessageWrapper(c.conn, MessageWrapper{Type: messages.StartStreamType, Message: m})
			c.setStateToStreamer()
			go c.webRTCStreamer.Start()
			c.startClients()
		} else {
			m, _ := json.Marshal(StartStreamInfoMessage{StreamStartSuccess: false})
			sendMessageWrapper(c.conn, MessageWrapper{Type: messages.StartStreamType, Message: m})
		}
		break
	}
}

func (c *Client) startClients() {
	for client := range c.hub.rooms[c.room].clients {
		c.startClient(client)
	}
}

func (c *Client) startClient(client *Client) {
	m, _ := json.Marshal(StreamerMessage{RoomHasStreamer: true})
	sendMessageWrapper(client.conn, MessageWrapper{Type: messages.JoinRoomSuccessType, Message: m})
	if client != c {
		client.setStateToViewer(c.hub.rooms[c.room])
		//go client.webRTCViewer.LateStart(c.hub.rooms[c.room].VideoStream, c.hub.rooms[c.room].AudioStream)
		go client.webRTCViewer.Start()
	}
}
func (c *Client) lateStartClient() {
	m, _ := json.Marshal(StreamerMessage{RoomHasStreamer: true})
	sendMessageWrapper(c.conn, MessageWrapper{Type: messages.JoinRoomSuccessType, Message: m})
	c.setStateToViewer(c.hub.rooms[c.room])
	go c.webRTCViewer.LateStart(c.hub.rooms[c.room].VideoStream, c.hub.rooms[c.room].AudioStream)
}

func (c *Client) pumpToViewer() {
	go func() {
		for {
			if c.webRTCStreamer != nil && c.webRTCStreamer.WebRtcVideoStream != nil {
				c.hub.rooms[c.room].VideoStream = <-c.webRTCStreamer.WebRtcVideoStream
				for client := range c.hub.rooms[c.room].clients {
					if client != c && client.state == Viewer {
						client.webRTCViewer.WebRtcVideoStream <- c.hub.rooms[c.room].VideoStream
					}
				}
			}
		}

	}()

	go func() {
		for {
			if c.webRTCStreamer != nil && c.webRTCStreamer.WebRtcAudioStream != nil {
				c.hub.rooms[c.room].AudioStream = <-c.webRTCStreamer.WebRtcAudioStream
				for client := range c.hub.rooms[c.room].clients {
					if client != c && client.state == Viewer {
						client.webRTCViewer.WebRtcAudioStream <- c.hub.rooms[c.room].AudioStream
					}
				}
			}
		}

	}()
}
