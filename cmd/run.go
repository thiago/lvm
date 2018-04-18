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
			detached := cmd.GlobalBool("detach")
			cacheFolder := cmd.GlobalString("cache")
			skipCommand := cmd.GlobalBool("skip-cmd")
			ports := cmd.GlobalStringSlice("port")
			userEnvs := cmd.GlobalStringSlice("env")
			keep := cmd.GlobalBool("keep")
			user := cmd.GlobalString("user")
			tag := cmd.GlobalString("tag")
			entrypoint := cmd.GlobalStringSlice("entrypoint")

			workingDir, err := os.Getwd()
			if err != nil {
				log.Fatal(err)
			}

			// Mount home and cache folders
			volumes := composeYAML.Volumes{}
			homeVolume := &composeYAML.Volume{Destination: homeFolder, Source: homeFolder}
			fakeHome := strings.Join([]string{cacheFolder, "home"}, string(os.PathSeparator))
			fakeHomeVolume := &composeYAML.Volume{
				Destination: fakeHome,
				Source:      fakeHome,
			}

			volumes.Volumes = append(volumes.Volumes, homeVolume, fakeHomeVolume)
			for _, cache := range s.Cache {
				m := strings.Join([]string{cacheFolder, "services", alias, s.GetImageTag(tag), cache}, string(os.PathSeparator))
				volumes.Volumes = append(volumes.Volumes, &composeYAML.Volume{Destination: cache, Source: m})
			}

			for _, v := range volumes.Volumes {
				if err := os.MkdirAll(v.Source, os.ModePerm); err != nil {
					log.Fatal(err)
				}
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
			envs = append(envs, "HOME="+fakeHome)
			envs = append(envs, s.Env...)
			envs = append(envs, userEnvs...)

			// Configure service
			cConfig := &composeConfig.ServiceConfig{
				Image:       s.GetImage(tag),
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

			if len(entrypoint) > 0 {
				cConfig.Entrypoint = entrypoint
			}

			// Create a empty project to setup context
			cContext := &ctx.Context{
				Context: project.Context{
					ProjectName: cmd.App.Name,
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

			// if flag keep, don't remove the container
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
