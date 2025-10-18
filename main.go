package main

import (
	"bufio"
	"fmt"
	"net"
	"sync/atomic"
	"time"

	"github.com/obsword78/MCServerConservationalist/lib"
)

type ProgramState = lib.ProgramState

func main() {
	state := &ProgramState{
        YAMLConfig:  lib.LoadYAMLConfig("MCServerConservationalist.yaml"),
        ServerProps: lib.LoadServerProps("server.properties"),
		ServerRunning: new(int32),
    }
    atomic.StoreInt32(state.ServerRunning, 0)

	for {
		state.ServerStarted = make(chan struct{})

		go WaitForValidTrigger(state)

		<-state.ServerStarted
		fmt.Println("Server start triggered → stopping listener")
		MonitorIdle(state)
	}
}

func WaitForValidTrigger(state *ProgramState) {
    ln, err := net.Listen("tcp", fmt.Sprintf(":%d", state.ServerProps.ServerPort))
    if err != nil {
        fmt.Println("Failed to create listener:", err)
        close(state.ServerStarted)
        return
    }
    defer ln.Close()

    fmt.Println("Waiting for players to trigger server…")
    for {
        if atomic.LoadInt32(state.ServerRunning) == 1 {
            return // stop accepting once server started
        }

        conn, err := ln.Accept()
        if err != nil {
            if opErr, ok := err.(*net.OpError); ok && !opErr.Temporary() {
                return
            }
            fmt.Println("Accept error:", err)
            continue
        }

        go handleConnection(state, conn)
    }
}


func MonitorIdle(ste *ProgramState) {
	idleSeconds := 0
	for {
		count := lib.GetPlayerCountRCON(fmt.Sprintf("localhost:%d", ste.ServerProps.RconPort), ste.ServerProps.RconPassword)
		if count <= 0 {
			idleSeconds++
			if idleSeconds >= ste.YAMLConfig.IdleTimeoutSeconds {
				fmt.Println("Idle timeout reached → stopping server")
				lib.SendRCONStop(fmt.Sprintf("localhost:%d", ste.ServerProps.RconPort), ste.ServerProps.RconPassword)
				atomic.StoreInt32(ste.ServerRunning, 0)
				break
			}
		} else {
			idleSeconds = 0
		}
		time.Sleep(time.Second)

		if (atomic.LoadInt32(ste.ServerRunning) == 0) {
			fmt.Println("Server is not running → stopping idle monitor")
			return
		}
	}
}

func handleConnection(state *ProgramState, conn net.Conn) {
    defer conn.Close()
    r := bufio.NewReader(conn)
    hs, err := lib.ReadHandshake(r)
    if err != nil {
        fmt.Println("Handshake error:", err)
        return
    }

    switch hs.NextState {
    case 1: // Status
        if err := lib.HandleStatus(conn, state.ServerProps, state.YAMLConfig); err != nil {
            fmt.Println("Status error:", err)
        }
    case 2: // Login
        if atomic.LoadInt32(state.ServerRunning) == 1 {
			fmt.Println("Server already running, rejecting login trigger")
            return
        }
        if err := lib.HandleLogin(state, r, conn); err != nil {
            fmt.Println("Login error:", err)
            return
        }
    default:
        fmt.Println("Unknown next state:", hs.NextState)
    }
}