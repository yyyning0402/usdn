package funcs

import (
	"errors"
	"fmt"
	"sync"
	"usdn/utils"
)

var (
	storage Queue
)

type Bucket struct {
	Wanip      string
	SerfClient Client
	Ping       Monitor
	Hosts      map[string]host
	merges     map[string]merged
	mlock      sync.Mutex
}

type merged struct {
	srcip string
	dstip string
	delay int
	lost  int
}

type host struct {
	srcip    string
	tunnelip string
	mac      string
	status   string
	update   int
}

func (b *Bucket) setMerges(data map[string]merged) {
	b.mlock.Lock()
	defer b.mlock.Unlock()
	b.merges = data
	return
}

func (b *Bucket) updateMerges(data map[string]merged) {
	b.mlock.Lock()
	defer b.mlock.Unlock()
	for k, v := range data {
		b.merges[k] = v
	}
}

func (b *Bucket) ListMerges() []merged {
	var tmp []merged
	for _, i := range b.merges {
		tmp = append(tmp, i)
	}
	return tmp
}

func New() (*Bucket, error) {
	var err error
	var b Bucket

	// use local ip
	ipa := utils.Iplist()
	if len(ipa) > 0 {
		b.Wanip = ipa[0]
	}

	// use config wan ip
	if utils.Config().Wanip != "" {
		b.Wanip = utils.Config().Wanip
	}

	// create serf agent and client
	b.SerfClient, err = ActiveSerf()
	if err != nil {
		return &b, errors.New(fmt.Sprintf("active serf error %v", err))
	}

	// if not enough members, rejoin member
	b.SerfClient.Rejoin()

	// create ping server
	b.Ping, err = ActiveMonitor()
	if err != nil {
		return &b, errors.New(fmt.Sprintf("active monitor error %v", err))
	}

	b.Hosts = make(map[string]host)

	// update member status
	go UpdateHost()

	return &b, nil
}

func (b *Bucket) Shutdown() {
	b.SerfClient.Shutdown()

	b.Ping.Stop()
}

type Queue struct {
	Data        []map[string]int
	Lock        sync.Mutex
	LastRequest int
}

func (this *Queue) Put(data map[string]int) {
	this.Lock.Lock()
	defer this.Lock.Unlock()
	this.Data = append(this.Data, data)
	return
}

func (this *Queue) Lrange(start int) []map[string]int {
	var res []map[string]int
	if len(this.Data) == 0 {
		return res
	}
	for _, i := range this.Data[start:] {
		res = append(res, i)
	}
	this.LastRequest = utils.CurTimeInt()
	return res
}

func (this *Queue) Clear() {
	this.Lock.Lock()
	defer this.Lock.Unlock()
	this.Data = []map[string]int{}
}
