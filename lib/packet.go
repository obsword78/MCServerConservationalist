package lib

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

type Handshake struct {
	ProtocolVersion int32
	Address         string
	Port            uint16
	NextState       int32
}

func readVarInt(r io.Reader) (int32, error) {
    var num int32
    var shift uint
    for {
        var b [1]byte
        if _, err := r.Read(b[:]); err != nil {
            return 0, err
        }
        num |= int32(b[0]&0x7F) << shift
        if (b[0] & 0x80) == 0 {
            break
        }
        shift += 7
        if shift > 35 {
            return 0, fmt.Errorf("VarInt too big")
        }
    }
    return num, nil
}

func WriteVarInt(buf *bytes.Buffer, value int32) {
	for {
		if (value & ^0x7F) == 0 {
			buf.WriteByte(byte(value))
			return
		}
		buf.WriteByte(byte(value&0x7F | 0x80))
		value >>= 7
	}
}

func readUShort(r io.Reader) (uint16, error) {
    var b [2]byte
    if _, err := io.ReadFull(r, b[:]); err != nil {
        return 0, err
    }
    return binary.BigEndian.Uint16(b[:]), nil
}

func readString(r io.Reader) (string, error) {
    length, err := readVarInt(r)
    if err != nil {
        return "", err
    }
    data := make([]byte, length)
    if _, err := io.ReadFull(r, data); err != nil {
        return "", err
    }
    return string(data), nil
}

func readPacket(r io.Reader) ([]byte, error) {
    // Step 1: Read packet length VarInt
    length, err := readVarInt(r)
    if err != nil {
        return nil, fmt.Errorf("failed to read length: %v", err)
    }

    data := make([]byte, length)
    if _, err := io.ReadFull(r, data); err != nil {
        return nil, fmt.Errorf("failed to read packet data: %v", err)
    }

    return data, nil
}

func ReadHandshake(r io.Reader) (*Handshake, error) {
	packetData, err := readPacket(r)
	if packetData == nil {
		return nil, err
	}

	buf := bytes.NewReader(packetData)
	packetID, _ := readVarInt(buf)
	if packetID != 0x00 {
		return nil, fmt.Errorf("unexpected packet ID: %d", packetID)
	}

	protocolVersion, _ := readVarInt(buf)
	address, _ := readString(buf)
	port, _ := readUShort(buf)
	nextState, _ := readVarInt(buf)

	return &Handshake{
		ProtocolVersion: protocolVersion,
		Address:         address,
		Port:            port,
		NextState:       nextState,
	}, nil
}