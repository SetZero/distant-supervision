package rtc

import (
	"../messages"
	"encoding/json"
)

type MessageType string

const (
	WebRTCOffer       MessageType = "webRtcOffer"
	IceCandidate                  = "iceCandidate"
	BitrateChangeType             = "bitrateChange"
)

type messageWrapper struct {
	Type    MessageType     `json:"type"`
	Message json.RawMessage `json:"message"`
}

type webRtcOffer struct {
	Offer string `json:"offer"`
}

type BitrateChangeOffer struct {
	Bitrate uint64 `json:"bitrate"`
}

type webRtcIceCanidate struct {
	Candidate string `json:"candidate"`
}

type OutputMessage struct {
	Data []byte
	Type messages.MessageType
}
