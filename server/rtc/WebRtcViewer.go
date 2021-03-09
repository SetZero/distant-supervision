package rtc

import (
	"../logger"
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
			WebRtcVideoStream: make(chan *webrtc.TrackLocalStaticRTP, 1),
			WebRtcAudioStream: make(chan *webrtc.TrackLocalStaticRTP, 1),
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

func (r *WebRTCViewer) LateStart(videoStream *webrtc.TrackLocalStaticRTP, audioStream *webrtc.TrackLocalStaticRTP) {
	if videoStream != nil {
		r.WebRtcVideoStream <- videoStream
		if audioStream != nil {
			r.WebRtcAudioStream <- audioStream
		}
		r.Start()
	}
}

func (r *WebRTCViewer) Start() {
	go r.startStream(<-r.WebRtcVideoStream, "video")
	go r.startOptionalStream(r.WebRtcAudioStream, "audio")

	r.peerConnection.OnICEConnectionStateChange(r.onICEConnectionStateChange)
	r.peerConnection.OnTrack(r.onTrack)
	r.peerConnection.OnICECandidate(r.onICECandidate)

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
	rtpSender, videoErr := r.peerConnection.AddTrack(track)
	logger.InfoLogger.Printf("[Viewer | %s] Track exists!\n", trackId)
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

func (r *WebRTCViewer) onTrack(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
	fmt.Println("Got Track")
}

func (r *WebRTCViewer) onICECandidate(candidate *webrtc.ICECandidate) {
	if candidate == nil {
		return
	}
	iceCandidate, err := json.Marshal(candidate.ToJSON())
	if err != nil || iceCandidate == nil {
		println("Error with Ice canidate")
		return
	}
	r.send <- OutputMessage{Data: iceCandidate, Type: messages.IceCandidate}
}

func (r *WebRTCViewer) onICEConnectionStateChange(connectionState webrtc.ICEConnectionState) {
	logger.InfoLogger.Printf("[Viewer] Connection State has changed %s \n", connectionState.String())
	if connectionState == webrtc.ICEConnectionStateFailed ||
		connectionState == webrtc.ICEConnectionStateDisconnected {
		logger.InfoLogger.Printf("[Viewer] Connection State has changed %s \n", connectionState.String())
		err := r.peerConnection.Close()
		if err != nil {
			logger.InfoLogger.Printf("Failed to Close Peer Connection: %s \n", err)
		}
	}

	if connectionState == webrtc.ICEConnectionStateConnected {
		//stats, _ := json.Marshal(r.peerConnection.GetStats())
		logger.InfoLogger.Println("Connected to ICE")
	}
}
