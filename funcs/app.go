package funcs

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
	"usdn/utils"
)

var (
	App *Bucket
)

func Run() {

	var err error
	App, err = New()
	if err != nil {
		utils.Logger.Errorln(err)
		return
	}

	// init ovs client
	if err := OvsInit(); err != nil {
		utils.Logger.Errorln(err)
		return
	}

	// wait for ping result
	time.Sleep(time.Duration(40 * time.Second))

	// wait
	Wait0()

	utils.Logger.Infoln("start cron")

	go Cron()

	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	fmt.Println("received", <-ch)
	App.Shutdown()
}

func SendEvent() {
	msg, err := App.Ping.ExportToSubject()
	if err != nil {
		utils.Logger.Errorln(err)
	}
	App.SerfClient.Event(msg)
}

func UpdateTags(tags map[string]string) {
	toUpdate := make(map[string]string)
	for k, v := range tags {
		toUpdate[k] = v
	}
	err := App.SerfClient.UpdateTags(toUpdate, []string{})
	if err != nil {
		utils.Logger.Errorln(err)
	}
}

func MergeEvent() (map[string]map[string]int, error) {

	merged := make(map[string]map[string]int)

	defer func() {
		err := recover()
		if err != nil {
			utils.Logger.Errorln(err)
		}
	}()

	oris := storage.Lrange(0)
	storage.Clear()

	for _, ori := range oris {
		srcip := utils.InetNtoA(ori["srcip"])
		delete(ori, "srcip")
		if _, ok := merged[srcip]; !ok {
			merged[srcip] = ori
		} else {
			for k, v := range ori {
				merged[srcip][k] = v
			}
		}
	}

	for k, v := range merged {
		for dst, out := range v {
			in := 3000
			if _, ok := merged[dst]; ok {
				if _, ok := merged[dst][k]; ok {
					in = merged[dst][k]
				}
			} else {
				merged[dst] = map[string]int{}
			}
			avg := (out + in) / 2
			merged[k][dst] = avg
			merged[dst][k] = avg
		}
	}

	return merged, nil
}

func MergeAndCalc() {

	merged, err := MergeEvent()
	if err != nil {
		utils.Logger.Errorf("merge event error %v", err)
		return
	}
	utils.Logger.Debugln("merge event %v", merged)

	if err := Calculate(merged); err != nil {
		utils.Logger.Errorf("calculate error %v", err)
		return
	}

}
