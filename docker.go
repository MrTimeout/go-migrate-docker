package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"regexp"
	"sync"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/moby/moby/client"
)

var (
	//imageRegexp we have to use it to get all the image names from a file instead of getting them from input as a parameter. TODO
	ImageRegexp = regexp.MustCompile(`^[ ]*image: (.+)$`) //nolint:deadcode
)

type DockerAction interface {
	grepImages(ctx context.Context, imageName string) (map[string][]string, error)

	migrateImage(ctx context.Context, imageIDs ...string) error
}

type DockerBuilder interface {
	buildTarget(host string) error

	buildFrom(host string) error
}

type DockerClient interface {
	DockerAction
	DockerBuilder
}

type DockerHost struct {
	from   *client.Client
	target []*client.Client
}

func (dh DockerHost) grepImages(ctx context.Context, imageName string) (map[string][]string, error) {
	var imageIds = map[string][]string{}

	re, err := regexp.Compile(imageName)
	if err != nil {
		return imageIds, err
	}

	images, err := dh.from.ImageList(ctx, types.ImageListOptions{All: true})
	if err != nil {
		return nil, err
	}

	for _, image := range images {
		if filteredImages := matchesOfStringArr(re, image.RepoTags); len(filteredImages) > 0 {
			imageIds[image.ID] = filteredImages
		}
	}

	return imageIds, err
}

func (dh DockerHost) migrateImage(ctx context.Context, imageIDs ...string) error {
	var wg sync.WaitGroup
	tar, err := dh.from.ImageSave(ctx, imageIDs)
	if err != nil {
		return err
	}

	tarContent, err := io.ReadAll(tar)
	if err != nil {
		return err
	}

	for _, cl := range dh.target {
		wg.Add(1)
		go func(c *client.Client) {
			defer wg.Done()
			res, err := c.ImageLoad(ctx, bytes.NewReader(tarContent), false)
			if err != nil {
				fmt.Println(err)
				return
			}

			if res.JSON {
				b, err := io.ReadAll(res.Body)
				if err != nil {
					fmt.Println(err)
					return
				}
				defer res.Body.Close()

				fmt.Println(string(b))
			}
		}(cl)
	}

	wg.Wait()

	return nil
}

func (d *DockerHost) buildFrom(host string) error {
	c, err := buildDockerClient(host)
	if err != nil {
		return err
	}

	d.from = c

	return nil
}

func (d *DockerHost) buildTarget(host string) error {
	c, err := buildDockerClient(host)
	if err != nil {
		return err
	}

	d.target = append(d.target, c)

	return nil
}

func buildDockerClient(host string) (*client.Client, error) {
	c, err := client.NewClientWithOpts(
		client.WithHost(host),
		client.WithVersion("1.41"),
		client.WithScheme("http"),
		client.WithTimeout(10*time.Second),
	)

	if err != nil {
		return nil, err
	}

	return c, err
}
