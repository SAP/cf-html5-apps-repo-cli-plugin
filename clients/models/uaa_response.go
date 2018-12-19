package models

// UAAResponse XSUAA response with token
type UAAResponse struct {
	AccessToken string `json:"access_token,omitempty"`
	TokenType   string `json:"token_type,omitempty"`
	ExpiresIn   int    `json:"expires_in,omitempty"`
	Scope       string `json:"scope,omitempty"`
	JTI         string `json:"jti,omitempty"`
}
