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

type AddBulkUsersRequest struct {
	TrafficGB      float64 `json:"traffic_gb"`
	ExpirationDays int     `json:"expiration_days"`
	Count          int     `json:"count"`
	Prefix         string  `json:"prefix"`
	StartNumber    int     `json:"start_number"`
	Unlimited      bool    `json:"unlimited"`
}

type EditUserRequest struct {
	NewUsername       *string `json:"new_username,omitempty"`
	NewPassword       *string `json:"new_password,omitempty"`
	NewTrafficLimit   *int    `json:"new_traffic_limit,omitempty"`
	NewExpirationDays *int    `json:"new_expiration_days,omitempty"`
	RenewPassword     bool    `json:"renew_password"`
	RenewCreationDate bool    `json:"renew_creation_date"`
	Blocked           *bool   `json:"blocked,omitempty"`
	UnlimitedIP       *bool   `json:"unlimited_ip,omitempty"`
	Note              *string `json:"note,omitempty"`
}

type UsernamesRequest struct {
	Usernames []string `json:"usernames"`
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

type ServerServicesStatusResponse struct {
	HysteriaServer      bool `json:"hysteria_server"`
	HysteriaWebpanel    bool `json:"hysteria_webpanel"`
	HysteriaIPLimit     bool `json:"hysteria_iplimit"`
	HysteriaNormalSub   bool `json:"hysteria_normal_sub"`
	HysteriaTelegramBot bool `json:"hysteria_telegram_bot"`
	HysteriaWarp        bool `json:"hysteria_warp"`
}

type VersionInfoResponse struct {
	CurrentVersion string  `json:"current_version"`
	CoreVersion    *string `json:"core_version"`
}

type VersionCheckResponse struct {
	IsLatest       bool   `json:"is_latest"`
	CurrentVersion string `json:"current_version"`
	LatestVersion  string `json:"latest_version"`
	Changelog      string `json:"changelog"`
}

type GetPortResponse struct {
	Port int `json:"port"`
}

type GetSNIResponse struct {
	SNI string `json:"sni"`
}

type GetObfsResponse struct {
	Obfs string `json:"obfs"`
}

type GetMasqueradeStatusResponse struct {
	Status string `json:"status"`
}

type ConfigFile map[string]interface{}

type IPLimitConfig struct {
	BlockDuration *int `json:"block_duration,omitempty"`
	MaxIPs        *int `json:"max_ips,omitempty"`
}

type IPLimitConfigResponse struct {
	BlockDuration *int `json:"block_duration"`
	MaxIPs        *int `json:"max_ips"`
}

type SetupDecoyRequest struct {
	Domain    string `json:"domain"`
	DecoyPath string `json:"decoy_path"`
}

type DecoyStatusResponse struct {
	Active bool    `json:"active"`
	Path   *string `json:"path"`
}

type ConfigureWarpRequest struct {
	All             bool `json:"all"`
	PopularSites    bool `json:"popular_sites"`
	DomesticSites   bool `json:"domestic_sites"`
	BlockAdultSites bool `json:"block_adult_sites"`
}

type WarpStatusResponse struct {
	AllTrafficViaWarp    bool `json:"all_traffic_via_warp"`
	PopularSitesViaWarp  bool `json:"popular_sites_via_warp"`
	DomesticSitesViaWarp bool `json:"domestic_sites_via_warp"`
	BlockAdultContent    bool `json:"block_adult_content"`
}

type TelegramStartRequest struct {
	Token          string `json:"token"`
	AdminID        string `json:"admin_id"`
	BackupInterval *int   `json:"backup_interval,omitempty"`
}

type BackupIntervalResponse struct {
	BackupInterval *int `json:"backup_interval"`
}

type SetIntervalRequest struct {
	BackupInterval int `json:"backup_interval"`
}

type DomainPortRequest struct {
	Domain string `json:"domain"`
	Port   int    `json:"port"`
}

type EditSubPathRequest struct {
	Subpath string `json:"subpath"`
}

type GetSubPathResponse struct {
	Subpath *string `json:"subpath"`
}

type IPStatusResponse struct {
	IPv4 *string `json:"ipv4"`
	IPv6 *string `json:"ipv6"`
}

type EditIPRequest struct {
	IPv4 *string `json:"ipv4,omitempty"`
	IPv6 *string `json:"ipv6,omitempty"`
}

type Node struct {
	Name      string  `json:"name"`
	IP        string  `json:"ip"`
	Port      *int    `json:"port"`
	SNI       *string `json:"sni"`
	PinSHA256 *string `json:"pinSHA256"`
	Obfs      *string `json:"obfs"`
	Insecure  *bool   `json:"insecure"`
}

type AddNodeRequest struct {
	Name      string  `json:"name"`
	IP        string  `json:"ip"`
	Port      *int    `json:"port"`
	SNI       *string `json:"sni"`
	PinSHA256 *string `json:"pinSHA256"`
	Obfs      *string `json:"obfs"`
	Insecure  *bool   `json:"insecure"`
}

type DeleteNodeRequest struct {
	Name string `json:"name"`
}

type NodeUserTraffic struct {
	Username            string  `json:"username"`
	UploadBytes         int64   `json:"upload_bytes"`
	DownloadBytes       int64   `json:"download_bytes"`
	Status              string  `json:"status"`
	OnlineCount         int     `json:"online_count"`
	AccountCreationDate *string `json:"account_creation_date"`
}

type NodesTrafficPayload struct {
	Users []NodeUserTraffic `json:"users"`
}

type ExtraConfigResponse struct {
	Name string `json:"name"`
	URI  string `json:"uri"`
}

type AddExtraConfigRequest struct {
	Name string `json:"name"`
	URI  string `json:"uri"`
}

type DeleteExtraConfigRequest struct {
	Name string `json:"name"`
}

type HTTPValidationError struct {
	Detail []ValidationError `json:"detail"`
}

type ValidationError struct {
	Loc  []interface{} `json:"loc"`
	Msg  string        `json:"msg"`
	Type string        `json:"type"`
}
