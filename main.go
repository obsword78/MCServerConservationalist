package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/obsword78/MCServerConservationalist/lib"
)
type ProgramState = lib.ProgramState

func main() {
    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	state := &ProgramState{
        YAMLConfig:  lib.LoadYAMLConfig("MCServerConservationalist.yaml"),
        ServerProps: lib.LoadServerProps("server.properties"),
		ServerRunning: new(int32),
    }
    atomic.StoreInt32(state.ServerRunning, 0)

    go func() {
        <-c
        fmt.Println("Received interrupt signal, exiting connections...")

        if state.PortListener != nil {
            state.PortListener.Close()
            state.PortListener = nil
        }
        if state.RCONClient != nil {
            state.RCONClient.Close()
            state.RCONClient = nil
        }
        
        atomic.StoreInt32(state.ServerRunning, 0)

        os.Exit(0)
    }()

	for {
		state.ServerStarted = make(chan struct{})

		go WaitForValidTrigger(state)

		<-state.ServerStarted

		MonitorIdle(state)

        state.RCONClient.Close()
        state.RCONClient = nil
	}
}

func WaitForValidTrigger(state *ProgramState) {
    ln, err := net.Listen("tcp", fmt.Sprintf(":%d", state.ServerProps.ServerPort))
    if err != nil {
        fmt.Println("Failed to create listener:", err)
        close(state.ServerStarted)
        return
    }
    state.PortListener = ln

    for {
        if atomic.LoadInt32(state.ServerRunning) == 0 {
            break
        }
        time.Sleep(time.Second)
    }

    fmt.Println("Server is sleeping, waiting for players to trigger server…")
    for {
        if atomic.LoadInt32(state.ServerRunning) == 1 {
            return
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
    rconClient, err := lib.NewRCONClient(fmt.Sprintf("localhost:%d", ste.ServerProps.RconPort), ste.ServerProps.RconPassword)
    if err != nil {
        fmt.Println("RCON connection error:", err)
        return
    }

    ste.RCONClient = rconClient
    idleSeconds := 0
    for {
        if ste.RCONClient.GetPlayerCount() == 0 {
            idleSeconds++
            if idleSeconds >= ste.YAMLConfig.IdleTimeoutSeconds {
                fmt.Println("Idle timeout reached → stopping server")
                ste.RCONClient.StopServer()
                break
            }
        } else {
            idleSeconds = 0
        }
        time.Sleep(time.Second)
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