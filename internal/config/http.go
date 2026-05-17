package config

import "fmt"

type HTTP struct {
	Scheme string
	Host   string
	Port   string
}

func (h *HTTP) BaseURL() string {
	if h.Port == "" {
		return fmt.Sprintf("%s://%s", h.Scheme, h.Host)
	}
	return fmt.Sprintf("%s://%s:%s", h.Scheme, h.Host, h.Port)
}
