package lib

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"io"
	"net"
	"os"
	"os/exec"
	"sync/atomic"

	"github.com/disintegration/imaging"
)

func HandleStatus(conn io.ReadWriter, srvProps *ServerProps, config *YAMLConfig) error {
	r := bufio.NewReader(conn)

	_, err := readVarInt(r)
	if err != nil {
		return err
	}
	packetID, _ := readVarInt(r)
	if packetID != 0x00 {
		return fmt.Errorf("unexpected status request packet ID: %d", packetID)
	}

	response := map[string]interface{}{
		"version": map[string]interface{}{
			"name":     fmt.Sprintf("%d", srvProps.ServerPort),
			"protocol": 773,
		},
		"description": map[string]interface{}{
			"text": config.MOTD,
		},
	}
	
	imgFile, err := os.Open(config.SleepingIcon)
	if err != nil {
		response["favicon"] = "" 
	} else {
		defer imgFile.Close()
		img, _, err := image.Decode(imgFile)
		if err != nil {
			response["favicon"] = ""
		} else {
			resized := imaging.Resize(img, 64, 64, imaging.Lanczos)
			var buf bytes.Buffer
			png.Encode(&buf, resized)
			b64 := base64.StdEncoding.EncodeToString(buf.Bytes())
			response["favicon"] = "data:image/png;base64," + b64
		}
	}
	
	data, _ := json.Marshal(response)

	var buf bytes.Buffer
	WriteVarInt(&buf, 0x00)            
	WriteVarInt(&buf, int32(len(data)))
	buf.Write(data)

	var final bytes.Buffer
	WriteVarInt(&final, int32(buf.Len()))
	final.Write(buf.Bytes())

	_, err = conn.Write(final.Bytes())
	if err != nil {
		return err
	}

	return nil
}

func readLoginStart(r io.Reader) (string, error) {
    length, err := readVarInt(r)
    if err != nil {
        return "", err
    }
    packetData := make([]byte, length)
    if _, err := io.ReadFull(r, packetData); err != nil {
        return "", err
    }

    buf := bytes.NewReader(packetData)
    packetID, _ := readVarInt(buf)
    if packetID != 0x00 {
        return "", fmt.Errorf("unexpected login packet ID: %d", packetID)
    }

    username, _ := readString(buf)
    return username, nil
}

func sendLoginMessage(conn net.Conn, msg string) error {
    jsonMsg := fmt.Sprintf(`{"text":"%s"}`, msg)
    msgBytes := []byte(jsonMsg)

    var buf bytes.Buffer
    WriteVarInt(&buf, 0x00)                 // Packet ID
    WriteVarInt(&buf, int32(len(msgBytes))) // String length
    buf.Write(msgBytes)

    var final bytes.Buffer
    WriteVarInt(&final, int32(buf.Len()))
    final.Write(buf.Bytes())

    _, err := conn.Write(final.Bytes())
    return err
}

func startMinecraftServer(jarPath string, ram string) error {
    cmd := exec.Command("java", "-Xmx"+ram, "-jar", jarPath, "--nogui")
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    return cmd.Start()
}


func HandleLogin(conn net.Conn, yamlCfg *YAMLConfig, serverRunningPointer *int32) error {
    username, err := readLoginStart(bufio.NewReader(conn))
    if err != nil {
        return fmt.Errorf("login error: %v", err)
    }

    fmt.Println("Player attempting login:", username)
    
	if !CanWake(username, yamlCfg) {
		return sendLoginMessage(conn, "You are not whitelisted to wake the server!")
	} else {
		sendLoginMessage(conn, "Server is starting, please reconnect in a moment!")
		go func() {
			if err := startMinecraftServer("server.jar", "4G"); err != nil {
				fmt.Println("Failed to start server:", err)
			}
			atomic.StoreInt32(serverRunningPointer, 1)
		}()
	}

	return nil
}

