package config

import "strings"

func SubscriptionPublicBase() string {
	return Get().SubscriptionPublicBase()
}

func SubscriptionURL(token string) string {
	return Get().SubscriptionURL(token)
}

func SubscriptionPath() string {
	return Get().SubscriptionPath()
}

func UsingLocalSubscriptionURL() bool {
	return Get().UsingLocalSubscriptionURL()
}

func (c Config) SubscriptionPublicBase() string {
	if v := strings.TrimSpace(c.SubDomain); v != "" {
		return strings.TrimRight(v, "/")
	}
	return c.LocalHTTPBase()
}

func (c Config) LocalHTTPBase() string {
	addr := c.HTTPAddr
	if addr == "" {
		addr = Default().HTTPAddr
	}
	if strings.HasPrefix(addr, ":") {
		return "http://127.0.0.1" + addr
	}
	if strings.HasPrefix(addr, "0.0.0.0") {
		return "http://" + strings.Replace(addr, "0.0.0.0", "127.0.0.1", 1)
	}
	if strings.Contains(addr, "://") {
		return strings.TrimRight(addr, "/")
	}
	return "http://" + strings.TrimRight(addr, "/")
}

func (c Config) LocalHealthURL() string {
	return c.LocalHTTPBase() + "/healthz"
}

func (c Config) SubscriptionPath() string {
	if c.SubPath == "" {
		return defaultSubPath
	}
	return c.SubPath
}

func (c Config) SubscriptionURL(token string) string {
	return c.SubscriptionPublicBase() + "/" + c.SubscriptionPath() + "/" + token
}

func (c Config) UsingLocalSubscriptionURL() bool {
	return strings.TrimSpace(c.SubDomain) == ""
}
