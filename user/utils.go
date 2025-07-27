package user

// UserAccount represents a user account with basic profile info
type UserAccount struct {
	Name       string `json:"name"`
	ID         string `json:"id"`
	ProfilePic string `json:"profile_pic"`
	Email      string `json:"email"`
}
