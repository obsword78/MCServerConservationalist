package lib

import (
	"fmt"
	"os"

	"github.com/magiconair/properties"
	"gopkg.in/yaml.v3"
)

type YAMLConfig struct {
	MOTD               string   `yaml:"motd"`
	IdleTimeoutSeconds int      `yaml:"idleTimeoutSeconds"`
	UseWhiteListJSON   bool     `yaml:"useWhiteListJson"`
	WakeWhiteList      []string `yaml:"wakeWhiteList"`
	SleepingIcon       string   `yaml:"sleepingIcon"`
	ServerVersion      string   `yaml:"serverVersion"`
	ServerJarPath      string   `yaml:"serverJarPath"`
}

type ServerProps struct {
	ServerPort   int
	RconPort     int
	RconPassword string
}
//=======================================================
func LoadYAMLConfig(path string) *YAMLConfig {
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Println("Config not found, creating default YAML")

		defaultCfg := YAMLConfig{
			MOTD:               "Server is sleeping - join to wake",
			IdleTimeoutSeconds: 30,
			UseWhiteListJSON:   true,
			WakeWhiteList:      []string{},
			SleepingIcon:       "sleeping.png",
			ServerVersion: "Unspecified Version",
			ServerJarPath: "server.jar",
		}

		bytes, _ := yaml.Marshal(defaultCfg)
		os.WriteFile(path, bytes, 0644)
		return &defaultCfg
	}

	var cfg YAMLConfig
	yaml.Unmarshal(data, &cfg)
	return &cfg
}

func LoadServerProps(path string) *ServerProps {
	p := properties.MustLoadFile(path, properties.UTF8)
	props := &ServerProps{
		ServerPort:   p.MustGetInt("server-port"),
		RconPort:     p.MustGetInt("rcon.port"),
		RconPassword: p.GetString("rcon.password", ""),
	}	
	return props
}

func CheckRCONEnabled(path string) bool {
	p := properties.MustLoadFile(path, properties.UTF8)
	return p.GetBool("enable-rcon", false)
}