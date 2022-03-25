package container

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	log "github.com/sirupsen/logrus"
)

var (
	RUNNING              string = "running"
	STOP                 string = "stopped"
	Exit                 string = "exited"
	DefaultInfoLodcation string = "/var/run/mydocker/%s/"
	ConfigName           string = "config.json"
	ContainerLogFile     string = "container.log"
)

type ContainerInfo struct {
	Pid         string   `json:"Pid"`
	Id          string   `json:"id"`
	Name        string   `json:"name"`
	Command     string   `json:"command"`
	CreatedTime string   `json:"createTime"`
	Status      string   `json:"status"`
	Volume      string   `json:"volume"`
	PortMapping []string `json:"portmapping"`
}

func RecordContainerInfo(containerID string, containerPID int, commandArray []string, containerName, volume string) (string, error) {

	id := containerID
	createTime := time.Now().Format("2006-01-02 15:04:04")
	command := strings.Join(commandArray, "")
	if containerName == "" {
		containerName = id
	}
	containerInfo := &ContainerInfo{
		Id:          id,
		Pid:         strconv.Itoa(containerPID),
		Command:     command,
		CreatedTime: createTime,
		Status:      RUNNING,
		Name:        containerName,
		Volume:      volume,
	}
	jsonBytes, err := json.Marshal(containerInfo)
	if err != nil {
		log.Errorf("Record container info err %v", err)
		return "", err
	}
	jsonStr := string(jsonBytes)
	dirUrl := fmt.Sprintf(DefaultInfoLodcation, containerName)
	if err := os.MkdirAll(dirUrl, 0777); err != nil {
		log.Errorf("Mkdir error %s error %v", dirUrl, err)
		return "", err
	}
	fileName := dirUrl + "/" + ConfigName
	file, err := os.Create(fileName)
	defer file.Close()
	if err != nil {
		log.Errorf("cannot create file %s error %v", fileName, err)
		return "", err
	}
	if _, err := file.WriteString(jsonStr); err != nil {
		log.Errorf("File write string err %v", err)
		return "", err
	}
	return containerName, nil

}
func DeleteContainerInfo(containerName string) {
	dirURL := fmt.Sprintf(DefaultInfoLodcation, containerName)
	if err := os.RemoveAll(dirURL); err != nil {
		log.Errorf("remove dir %s error %v", dirURL, err)
	}
}
func ListContainers() {
	dirURL := fmt.Sprintf(DefaultInfoLodcation, "")
	dirURL = dirURL[:len(dirURL)-1]
	files, err := ioutil.ReadDir(dirURL)
	if err != nil {
		log.Errorf("Read dir %s error %v", dirURL, err)
		return
	}
	var containers []*ContainerInfo
	for _, file := range files {
		tmpContainer, err := getContainerInfo(file)
		if err != nil {
			log.Errorf("get container info err %v", err)
			continue
		}
		containers = append(containers, tmpContainer)
	}
	w := tabwriter.NewWriter(os.Stdout, 12, 1, 3, ' ', 0)
	fmt.Fprint(w, "ID\tName\tSTATUS\tCOMMAND\tCreated\n")
	for _, item := range containers {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			item.Id,
			item.Name,
			item.Pid,
			item.Status,
			item.Command,
			item.CreatedTime,
		)
	}
	if err := w.Flush(); err != nil {
		log.Errorf("Flush err %v", err)
		return
	}
}
func getContainerInfo(file os.FileInfo) (*ContainerInfo, error) {
	containerName := file.Name()
	configFileDir := fmt.Sprintf(DefaultInfoLodcation, containerName)
	configFileDir = configFileDir + ConfigName
	content, err := ioutil.ReadFile(configFileDir)
	if err != nil {
		log.Errorf("Read file %s err %v", configFileDir, err)
		return nil, err
	}
	var containerInfo ContainerInfo
	if err := json.Unmarshal(content, &containerInfo); err != nil {
		log.Errorf("Json unmarshal error %v", err)
		return nil, err
	}
	return &containerInfo, nil
}
func LogContainer(containerName string) {
	dirURL := fmt.Sprintf(DefaultInfoLodcation, containerName)
	logFileLocation := dirURL + ContainerLogFile
	file, err := os.Open(logFileLocation)
	defer file.Close()
	if err != nil {
		log.Errorf("log file %s err %v", logFileLocation, err)
		return
	}
	content, err := ioutil.ReadAll(file)
	if err != nil {
		log.Errorf("log file reaf %s err %v", logFileLocation, err)
		return
	}
	fmt.Fprint(os.Stdout, string(content))
}
