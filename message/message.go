package message

const (
	PlatformYoutube = "Youtube"
	PlatformTwitch  = "Twitch"
)

type Message struct {
	UID    string
	Author string

	Text      string
	Timestamp string

	Platform string
}
