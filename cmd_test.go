package main

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

const DaemonNotRunning = "daemonnotrunning"

var ErrDaemonNotRunning = errors.New("daemon not running")

type DumbDockerHost struct {
}

func (DumbDockerHost) buildTarget(host string) error {
	return nil
}

func (DumbDockerHost) buildFrom(host string) error {
	return nil
}

func (DumbDockerHost) grepImages(ctx context.Context, imageName string) (map[string][]string, error) {
	if imageName == DaemonNotRunning {
		return nil, ErrDaemonNotRunning
	}
	return map[string][]string{"123456": {"alpine:latest", "alpine:3.16"}}, nil
}

func (DumbDockerHost) migrateImage(ctx context.Context, imageIDs ...string) error {
	return nil
}

func TestRootCmdPreRunE(t *testing.T) {
	t.Run("source host is empty so err parsing args is returned", func(t *testing.T) {
		cmd := NewRootCmd(&DockerHost{})
		cmd.SetArgs([]string{"--source-host", "  ", "--destination-host", "somethinghere"})
		got := cmd.Execute()

		assert.ErrorIs(t, got, ErrParsingArgs)
	})

	t.Run("when source host is invalid", func(t *testing.T) {
		cmd := NewRootCmd(&DockerHost{})
		cmd.SetArgs([]string{"--source-host", "192.168.56.2", "destination-host", "somethinghere"})
		got := cmd.Execute()

		assert.ErrorContains(t, got, "unable to parse docker host `192.168.56.2`")
	})

	t.Run("when some target host is invalid", func(t *testing.T) {
		cmd := NewRootCmd(&DockerHost{})
		cmd.SetArgs([]string{"--source-host", "tcp://192.168.56.2:2375", "--destination-host", "tcp://192.168.56.3:2375", "--destination-host", "192.168.56.4"})
		got := cmd.Execute()

		assert.ErrorContains(t, got, "unable to parse docker host `192.168.56.4`")
	})

	t.Run("when all is correct", func(t *testing.T) {
		cmd := NewRootCmd(&DumbDockerHost{})
		cmd.SetArgs([]string{"--source-host", "tcp://192.168.56.2:2375", "--destination-host", "tcp://192.168.56.3:2375", "--destination-host", "tcp://192.168.56.4:2375"})
		got := cmd.Execute()

		assert.Nil(t, got)
	})
}

func TestRootCmdRun(t *testing.T) {
	var grepImagesTestCases = []struct {
		description, imageName string
		gotErr                 error
	}{
		{
			description: "daemon is not running",
			imageName:   DaemonNotRunning,
			gotErr:      ErrDaemonNotRunning,
		},
		{
			description: "got images succesfully",
			imageName:   ".*",
			gotErr:      nil,
		},
	}
	cmd := NewRootCmd(DumbDockerHost{})
	args := []string{
		"--source-host", "tcp://192.168.56.2:2375",
		"--destination-host", "tcp://192.168.56.3:2375",
		"--destination-host", "tcp://192.168.56.4:2375",
	}

	for _, tCase := range grepImagesTestCases {
		t.Run(tCase.description, func(t *testing.T) {
			b := bytes.NewBufferString("")
			cmd.SetArgs(append(args, "--image-pattern", tCase.imageName))
			cmd.SetErr(b)

			cmd.Execute() //nolint:errcheck

			if tCase.gotErr != nil {
				assert.Contains(t, b.String(), tCase.gotErr.Error())
			} else {
				assert.Empty(t, b.String())
			}
		})
	}
}
