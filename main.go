package main

import (
        "flag"
        "fmt"

        pluginpb "github.com/dsrvlabs/vatz-proto/plugin/v1"
        "github.com/dsrvlabs/vatz/sdk"
	"github.com/rs/zerolog/log"
        "golang.org/x/net/context"
        "google.golang.org/protobuf/types/known/structpb"

        "bufio"
        "runtime"
	"io"
        "io/ioutil"
        "os"
	"os/exec"
        "strconv"
        "strings"
        "sync"
	"github.com/hpcloud/tail"
)

const (
        // Default values.
        defaultAddr = "127.0.0.1"
        defaultPort = 9095

        pluginName = "mev-monitor"
        procDir = "/proc"
        procName = "mev-boost"
	    pathScript = "/root/bin/mev/mev-boost/mev_boost_start.sh"
		pathLog = "/var/log/mev-boost.log"
)

var (
        addr string
        port int
        pidMevBoost int
	cnt int
)

func init() {
        flag.StringVar(&addr, "addr", defaultAddr, "IP Address(e.g. 0.0.0.0, 127.0.0.1)")
        flag.IntVar(&port, "port", defaultPort, "Port number, default 9095")

        flag.Parse()
	cnt = 0
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
        res := checkProcName()
        var msg string
	var hostname string

        if res {
                state = pluginpb.STATE_SUCCESS
		hostname, _ = os.Hostname()
                msg = fmt.Sprintf("[%s]mev-boost is alive(%d)", hostname, pidMevBoost)
                //log.Info().Str("module", pluginName).Msg(msg)
        } else {
                severity = pluginpb.SEVERITY_CRITICAL
		hostname, _ = os.Hostname()
                msg = fmt.Sprintf("[%s]mev-boost is down", hostname)
                //log.Info().Str("module", pluginName).Msg("mev-boost is down")
		buf, err := os.ReadFile(pathScript)
		//fmt.Println(string(buf))
		if err != nil {
			errMsg := fmt.Sprintf("[%s] %s", hostname, err)
			log.Info().Str("module", pluginName).Msg(errMsg)
			goto errExit
		}
		exec_command(string(buf))
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			t, err := tail.TailFile(pathLog, tail.Config{
				Follow: true,
				Location: &tail.SeekInfo{Offset: 0, Whence: io.SeekEnd},
			})
			if err != nil {
				fmt.Println(err)
			}
			for line := range t.Lines {
				fmt.Println(line.Text)
				checkRunSuccess(line.Text)
				if cnt == 2 {
					cnt = 0
					break
				}
			}
		}()
		wg.Wait()
        }
errExit:
	ret := sdk.CallResponse{
                FuncName:   "mev-boost-liveness",
                Message:    msg,
                Severity:   severity,
                State:      state,
                AlertTypes: []pluginpb.ALERT_TYPE{pluginpb.ALERT_TYPE_DISCORD},
        }

        return ret, nil
}

func checkProcName() bool {
        //Use cpus for go routine.
        runtime.GOMAXPROCS(runtime.NumCPU()/2)
        res := false
        if os.Chdir(procDir) != nil {
                fmt.Println("/proc unavailable.")
                return res
        }

        files, err := ioutil.ReadDir(".")
        if err != nil {
                fmt.Println("Unable to read /proc directory.")
                return res
        }

        var wg sync.WaitGroup
        for _, file := range files {
                wg.Add(1)
                go func(procName string, file os.FileInfo) {
                        defer wg.Done()
                        if isExistProc(procName, file) {
                                res = true
                        }
			//fmt.Println("res = ", res)
                }(procName, file)
        }
        wg.Wait()
        return res
}

func isExistProc(procName string, file os.FileInfo) bool {
        if !file.IsDir() {
                return false
        }

        pid, err := strconv.Atoi(file.Name())
        if err != nil {
                return false
        }

        f, err := os.Open(file.Name() + "/stat")
        if err != nil {
                fmt.Println("unable to open", file.Name())
        }
        defer f.Close()

        r := bufio.NewReader(f)
        scanner := bufio.NewScanner(r)
        for scanner.Scan() {
                if strings.Contains(scanner.Text(), procName) {
                        //fmt.Println(pid)
                        pidMevBoost = pid
                        return true
                }
        }
        return false
}

func exec_command(program string) {
	cmd := exec.Command("bash", "-c", program)
	cmd.Stdin = os.Stdin;
	cmd.Stdout = os.Stdout;
	cmd.Stderr = os.Stderr;
	err := cmd.Start()
	if err != nil {
		fmt.Printf("%v\n", err)
	}
}

func checkRunSuccess(str string) {
	logMap := map[int]string{
		0: "level=debug msg=registerValidator method=registerValidator module=service",
		1: "method=POST module=service path=/eth/v1/builder/validators status=200",
	}
	for _, value := range logMap {
		if strings.Contains(str, value) == true {
			fmt.Println(str, value)
			cnt++
			break
		}
	}
}
