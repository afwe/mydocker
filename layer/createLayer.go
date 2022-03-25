package layer

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	log "github.com/sirupsen/logrus"
)

func NewWorkSpace(containerName, imageName, volume string) {
	CreateReadOnlyLayer(imageName)
	CreateWriteLayer(containerName)
	CreateMountPoint(containerName, imageName)
	if volume != "" {
		volumeURLs := volumeUrlExtract(volume)
		length := len(volumeURLs)
		if length == 2 && volumeURLs[0] != "" && volumeURLs[1] != "" {
			MountVolume(containerName, volumeURLs)
			log.Infof("%q", volumeURLs)
		} else {
			log.Infof("Volume parameter is not correct")
		}
	}
}
func MountVolume(containerName string, volumeURLs []string) error {
	parentUrl := volumeURLs[0]
	if err := os.Mkdir(parentUrl, 0777); err != nil {
		log.Infof("mkdir parent dir %s error %v", parentUrl, err)
	}
	containerUrl := volumeURLs[1]
	mntURL := fmt.Sprintf(MntUrl, containerName)
	containerVolumeURL := mntURL + containerUrl
	if err := os.Mkdir(containerVolumeURL, 0777); err != nil {
		log.Infof("mkdir container dir %s error %v", containerVolumeURL, err)
	}
	dirs := "dirs=" + parentUrl
	dirs = dirs + ",xino=/dev/shm/aufs.xino"
	cmd := exec.Command("mount", "-t", "aufs", "-o", dirs, "none", containerVolumeURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("mount volume fail %v", err)
	}
	return nil
}
func volumeUrlExtract(volume string) []string {
	var volumeURLs []string
	volumeURLs = strings.Split(volume, ":")
	return volumeURLs
}
func CreateReadOnlyLayer(imageName string) {
	unTarFoldURL := RootUrl + "/" + imageName
	imageURL := RootUrl + "/" + imageName + ".tar"
	exist, err := PathExists(unTarFoldURL)
	if err != nil {
		log.Infof("Fatal to panduan dir %s exists %v", unTarFoldURL, err)
	}
	if exist == false {
		if err := os.MkdirAll(unTarFoldURL, 0777); err != nil {
			log.Errorf("Mkdir %s error,%v", imageURL, err)
		}
	}
	if _, err := exec.Command("tar", "-xvf", imageURL, "-C", unTarFoldURL).CombinedOutput(); err != nil {
		log.Errorf("unTar dir %s error %v", imageURL, err)
	}
}
func CreateWriteLayer(containerName string) {
	writeURL := fmt.Sprintf(WriteLayerUrl, containerName)
	if err := os.MkdirAll(writeURL, 0777); err != nil {
		log.Errorf("Mkdir %s error, %v", writeURL, err)
	}
}

func CreateMountPoint(containerName, imageName string) {
	mntURL := fmt.Sprintf(MntUrl, containerName)
	mntURL = mntURL + "/"
	if err := os.Mkdir(mntURL, 0777); err != nil {
		log.Errorf("Mkdir %s error, %v", mntURL, err)
	}
	tmpWriteLayer := fmt.Sprintf(WriteLayerUrl, containerName)
	tmpImageLocation := RootUrl + "/" + imageName
	dirs := "dirs=" + tmpWriteLayer + ":" + tmpImageLocation
	dirs = dirs + ",xino=/dev/shm/aufs.xino"
	cmd := exec.Command("mount", "-t", "aufs", "-o", dirs, "none", mntURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("%v", err)
	}
}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, nil
}
