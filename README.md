# Backend Web Service for Waiting Time Analysis CLI Tool

![CI Status](https://github.com/AutomatedProcessImprovement/waiting-time-backend/actions/workflows/main.yml/badge.svg) [![Coverage Status](https://coveralls.io/repos/github/AutomatedProcessImprovement/waiting-time-backend/badge.svg?branch=main)](https://coveralls.io/github/AutomatedProcessImprovement/waiting-time-backend?branch=main) [![Go Reference](https://pkg.go.dev/badge/github.com/AutomatedProcessImprovement/waiting-time-backend.svg)](https://pkg.go.dev/github.com/AutomatedProcessImprovement/waiting-time-backend)

## Deployment

Use `deploy.bash` script to set up and launch the production deployment. SSH access to the production machine is required.

Use `docker_install.bash` to install Docker in production.

## Local development

Use `run_dev.bash` script to start from scratch. It does the following:

- builds the Go application for Linux, Darwin and Windows platforms;
- builds Docker images;
- runs the `docker compose` for you, so the local deployment is ready at http://localhost:8080/.

It's possible to just start the compiled binary or run the software with `go run`, but the downstream `waiting-time-analysis` CLI tool wouldn't be available then.