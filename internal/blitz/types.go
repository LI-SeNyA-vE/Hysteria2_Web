package blitz

type AddUserRequest struct {
	Username       string  `json:"username"`
	Password       *string `json:"password,omitempty"`
	TrafficLimit   int     `json:"traffic_limit"`
	ExpirationDays int     `json:"expiration_days"`
	Unlimited      bool    `json:"unlimited"`
}

type DetailResponse struct {
	Detail string `json:"detail"`
}

type UserInfo struct {
	Username         string `json:"username"`
	Password         string `json:"password"`
	MaxDownloadBytes int64  `json:"max_download_bytes"`
	ExpirationDays   int    `json:"expiration_days"`
	Blocked          bool   `json:"blocked"`
	UploadBytes      *int64 `json:"upload_bytes"`
	DownloadBytes    *int64 `json:"download_bytes"`
	OnlineCount      int    `json:"online_count"`
}

type HTTPValidationError struct {
	Detail []ValidationError `json:"detail"`
}

type ValidationError struct {
	Loc  []interface{} `json:"loc"`
	Msg  string        `json:"msg"`
	Type string        `json:"type"`
}
