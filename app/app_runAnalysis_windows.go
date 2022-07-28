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

	// capture stdout and stderr
	cmd.Stdout = app.logger.Writer()
	var buf bytes.Buffer
	errWriter := io.MultiWriter(app.logger.Writer(), &buf)
	cmd.Stderr = errWriter

	// interrupt the command if the context is cancelled
	go func() {
		select {
		case <-ctx.Done():
			// NOTE: Windows specific code. Not sure if it kills child processes
			if err = cmd.Process.Kill(); err != nil {
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
