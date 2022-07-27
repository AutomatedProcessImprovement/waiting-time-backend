//go:build linux || darwin

package app

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/AutomatedProcessImprovement/waiting-time-backend/model"
	"io"
	"os/exec"
	"path"
	"syscall"
)

func (app *Application) runAnalysis(ctx context.Context, eventLogName string, job *model.Job) error {
	jobDir, err := abspath(job.Dir)
	if err != nil {
		return err
	}

	eventLogPath := path.Join(jobDir, eventLogName)
	scriptName := "run_analysis.bash"
	if app.config.DevelopmentMode {
		scriptName = "run_analysis_dev.bash"
	}
	args := fmt.Sprintf("bash %s %s %s", scriptName, eventLogPath, jobDir)

	cmd := exec.CommandContext(ctx, "sh", "-c", args)

	// sets process group ID to kill all processes in the group later on cancel if needed
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	// capture stdout and stderr
	cmd.Stdout = app.logger.Writer()
	var buf bytes.Buffer
	errWriter := io.MultiWriter(app.logger.Writer(), &buf)
	cmd.Stderr = errWriter

	// interrupt the command if the context is cancelled
	go func() {
		select {
		case <-ctx.Done():
			if err = syscall.Kill(-1*cmd.Process.Pid, syscall.SIGKILL); err != nil {
				app.logger.Printf("Error cancelling job: %s", err.Error())
			}
		}
	}()

	if err = cmd.Start(); err != nil {
		return errors.New(fmt.Sprintf("error starting analysis: %s", err.Error()))
	}

	app.logger.Printf("Job %s executing", job.ID)

	if err = cmd.Wait(); err != nil {
		err = fmt.Errorf("error executing analysis: %s; stderr: %s", err.Error(), buf.String())
	}
	return err
}
