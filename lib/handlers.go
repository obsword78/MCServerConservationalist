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
	"path/filepath"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/disintegration/imaging"
)

func HandleStatus(conn io.ReadWriter, state *ProgramState) error {
	// 3️⃣ Build JSON response
	response := map[string]interface{}{
		"version": map[string]interface{}{
			"name":     state.YAMLConfig.ServerVersion,
			"protocol": 767,                  // 1.21.1 protocol version (adjust if needed)
		},
		"description": map[string]interface{}{
			"text": state.YAMLConfig.MOTD,
		},
	}

	imgFile, err := os.Open(filepath.Join(state.ExeDir, state.YAMLConfig.SleepingIcon))
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

func startMinecraftServer(jarPath, ram string, state *ProgramState) error {
    cmd := exec.Command("java", "-Xmx"+ram, "-jar", filepath.Join(state.ExeDir, jarPath), "--nogui")
	cmd.Dir = state.ExeDir
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr

	if state.PortListener != nil {
		fmt.Println("\033[1;32mPort closed.\033[0m")
		state.PortListener.Close()
		state.PortListener = nil
	}

    state.MinecraftProcess = cmd
    if err := cmd.Start(); err != nil {
        return err
    }

    go func() {
        err := cmd.Wait()
        if err != nil {
            fmt.Println("Minecraft server crashed:", err)
        } else {
            fmt.Println("Minecraft server stopped gracefully")
        }
        atomic.StoreInt32(state.ServerRunning, 0)

		if cmd.ProcessState.ExitCode() != 0 {
			p, _ := os.FindProcess(os.Getpid())
   			p.Signal(syscall.SIGINT)
		}
    }()

	for {
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("localhost:%d", state.ServerProps.RconPort), 2*time.Second)
		if err == nil {
			fmt.Println("\033[1;32mServer is ready!\033[0m")
			conn.Close()
			break
		}
		if atomic.LoadInt32(state.ServerRunning) == 0 {
			fmt.Println("\033[1;31mServer process ended before RCON became available\033[0m")
			conn.Close()
			break
		}
		time.Sleep(time.Second)
	}
	

    return nil
}

func HandleLogin(state *ProgramState, r *bufio.Reader, w io.Writer) error {
    username, err := readLoginStart(r)
    if err != nil {
        return fmt.Errorf("login error: %v", err)
    }

    if !CanWake(username, state) {
        return sendLoginMessage(w, "You are not whitelisted!")
    }

    sendLoginMessage(w, "Server starting… please reconnect in a moment")

    go func() {
		atomic.StoreInt32(state.ServerRunning, 1)
        if err := startMinecraftServer("server.jar", "4G", state); err != nil {
            fmt.Println("\033[1;31mFailed to start server:", err, "\033[0m")
        }

        select {
        case state.ServerStarted <- struct{}{}:
        default:
        }
    }()

    return nil
}

