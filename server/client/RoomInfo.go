package client

import (
	"github.com/pion/webrtc"
	"sync"
)

type RoomInfo struct {
	roomName    string
	clients     map[*Client]bool
	streamer    *Client
	mu          sync.Mutex
	VideoStream *webrtc.TrackLocalStaticRTP
	AudioStream *webrtc.TrackLocalStaticRTP
}

func (ri *RoomInfo) GetClientsInRoom() int {
	if ri == nil {
		return 1
	}

	clients := 1
	if ri.clients != nil {
		clients = len(ri.clients)
	}
	if ri.streamer != nil {
		clients += 1
	}
	return clients
}