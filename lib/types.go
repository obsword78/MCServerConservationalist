package lib

import "os/exec"

type ProgramState struct {
    YAMLConfig      *YAMLConfig
    ServerProps     *ServerProps
    ServerRunning   *int32        
	ServerStarted   chan struct{}
    MinecraftProcess *exec.Cmd
}