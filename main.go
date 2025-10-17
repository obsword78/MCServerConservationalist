package main

import (

	// "time"

	"bufio"
	"fmt"
	"net"
	"sync/atomic"
	"time"

	"github.com/obsword78/MCServerConservationalist/lib"
)

var yamlCfg *lib.YAMLConfig
var srvProps *lib.ServerProps
var serverRunning = new(int32)

func main() {
	yamlCfg = lib.LoadYAMLConfig("MCServerConservationalist.yaml")
	srvProps = lib.LoadServerProps("server.properties")
	atomic.StoreInt32(serverRunning, 0)

	for {
		WaitForValidTrigger()
		MonitorIdle()
	}
}

func WaitForValidTrigger() {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", srvProps.ServerPort))
	if err != nil {
		fmt.Println("Error creating listener:", err)
		return
	}
	defer ln.Close()

	for {
		if atomic.LoadInt32(serverRunning) == 1 {
            fmt.Println("Server start triggered → stopping listener")
            break
        }

		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		go handleConnection(conn) // handle each client concurrently
	}
}

func MonitorIdle() {
	idleSeconds := 0
	for {
		count := lib.GetPlayerCountRCON(fmt.Sprintf("localhost:%d", srvProps.RconPort), srvProps.RconPassword)
		if count <= 0 {
			idleSeconds++
			if idleSeconds >= yamlCfg.IdleTimeoutSeconds {
				fmt.Println("Idle timeout reached → stopping server")
				lib.SendRCONStop(fmt.Sprintf("localhost:%d", srvProps.RconPort), srvProps.RconPassword)
				atomic.StoreInt32(serverRunning, 0)
				break
			}
		} else {
			idleSeconds = 0
		}
		time.Sleep(time.Second)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	fmt.Println("Client connected:", conn.RemoteAddr())

	r := bufio.NewReader(conn)
	hs, err := lib.ReadHandshake(r)
	if err != nil {
		fmt.Println("Handshake error:", err)
		return
	}

	switch hs.NextState {
	case 1: // Status
		if err := lib.HandleStatus(conn, srvProps, yamlCfg); err != nil {
			fmt.Println("Status error:", err)
		}
	case 2: // Login
		if atomic.LoadInt32(serverRunning) == 1 {
			fmt.Println("Server already running, rejecting login attempt")
		} else if err := lib.HandleLogin(conn, yamlCfg, serverRunning); err != nil {
			fmt.Println("Status error:", err)
		}
	default:
		fmt.Println("Unknown next state:", hs.NextState)
	}
}