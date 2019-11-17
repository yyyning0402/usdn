package funcs

import (
	"fmt"
	"net"
	"os"
	"time"
	"usdn/utils"

	"github.com/digineo/go-ping"
	"github.com/digineo/go-ping/monitor"
)

var (
	pinger *ping.Pinger
)

type Monitor struct {
	*monitor.Monitor
}

func ActiveMonitor() (Monitor, error) {
	targets := make(map[string]host)

	m := newMonitor()
	m.AddOrCleanTargets(targets)

	return m, nil
}

func newMonitor() Monitor {
	pingInterval := time.Duration(utils.Config().Ping.Interval) * time.Second
	pingTimeout := time.Duration(utils.Config().Ping.TimeOut) * time.Second
	size := 100
	utils.Logger.Debugf("init monitor interval %v timeout %v size %v", pingInterval, pingTimeout, size)

	// Bind to sockets
	if p, err := ping.New("0.0.0.0", "::"); err != nil {
		utils.Logger.Errorln("Unable to bind: %s\nRunning as root?\n", err)
		os.Exit(2)
	} else {
		pinger = p
	}
	pinger.SetPayloadSize(uint16(size))

	return Monitor{
		monitor.New(pinger, pingInterval, pingTimeout),
	}
}

func (m Monitor) AddOrCleanTargets(targets map[string]host) {

	ct := make(map[string]string)
	for k, _ := range targets {
		// Add targets
		ipAddr, err := net.ResolveIPAddr("", k)
		if err != nil {
			fmt.Printf("invalid target '%s': %s", k, err)
			continue
		}
		m.AddTargetDelayed(k, *ipAddr, 10*time.Millisecond)
		if k != App.SerfClient.Name {
			ct[k] = ""
		}
	}

	m.CleanTarget(ct)
}

func (m Monitor) ExportToSubject() (map[string]int, error) {
	pingMetrics := make(map[string]int)
	defer func() {
		err := recover()
		if err != nil {
			utils.Logger.Errorln(err)
		}
	}()

	for i, metrics := range m.ExportAndClear() {

		// if math.IsNaN(float64(metrics.Mean)) {
		// 	pingMetrics[i] = 3000
		// 	continue
		// } else {
		// 	pingMetrics[i] = int(metrics.Mean)
		// }
		lostper := metrics.PacketsLost / metrics.PacketsSent * 100
		utils.Logger.Debugf("%s lost %v sent %v mean %v", i, metrics.PacketsLost, metrics.PacketsSent, metrics.Mean)
		if lostper >= utils.Config().Ping.Lostper && lostper != 0 {
			pingMetrics[i] = 3000
		} else {
			pingMetrics[i] = int(metrics.Mean)
		}
	}
	return pingMetrics, nil
}
