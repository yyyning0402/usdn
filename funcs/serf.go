package funcs

import (
	"fmt"
	"net"
	"time"

	"encoding/json"
	"usdn/utils"

	"github.com/hashicorp/serf/client"
	"github.com/hashicorp/serf/cmd/serf/command/agent"
	"github.com/hashicorp/serf/serf"
)

type Client struct {
	*client.RPCClient
	Quit chan int
	Name string
}

func ActiveSerf() (Client, error) {

	var (
		c   Client
		err error
	)
	// active serf agent
	if utils.Config().SerfA.Enable {
		s, err := NewSerfAgent()
		if err != nil {
			return c, err
		}
		err = s.Start()
		if err != nil {
			return c, err
		}
	}

	time.Sleep(time.Duration(3 * time.Second))

	// create serf client
	c, err = newSerfClient()
	if err != nil {
		return c, err
	}

	// join members
	_, err = c.Join(*utils.Config().MemberList, true)
	if err != nil {
		utils.Logger.Warningln(err)
	}

	members, err := c.Members()
	if err != nil {
		return c, err
	}
	for _, m := range members {
		utils.Logger.Debugf("node %s ip %s status %s", m.Name, string(m.Addr), m.Status)
	}

	// start listen event
	go c.Listen()

	return c, nil
}

func newSerfClient() (Client, error) {
	var (
		c    Client
		name string
	)

	ips := utils.Iplist()
	if len(ips) > 0 {
		name = ips[0]
	}

	if utils.Config().SerfC.Name != "" {
		name = utils.Config().SerfC.Name
	}

	if utils.Config().Wanip != "" {
		name = utils.Config().Wanip
	}

	rpcaddr := fmt.Sprintf("%s:%v", utils.Config().SerfC.RpcAddr, utils.Config().SerfC.RpcPort)
	rpc, err := client.NewRPCClient(rpcaddr)
	if err != nil {
		return c, err
	}
	c = Client{rpc, make(chan int), name}

	return c, nil
}

// send event message
func (c *Client) Event(msg map[string]int) {

	events := split_event(msg, utils.Config().SerfC.ELen)

	for _, e := range events {
		b, err := json.Marshal(e)
		if err != nil {
			utils.Logger.Errorf("marshal event error %v", err)
			continue
		}

		utils.Logger.Debugln("send event %v", e)
		err = c.UserEvent("usdn", b, false)
		if err != nil {
			utils.Logger.Errorf("tigger event error %v", err)
		}
		time.Sleep(time.Duration(1 * time.Second))
	}
}

// number of members available is greater than 1 break loop
func (c *Client) Rejoin() {

	for {
		members, err := c.Members()
		var cnt int
		for _, m := range members {
			if m.Status == "alive" {
				cnt += 1
			}
		}
		if cnt > 1 {
			return
		}
		utils.Logger.Warningln("not enough members, rejoin member")
		time.Sleep(time.Duration(3 * time.Second))
		if len(*utils.Config().MemberList) == 0 {
			continue
		}
		_, err = c.Join(*utils.Config().MemberList, true)
		if err != nil {
			utils.Logger.Warningln(err)
		}
	}
}

// receive the event msg
func (c *Client) Listen() {

	// defer func() {
	// 	err := recover()
	// 	if err != nil {
	// 		utils.Logger.Warningln(err)
	// 	}
	// }()

	for {
		ch := make(chan map[string]interface{})
		_, err := c.Stream("user:usdn", ch)
		if err != nil {
			utils.Logger.Errorln("listen stream", err)
		}
		select {
		case <-c.Quit:
			utils.Logger.Debugln("close channel")
			return
		case resp := <-ch:
			data := make(map[string]int)
			err := json.Unmarshal(resp["Payload"].([]byte), &data)
			if err != nil {
				utils.Logger.Errorf("unmarshal payload error %v", err)
				continue
			}
			storage.Put(data)
		}
	}
}

// close goroutine and client
func (c *Client) Shutdown() {
	c.Quit <- 1
	if utils.Config().SerfC.CloseAgent || utils.Config().SerfA.Enable {
		err := c.Leave()
		if err != nil {
			utils.Logger.Errorln("leave", err)
		}
	}
	err := c.Close()
	if err != nil {
		utils.Logger.Errorln("close", err)
	}
}

