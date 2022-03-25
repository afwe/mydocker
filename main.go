package main

import (
	"mydocker/cgroup"
	"mydocker/container"
	"mydocker/layer"
	"mydocker/network"
	"mydocker/subsystems"
	"mydocker/utils"
	"os"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

const usage = `mydocker is a daoban docker`

func main() {
	app := cli.NewApp()
	app.Name = "mydocker"
	app.Usage = usage
	app.Commands = []cli.Command{
		initCommand,
		runCommand,
		listCommand,
		logCommand,
		execCommand,
		stopCommand,
		removeCommand,
		networkCommand,
	}
	app.Before = func(context *cli.Context) error {
		log.SetFormatter(&log.JSONFormatter{})
		log.SetOutput(os.Stdout)
		return nil
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func Run(tty bool, comArray []string, res *subsystems.ResourceConfig, volume, containerName, imageName, nw string, portmapping []string) {
	containerID := utils.RandStringBytes(10)
	if containerName == "" {
		containerName = containerID
	}
	parent, writePipe := container.NewParentProcess(tty, containerName, volume, imageName)
	if parent == nil {
		log.Errorf("New parent process error")
		return
	}
	if err := parent.Start(); err != nil {
		log.Errorf(err.Error())
	}
	containerName, err := container.RecordContainerInfo(containerID, parent.Process.Pid, comArray, containerName, volume)
	if err != nil {
		log.Errorf("record container info err %v", err)
		return
	}
	cgroupManager := cgroup.NewCgroupManager("mydocker-cgroup")
	defer cgroupManager.Destroy()
	cgroupManager.Set(res)
	cgroupManager.Apply(parent.Process.Pid)
	if nw != "" {
		network.Init()
		containerInfo := &container.ContainerInfo{
			Id:          containerID,
			Pid:         strconv.Itoa(parent.Process.Pid),
			Name:        containerName,
			PortMapping: portmapping,
		}
		if err := network.Connect(nw, containerInfo); err != nil {
			log.Errorf("err connect network %v", err)
		}
	}
	sendInitCommand(comArray, writePipe)
	if tty {
		parent.Wait()
		container.DeleteContainerInfo(containerName)
		layer.DeleteWorkSpace(volume, containerName)
	}
	os.Exit(0)
}

func sendInitCommand(comArray []string, writePipe *os.File) {
	command := strings.Join(comArray, " ")
	log.Infof("command all is %s", command)
	writePipe.WriteString(command)
	writePipe.Close()
}
