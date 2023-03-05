package redis

import (
	"github.com/p4gefau1t/trojan-go/config"
)

type RedisConfig struct {
	Enabled    bool   `json:"enabled" yaml:"enabled"`
	ServerHost string `json:"server_addr" yaml:"server-addr"`
	ServerPort int    `json:"server_addr" yaml:"server-addr"`
	CheckRate  int    `json:"check_rate" yaml:"check-rate"`
}

type Config struct {
	Redis RedisConfig `json:"redis" yaml:"redis"`
}

func init() {
	config.RegisterConfigCreator(Name, func() interface{} {
		return &Config{
			Redis: RedisConfig{
				ServerHost: "",
				ServerPort: 6379,
				CheckRate:  30,
			},
		}
	})
}
