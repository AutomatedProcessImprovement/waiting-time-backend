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

	// custom column mapping if it was provided with the API request

	var columnMapping string

	if job.ColumnMapping != nil {
		b, err := json.Marshal(job.ColumnMapping)
		if err != nil {
			return fmt.Errorf("error marshalling column mapping: %s", err.Error())
		}
		columnMapping = string(b)
	}

	// analysis script name

	var scriptName string

	if app.config.DevelopmentMode {
		// NOTE: we don't use columns mapping in development mode, modify the log accordingly
		scriptName = "run_analysis_dev.bash"
	} else if columnMapping == "" {
		scriptName = "run_analysis.bash"
	} else {
		scriptName = "run_analysis_columns.bash"
	}

	// shell command

	var args string

	eventLogPath := path.Join(jobDir, eventLogName)
	if columnMapping == "" {
		args = fmt.Sprintf("bash %s %s %s", scriptName, eventLogPath, jobDir)
	} else {
		args = fmt.Sprintf("bash %s %s %s %q", scriptName, eventLogPath, jobDir, columnMapping)
	}

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
				app.logger.Printf("Cannot cancel the job: %s. But it might be okay if the job finished successfully", err.Error())
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
