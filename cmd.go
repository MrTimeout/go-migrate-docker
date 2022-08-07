package main

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var (
	ErrParsingArgs = errors.New("from or targethost has not right values")

	sourceHost       string
	destinationHosts []string
	imagePattern     string
)

type PreRunE func(cmd *cobra.Command, args []string) error

type Run func(cmd *cobra.Command, args []string)

func NewRootCmd(docker DockerClient) *cobra.Command {
	cmd := &cobra.Command{
		Use: "go-migrate-docker",
		Short: `Allows us to migrate images from one docker daemon to others. 
	This can be a 'good' solution when no general repository is set`,
		Long: `Allows us to migrate images from one docker daemon to others. 
	This can be a 'good' solution when no general repository is set. 
	By now, we can only migrate all the images from one daemon to others, user can't select which ones he wants.
	
	Hosts must be in the following format: tcp://[ip|dns]:(2375|2376,ca=/path/to/ca,cert=/path/to/cert,key=/path/to/key)
	tcp means that we are using the TCP protocol with the port 2375 (http) or 2376 (https)
	`,
		Example: `
	go-migrate-docker --source-host "tcp://192.168.56.2:2375" --destination-host "tcp://192.168.56.3:2376,ca=~/ca-file,cert=~/cert-file,key=~/cert-key"

	go-migrate-docker --source-host "tcp://192.168.56.2:2375" --destination-host "tcp://192.168.56.3:2375" --destination-host "tcp://192.168.56.4:2375"
	`,
		PreRunE: rootPreRunE(docker),
		Run:     rootRun(docker),
	}

	cmd.PersistentFlags().StringVarP(&sourceHost, "source-host", "s", "", "Host from where we are getting the images to migrate them to the target hosts")
	cmd.PersistentFlags().StringArrayVarP(&destinationHosts, "destination-host", "d", []string{}, "hosts where all the images are going to")
	cmd.PersistentFlags().StringVarP(&imagePattern, "image-pattern", "p", ".*", "image pattern 'regexp.Regexp style' to fetch images from 'from-host' to 'target-hosts'. By default, value is set to .*, which means all images.")

	cmd.MarkFlagRequired("from")        //nolint:errcheck
	cmd.MarkFlagRequired("target-host") //nolint:errcheck

	return cmd
}

func rootPreRunE(docker DockerBuilder) PreRunE {
	return func(cmd *cobra.Command, args []string) error {
		if strings.TrimSpace(sourceHost) == "" {
			return ErrParsingArgs
		}

		if err := docker.buildFrom(sourceHost); err != nil {
			return err
		}

		for _, host := range destinationHosts {
			if err := docker.buildTarget(host); err != nil {
				return err
			}
		}

		return nil
	}
}

func rootRun(docker DockerAction) Run {
	return func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()

		images, err := docker.grepImages(ctx, imagePattern)
		if err != nil {
			cmd.PrintErrln(err)
			return
		}

		for imageId, imageRepoTags := range images {
			temp := append(imageRepoTags, imageId)
			if err := docker.migrateImage(ctx, temp...); err != nil {
				cmd.PrintErrln(err)
			}
		}
	}
}
