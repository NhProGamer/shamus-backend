package entities

// Channel represents the WebSocket message channel
// Channels define the direction and purpose of messages
type Channel string

const (
	// Server → Client channels
	ChannelNotification Channel = "notification" // Informational, no response expected
	ChannelPrompt       Channel = "prompt"       // Request for player action with timeout

	// Client → Server channels
	ChannelResponse Channel = "response" // Response to a Prompt
	ChannelCommand  Channel = "command"  // Free-form player command (chat, settings, etc.)
)

// IsServerToClient returns true if the channel is used for server-to-client messages
func (c Channel) IsServerToClient() bool {
	return c == ChannelNotification || c == ChannelPrompt
}

// IsClientToServer returns true if the channel is used for client-to-server messages
func (c Channel) IsClientToServer() bool {
	return c == ChannelResponse || c == ChannelCommand
}
