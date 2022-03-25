package main

import (
	"fmt"
	"os"

	"mydocker/container"
	"mydocker/network"
	"mydocker/subsystems"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var runCommand = cli.Command{
	Name: "run",
	Usage: `zuo yi ge rong qi 
		mydocker run -ti [command]`,
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "ti",
			Usage: "enable tty",
		},
		cli.BoolFlag{
			Name:  "d",
			Usage: "detach containr",
		},
		cli.StringFlag{
			Name:  "m",
			Usage: "set memory limit",
		},
		cli.StringFlag{
			Name:  "v",
			Usage: "volumes",
		},
		cli.StringFlag{
			Name:  "name",
			Usage: "container name",
		},
		cli.StringFlag{
			Name:  "net",
			Usage: "container network",
		},
		cli.StringSliceFlag{
			Name:  "p",
			Usage: "port mapping",
		},
	},
	Action: func(context *cli.Context) error {
		if len(context.Args()) < 1 {
			return fmt.Errorf("missing container command")
		}
		var cmdArray []string
		for _, arg := range context.Args() {
			cmdArray = append(cmdArray, arg)
		}
		imageName := cmdArray[0]
		cmdArray = cmdArray[1:]
		tty := context.Bool("ti")
		detach := context.Bool("d")
		if tty && detach {
			return fmt.Errorf("ti and d can not yiqi")
		}
		memory_limit := context.String("m")
		volume := context.String("v")
		resourceCfg := subsystems.ResourceConfig{MemoryLimit: memory_limit}
		containerName := context.String("name")
		network := context.String("net")
		portmapping := context.StringSlice("p")
		Run(tty, cmdArray, &resourceCfg, volume, containerName, imageName, network, portmapping)
		return nil
	},
}
var initCommand = cli.Command{
	Name:  "init",
	Usage: "init daoban docker",
	Action: func(context *cli.Context) error {
		log.Infof("init come on")
		cmd := context.Args().Get(0)
		log.Infof("command %s", cmd)
		err := container.RunContainerInitProcess()
		return err
	},
}

var listCommand = cli.Command{
	Name:  "ps",
	Usage: "list all containers",
	Action: func(context *cli.Context) error {
		container.ListContainers()
		return nil
	},
}
var logCommand = cli.Command{
	Name:  "log",
	Usage: "print logs of a container",
	Action: func(context *cli.Context) error {
		if len(context.Args()) < 1 {
			return fmt.Errorf("input the container name")
		}
		containerName := context.Args().Get(0)
		container.LogContainer(containerName)
		return nil
	},
}
var execCommand = cli.Command{
	Name:  "exec",
	Usage: "exec a command into container",
	Action: func(context *cli.Context) error {
		if os.Getenv("ENV_EXEC_PID") != "" {
			log.Infof("PID CALLBACK PID %s", os.Getpid())
			return nil
		}
		if len(context.Args()) < 2 {
			return fmt.Errorf("Missing container name of command, %v", context.Args())
		}
		containerName := context.Args().Get(0)
		var commandArray []string
		for _, arg := range context.Args().Tail() {
			commandArray = append(commandArray, arg)
		}
		container.ExecContainer(containerName, commandArray)
		return nil
	},
}
var stopCommand = cli.Command{
	Name:  "stop",
	Usage: "stop a container",
	Action: func(context *cli.Context) error {
		if len(context.Args()) < 1 {
			return fmt.Errorf("missing container name")
		}
		containerName := context.Args().Get(0)
		container.StopContainer(containerName)
		return nil
	},
}
var removeCommand = cli.Command{
	Name:  "remove",
	Usage: "remove a container",
	Action: func(context *cli.Context) error {
		if len(context.Args()) < 1 {
			return fmt.Errorf("missing container name")
		}
		containerName := context.Args().Get(0)
		container.RemoveContainer(containerName)
		return nil
	},
}
var commitCommand = cli.Command{
	Name:  "commit",
	Usage: "commit a container into image",
	Action: func(context *cli.Context) error {
		if len(context.Args()) < 2 {
			return fmt.Errorf("missing container name and image name")
		}
		containerName := context.Args().Get(0)
		imageName := context.Args().Get(1)
		container.CommitContainer(containerName, imageName)
		return nil
	},
}
var networkCommand = cli.Command{
	Name:  "network",
	Usage: "container network commands",
	Subcommands: []cli.Command{
		{
			Name:  "create",
			Usage: "create a container network",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "driver",
					Usage: "network driver",
				},
				cli.StringFlag{
					Name:  "subnet",
					Usage: "subnet cidr",
				},
			},
			Action: func(context *cli.Context) error {
				if len(context.Args()) < 1 {
					return fmt.Errorf("missing network name")
				}
				network.Init()
				err := network.CreateNetwork(context.String("driver"),
					context.String("subnet"), context.Args()[0])
				if err != nil {
					return fmt.Errorf("create network error %+v", err)
				}
				return nil
			},
		},
		{
			Name:  "list",
			Usage: "list container network",
			Action: func(context *cli.Context) error {
				network.Init()
				network.ListNetwork()
				return nil
			},
		},
		{
			Name:  "remove",
			Usage: "remove container network",
			Action: func(context *cli.Context) error {
				if len(context.Args()) < 1 {
					return fmt.Errorf("missing network name")
				}
				network.Init()
				err := network.DeleteNetwork(context.Args()[0])
				if err != nil {
					return fmt.Errorf("remove network error:%+v", err)
				}
				return nil
			},
		},
	},
}
