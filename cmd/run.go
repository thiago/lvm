package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	composeConfig "github.com/docker/libcompose/config"
	composeDocker "github.com/docker/libcompose/docker"
	"github.com/docker/libcompose/docker/ctx"
	composeService "github.com/docker/libcompose/docker/service"
	"github.com/docker/libcompose/project"
	"github.com/docker/libcompose/project/options"
	composeYAML "github.com/docker/libcompose/yaml"
	"gopkg.in/urfave/cli.v1"
)

func aliasCmd(alias string, s service) cli.Command {
	cmdContext := context.Background()
	c := cli.Command{
		Name:            alias,
		Description:     s.Long,
		Usage:           fmt.Sprintf("%s (%s)", s.Short, s.Image),
		Aliases:         s.Aliases,
		Category:        s.Category,
		SkipFlagParsing: true,
		Action: func(cmd *cli.Context) error {
			skipCommand := cmd.GlobalBool("skip-cmd")
			ports := cmd.GlobalStringSlice("port")
			userEnvs := cmd.GlobalStringSlice("env")
			detached := cmd.GlobalBool("d")
			keep := cmd.GlobalBool("keep")
			user := cmd.GlobalString("user")
			workingDir, err := os.Getwd()
			if err != nil {
				log.Fatal(err)
			}

			// Mount home and cache folders
			volumes := composeYAML.Volumes{}
			volumes.Volumes = append(volumes.Volumes, &composeYAML.Volume{Destination: homeFolder, Source: homeFolder})
			for _, cache := range s.Cache {
				volumes.Volumes = append(volumes.Volumes, &composeYAML.Volume{Destination: cache, Source: strings.Join([]string{cacheFolder, alias, cache}, string(os.PathSeparator))})
			}

			// Prepare command
			var command []string
			if !skipCommand {
				command = append(command, cmd.Parent().Args().First())
			}

			command = append(command, cmd.Args()...)
			if s.PreCmd != "" {
				command = append([]string{s.PreCmd}, strings.Join(command, " "))
				command = []string{"sh", "-c", strings.Join(command, "\n")}
			}

			// append environments
			var envs []string
			for _, v := range s.Env {
				envs = append(envs, v)
			}
			for _, v := range userEnvs {
				envs = append(envs, v)
			}

			// Configure service
			cConfig := &composeConfig.ServiceConfig{
				Image:       s.Image,
				Command:     command,
				WorkingDir:  workingDir,
				Volumes:     &volumes,
				Environment: envs,
				NetworkMode: "host",
				User:        user,
			}

			// If port is provided run container in bridge mode
			if len(ports) > 0 {
				cConfig.NetworkMode = "bridge"
				cConfig.Ports = ports
			}

			// If has entrypoint
			if len(s.Entrypoint) > 0 {
				cConfig.Entrypoint = s.Entrypoint
			}

			// Create a empty project to setup context
			cContext := &ctx.Context{
				Context: project.Context{
					ProjectName: "lvm",
				},
			}
			_, err = composeDocker.NewProject(cContext, nil)
			if err != nil {
				log.Fatal(err)
			}

			// Run service
			s := composeService.NewService(alias, cConfig, cContext)
			exit, err := s.Run(cmdContext, command, options.Run{Detached: detached})
			if err != nil {
				log.Println(err)
			}

			// if flag keep, do not remove container
			if !keep {
				err = s.Delete(cmdContext, options.Delete{})
				if err != nil {
					log.Println(err)
				}
			}
			os.Exit(exit)
			return nil
		},
	}
	return c
}
