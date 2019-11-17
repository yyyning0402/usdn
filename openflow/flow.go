package openflow

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/digitalocean/go-openvswitch/ovs"
	//"reflect"
)

type Yancnum struct {
	Src_ip string
	Des_ip string
	Yy     int
}
type Yanc struct {
	src_ip int
	des_ip int
	yy     int
}

type Table struct {
	Flow          *ovs.Flow
	Current_table int
}

// func NewTable (tid int,hardware string)Table{
// 	var actions []ovs.Action
// 	if tid == 0 {
// 		actions = []ovs.Action{ovs.Resubmit(0,1)}
// 	} else {
// 		actions = []ovs.Action{ovs.Output(1) }
// 	}
// 	log.Println(actions)
// 	return Table{
// 		flow: &ovs.Flow{
// 			Table : 0,
// 			Matches:[]ovs.Match{
// 				ovs.DataLinkDestination(hardware),
// 			},
// 			Actions: actions,
// 		},
// 		current_table : 1,
// 	}
// }
func NewTable(table int, hardware string) Table {
	var actions []ovs.Action
	actions = []ovs.Action{ovs.Resubmit(0, table)}
	//actions = []ovs.Action{ovs.Output(out)}
	return Table{
		Flow: &ovs.Flow{
			Table: table,
			Matches: []ovs.Match{
				ovs.DataLinkDestination(hardware),
			},
			Actions: actions,
		},
		Current_table: table,
	}
}

func NewTable0(table int) Table {
	var actions []ovs.Action
	actions = []ovs.Action{ovs.Resubmit(0, table)}
	//actions = []ovs.Action{ovs.Output(out)}
	return Table{
		Flow: &ovs.Flow{
			Table:   table,
			Actions: actions,
		},
		Current_table: table,
	}
}

func (t Table) Change(c *ovs.Client, tid int) {
	actions := []ovs.Action{ovs.Resubmit(0, tid)}
	t.Flow.Actions = actions
	t.Current_table = tid
	c.OpenFlow.AddFlow("br0", t.Flow)
}

func (t Table) Add(c *ovs.Client, out int) {
	actions := []ovs.Action{ovs.Output(out)}
	t.Flow.Actions = actions
	c.OpenFlow.AddFlow("br0", t.Flow)
}

// len(a)
// for k,_ :=range(a){
// 	if _,ok := b[k]; !ok{

// }
// }
func Addflow(c *ovs.Client, bridge string, hardware string, out int, table int) error {
	//t0 := NewTable(0,hardware)
	err := c.OpenFlow.AddFlow(bridge, &ovs.Flow{
		//Priority: 100,
		//Protocol: ovs.ProtocolIPv4,
		Table: table,
		Matches: []ovs.Match{
			ovs.DataLinkDestination(hardware),

			// ovs.NetworkSource("169.254.169.254"),
			// ovs.NetworkDestination("169.254.0.0/16"),
		},
		Actions: []ovs.Action{ovs.Output(out)},
		//Actions:  []ovs.Action{ovs.Resubmit(0,2)},
	})
	if err != nil {
		log.Println("failed to add flow: ", err)
	}
	return err
}

func Delflow(table int) {
	var cmdline string
	if table == 1 {
		cmdline = "ovs-ofctl del-flows br0 table=2 -O OpenFlow13"
	} else {
		cmdline = "ovs-ofctl del-flows br0 table=1 -O OpenFlow13"
	}
	Cmd(cmdline)
}
func GetPorts(c *ovs.Client, port string) (int, error) {
	// ww, _ := c.VSwitch.ListPorts("br0")
	// for index, value := range ww {
	// 	if value == port {
	// 		return index + 1, nil
	// 	}
	// }
	// var err error = errors.New("not exist")
	cmd := fmt.Sprintf("ovs-ofctl show br0 | grep %s | awk -F '(' '{print $1}'", port)
	restr, err := Cmd(cmd)
	if err != nil {
		return -1, err
	}
	strip := strings.Replace(restr, "\n", "", 1)
	strip = strings.Replace(strip, " ", "", 1)

	res, err := strconv.Atoi(strip)
	if err != nil {
		return -1, err
	}
	return res, nil
}

func Transfer(c *ovs.Client, st []Yancnum) []Yanc {
	var f []Yanc
	log.Println("in Transfer ", st)
	for _, i := range st {
		log.Println(i)
		src, _ := GetPorts(c, i.Src_ip)
		dst, _ := GetPorts(c, i.Des_ip)
		des := i.Yy
		log.Println(src, dst, des)
		f = append(f, Yanc{src_ip: src, des_ip: dst, yy: des})
	}
	return f
}
