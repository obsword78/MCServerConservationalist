package main

import (
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
var serverStarted = make(chan struct{}) // channel trigger

func main() {
	yamlCfg = lib.LoadYAMLConfig("MCServerConservationalist.yaml")
	srvProps = lib.LoadServerProps("server.properties")

	ln, _ := net.Listen("tcp", fmt.Sprintf(":%d", srvProps.ServerPort))
	defer ln.Close()

	for {
		atomic.StoreInt32(serverRunning, 0)
		serverStarted = make(chan struct{})

		fmt.Println("Waiting for valid trigger...")		
		go func() {
			for {
				conn, _ := ln.Accept()
				go handleConnection(conn)
			}
		}()

		<-serverStarted
		fmt.Println("Server triggered → starting idle monitor")
		MonitorIdle()
	}
}

func MonitorIdle() {
	idleSeconds := 0
	for {
		count := lib.GetPlayerCountRCON(fmt.Sprintf("localhost:%d", srvProps.RconPort), srvProps.RconPassword)
		fmt.Println(count)
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

	if atomic.LoadInt32(serverRunning) == 1 {
		fmt.Println("Server already running, rejecting login attempt")
		return
	}

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
		} else {
			if err := lib.HandleLogin(r, conn, yamlCfg, serverRunning); err != nil {
				fmt.Println("Login error:", err)
			}
			select {
			case <-serverStarted:
			default:
				close(serverStarted)
			}
		}
	default:
		fmt.Println("Unknown next state:", hs.NextState)
	}
}