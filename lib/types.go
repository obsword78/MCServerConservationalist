package lib

import (
	"net"
	"os/exec"
)

type ProgramState struct {
    ExeDir        string
    YAMLConfig      *YAMLConfig
    ServerProps     *ServerProps

    ServerRunning   *int32        
	ServerStarted   chan struct{}
    
    MinecraftProcess *exec.Cmd
    PortListener   net.Listener
    RCONClient     *RCONClient
}