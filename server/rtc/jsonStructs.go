package rtc

import (
	"../messages"
	"encoding/json"
)

type MessageType string

const(
	WebRTCOffer  MessageType = "webRtcOffer"
	IceCandidate             = "iceCandidate"
)

type messageWrapper struct {
	Type MessageType `json:"type"`
	Message json.RawMessage `json:"message"`
}

type webRtcOffer struct {
	Offer string `json:"offer"`
}

type webRtcIceCanidate struct {
	Candidate string `json:"candidate"`
}

type OutputMessage struct {
	Data []byte
	Type messages.MessageType
}