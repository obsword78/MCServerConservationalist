package lib

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"strings"
	"time"
)

func SendRCONStop(addr, password string) {
	conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
	if err != nil {
		fmt.Println("RCON connect failed:", err)
		return
	}
	defer conn.Close()

	if !rconAuth(conn, password) {
		fmt.Println("RCON auth failed")
		return
	}

	rconCommand(conn, "stop")
}

func GetPlayerCountRCON(addr, password string) int {
	conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
	if err != nil {
		return 0
	}
	defer conn.Close()

	if !rconAuth(conn, password) {
		return 0
	}

	out := rconCommand(conn, "list")
	var count int
	fmt.Sscanf(out, "There are %d/", &count)
	return count
}

func rconAuth(conn net.Conn, password string) bool {
	return rconSend(conn, 3, 1, password) != ""
}

func rconCommand(conn net.Conn, cmd string) string {
	return rconSend(conn, 2, 2, cmd)
}

func rconSend(conn net.Conn, kind, id int32, payload string) string {
	length := int32(len(payload) + 10)
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, length)
	binary.Write(buf, binary.LittleEndian, id)
	binary.Write(buf, binary.LittleEndian, kind)
	buf.WriteString(payload)
	buf.WriteByte(0)
	buf.WriteByte(0)
	conn.Write(buf.Bytes())

	resp := make([]byte, 4096)
	n, _ := conn.Read(resp)
	resp = resp[:n]

	if n < 12 {
		return ""
	}

	return strings.TrimRight(string(resp[12:]), "\x00")
}
