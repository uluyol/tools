// Simple tool to run programs within a docker container. Useful while hacking.

package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
)

var (
	argBase  = flag.String("base", os.Getenv("DHACK_BASE"), "base docker image (DHACK_BASE)")
	argMount = flag.String("mount", os.Getenv("DHACK_MOUNT"), "mount point for directory (DHACK_MOUNT)")
	argDir   = flag.String("dir", os.Getenv("PWD"), "directory to mount (PWD)")
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s commands", os.Args[0])
	flag.PrintDefaults()
}

func main() {
	flag.Usage = usage
	flag.Parse()
	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(1)
	}
	args := []string{"run", "-v", fmt.Sprintf("%s:%s", *argDir, *argMount), "-i", "-t", *argBase}
	args = append(args, flag.Args()...)
	fmt.Print("docker")
	for _, a := range args {
		fmt.Printf(" %s", a)
	}
	fmt.Println()
	cmd := exec.Command("docker", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(255)
	}
}
