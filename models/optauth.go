package models

type OAuthLoginRequest struct {
	FirebaseToken string `json:"firebase_token" binding:"required"`
	Provider      string `json:"provider" binding:"required"` // "google" or "facebook"
}

type OAuthLoginResponse struct {
	Token           string `json:"token"`
	UserID          string `json:"user_id"`
	Email           string `json:"email"`
	ProfileComplete bool   `json:"profile_complete"`
}
