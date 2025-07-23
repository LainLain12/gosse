package chat

// ProfilePicMessage represents a public chat message with profile picture and metadata
type ProfilePicMessage struct {
	ID         string `json:"id"`
	Message    string `json:"message"`
	ProfilePic string `json:"profile_pic"`
	CreatedAt  string `json:"created_at"`
}