// Broadcast is delivered via udp;
// Due to packet size limitations, data needs to be split
func split_event(data map[string]int, max int) []map[string]int {

	var (
		data_sli []map[string]int
	)
	tmp_map := make(map[string]int)
	srcip := utils.InetAtoN(App.Wanip)
	length_srcip := len("srcip") + len(fmt.Sprintf("%v", srcip))
	length := 1 + length_srcip + 6

	for k, v := range data {
		tmpl := length + len(k) + len(fmt.Sprintf("%v", v)) + 6
		if tmpl > max {
			tmp_map["srcip"] = srcip
			data_sli = append(data_sli, tmp_map)
			tmp_map = map[string]int{k: v}
			length = 1 + len(k) + len(fmt.Sprintf("%v", v)) + length_srcip + 6
		} else {
			length = tmpl
			tmp_map[k] = v
		}
	}
	tmp_map["srcip"] = srcip
	data_sli = append(data_sli, tmp_map)
	return data_sli
}

type Serf struct {
	Name  string
	Agent *agent.Agent
	IPC   *agent.AgentIPC
}

var (
	bindaddr      string = "0.0.0.0"
	bindport      int    = 7946
	rpcaddr       string = "127.0.0.1"
	rpcport       int    = 7373
	advertiseaddr string
	advertiseport int = 7946
)

// serf agent requires two types of configuration files
func GenerateConfig() (*agent.Config, *serf.Config) {
	config := agent.DefaultConfig()
	config.UserEventSizeLimit = utils.Config().SerfA.ELen
	serfConfig := serf.DefaultConfig()
	serfConfig.UserEventSizeLimit = utils.Config().SerfA.ELen

	ips := utils.Iplist()
	if len(ips) > 0 {
		config.NodeName = ips[0]
	}

	if utils.Config().SerfA.Name != "" {
		config.NodeName = utils.Config().SerfA.Name
	}

	if utils.Config().Wanip != "" {
		config.NodeName = utils.Config().Wanip
	}

	if utils.Config().SerfA.BindAddr != "" {
		bindaddr = utils.Config().SerfA.BindAddr
		bindport = utils.Config().SerfA.BindPort
	}
	config.BindAddr = fmt.Sprintf("%s:%v", bindaddr, bindport)

	advertiseaddr = config.NodeName
	advertiseport = bindport
	if utils.Config().SerfA.AdvertiseAddr != "" {
		advertiseaddr = utils.Config().SerfA.AdvertiseAddr
		advertiseport = utils.Config().SerfA.AdvertisePort
	}
	config.AdvertiseAddr = fmt.Sprintf("%s:%v", advertiseaddr, advertiseport)

	if utils.Config().SerfA.RpcAddr != "" {
		rpcaddr = utils.Config().SerfA.RpcAddr
		rpcport = utils.Config().SerfA.RpcPort
	}
	config.RPCAddr = fmt.Sprintf("%s:%v", rpcaddr, rpcport)

	serfConfig.MemberlistConfig.BindAddr = bindaddr
	serfConfig.MemberlistConfig.BindPort = bindport
	serfConfig.MemberlistConfig.AdvertiseAddr = advertiseaddr
	serfConfig.MemberlistConfig.AdvertisePort = advertiseport
	serfConfig.NodeName = config.NodeName
	return config, serfConfig
}

func NewSerfAgent() (Serf, error) {
	var serf Serf

	config, serfConfig := GenerateConfig()

	// create serf agent
	client, err := agent.Create(config, serfConfig, utils.SerfLog)
	if err != nil {
		return serf, err
	}

	// Initialize rpc listener
	rpcListener, err := net.Listen("tcp", "127.0.0.1:7373")
	if err != nil {
		return serf, err
	}
	logWriter := agent.NewLogWriter(512)
	ipc := agent.NewAgentIPC(client, "", rpcListener, utils.SerfLog, logWriter)

	return Serf{
		Agent: client,
		IPC:   ipc,
	}, nil
}

func (s *Serf) Start() error {

	err := s.Agent.Start()
	if err != nil {
		return err
	}
	return nil
}
