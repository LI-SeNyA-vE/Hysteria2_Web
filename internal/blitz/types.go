package blitz

type AddUserRequest struct {
	Username       string  `json:"username"`
	Password       *string `json:"password,omitempty"`
	TrafficLimit   int     `json:"traffic_limit"`
	ExpirationDays int     `json:"expiration_days"`
	CreationDate   *string `json:"creation_date,omitempty"`
	Unlimited      bool    `json:"unlimited"`
	Note           *string `json:"note,omitempty"`
}

type DetailResponse struct {
	Detail string `json:"detail"`
}

type UserInfo struct {
	Username            string  `json:"username"`
	Password            string  `json:"password"`
	MaxDownloadBytes    int64   `json:"max_download_bytes"`
	ExpirationDays      int     `json:"expiration_days"`
	AccountCreationDate *string `json:"account_creation_date"`
	Blocked             bool    `json:"blocked"`
	UnlimitedUser       bool    `json:"unlimited_user"`
	Note                *string `json:"note"`
	Status              *string `json:"status"`
	UploadBytes         *int64  `json:"upload_bytes"`
	DownloadBytes       *int64  `json:"download_bytes"`
	OnlineCount         int     `json:"online_count"`
}

type UserURIResponse struct {
	Username  string    `json:"username"`
	IPv4      *string   `json:"ipv4"`
	IPv6      *string   `json:"ipv6"`
	Nodes     []NodeURI `json:"nodes"`
	NormalSub *string   `json:"normal_sub"`
	Error     *string   `json:"error"`
}

type NodeURI struct {
	Name string `json:"name"`
	URI  string `json:"uri"`
}

type ServerStatusResponse struct {
	Uptime                  string `json:"uptime"`
	BootTime                string `json:"boot_time"`
	ServerIPv4              string `json:"server_ipv4"`
	ServerIPv6              string `json:"server_ipv6"`
	CPUUsage                string `json:"cpu_usage"`
	RAMUsage                string `json:"ram_usage"`
	TotalRAM                string `json:"total_ram"`
	OnlineUsers             int    `json:"online_users"`
	UploadSpeed             string `json:"upload_speed"`
	DownloadSpeed           string `json:"download_speed"`
	TCPConnections          int    `json:"tcp_connections"`
	UDPConnections          int    `json:"udp_connections"`
	RebootUploadedTraffic   string `json:"reboot_uploaded_traffic"`
	RebootDownloadedTraffic string `json:"reboot_downloaded_traffic"`
	RebootTotalTraffic      string `json:"reboot_total_traffic"`
	UserUploadedTraffic     string `json:"user_uploaded_traffic"`
	UserDownloadedTraffic   string `json:"user_downloaded_traffic"`
	UserTotalTraffic        string `json:"user_total_traffic"`
}

type HTTPValidationError struct {
	Detail []ValidationError `json:"detail"`
}

type ValidationError struct {
	Loc  []interface{} `json:"loc"`
	Msg  string        `json:"msg"`
	Type string        `json:"type"`
}
