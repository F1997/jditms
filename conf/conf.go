package conf

import (
	"fmt"
	"net"
	"os"
	"strings"

	"jditms/alert/aconf"
	"jditms/center/cconf"
	"jditms/pkg/cfg"
	"jditms/pkg/httpx"
	"jditms/pkg/logx"
	"jditms/pkg/ormx"
	"jditms/pushgw/pconf"
	"jditms/storage"
)

type ConfigType struct {
	Global    GlobalConfig
	Log       logx.Config
	HTTP      httpx.Config
	DB        ormx.DBConfig
	Redis     storage.RedisConfig
	CenterApi CenterApi

	Pushgw pconf.Pushgw
	Alert  aconf.Alert
	Center cconf.Center
	Ibex   Ibex
}

type CenterApi struct {
	Addrs         []string
	BasicAuthUser string
	BasicAuthPass string
	Timeout       int64
}

type GlobalConfig struct {
	RunMode string
}

type Ibex struct {
	Enable    bool
	RPCListen string
	Output    Output
}

type Output struct {
	ComeFrom string
	AgtdPort int
}

func InitConfig(configDir, cryptoKey string) (*ConfigType, error) {
	var config = new(ConfigType)

	if err := cfg.LoadConfigByDir(configDir, config); err != nil {
		return nil, fmt.Errorf("failed to load configs of directory: %s error: %s", configDir, err)
	}

	config.Pushgw.PreCheck()
	config.Alert.PreCheck(configDir)
	config.Center.PreCheck()

	err := decryptConfig(config, cryptoKey)
	if err != nil {
		return nil, err
	}

	if config.Alert.Heartbeat.IP == "" {
		// auto detect
		config.Alert.Heartbeat.IP = fmt.Sprint(GetOutboundIP())
		if config.Alert.Heartbeat.IP == "" {
			hostname, err := os.Hostname()
			if err != nil {
				fmt.Println("failed to get hostname:", err)
				os.Exit(1)
			}

			if strings.Contains(hostname, "localhost") {
				fmt.Println("Warning! hostname contains substring localhost, setting a more unique hostname is recommended")
			}

			config.Alert.Heartbeat.IP = hostname
		}
	}

	config.Alert.Heartbeat.Endpoint = fmt.Sprintf("%s:%d", config.Alert.Heartbeat.IP, config.HTTP.Port)

	return config, nil
}

func GetOutboundIP() net.IP {
	conn, err := net.Dial("udp", "223.5.5.5:80")
	if err != nil {
		fmt.Println("auto get outbound ip fail:", err)
		return []byte{}
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}
