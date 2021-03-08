package rtc

import (
	"../messages"
	"encoding/json"
	"fmt"
	"github.com/pion/webrtc"
)

type WebRTCViewer struct {
	peerConnection    *webrtc.PeerConnection
	send              chan OutputMessage
	recv              chan []byte
	WebRtcVideoStream chan *webrtc.TrackLocalStaticRTP
	WebRtcAudioStream chan *webrtc.TrackLocalStaticRTP
	outputTrack       *webrtc.TrackLocalStaticRTP
}

func NewWebRTCViewer() *WebRTCViewer {
	connectionInfo, err := createPeerConnection(false)
	if err == nil && connectionInfo != nil {
		rtc := &WebRTCViewer{send: make(chan OutputMessage, 8192), recv: make(chan []byte, 8192),
			peerConnection:    connectionInfo.peerConnection,
			WebRtcVideoStream: make(chan *webrtc.TrackLocalStaticRTP),
			WebRtcAudioStream: make(chan *webrtc.TrackLocalStaticRTP),
			outputTrack:       connectionInfo.outputTrack}
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
	// TODO: There is another channel which is likely blocking...
	go r.startStream(<-r.WebRtcVideoStream, "video")
	go r.startOptionalStream(r.WebRtcAudioStream, "audio")

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
		if candidate == nil {
			return
		}
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
		fmt.Println("[Viewer] got message of type: ", m.Type)
		if err == nil {
			switch m.Type {
			case WebRTCOffer:
				fmt.Println("Got offer from viewer")
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
					fmt.Errorf("Error: %s\n", err)
					r.peerConnection.Close()
					return
				}
				jsonAnswer, err := json.Marshal(answer)
				if err != nil {
					fmt.Errorf("Error: %s\n", err)
					r.peerConnection.Close()
					return
				}
				r.send <- OutputMessage{Data: jsonAnswer, Type: messages.AnswerType}

				if err = r.peerConnection.SetLocalDescription(answer); err != nil {
					fmt.Errorf("Error: %s\n", err)
					r.peerConnection.Close()
					return
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

func (r *WebRTCViewer) startStream(track *webrtc.TrackLocalStaticRTP, trackId string) {
	fmt.Println("Track; ", track)
	rtpSender, videoErr := r.peerConnection.AddTrack(track)
	fmt.Printf("[Viewer | %s] Track exists!\n", trackId)
	if videoErr != nil {
		fmt.Errorf("Error: %s\n", videoErr)
		r.peerConnection.Close()
		return
	}
	go func() {
		rtcpBuf := make([]byte, 1500)
		for {
			if _, _, rtcpErr := rtpSender.Read(rtcpBuf); rtcpErr != nil {
				return
			}
		}
	}()
}

func (r *WebRTCViewer) startOptionalStream(stream chan *webrtc.TrackLocalStaticRTP, trackId string) {
	r.startStream(<-stream, trackId)
}
