package subsystems

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"strings"
)

type ResourceConfig struct {
	MemoryLimit string
	CpuSharre   string
	CpuSet      string
}
type Subsystem interface {
	Name() string
	Set(path string, res *ResourceConfig) error
	Apply(path string, pid int) error
	Remove(path string) error
}

var (
	SubsystemsIns = []Subsystem{
		&MemorySubSystem{},
	}
)

func FindCgroupMountpoint(subsystem string) string {
	f, err := os.Open("/proc/self/mountinfo")
	if err != nil {
		return ""
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		txt := scanner.Text()
		fields := strings.Split(txt, " ")
		for _, opt := range strings.Split(fields[len(fields)-1], ",") {
			if opt == subsystem {
				return fields[4]
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return ""
	}
	return ""
}

func GetCgroupPath(subsystem string, cgroupPath string, autoCreate bool) (string, error) {
	cgroupRoot := FindCgroupMountpoint(subsystem)
	if _, err := os.Stat(path.Join(cgroupRoot, cgroupPath)); err == nil || (autoCreate && os.IsNotExist(err)) {
		if os.IsNotExist(err) {
			if err := os.Mkdir(path.Join(cgroupRoot, cgroupPath), 0777); err == nil {

			} else {
				return "", fmt.Errorf("error create cgroup %v", err)
			}
		}
		return path.Join(cgroupRoot, cgroupPath), nil
	} else {
		return "", fmt.Errorf("cgroup path error %v", err)
	}
}
