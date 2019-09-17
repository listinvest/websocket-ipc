// Author :		Eric<eehsiao@gmail.com>

package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	ipc "github.com/eehsiao/websocket-ipc"
	"github.com/takama/daemon"
)

type Service struct {
	daemon.Daemon
}

var (
	dependencies   = []string{}
	stdlog, errlog *log.Logger
	service        *Service
	wsIpc          *ipc.IPC

	verNo = "v0.0.1"
)

func init() {
	stdlog = log.New(os.Stdout, "", log.Ldate|log.Ltime)
	errlog = log.New(os.Stderr, "", log.Ldate|log.Ltime)
}

func main() {
	args := os.Args
	if len(args) == 1 {
		errlog.Println("need a command ex: example start")
		return
	}

	srv, err := daemon.New("deamon-example", "deamon-example", dependencies...)
	if err != nil {
		errlog.Println("Error: ", err)
		os.Exit(1)
	}
	service = &Service{srv}
	if service == nil {
		errlog.Println("Error: no service")
		os.Exit(1)
	}

	switch args[1] {
	case "install":
		if status, err := service.Install(); err != nil {
			errlog.Println(status, "\nError: ", err)
			os.Exit(1)
		}
	case "remove":
		if status, err := service.Remove(); err != nil {
			errlog.Println(status, "\nError: ", err)
			os.Exit(1)
		}
	case "start":
		if status, err := service.Start(); err != nil {
			errlog.Println(status, "\nError: ", err)
			os.Exit(1)
		}

		wsIpc = ipc.NewIpc(ipcCmd, stdlog, errlog)
		go wsIpc.WsHandel()
		status := bgLoop()
		stdlog.Println(status)
	default:
		if msg, err := ipc.SendCmd(args[1]); err == nil {
			fmt.Printf("%s\n", msg)
		}
	}
}

func bgLoop() string {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, os.Kill, syscall.SIGTERM)

	for {
		select {
		case sysInterrupt := <-interrupt:
			stdlog.Println("Got signal:", sysInterrupt)
			if status, err := service.Stop(); err != nil {
				errlog.Println(status, "\nError: ", err)
				os.Exit(1)
			}
			if sysInterrupt == os.Interrupt {
				return "Daemon was interruped by system signal"
			}
			return "Daemon was killed"
		case client := <-wsIpc.WsClient:
			go ipcCmd(client)
		}
	}
}

func ipcCmd(client *ipc.Client) {
	stdlog.Println(string(client.Msg))
	switch string(client.Msg) {
	case "stop":
		if status, err := service.Stop(); err != nil {
			errlog.Println(status, "\nError: ", err)
		}
		os.Exit(1)
	case "version":
		client.Ws.WriteJSON("{'Version':" + verNo + "}")
	default:
		client.Ws.WriteJSON("{'echo':'" + string(client.Msg) + "'}")
	}
}