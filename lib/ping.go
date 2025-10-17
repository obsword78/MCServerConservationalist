package lib

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"os"
)

// PingResponse structure for modern Minecraft clients
type PingResponse struct {
	Version            Version     `json:"version"`
	Players            Players     `json:"players"`
	Description        Description `json:"description"`
	Favicon            string      `json:"favicon,omitempty"`
	EnforcesSecureChat bool        `json:"enforcesSecureChat"`
}

type Version struct {
	Name     string `json:"name"`
	Protocol int    `json:"protocol"`
}

type Players struct {
	Max    int           `json:"max"`
	Online int           `json:"online"`
	Sample []PlayerSample `json:"sample,omitempty"`
}

type PlayerSample struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

type Description struct {
	Text string `json:"text"`
}

// SendPing dynamically fetches player count and sends JSON ping
func SendJSONPing(conn net.Conn, motd, iconPath string, online, maxPlayers int) {
	resp := PingResponse{
		Version: Version{
			Name:     "1.21.8",
			Protocol: 772,
		},
		Players: Players{
			Max:    maxPlayers,
			Online: online,
			Sample: nil, // empty: we don’t fake player names
		},
		Description: Description{
			Text: motd,
		},
		EnforcesSecureChat: false,
	}

	if iconPath != "" {
		data, err := os.ReadFile(iconPath)
		if err == nil {
			resp.Favicon = "data:image/png;base64," + base64.StdEncoding.EncodeToString(data)
		}
	}

	jsonBytes, err := json.Marshal(resp)
	if err != nil {
		fmt.Println("Failed to marshal ping response:", err)
		return
	}

	conn.Write(jsonBytes)
}
