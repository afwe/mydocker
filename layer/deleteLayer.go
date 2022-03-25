package layer

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	log "github.com/sirupsen/logrus"
)

func DeleteWorkSpace(volume, containerName string) {
	if volume != "" {
		volumeURLs := strings.Split(volume, ":")
		length := len(volumeURLs)
		if length == 2 && volumeURLs[0] != "" && volumeURLs[1] != "" {
			DeleteMountPointWithVolume(containerName, volumeURLs)
		} else {
			DeleteMountPoint(containerName)
		}
	} else {
		DeleteMountPoint(containerName)
	}
	DeleteWriteLayer(containerName)
}
func DeleteMountPointWithVolume(containerName string, volumeURLs []string) {
	mntURL := fmt.Sprintf(MntUrl, containerName)
	containerUrl := mntURL + volumeURLs[1]
	cmd := exec.Command("umount", "-A", containerUrl)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("Unmount volume fail %v", err)
	}
	cmd = exec.Command("umount", "-A", mntURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("Unmount mntpoint fail %v", err)
	}
	if err := os.RemoveAll(mntURL); err != nil {
		log.Infof("Remove mountpoint dir %s error %v", mntURL, err)
	}
	if err := os.RemoveAll(volumeURLs[0]); err != nil {
		log.Infof("Remove mountpoint dir %s error %v", mntURL, err)
	}
}
func DeleteMountPoint(containerName string) error {
	mntURL := fmt.Sprintf(MntUrl, containerName)
	cmd := exec.Command("umount", "-A", mntURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("%v", err)
	}
	if err := os.RemoveAll(mntURL); err != nil {
		log.Errorf("Remove mntdir %s error %v", mntURL, err)
	}
	return nil
}
func DeleteWriteLayer(containerName string) {
	writeURL := fmt.Sprintf(WriteLayerUrl, containerName)
	if err := os.RemoveAll(writeURL); err != nil {
		log.Errorf("Remove writeURL %s error %v", writeURL, err)
	}
}
