package cmd

import (
	"fmt"
	"log"
	"os"
	"sort"
	"strings"

	"syscall"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
	"gopkg.in/urfave/cli.v1"
)

var cfgFile string
var cacheFolder string
var homeFolder string

type config struct {
	Services map[string]service
}

// Service represents the command definition
type service struct {
	Image      string
	Short      string
	Long       string
	Aliases    []string
	Env        []string
	Cache      []string
	PreCmd     string
	Entrypoint []string
	Category   string
}

// RootCmd represents the base command when called without any subcommands
var RootCmd *cli.App

// Execute command
func Execute(version string) {
	RootCmd.Version = version
	err := RootCmd.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func init() {
	RootCmd = cli.NewApp()
	RootCmd.HideVersion = true
	RootCmd.Name = "lvm"
	RootCmd.Usage = "language version manager"
	RootCmd.Version = "0.0.1"

	RootCmd.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:   "d",
			EnvVar: "LVM_DETACHED",
			Usage:  "Detached mode: Run container in the background, print new container name.",
		},
		cli.StringFlag{
			Name:   "user, u",
			EnvVar: "LVM_USER",
			Value:  fmt.Sprintf("%d:%d", syscall.Getuid(), syscall.Getgid()),
			Usage:  "Username or UID (format: <name|uid>[:<group|gid>])",
		},
		cli.StringSliceFlag{
			Name:   "port, p",
			EnvVar: "LVM_PORTS",
			Value:  nil,
			Usage:  "Expose ports in bridge mode.",
		},
		cli.BoolFlag{
			Name:   "keep, k",
			EnvVar: "LVM_KEEP",
			Usage:  "Don't remove container after exit.",
		},
		cli.StringSliceFlag{
			Name:   "env, e",
			EnvVar: "LVM_ENVS",
			Value:  nil,
			Usage:  "Set environment variables",
		},
		cli.BoolFlag{
			Name:   "skip-cmd, s",
			EnvVar: "LVM_SKIP_CMD",
			Usage:  "Skip command name in command prefix",
		},
	}

	sort.Sort(cli.FlagsByName(RootCmd.Flags))
	sort.Sort(cli.CommandsByName(RootCmd.Commands))

	// Find home directory.
	home, err := homedir.Dir()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	homeFolder = home

	cacheFolder = strings.Join([]string{home, ".lvm", "cache"}, string(os.PathSeparator))

	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Search config in home directory with name ".lvm" (without extension).
		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.SetConfigName("lvm")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		//fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
	var c config
	if err := viper.Unmarshal(&c); err != nil {
		panic(err)
	}

	for alias, service := range c.Services {
		RootCmd.Commands = append(RootCmd.Commands, aliasCmd(alias, service))
	}
}
