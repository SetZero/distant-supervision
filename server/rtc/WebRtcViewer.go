package rtc

import (
	"encoding/json"
	"fmt"
	"github.com/pion/rtp"
	"github.com/pion/webrtc"
)

type WebRTCViewer struct {
	peerConnection *webrtc.PeerConnection
	send           chan []byte
	recv           chan []byte
	WebRtcStream   chan *rtp.Packet
	outputTrack    *webrtc.TrackLocalStaticRTP
}

func NewWebRTCViewer() *WebRTCViewer {
	connectionInfo, err := createPeerConnection(false)
	if err == nil && connectionInfo != nil {
		rtc := &WebRTCViewer{send: make(chan []byte, 16384), recv: make(chan []byte, 16384), peerConnection: connectionInfo.peerConnection, WebRtcStream: make(chan *rtp.Packet), outputTrack: connectionInfo.outputTrack}
		return rtc
	} else {
		return nil
	}
}

func (r *WebRTCViewer) Send() chan []byte {
	return r.send
}

func (r *WebRTCViewer) Recv() chan []byte {
	return r.recv
}

func (r *WebRTCViewer) Start() {

	r.peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		fmt.Printf("[Viewer] Connection State has changed %s \n", connectionState.String())
		if connectionState == webrtc.ICEConnectionStateFailed ||
			connectionState == webrtc.ICEConnectionStateDisconnected {
			fmt.Println("TODO: Close stuff")
		}

		if connectionState == webrtc.ICEConnectionStateConnected {
			go func() {
				for connectionState != webrtc.ICEConnectionStateFailed &&
					connectionState != webrtc.ICEConnectionStateDisconnected {
					stream := <-r.WebRtcStream
					err := r.outputTrack.WriteRTP(stream)
					if err != nil {
						fmt.Println("[Viewer] Error: ", err)
					}
				}
			}()
		}
	})

	for {
		webRTCMessage := <-r.recv

		var m messageWrapper
		err := json.Unmarshal(webRTCMessage, &m)
		if err == nil {
			switch m.Type {
			case webRTCOffer:
				var offerMessage webRtcOffer
				json.Unmarshal(m.Message, &offerMessage)
				offer := webrtc.SessionDescription{}
				Decode(offerMessage.Offer, &offer)
				err = r.peerConnection.SetRemoteDescription(offer)
				if err != nil {
					fmt.Println("Error: ", err)
				}
				answer, err := r.peerConnection.CreateAnswer(nil)
				if err != nil {
					panic(err)
				}
				jsonAnswer, err := json.Marshal(answer)
				if err != nil {
					panic(err)
				}
				r.send <- jsonAnswer

				if err = r.peerConnection.SetLocalDescription(answer); err != nil {
					panic(err)
				}
				break
			case iceCandidate:
				var iceandidate webrtc.ICECandidateInit
				json.Unmarshal(m.Message, &iceandidate)
				err := r.peerConnection.AddICECandidate(iceandidate)
				if err != nil {
					panic(err)
				}
				break
			}
		} else {
			fmt.Println("Error: ", err)
		}
	}
}
