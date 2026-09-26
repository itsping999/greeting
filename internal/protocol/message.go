package protocol

import "regexp"

const (
	MQTTRequestTopic   = "greeting/requests"
	MQTTReplyPrefix    = "greeting/replies/"
	RabbitExchange     = "greeting"
	RabbitRequestKey   = "greeting.request"
	RabbitReplyPrefix  = "greeting.reply."
	RabbitRequestQueue = "greeting.requests"
)

var requestIDPattern = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,64}$`)

func ValidRequestID(id string) bool { return requestIDPattern.MatchString(id) }

type GreetingRequest struct {
	RequestID string `json:"request_id"`
	Name      string `json:"name"`
}
type GreetingReply struct {
	RequestID string `json:"request_id"`
	Message   string `json:"message,omitempty"`
	ID        int64  `json:"id,omitempty"`
	Error     string `json:"error,omitempty"`
}

// ChatMessage is a plain JSON WebSocket message. Type 1 requests a greeting.
type ChatMessage struct {
	Type    int    `json:"type"`
	Name    string `json:"name,omitempty"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}
