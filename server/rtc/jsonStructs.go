package rtc

import "encoding/json"

type MessageType string

const(
	webRTCOffer MessageType = "webRtcOffer"
	iceCandidate = "iceCandidate"
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