package config

import "github.com/kpture/kpture/pkg/wireshark"

type AppConfig struct {
	*wireshark.Wireshark `mapstructure:"Wireshark"`
}
