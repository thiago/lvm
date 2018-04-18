# LVM - Language Version Manager

Facilitate the use of different language versions so they work as if they
were installed directly on the host.

When executing a command, LVM creates a container with `PWD` mounted with
network `host` by default, creates a "fake" user home to mount on every
command and set `HOME` variable to fake user home.

# Usage

LVM is configured with `.yml|.json` file. By default search `lvm.yml` in the
current directory or `$HOME`.

You can creates any service to use as a command from cli.

The configuration has this structure

```golang

type config struct {
    Services map[string]struct {
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
}
```

#### Configuration example

```yaml
Services:
  node:
    Image: node
    Short: NodeJS image
    Aliases:
      - npm
      - yarn
```

For more examples, please see [lvm.yml](https://github.com/thiago/lvm/blob/master/lvm.yml).

Download the default configuration

    $ curl https://github.com/thiago/lvm/blob/master/lvm.yml -o $HOME/lvm.yml


#### Using services

Execute something like:

    $ lvm node -v
or
    $ lvm npm install -g gulp

Use `-s` to skip command and execute arguments directly

    $ lvm -s npm gulp -v

Use `sh` (or `bash` if the image has) to access container

    $ lvm -s npm sh

You can change network mode to `bridge` passing the ports to the command line

    $ lvm -p 80:9090 python -m http.server 9090

Change tag of image service

    $ lvm -t 6 node -v

Global options

```sh
   --cache value	Cache folder (default: "/Users/trsouz/.lvm") [$LVM_CACHE]
   -d, --detach		Detached mode: Run container in the background, print new container name. [$LVM_DETACH]
   -e value, --env value	Set environment variables [$LVM_ENVS]
   --entrypoint value	Overwrite the default ENTRYPOINT of the image [$LVM_ENTRYPOINT]
   -k, --keep		Don't remove the container after exit [$LVM_KEEP]
   -p value, --port value	Expose ports in bridge mode [$LVM_PORTS]
   -s, --skip-cmd	Skip command name in command prefix [$LVM_SKIP_CMD]
   -t value, --tag value	Overwrite the default tag of the image [$LVM_TAG]
   -u value, --user value	Username or UID (format: <name|uid>[:<group|gid>]) (default: "$(id -u):$(id -g)") [$LVM_USER]
   --help, -h		show help
   --version		print the version
```

# Development

    $ make deps
    $ make binary

Run the lint tools to ensure that code is correct

    $ make depslint
    $ make lint

For more commands, please see `make help`.

**Please let me know if you have any questions or suggestions**. *[issues](https://github.com/thiago/lvm/issues)*