package messages

type MessageType string

const (
	ErrorMessageType    MessageType = "error"
	JoinMessageType                 = "joinMessage"
	JoinRoomSuccessType             = "joinedMessage"
	StartStreamType                 = "startStream"
	RequestStreamerType             = "requestStreamer"
	AnswerType                      = "answer"
	IceCandidate					= "newIceCandidate"
)
