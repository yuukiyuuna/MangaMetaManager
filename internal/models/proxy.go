package models

import "gorm.io/gorm"

type ProxySettings struct {
	gorm.Model
	Enabled        bool   `json:"enabled"`
	Type           string `json:"type"` // http, https, socks5
	Host           string `json:"host"`
	Port           int    `json:"port"`
	Username       string `json:"username"`
	Password       string `json:"-"` // Hidden in JSON
	NoProxy        string `json:"noProxy"`
	TimeoutSeconds int    `json:"timeoutSeconds"`
}

type ProviderProxyStrategy struct {
	gorm.Model
	ProviderID string `json:"providerId" gorm:"uniqueIndex"`
	Strategy   string `json:"strategy"` // inherit, disabled, custom
	// Custom settings if strategy is custom
	Type           string `json:"type"`
	Host           string `json:"host"`
	Port           int    `json:"port"`
	Username       string `json:"username"`
	Password       string `json:"-"`
	TimeoutSeconds int    `json:"timeoutSeconds"`
}
