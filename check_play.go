package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"syscall"
)

var (
	flpid    *int    = flag.Int("pid", 0, "pid value")
	flconfig *string = flag.String("c", "/etc/check_play.json", "Config file")
)

type Application struct {
	Name     string `json:"name"`
	Home     string `json:"home"`
	StartCmd string `json:"startcmd"`
	User     string `json:"user"`
}

type Configuration struct {
	MaxFailure   int           `json:"maxfailure"`
	Applications []Application `json:"apps"`
}

func getConfig(config string) (configuration Configuration, err error) {
	configuration = Configuration{}
	fconfig, err := ioutil.ReadFile(config)
	if err != nil {
		return
	}

	err = json.Unmarshal(fconfig, &configuration)

	if err != nil {
		return
	}

	return
}

func checkpid(pid int) (err error) {
	p, err := os.FindProcess(pid)

	if err != nil {
		return
	}

	err = p.Signal(syscall.Signal(0))

	if err != nil && err.Error() == "no such process" {
		return
	}

	return
}

func checkrunapp(app Application) (err error) {
	pidfile := fmt.Sprintf("%s/server.pid", app.Home)
	bpidapp, err := ioutil.ReadFile(pidfile)
	if err != nil {
		return
	}

	pidapp, err := strconv.Atoi(strings.TrimSpace(string(bpidapp)))

	if err != nil {
		return
	}

	err = checkpid(int(pidapp))
	if err != nil {
		return
	}
	return
}

func main() {

	flag.Parse()

	if syscall.Getuid() != 0 {
		fmt.Printf("I need to root !")
		os.Exit(3)
	}

	config, err := getConfig(*flconfig)

	if err != nil {
		fmt.Printf("err:%s\n", err)
		os.Exit(3)
	}

	nbfailure := 0
	msg := ""
	for _, app := range config.Applications {
		err := checkrunapp(app)
		if err != nil {
			nbfailure += 1
			msg += app.Name + " "
		}
	}

	if nbfailure != 0 {
		fmt.Printf(msg + "is down")
	} else {
		fmt.Printf("All Play! apps run")
	}

	if nbfailure >= config.MaxFailure {
		os.Exit(2)
	} else if nbfailure <= config.MaxFailure && nbfailure != 0 {
		os.Exit(1)
	}

	os.Exit(0)

}
