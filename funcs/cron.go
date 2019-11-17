package funcs

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"usdn/utils"
)

type Channel struct {
	Name  string
	Fn    func()
	Start int
	End   int
	Last  int
}

func NewSplitMinute() map[string]*Channel {
	tasks := map[string]map[string]func(){
		"event": map[string]func(){
			"0-10-15": SendEvent,
		},
		"merge": map[string]func(){
			"20-23-45": MergeAndCalc,
		},
		"commit": map[string]func(){
			"50-50-53": Commit,
		},
	}
	channels := make(map[string]*Channel)
	for k, v := range tasks {
		for t, f := range v {
			timesli := strings.Split(t, "-")
			min, _ := strconv.Atoi(timesli[0])
			max, _ := strconv.Atoi(timesli[1])
			end, _ := strconv.Atoi(timesli[2])
			start := utils.Random(min, max)
			channels[k] = &Channel{
				Name:  k,
				Fn:    f,
				Start: start,
				End:   end,
			}
		}
	}
	return channels
}

func Cron() {
	onemin := NewSplitMinute()
	quit := make(chan int)
	utils.Logger.Debugln(onemin)
	ticker := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-ticker.C:
			timestamp := utils.CurTimeInt()

			rem := timestamp % 60
			for k, v := range onemin {
				// go func(rem int, k string, v *Channel) {
				// 	if rem >= v.Start && rem <= v.End {
				// 		utils.Logger.Debugf("start job %s", v.Name)
				// 		utils.Logger.Debugf("cron %s last %v", k, v.Last)
				// 		if timestamp > v.Last {
				// 			v.Fn()
				// 			v.Last = timestamp + v.End - v.Start
				// 		}
				// 	}
				// }(rem, k, v)
				if rem >= v.Start && rem <= v.End {
					// utils.Logger.Debugf("start job %s", v.Name)
					// utils.Logger.Debugf("cron %s last %v", k, v.Last)
					if timestamp > v.Last {
						utils.Logger.Debugf("start cron %s last %v", k, v.Last)
						go v.Fn()
						v.Last = timestamp + v.End - v.Start
					}
				}
			}
		case <-quit:
			fmt.Println("work well .")
			ticker.Stop()
			return
		default:
			time.Sleep(time.Duration(1 * time.Second))
		}
	}
	quit <- 1
}

func UpdateHost() {
	t := time.NewTicker(time.Second * time.Duration(10))
	defer t.Stop()
	for {
		<-t.C
		toUpdate := make(map[string]host)
		members, err := App.SerfClient.Members()
		if err != nil {
			utils.Logger.Warningf("get members error %v", err)
			continue
		}
		for _, m := range members {
			var (
				tunnelip, mac string
				ok            bool
			)
			srcip := string(m.Addr)
			if tunnelip, ok = m.Tags["tunnelip"]; !ok {
				tunnelip = ""
			}

			if mac, ok = m.Tags["mac"]; !ok {
				mac = ""
			}

			cur := host{
				srcip:    srcip,
				tunnelip: tunnelip,
				mac:      mac,
				status:   m.Status,
				update:   utils.CurTimeInt(),
			}
			if !comparehost(m.Name, cur) {
				toUpdate[m.Name] = cur
			}
		}
		for k, v := range toUpdate {
			App.Hosts[k] = v
		}

		App.Ping.AddOrCleanTargets(App.Hosts)
	}
}

func Wait0() {
	ticker := time.NewTicker(time.Second * time.Duration(1))
	defer ticker.Stop()

	for {
		<-ticker.C
		timestamp := utils.CurTimeInt()
		rem := timestamp % 60
		if rem >= 0 && rem <= 5 {
			break
		}
		utils.Logger.Debugf("wait rem %v", rem)
		time.Sleep(time.Duration(1 * time.Second))
	}
}

func comparehost(name string, new host) bool {
	olds := App.Hosts
	old, ok := olds[name]
	if !ok {
		return false
	}
	if new.status != old.status {
		return false
	}

	if old.tunnelip != new.tunnelip && new.tunnelip != "" {
		return false
	}

	if old.mac != new.mac && new.mac != "" {
		return false
	}
	return true
}
