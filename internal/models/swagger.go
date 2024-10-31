// internal/models/swagger.go
package models

// Error representa un error de la API
type Error struct {
	Message string `json:"error" example:"Descripción del error"`
}

// FollowResponse representa la respuesta al seguir a un usuario
type FollowResponse struct {
	Message     string `json:"message" example:"Usuario seguido exitosamente"`
	UserID      string `json:"user_id" example:"123"`
	FollowingID string `json:"following_id" example:"456"`
}

// TimelineResponse representa la respuesta del timeline
type TimelineResponse struct {
	UserID string  `json:"user_id" example:"123"`
	Page   int     `json:"page" example:"1"`
	Limit  int     `json:"limit" example:"10"`
	Count  int     `json:"count" example:"5"`
	Tweets []Tweet `json:"tweets"`
}
