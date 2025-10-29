package lib

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"strings"
	"time"
)

type RCONClient struct {
	conn net.Conn
}

func NewRCONClient(addr, password string) (*RCONClient, error) {
	conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
	if err != nil {
		conn.Close()
		return nil, err
	}
	if !rconAuth(conn, password) {
		conn.Close()
		return nil, fmt.Errorf("RCON auth failed")
	}
	return &RCONClient{conn: conn}, nil
}

func (r *RCONClient) GetPlayerCount() int {
	out := rconCommand(r.conn, "list")
	var count int
	fmt.Sscanf(out, "There are %d/", &count)
	return count
}

func (r *RCONClient) StopServer() {
	rconCommand(r.conn, "stop")
}

func (r *RCONClient) Close() {
	r.conn.Close()
}

func rconAuth(conn net.Conn, password string) bool {
    _, respID := rconSend(conn, 3, 1, password)
    return respID == 1
}

func rconCommand(conn net.Conn, cmd string) string {
    out, _ := rconSend(conn, 2, 2, cmd)
    return out
}

func rconSend(conn net.Conn, kind, id int32, payload string) (string, int32) {
    length := int32(len(payload) + 9)
    buf := new(bytes.Buffer)
    binary.Write(buf, binary.LittleEndian, length)
    binary.Write(buf, binary.LittleEndian, id)
    binary.Write(buf, binary.LittleEndian, kind)
    buf.WriteString(payload)
    buf.WriteByte(0)
    conn.Write(buf.Bytes())

    resp := make([]byte, 4096)
    n, _ := conn.Read(resp)
    resp = resp[:n]


    if n < 12 {
        return "", 0
    }

    respID := int32(binary.LittleEndian.Uint32(resp[4:8]))
    return strings.TrimRight(string(resp[12:]), "\x00"), respID
}