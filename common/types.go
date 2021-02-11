package common

type Addr struct {
	Isdn bool
	Host string
	Port string
}

type Strategy struct {
	DNS           string `json:"dns,omitempty"`        // "8.8.8.8:53"
	Next          string `json:"next,omitempty"`       // "1.1.1.1:443"
	Block         string `json:"block,omitempty"`      // "true","false"
	IPRange       string `json:"ip_range,omitempty"`   // "192.168.1.1/28"
	PortRange     string `json:"port_range,omitempty"` //80, 443, 8080-8082
	DomainPrefix  string `json:"domain_prefix,omitempty"`
	DomainSuffix  string `json:"domain_suffix,omitempty"`
	DomainContain string `json:"domain_contain,omitempty"`
}

type ServerConfig struct {
	LogLevel string `json:"log,omitempty"`
	//
	Addr        string `json:"listen,omitempty"`
	ServiceName string `json:"service_name,omitempty"`
	//
	TLS     string `json:"tls,omitempty"`
	TLSCert string `json:"tls_cert,omitempty"`
	TLSKey  string `json:"tls_key,omitempty"`
}

type ClientConfig struct {
	LogLevel string `json:"log,omitempty"`
	//
	Local       string `json:"local,omitempty"`
	Remote      string `json:"remote,omitempty"`
	ServiceName string `json:"service_name,omitempty"`
	//
	TLS       string `json:"tls,omitempty"`
	TLSCA     string `json:"tls_ca,omitempty"`
	TLSSNI    string `json:"tls_sni,omitempty"`
	TLSVerify string `json:"tls_verify,omitempty"`
	//
	Compress string `json:"compress,omitempty"`
	//
	StrategyGroup []*Strategy `json:"strategy,omitempty"`
}
