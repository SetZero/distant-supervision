package rtc

import (
	"github.com/pion/rtp"
	"github.com/pion/webrtc"
)

type WebRTCViewer struct {
	peerConnection *webrtc.PeerConnection
	Send           chan []byte
	Recv           chan []byte
	WebRtcStream   chan *rtp.Packet
}

func NewWebRTCViewer() *WebRTCViewer {
	rtc := &WebRTCViewer{Send: make(chan []byte, 16384), Recv: make(chan []byte, 16384), peerConnection: nil}
	return rtc
}