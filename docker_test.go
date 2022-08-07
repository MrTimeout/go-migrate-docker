package main

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

const (
	alpineImageID = "e66264b98777"
	golangImageID = "759ab1463be2"
)

var ctx = context.Background()

func TestGrepImages(t *testing.T) {
	testResponseImages := func(t *testing.T, imageName string, want []string) {
		server(t)
		dh := DockerHost{}
		err := dh.buildFrom("tcp://127.0.0.1:2375")
		if err != nil {
			t.Fatal(err)
		}

		ctxTimeout, cancel := context.WithTimeout(ctx, 20*time.Second)
		defer cancel()

		waitUntilServerUpAndRunning(t)
		imageIDs, err := dh.grepImages(ctxTimeout, imageName)
		if err != nil {
			t.Fatal(err)
		}

		for _, w := range want {
			assert.Contains(t, imageIDs, w)
		}
	}

	t.Run("grep all images from docker image repository succesfully", func(t *testing.T) {
		testResponseImages(t, ".*", []string{golangImageID, alpineImageID})
	})

	t.Run("grep alpine images from repo succesfully", func(t *testing.T) {
		testResponseImages(t, "alpine.*", []string{alpineImageID})
	})

	t.Run("grep golang images from repo successfully", func(t *testing.T) {
		testResponseImages(t, "golang.*", []string{golangImageID})
	})

	t.Run("docker daemon is not running", func(t *testing.T) {
		want := "Cannot connect to the Docker daemon at tcp://127.0.0.1:2375. Is the docker daemon running?"
		dh := DockerHost{}
		err := dh.buildFrom("tcp://127.0.0.1:2375")
		if err != nil {
			t.Fatal(err)
		}

		ctxTimeout, cancel := context.WithTimeout(ctx, 20*time.Second)
		defer cancel()
		_, got := dh.grepImages(ctxTimeout, "e66264b98777")

		assert.ErrorContains(t, got, want)
	})

	// We have to fix this test and the method migrateImage, because when getting error, we are not displaying it
	t.Run("docker persist images succesfully", func(t *testing.T) {
		server(t)
		dh := DockerHost{}
		dh.buildFrom("tcp://127.0.0.1:2375")   //nolint:errcheck
		dh.buildTarget("tcp://127.0.0.1:2375") //nolint:errcheck

		ctxTimeout, cancel := context.WithTimeout(ctx, 20*time.Second)
		defer cancel()

		waitUntilServerUpAndRunning(t)
		got := dh.migrateImage(ctxTimeout, "e66264b98777", "alpine:latest", "alpine:3.16")

		assert.Nil(t, got)
	})
}

func TestBuildClient(t *testing.T) {
	t.Run("build from host", func(t *testing.T) {
		dh := DockerHost{}

		got := dh.buildFrom("tcp://192.168.56.2:2375")

		assert.Nil(t, got)
		assert.NotNil(t, dh.from)
	})

	t.Run("build target host", func(t *testing.T) {
		dh := DockerHost{}
		input := []string{
			"tcp://192.168.56.3:2375",
			"tcp://192.168.56.4:2375",
			"tcp://192.168.56.5:2375",
		}

		for _, i := range input {
			err := dh.buildTarget(i)
			if err != nil {
				t.Fatal(err)
			}
		}

		assert.Len(t, dh.target, len(input))
	})

	t.Run("incorrect host as input for docker client in buildFrom", func(t *testing.T) {
		dh := DockerHost{}
		input := "192.168.56.100"

		err := dh.buildFrom(input)

		assert.ErrorContains(t, err, "unable to parse docker host `192.168.56.100`")
	})

	t.Run("incorrect host as input for docker client in target", func(t *testing.T) {
		dh := DockerHost{}
		input := "192.168.56.100"

		err := dh.buildTarget(input)

		assert.ErrorContains(t, err, "unable to parse docker host `192.168.56.100`")
	})
}

func server(t *testing.T) {
	g := gin.Default()

	g.GET("/ping", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, struct{ Msg string }{Msg: "OK"})
	})

	g.GET("/v1.41/images/json", func(ctx *gin.Context) {
		if v, ok := ctx.GetQuery("all"); ok && v == "1" {
			ctx.JSON(http.StatusOK, []types.ImageSummary{
				{
					ID:       alpineImageID,
					RepoTags: []string{"alpine:latest", "alpine:3.16"},
				},
				{
					ID:       golangImageID,
					RepoTags: []string{"golang:latest", "golang:1.8-alpine"},
				},
			})
		}
		ctx.JSON(http.StatusBadRequest, struct{ msg string }{msg: "bye bye"})
	})

	g.GET("/v1.41/images/get", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, struct{ Msg string }{Msg: "image get"})
	})

	g.GET("/v1.41/images/load", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, struct{ Msg string }{Msg: "image loaded"})
	})

	srv := &http.Server{
		Addr:    "127.0.0.1:2375",
		Handler: g,
	}

	go func() {
		srv.ListenAndServe() //nolint:errcheck
	}()
	t.Cleanup(func() {
		ctxTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		srv.Shutdown(ctxTimeout) //nolint:errcheck
	})
}

func waitUntilServerUpAndRunning(t *testing.T) {
	for i := 0; i < 3; i++ {
		if res, err := http.Get("http://localhost:2375/ping"); err == nil {
			defer res.Body.Close()
			b, err := io.ReadAll(res.Body)
			if err != nil {
				continue
			}
			if strings.Contains(string(b), "OK") {
				return
			}
		}
		time.Sleep(300 * time.Millisecond)
	}
	t.Fatal(errors.New("server is not up and running"))
}
