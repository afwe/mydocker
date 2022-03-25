package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {
	pid := os.Getpid()
	fmt.Println(pid)
	cmd := exec.Command("/proc/self/exe")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
	pid = os.Getpid()
	fmt.Println(pid)
}
