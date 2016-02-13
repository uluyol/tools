package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

var (
	command  string
	timeout  uint64
	username string
	hosts    []string
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s command timeout user hosts...\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "timeout should be in seconds (can be a float)\n")
	os.Exit(1)
}

func parseArgs() {
	if len(os.Args) < 5 {
		usage()
	}
	command = os.Args[1]
	timeoutf, err := strconv.ParseFloat(os.Args[2], 32)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error while parsing timeout: %v\n", err)
		usage()
	}
	timeout = uint64(timeoutf * 1000) // convert to ms
	username = os.Args[3]
	hosts = os.Args[4:]
}

type Req struct {
	Action  string
	Cmd     string
	Timeout uint64
	Hosts   []string
}

type InitializeComplete struct {
	InitializeComplete bool
}

type ConnectionProgress struct {
	ConnectedHost string
}

type UserError struct {
	IsCritical bool
	ErrorMsg   string
}

type Reply struct {
	Hostname string
	Stdout   string
	Stderr   string
	Success  bool
	ErrMsg   string
}

type FinalReply struct {
	TotalTime     float64
	TimedOutHosts map[string]bool
}

type Resp struct {
	Type               string
	InitializeComplete `json:",omitempty,inline"`
	ConnectionProgress `json:",omitempty,inline"`
	UserError          `json:",omitempty,inline"`
	Reply              `json:",omitempty,inline"`
	FinalReply         `json:",omitempty,inline"`
}

func main() {
	parseArgs()
	cmd := exec.Command("GoSSHa", "-l", username)
	send, err := cmd.StdinPipe()
	if err != nil {
		die("failed to create stdin pipe: %v\n", err)
	}
	recv, err := cmd.StdoutPipe()
	if err != nil {
		die("failed to create stdout pipe: %v\n", err)
	}
	if err := cmd.Start(); err != nil {
		die("failed to start parallel ssh agent: %v\n", err)
	}
	scanner := bufio.NewScanner(recv)
	if !scanner.Scan() {
		die("error occurred while initializing parallel ssh agent: %v\n", scanner.Err())
	}
	var resp Resp
	if err := json.Unmarshal(scanner.Bytes(), &resp); err != nil {
		die("error decoding response: %v\n", err)
	}
	if resp.Type != "InitializeComplete" || !resp.InitializeComplete.InitializeComplete {
		die("error occurred while initilizing parallel ssh agent\n")
	}
	enc := json.NewEncoder(send)
	req := Req{
		Action:  "ssh",
		Cmd:     command,
		Timeout: timeout,
		Hosts:   hosts,
	}
	if err := enc.Encode(&req); err != nil {
		die("error occurred while encoding request: %v\n", err)
	}
	cleanExit := true
	for scanner.Scan() {
		var resp Resp
		if err := json.Unmarshal(scanner.Bytes(), &resp); err != nil {
			fmt.Fprintf(os.Stderr, "error occured decoding response from parallel ssh agent: %v\n", err)
			cleanExit = false
			continue
		}
		switch resp.Type {
		case "ConnectionProgress":
			fmt.Printf("Connected to %s\n", resp.ConnectionProgress.ConnectedHost)
		case "UserError":
			if resp.UserError.IsCritical {
				fmt.Fprintf(os.Stderr, "Error: %s\n", resp.UserError.ErrorMsg)
				cmd.Process.Kill()
				cleanExit = false
			} else {
				fmt.Fprintf(os.Stderr, "Warning: %s\n", resp.UserError.ErrorMsg)
			}
		case "Reply":
			fmt.Printf("==== Response from %s ====\n", resp.Reply.Hostname)
			fmt.Printf(" - Success: %t\n", resp.Reply.Success)
			if resp.Reply.ErrMsg != "" {
				fmt.Printf(" - ErrMsg: %s\n", resp.Reply.ErrMsg)
			}
			if resp.Reply.Stdout != "" {
				fmt.Println(" - Stdout:")
				fmt.Println(resp.Reply.Stdout)
			}
			if resp.Reply.Stderr != "" {
				fmt.Println(" - Stderr:")
				fmt.Println(resp.Reply.Stderr)
			}
		case "FinalReply":
			fmt.Println("==== Summary ====")
			fmt.Printf("- Execution Time: %f\n", resp.FinalReply.TotalTime)
			var timedout []string
			for h, t := range resp.FinalReply.TimedOutHosts {
				if t {
					timedout = append(timedout, h)
				}
			}
			if len(timedout) > 0 {
				fmt.Printf("- Timed out: %s\n", strings.Join(timedout, " "))
			}
			cmd.Process.Kill()
		default:
			fmt.Printf("Unexpected response: %s\n", scanner.Text())
			cmd.Process.Kill()
		}
	}
	if !cleanExit {
		os.Exit(2)
	}
}

func die(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format, args)
	os.Exit(1)
}
