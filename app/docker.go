package app

import (
	"context"
	"log"
	"os/exec"
)

type DockerWorker struct {
	Image     string
	Container string
	Env       []string
	WorkDir   string
}

func DefaultDockerWorker() *DockerWorker {
	return &DockerWorker{
		Image:     "nokal/process-waste:latest",
		Container: "wt_container",
		Env:       []string{"RSCRIPT_BIN_PATH=/usr/bin/Rscript"},
		WorkDir:   "/usr/src/app",
	}
}

func (w *DockerWorker) Run(ctx context.Context, logger *log.Logger, cmd string) ([]byte, error) {
	logger.Printf("Command: sh -c %s", cmd)
	return exec.CommandContext(ctx, "sh", "-c", cmd).Output()
}

func (w *DockerWorker) UpdateImage(ctx context.Context) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "docker", "pull", w.Image)
	return cmd.Output()
}
