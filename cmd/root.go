package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"syscall"

	"log"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
	"gopkg.in/urfave/cli.v1"
)

var homeFolder string

type config struct {
	Services map[string]service
}

type service struct {
	// Image of container
	Image string
	// Short description of service
	Short string
	// Long description of service
	Long string
	// Aliases is a list of subcommands for service like npm, pip, ...
	Aliases []string
	// Env is a list of environment variables to set in container
	Env []string
	// Cache is a list of folders to cache
	Cache []string
	// PreCmd is executed before command
	PreCmd string
	// Entrypoint overwrite image ENTRYPOINT
	Entrypoint []string
	// Category is a way to categorize commands
	Category string
}

func (s *service) GetImage(tag string) string {
	if tag == "" {
		return s.Image
	}
	i := strings.SplitN(s.Image, ":", 2)
	if len(i) == 0 {
		log.Fatal("Ensure image exists")
	}
	return fmt.Sprintf("%s:%s", i[0], tag)
}

func (s *service) GetImageTag(tag string) string {
	if tag != "" {
		return tag
	}
	i := strings.SplitN(s.Image, ":", 2)
	if len(i) > 1 {
		return i[1]
	}
	return "latest"
}

func init() {
	// Find home directory.
	home, err := homedir.Dir()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	homeFolder = home
}

// App return a new cli App
func App(name, version string) *cli.App {
	getEnvName := func(a string) string {
		return fmt.Sprintf("%s_%s", strings.ToUpper(name), strings.ToUpper(a))
	}

	cli.VersionFlag = cli.BoolFlag{
		Name:  "version",
		Usage: "print the version",
	}

	app := cli.NewApp()
	app.Name = name
	app.Usage = "language version manager"
	app.Version = version
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:   "d, detach",
			EnvVar: getEnvName("detach"),
			Usage:  "Detached mode: Run container in the background, print new container name.",
		},
		cli.StringFlag{
			Name:   "u, user",
			EnvVar: getEnvName("user"),
			Value:  fmt.Sprintf("%d:%d", syscall.Getuid(), syscall.Getgid()),
			Usage:  "Username or UID (format: <name|uid>[:<group|gid>])",
		},
		cli.StringFlag{
			Name:   "t, tag",
			EnvVar: getEnvName("tag"),
			Usage:  "Overwrite the default tag of the image",
		},
		cli.StringSliceFlag{
			Name:   "entrypoint",
			EnvVar: getEnvName("entrypoint"),
			Usage:  "Overwrite the default ENTRYPOINT of the image",
		},
		cli.StringSliceFlag{
			Name:   "p, port",
			EnvVar: getEnvName("ports"),
			Value:  nil,
			Usage:  "Expose ports in bridge mode",
		},
		cli.BoolFlag{
			Name:   "k, keep",
			EnvVar: getEnvName("keep"),
			Usage:  "Don't remove the container after exit",
		},
		cli.StringSliceFlag{
			Name:   "e, env",
			EnvVar: getEnvName("envs"),
			Value:  nil,
			Usage:  "Set environment variables",
		},
		cli.BoolFlag{
			Name:   "s, skip-cmd",
			EnvVar: getEnvName("skip_cmd"),
			Usage:  "Skip command name in command prefix",
		},
		cli.StringFlag{
			Name:   "cache",
			EnvVar: getEnvName("cache"),
			Value:  strings.Join([]string{homeFolder, "." + name}, string(os.PathSeparator)),
			Usage:  "Cache folder",
		},
	}

	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))

	if e := os.Getenv(getEnvName("CONFIG")); e != "" {
		// Use config file from the flag.
		viper.SetConfigFile(os.Getenv(getEnvName("CONFIG")))
	} else {
		// Search config in home directory with name ".lvm" (without extension).
		viper.AddConfigPath(homeFolder)
		viper.AddConfigPath(".")
		viper.SetConfigName("lvm")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
	var c config
	if err := viper.Unmarshal(&c); err != nil {
		log.Fatalln(err)
	}

	for alias, service := range c.Services {
		app.Commands = append(app.Commands, aliasCmd(alias, service))
	}
	return app
}
