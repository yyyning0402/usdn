package utils

import (
	"encoding/json"
	"sync"

	log "github.com/sirupsen/logrus"
	"github.com/toolkits/file"
)

type GlobalConfig struct {
	Debug      bool      `json:"debug"`
	Tunnelip   string    `json:"tunnel_ip"`
	Wanip      string    `json:"wanip"`
	Interval   int       `json:"interval"`
	SerfC      *SerfC    `json:"serf_client"`
	SerfA      *SerfA    `json:"serf_agent"`
	MemberList *[]string `json:"member_list"`
	Ping       *Ping     `json:"ping"`
}

type Ping struct {
	Interval int `json:"interval"`
	TimeOut  int `json:"timeout"`
	Lostper  int `json:"lostpercent"`
}

type SerfC struct {
	Name       string `json:"node_name"`
	RpcAddr    string `json:"rpc_addr"`
	RpcPort    int    `json:"rpc_port"`
	ELen       int    `json:"event_len"`
	CloseAgent bool   `json:"close_agent"`
}

type SerfA struct {
	Name          string `json:"name"`
	Enable        bool   `json:"enable"`
	ELen          int    `json:"event_len"`
	RpcAddr       string `json:"rpc_addr"`
	RpcPort       int    `json:"rpc_port"`
	BindAddr      string `json:"bind_addr"`
	BindPort      int    `json:"bind_port"`
	AdvertiseAddr string `json:"advertise_addr"`
	AdvertisePort int    `json:"advertise_port"`
}

var (
	ConfigFile string
	config     *GlobalConfig
	lock       = new(sync.RWMutex)
)

func ParseConfig(cfg string) {
	if cfg == "" {
		log.Fatalln("use -c to specify configuration file")
	}

	if !file.IsExist(cfg) {
		log.Fatalln("config file:", cfg, "is not existent. maybe you need `mv cfg.example.json cfg.json`")
	}

	ConfigFile = cfg

	configContent, err := file.ToTrimString(cfg)
	if err != nil {
		log.Fatalln("read config file:", cfg, "fail:", err)
	}

	var c GlobalConfig
	err = json.Unmarshal([]byte(configContent), &c)
	if err != nil {
		log.Fatalln("parse config file:", cfg, "fail:", err)
	}

	lock.Lock()
	defer lock.Unlock()

	config = &c

	log.Infoln("read config file:", cfg, "successfully")
}

func Config() *GlobalConfig {
	lock.RLock()
	defer lock.RUnlock()
	return config
}
