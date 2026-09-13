package runtime

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

type Runtime interface {
	Deploy(context.Context, string, string, string, string) (string, error)
	Stop(context.Context, string) error
	Logs(context.Context, string) (string, error)
}

type Docker struct{}

func NewDocker() Docker {
	return Docker{}
}

func (Docker) Deploy(ctx context.Context, sourceDir, image, containerName, port string) (string, error) {
	if output, err := run(ctx, sourceDir, "docker", "build", "-t", image, "."); err != nil {
		return "", fmt.Errorf("build image: %w: %s", err, output)
	}
	if output, err := run(ctx, sourceDir, "docker", "rm", "-f", containerName); err != nil && !strings.Contains(output, "No such container") {
		return "", fmt.Errorf("replace container: %w: %s", err, output)
	}
	output, err := run(ctx, sourceDir, "docker", "run", "-d", "--name", containerName, "-p", port+":"+port, image)
	if err != nil {
		return "", fmt.Errorf("start container: %w: %s", err, output)
	}
	return strings.TrimSpace(output), nil
}

func (Docker) Stop(ctx context.Context, containerName string) error {
	output, err := run(ctx, "", "docker", "stop", containerName)
	if err != nil && !strings.Contains(output, "No such container") {
		return fmt.Errorf("stop container: %w: %s", err, output)
	}
	return nil
}

func (Docker) Logs(ctx context.Context, containerName string) (string, error) {
	output, err := run(ctx, "", "docker", "logs", "--tail", "200", containerName)
	if err != nil {
		return "", fmt.Errorf("read logs: %w: %s", err, output)
	}
	return output, nil
}

func run(ctx context.Context, directory, name string, args ...string) (string, error) {
	command := exec.CommandContext(ctx, name, args...)
	command.Dir = directory
	var output bytes.Buffer
	command.Stdout = &output
	command.Stderr = &output
	if err := command.Run(); err != nil {
		return output.String(), err
	}
	return output.String(), nil
}
