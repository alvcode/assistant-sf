package dto

type LoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type LoginSuccessResponse struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}

type ErrorResponse struct {
	Message string `json:"message"`
	Status  int    `json:"status"`
	Code    int    `json:"code"`
}
