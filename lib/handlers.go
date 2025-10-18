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
	"os"
	"os/exec"
	"sync/atomic"

	"github.com/disintegration/imaging"
)

func HandleStatus(conn io.ReadWriter, srvProps *ServerProps, config *YAMLConfig) error {
	// 3️⃣ Build JSON response
	response := map[string]interface{}{
		"version": map[string]interface{}{
			"name":     config.ServerVersion,
			"protocol": 767,                  // 1.21.1 protocol version (adjust if needed)
		},
		"description": map[string]interface{}{
			"text": config.MOTD,
		},
	}

	imgFile, err := os.Open(config.SleepingIcon)
	if err == nil {
		defer imgFile.Close()
		img, _, err := image.Decode(imgFile)
		if err == nil {
			resized := imaging.Resize(img, 64, 64, imaging.Lanczos)
			var buf bytes.Buffer
			png.Encode(&buf, resized)
			b64 := base64.StdEncoding.EncodeToString(buf.Bytes())
			response["favicon"] = "data:image/png;base64," + b64
		}
	}

	data, _ := json.Marshal(response)

	// 5️⃣ Write Response Packet (ID 0x00)
	var pkt bytes.Buffer
	WriteVarInt(&pkt, 0x00)           // Packet ID
	WriteVarInt(&pkt, int32(len(data))) // Length of JSON string
	pkt.Write(data)

	var final bytes.Buffer
	WriteVarInt(&final, int32(pkt.Len())) // Total packet length
	final.Write(pkt.Bytes())

	if _, err := conn.Write(final.Bytes()); err != nil {
		return fmt.Errorf("failed to write status response: %w", err)
	}

	return nil
}

func readLoginStart(r *bufio.Reader) (string, error) {
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


func sendLoginMessage(conn io.Writer, msg string) error {
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

func HandleLogin(r *bufio.Reader, w io.Writer, yamlCfg *YAMLConfig, serverRunningPointer *int32) error {
	username, err := readLoginStart(r) // <- uses the same buffered reader
	if err != nil {
		return fmt.Errorf("login error: %v", err)
	}

	fmt.Println("Player attempting login:", username)
    
	if !CanWake(username, yamlCfg) {
		return sendLoginMessage(w, "You are not whitelisted to wake the server!")
	} else {
		sendLoginMessage(w, "Server is starting, please reconnect in a moment!")
		go func() {
			if err := startMinecraftServer("server.jar", "4G"); err != nil {
				fmt.Println("Failed to start server:", err)
			}
			atomic.StoreInt32(serverRunningPointer, 1)
		}()
	}

	return nil
}


