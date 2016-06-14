package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/pflag"
)

var (
	command    []string
	timeoutDur time.Duration
	timeoutMS  uint64
	username   string
	hosts      []string
)

func init() {
	pflag.DurationVarP(&timeoutDur, "timeout", "t", 10*time.Second, "timeout for the command to complete")
	pflag.StringVarP(&username, "user", "u", "root", "user to login with")
	pflag.StringSliceVarP(&hosts, "hosts", "h", nil, "hosts to run command on")
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [options] -- command\n", os.Args[0])
	pflag.PrintDefaults()
}

func parseArgs() {
	pflag.Usage = usage
	pflag.Parse()
	if pflag.NArg() < 1 {
		log.Fatal("must specify a command")
	}
	command = pflag.Args()
	timeoutMS = uint64(timeoutDur / time.Millisecond)
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

func run() error {
	cmd := exec.Command("GoSSHa", "-l", username)
	send, err := cmd.StdinPipe()
	if err != nil {
		return errors.Wrap(err, "failed to create stdin pipe")
	}
	recv, err := cmd.StdoutPipe()
	if err != nil {
		return errors.Wrap(err, "failed to create stdout pipe")
	}
	if err := cmd.Start(); err != nil {
		return errors.Wrap(err, "failed to start parallel ssh agent")
	}
	defer cmd.Process.Kill()
	scanner := bufio.NewScanner(recv)
	if !scanner.Scan() {
		return errors.Wrap(scanner.Err(), "error occurred during parallel ssh agent init")
	}
	var resp Resp
	if err := json.Unmarshal(scanner.Bytes(), &resp); err != nil {
		return errors.Wrap(err, "error decoding response")
	}
	if resp.Type != "InitializeComplete" || !resp.InitializeComplete.InitializeComplete {
		return errors.New("error occurred during parallel ssh agent init")
	}
	enc := json.NewEncoder(send)
	req := Req{
		Action:  "ssh",
		Cmd:     strings.Join(command, " "),
		Timeout: timeoutMS,
		Hosts:   hosts,
	}
	if err := enc.Encode(&req); err != nil {
		return errors.Wrap(err, "error occurred while encoding request: %v\n")
	}
	var failedHosts []string
	for scanner.Scan() {
		var resp Resp
		if err := json.Unmarshal(scanner.Bytes(), &resp); err != nil {
			errors.Wrap(err, "error occured decoding response from parallel ssh agent")
			continue
		}
		switch resp.Type {
		case "ConnectionProgress":
			log.Printf("[%s] connected\n", resp.ConnectionProgress.ConnectedHost)
		case "UserError":
			if resp.UserError.IsCritical {
				return errors.New("critical error: " + resp.UserError.ErrorMsg)
			} else {
				log.Printf("warning: %s\n", resp.UserError.ErrorMsg)
			}
		case "Reply":
			status := "success"
			if !resp.Reply.Success {
				status = "failure"
				failedHosts = append(failedHosts, resp.Reply.Hostname)
			}
			log.Printf("[%s] %s", resp.Reply.Hostname, status)
			if resp.Reply.ErrMsg != "" {
				log.Printf("[%s] error message: %s", resp.Reply.Hostname, resp.Reply.ErrMsg)
			}
			if resp.Reply.Stdout != "" {
				log.Printf("[%s] stdout:\n%s", resp.Reply.Stdout)
			}
			if resp.Reply.Stderr != "" {
				log.Printf("[%s] stderr:\n%s", resp.Reply.Stderr)
			}
		case "FinalReply":
			fmt.Println("summary:")
			fmt.Printf("\texecution time: %f s\n", resp.FinalReply.TotalTime)
			var timedout []string
			for h, t := range resp.FinalReply.TimedOutHosts {
				if t {
					timedout = append(timedout, h)
				}
			}
			if len(timedout) > 0 {
				fmt.Printf("\ttimed out: %s\n", strings.Join(timedout, " "))
				return errors.New("some hosts timed out")
			}
			if len(failedHosts) > 0 {
				return errors.Errorf("failed hosts: %s", strings.Join(failedHosts, " "))
			}
			return nil
		default:
			return errors.New("unexpected response: " + scanner.Text())
		}
	}
	return nil
}

func main() {
	log.SetPrefix("psshcmd: ")
	log.SetFlags(0)
	parseArgs()
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(2)
	}
}
