package container

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mydocker/layer"
	_ "mydocker/nsenter"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"

	log "github.com/sirupsen/logrus"
)

const ENV_EXEC_PID = "mydocker_pid"
const ENV_EXEC_CMD = "mydocker_cmd"

func getContainerPidByName(containerName string) (string, error) {
	dirURL := fmt.Sprintf(DefaultInfoLodcation, containerName)
	configFilePath := dirURL + ConfigName
	contentBytes, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return "", err
	}
	var containerInfo ContainerInfo
	if err := json.Unmarshal(contentBytes, &containerInfo); err != nil {
		return "", err
	}
	return containerInfo.Pid, nil
}
func ExecContainer(containerName string, comArray []string) {
	pid, err := getContainerPidByName(containerName)
	if err != nil {
		log.Errorf("exec container getcpn %s err %v", containerName, err)
		return
	}
	cmdStr := strings.Join(comArray, " ")
	log.Infof("container pid %s", pid)
	log.Infof("command %s", cmdStr)
	cmd := exec.Command("/proc/self/exe", "exec")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	os.Setenv(ENV_EXEC_PID, pid)
	os.Setenv(ENV_EXEC_CMD, cmdStr)
	if err := cmd.Run(); err != nil {
		log.Errorf("exec container % s err %v", containerName, err)
	}
}
func getContainerInfoByName(containerName string) (*ContainerInfo, error) {
	dirURL := fmt.Sprintf(DefaultInfoLodcation, containerName)
	configFilePath := dirURL + ConfigName
	contentBytes, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		log.Errorf("read file %s err %v", configFilePath, err)
	}
	var containerInfo ContainerInfo
	if err := json.Unmarshal(contentBytes, &containerInfo); err != nil {
		log.Errorf("getcontanerinfo unmarshal fail")
		return nil, err
	}
	return &containerInfo, nil
}
func StopContainer(containerName string) {
	pid, err := getContainerPidByName(containerName)
	if err != nil {
		log.Errorf(" get container pid by name %s err %v", containerName, err)
		return
	}
	pidInt, err := strconv.Atoi(pid)
	if err != nil {
		log.Errorf("conver pid atoi fail %v", err)
		return
	}
	if err := syscall.Kill(pidInt, syscall.SIGTERM); err != nil {
		log.Errorf("stop %s err %v", containerName, err)
		return
	}
	containerInfo, err := getContainerInfoByName(containerName)
	if err != nil {
		log.Errorf("get container%s info err %v", containerName, err)
		return
	}
	containerInfo.Status = STOP
	containerInfo.Pid = " "
	newContainerBytes, err := json.Marshal(containerInfo)
	if err != nil {
		log.Errorf("json marshal %s fail %v", containerName, err)
		return
	}
	dirURL := fmt.Sprintf(DefaultInfoLodcation, containerName)
	configFilePath := dirURL + ConfigName
	if err := ioutil.WriteFile(configFilePath, newContainerBytes, 0777); err != nil {
		log.Errorf("writefile %s err %v", configFilePath, err)
	}
}
func RemoveContainer(containerName string) {
	containerInfo, err := getContainerInfoByName(containerName)
	if err != nil {
		log.Errorf("get container %s info err %v", containerName, err)
		return
	}
	if containerInfo.Status != STOP {
		log.Errorf("cannt rm running container")
		return
	}
	dirURL := fmt.Sprintf(DefaultInfoLodcation, containerName)
	if err := os.RemoveAll(dirURL); err != nil {
		log.Errorf("remove files in %s error %v", dirURL, err)
		return
	}
}
func CommitContainer(containerName, imageName string) {
	mntURL := fmt.Sprintf(layer.MntUrl, containerName)
	mntURL += "/"
	imageTar := layer.RootUrl + "/" + imageName + ".tar"
	if _, err := exec.Command("tar", "-czf", imageTar, "-c", mntURL, ".").CombinedOutput(); err != nil {
		log.Errorf("tar folder %s error %v", mntURL, err)
	}
}
