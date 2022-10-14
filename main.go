package main

import (
	"flag"
	"fmt"

	pluginpb "github.com/dsrvlabs/vatz-proto/plugin/v1"
	"github.com/dsrvlabs/vatz/sdk"
	"github.com/rs/zerolog/log"
	"golang.org/x/net/context"
	"google.golang.org/protobuf/types/known/structpb"

	"net/http"
	"os"
	"os/exec"
)

const (
	// Default values.
	defaultAddr = "127.0.0.1"
	defaultPort = 9095

	pluginName      = "mev-monitor"
	dockerContainer = "mev-boost-v1.3.2"
)

var (
	addr          string
	mevDockerName string
	port          int
)

func init() {
	flag.StringVar(&addr, "addr", defaultAddr, "IP Address(e.g. 0.0.0.0, 127.0.0.1)")
	flag.StringVar(&mevDockerName, "mev-boost-version", dockerContainer, "Please check your mev docker container name")
	flag.IntVar(&port, "port", defaultPort, "Port number, default 9095")

	flag.Parse()
}

func main() {
	p := sdk.NewPlugin(pluginName)
	p.Register(pluginFeature)

	ctx := context.Background()
	if err := p.Start(ctx, addr, port); err != nil {
		fmt.Println("exit")
	}
}

func pluginFeature(info, option map[string]*structpb.Value) (sdk.CallResponse, error) {
	// TODO: Fill here.
	severity := pluginpb.SEVERITY_INFO
	state := pluginpb.STATE_NONE
	var msg string
	var hostname string
	url := "http://localhost:18550/eth/v1/builder/status"

	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != http.StatusOK {
		severity = pluginpb.SEVERITY_CRITICAL
		hostname, _ = os.Hostname()
		msg = fmt.Sprintf("[%s]mev-boost is down", hostname)
		log.Info().Str("moudle", "plugin").Msg(msg)
		exec_command("docker stop " + mevDockerName)
		exec_command("docker start " + mevDockerName)
	} else {
		state = pluginpb.STATE_SUCCESS
		hostname, _ = os.Hostname()
		msg = fmt.Sprintf("[%s]mev-boost is alive", hostname)
		log.Info().Str("moudle", "plugin").Msg(msg)
	}

	ret := sdk.CallResponse{
		FuncName:   "mev-boost-liveness",
		Message:    msg,
		Severity:   severity,
		State:      state,
		AlertTypes: []pluginpb.ALERT_TYPE{pluginpb.ALERT_TYPE_DISCORD},
	}

	resp.Body.Close()

	return ret, nil
}

func exec_command(program string) {
	cmd := exec.Command("bash", "-c", program)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Start()
	if err != nil {
		fmt.Printf("%v\n", err)
	}
}
