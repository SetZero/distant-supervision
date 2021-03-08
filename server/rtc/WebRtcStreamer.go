package rtc

import (
	"../messages"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/pion/rtcp"
	"github.com/pion/webrtc"
	"io"
	"strings"
	"time"
)

type WebRTCStreamer struct {
	peerConnection    *webrtc.PeerConnection
	send              chan OutputMessage
	recv              chan []byte
	WebRtcVideoStream chan *webrtc.TrackLocalStaticRTP
	WebRtcAudioStream chan *webrtc.TrackLocalStaticRTP
	track             *webrtc.TrackRemote
	currentBitrate    uint64
}

func NewWebRTCStreamer() *WebRTCStreamer {
	connectionInfo, err := createPeerConnection(true)
	if err == nil && connectionInfo != nil {
		rtc := &WebRTCStreamer{send: make(chan OutputMessage, 8192),
			recv:              make(chan []byte, 8192),
			peerConnection:    connectionInfo.peerConnection,
			WebRtcVideoStream: make(chan *webrtc.TrackLocalStaticRTP),
			WebRtcAudioStream: make(chan *webrtc.TrackLocalStaticRTP)}
		return rtc
	} else {
		return nil
	}
}

func (r *WebRTCStreamer) Send() chan OutputMessage {
	return r.send
}

func (r *WebRTCStreamer) Recv() chan []byte {
	return r.recv
}

func (r *WebRTCStreamer) Start() {
	if _, err := r.peerConnection.AddTransceiverFromKind(webrtc.RTPCodecTypeVideo); err != nil {
		panic(err)
	}

	if _, err := r.peerConnection.AddTransceiverFromKind(webrtc.RTPCodecTypeVideo); err != nil {
		panic(err)
	}


	r.peerConnection.OnTrack(func(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
		fmt.Println("Track: ", track)
		r.track = track
		go func() {
			ticker := time.NewTicker(3 * time.Second)
			for range ticker.C {
				if writeErr := r.peerConnection.WriteRTCP([]rtcp.Packet{&rtcp.PictureLossIndication{MediaSSRC: uint32(track.SSRC())}}); writeErr != nil {
					fmt.Println(writeErr)
				}
				// Send a remb message with a very high bandwidth to trigger chrome to send also the high bitrate stream
				if writeErr := r.peerConnection.WriteRTCP([]rtcp.Packet{&rtcp.ReceiverEstimatedMaximumBitrate{Bitrate: r.currentBitrate, SenderSSRC: uint32(track.SSRC())}}); writeErr != nil {
					fmt.Println(writeErr)
				}
			}
		}()

		var localTrack *webrtc.TrackLocalStaticRTP

		fmt.Printf("Track has started, of type %d: %s \n", track.PayloadType(), track.Codec().MimeType)
		if strings.HasPrefix(track.Codec().MimeType, "video") {
			localTrack = r.setupTrack(track, r.WebRtcVideoStream, "video")
		} else if strings.HasPrefix(track.Codec().MimeType, "audio") {
			localTrack = r.setupTrack(track, r.WebRtcAudioStream, "audio")
		}

		rtpBuf := make([]byte, 1500)
		for {
			i, _, readErr := track.Read(rtpBuf)
			if readErr != nil {
				panic(readErr)
			}

			// ErrClosedPipe means we don't have any subscribers, this is ok if no peers have connected yet
			if _, err := localTrack.Write(rtpBuf[:i]); err != nil && !errors.Is(err, io.ErrClosedPipe) {
				panic(err)
			}
		}
	})

	r.peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		fmt.Printf("[Streamer] Connection State has changed %s \n", connectionState.String())
		if connectionState == webrtc.ICEConnectionStateFailed ||
			connectionState == webrtc.ICEConnectionStateDisconnected {
			fmt.Println("Done writing media files")
		}
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
		if err == nil {
			switch m.Type {
			case WebRTCOffer:
				fmt.Println("check")
				var offerMessage webRtcOffer
				json.Unmarshal(m.Message, &offerMessage)
				offer := webrtc.SessionDescription{}
				Decode(offerMessage.Offer, &offer)
				err = r.peerConnection.SetRemoteDescription(offer)
				if err != nil {
					fmt.Println("[Streamer] Error: ", err)
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
			case BitrateChangeType:
				var bitrateOffer BitrateChangeOffer
				json.Unmarshal(m.Message, &bitrateOffer)
				if bitrateOffer.Bitrate > 0 && r.track != nil {
					r.currentBitrate = bitrateOffer.Bitrate * 1024
					fmt.Println("Changed Bitrate to: ", bitrateOffer.Bitrate*1024)
				}
				break
			}
		} else {
			fmt.Println("Error: ", err)
		}
	}
}

func (r *WebRTCStreamer) setupTrack(track *webrtc.TrackRemote, input chan *webrtc.TrackLocalStaticRTP, trackId string) *webrtc.TrackLocalStaticRTP {
	localTrack, newTrackErr := webrtc.NewTrackLocalStaticRTP(track.Codec().RTPCodecCapability, trackId, "pion")
	if newTrackErr != nil {
		panic(newTrackErr)
	}
	fmt.Println("Sending Track: ", trackId)
	input <- localTrack
	return localTrack
}
