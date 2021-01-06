package rtc

import (
	"../messages"
	"encoding/json"
	"fmt"
	"github.com/pion/webrtc"
)

type WebRTCViewer struct {
	peerConnection *webrtc.PeerConnection
	send           chan OutputMessage
	recv           chan []byte
	WebRtcStream   chan *webrtc.TrackLocalStaticRTP
	outputTrack    *webrtc.TrackLocalStaticRTP
}

func NewWebRTCViewer() *WebRTCViewer {
	connectionInfo, err := createPeerConnection(false)
	if err == nil && connectionInfo != nil {
		rtc := &WebRTCViewer{send: make(chan OutputMessage, 16384), recv: make(chan []byte, 16384), peerConnection: connectionInfo.peerConnection, WebRtcStream: make(chan *webrtc.TrackLocalStaticRTP), outputTrack: connectionInfo.outputTrack}
		return rtc
	} else {
		return nil
	}
}

func (r *WebRTCViewer) Send() chan OutputMessage {
	return r.send
}

func (r *WebRTCViewer) Recv() chan []byte {
	return r.recv
}

func (r *WebRTCViewer) Start() {

	rtpSender, err := r.peerConnection.AddTrack(<-r.WebRtcStream)
	if err != nil {
		panic(err)
	}
	go func() {
		rtcpBuf := make([]byte, 1500)
		for {
			if _, _, rtcpErr := rtpSender.Read(rtcpBuf); rtcpErr != nil {
				return
			}
		}
	}()

	r.peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		fmt.Printf("[Viewer] Connection State has changed %s \n", connectionState.String())
		if connectionState == webrtc.ICEConnectionStateFailed ||
			connectionState == webrtc.ICEConnectionStateDisconnected {
			fmt.Println("TODO: Close stuff")
		}

		if connectionState == webrtc.ICEConnectionStateConnected {
			stats, _ := json.Marshal(r.peerConnection.GetStats())
			fmt.Println("Connected... ", string(stats))
		}
	})

	r.peerConnection.OnTrack(func(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
		fmt.Println("Got Track")
	})

	r.peerConnection.OnICECandidate(func(candidate *webrtc.ICECandidate) {
		fmt.Println("Canidate: ", candidate)
		if candidate == nil { return }
		 iceCandidate, err := json.Marshal(candidate.ToJSON())
		if err != nil || iceCandidate == nil {
			println("Error with Ice canidate")
			return
		}
		r.send <- OutputMessage{Data: iceCandidate, Type: messages.IceCandidate}
	})

	for {
		webRTCMessage := <-r.recv

		var m messageWrapper
		err := json.Unmarshal(webRTCMessage, &m)
		if err == nil {
			switch m.Type {
			case WebRTCOffer:
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
				r.send <- OutputMessage{Data: jsonAnswer, Type: messages.AnswerType}

				if err = r.peerConnection.SetLocalDescription(answer); err != nil {
					panic(err)
				}
				break
			case IceCandidate:
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
