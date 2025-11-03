package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/obsword78/MCServerConservationalist/lib"
)
type ProgramState = lib.ProgramState

func main() {
    exePath, err := os.Executable()
    if err != nil {
        log.Fatal(err)
    }
    exeDir := filepath.Dir(exePath)

    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	state := &ProgramState{
        ExeDir:     exeDir,
        YAMLConfig:  lib.LoadYAMLConfig(filepath.Join(exeDir, "MCServerConservationalist.yaml")),
        ServerProps: lib.LoadServerProps(filepath.Join(exeDir, "server.properties")),
		ServerRunning: new(int32),
    }
    atomic.StoreInt32(state.ServerRunning, 0)

    go func() {
        <-c
        fmt.Println("\033[1;32mReceived interrupt signal, exiting connections...\033[0m")

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

    fmt.Println("\033[1;33mMCServerConservationalist CLI started. Please only close this program through CTRL + C to safely close ports.\033[0m")
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

    fmt.Println("\033[1;32mServer is sleeping, waiting for players to trigger server…\033[0m")
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

        if atomic.LoadInt32(state.ServerRunning) == 1 {
            return
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
        if ste.RCONClient.GetPlayerCount() == 0  {
            idleSeconds++
            if idleSeconds >= ste.YAMLConfig.IdleTimeoutSeconds {
                fmt.Println("\033[1;34mIdle timeout reached → stopping server\033[0m")
                ste.RCONClient.StopServer()
                break
            }
        } else {
            idleSeconds = 0
        }
        time.Sleep(time.Second)
        if atomic.LoadInt32(ste.ServerRunning) == 0 {
            fmt.Println("\033[1;34mServer stopped manually → ending idle monitor\033[0m")
            break
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
        if err := lib.HandleStatus(conn, state); err != nil {
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