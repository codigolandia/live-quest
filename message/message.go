// message is a common package to provide an abtraction over the chat interface
// of different platforms.
package message

import "time"

const (
	PlatformYoutube = "Youtube"
	PlatformTwitch  = "Twitch"
)

// Message is a set of attributes of a viewer message.
// It's commont o all supported platforms.
type Message struct {
	// UID is a unique identifier that is platform dependent.
	// It is not guaranteed to be unique across platforms.
	UID string

	// Author is the display name of the author of the message.
	// as provided by the platform.
	Author string

	// Text is the display text of the message as provided by the platform.
	Text string
	// Timestamp when the message was received.
	Timestamp time.Time

	// Platform is the name of the source platform.
	Platform string
}
