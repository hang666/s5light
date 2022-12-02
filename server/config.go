package server

import (
	"fmt"

	"github.com/spf13/viper"
)

var Accounts []*AccountStruct

var tcp_timeout_default = 60
var udp_timeout_default = 60

type AccountStruct struct {
	Username     string   `yaml:"username"  mapstructure:"username"`
	Password     string   `yaml:"password"  mapstructure:"password"`
	BindAddress  string   `yaml:"bind_address"  mapstructure:"bind_address"`
	ReqAddress   string   `yaml:"req_address"  mapstructure:"req_address"`
	Whitelist    []string `yaml:"whitelist"  mapstructure:"whitelist"`
	TCPTimeout   int      `yaml:"tcp_timeout"  mapstructure:"tcp_timeout"`
	UDPTimeout   int      `yaml:"udp_timeout" mapstructure:"udp_timeout"`
	WhitelistMap WhitelistMapType
}

type WhitelistMapType map[string]bool

var customConfigPath string = ""

func SetConfigPath(path string) {
	customConfigPath = path
}

func ReadConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	if customConfigPath == "" {
		viper.AddConfigPath("/etc/s5light/")
		viper.AddConfigPath("$HOME/.s5light")
		viper.AddConfigPath(".")
	} else {
		viper.SetConfigFile(customConfigPath)
	}
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}
	viper.UnmarshalKey("accounts", &Accounts)

	for _, acc := range Accounts {
		if acc.TCPTimeout == 0 {
			acc.TCPTimeout = tcp_timeout_default
		}
		if acc.UDPTimeout == 0 {
			acc.UDPTimeout = udp_timeout_default
		}
		wMap := make(map[string]bool)
		for _, w := range acc.Whitelist {
			if w != "" {
				wMap[w] = true
			}
		}
		acc.WhitelistMap = wMap
	}
}
