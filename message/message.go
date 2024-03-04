// message is a common package to provide an abtraction over the chat interface
// on different platforms.
package message

import "time"

const (
	PlatformYoutube = "Youtube"
	PlatformTwitch  = "Twitch"
)

// Message is a set of attributes of a viewer message.
// It's common to all supported platforms.
type Message struct {

	// UID is a unique identifier that is platform dependent.
	// It is not guaranteed to be unique across platforms.
	UID string `json:"uid"`

	// Author is the messages's author display name, as provided
	// by the platform.
	Author string `json:"author"`

	// Text is the message value as provided by the platform.
	Text string `json:"text"`
	// Timestamp when the message was received.
	Timestamp time.Time `json:"timestamp"`

	// Platform is the name of the source platform.
	Platform string `json:"platform"`
}
