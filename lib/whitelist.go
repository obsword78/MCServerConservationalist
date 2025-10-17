package lib

import (
	"encoding/json"
	"os"
	"strings"
)

type WhitelistEntry struct {
	UUID string `json:"uuid"`
	Name string `json:"name"`
}
//====================================================
func CanWake(player string, cfg *YAMLConfig) bool {
	player = strings.ToLower(player)
	if cfg.UseWhiteListJSON {
		allowed := readWhitelist("whitelist.json")
		return allowed[player]
	} else {
		for _, name := range cfg.WakeWhiteList {
			if strings.ToLower(name) == player {
				return true
			}
		}
		return false
	}
}
//====================================================
func readWhitelist(path string) map[string]bool {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var entries []WhitelistEntry
	json.Unmarshal(file, &entries)

	allowed := make(map[string]bool)
	for _, e := range entries {
		allowed[strings.ToLower(e.Name)] = true
	}
	return allowed
}