package funcs

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"usdn/djstl"
	"usdn/openflow"
	"usdn/utils"

	"github.com/digitalocean/go-openvswitch/ovs"

	//"reflect"
	"github.com/albertorestifo/dijkstra"
)

var ovsflows OvsFlow

type OvsFlow struct {
	CurTable int
	Client   *ovs.Client
	Status   bool
	NewPath  map[string]string
}

func OvsInit() error {
	utils.SqlCreate()

	veth0, veth1, err := openflow.Veth("veth0", "veth1")
	if err != nil {
		log.Println(err)
	}

	utils.Logger.Infoln(veth0.Attrs().Name, ":", veth0.Attrs().HardwareAddr)
	utils.Logger.Infoln(veth1.Attrs().Name, ":", veth1.Attrs().HardwareAddr)
	//set ip for veth0
	//local_ip from api
	local_ip := utils.Config().Tunnelip
	tunnelmac := veth0.Attrs().HardwareAddr.String()
	UpdateTags(map[string]string{
		"mac":      tunnelmac,
		"tunnelip": local_ip,
	})
	if err := openflow.Setip(local_ip, "veth0"); err != nil {
		return errors.New(fmt.Sprintln("set veth0 tunnel ip error", err))
	}
	utils.Logger.Infoln(veth0.Attrs().Name, ":", local_ip)

	ovsflows.Client = ovs.New(
		// Prepend "sudo" to all commands.
		ovs.Sudo(),
	)

	ovsflows.CurTable = 0
	//create bridge br0
	openflow.CreateBridge(ovsflows.Client, "br0")

	//add port veth1 to br0
	openflow.AddPort(ovsflows.Client, "br0", "veth1")

	out, err := openflow.GetPorts(ovsflows.Client, "veth1")
	if err != nil {
		return errors.New(fmt.Sprintln("get veth1 ports error", err))
	}

	//表
	utils.Logger.Infoln("veth1 out to :", out)

	err = openflow.Addflow(ovsflows.Client, "br0", tunnelmac, out, 0)
	if err != nil {
		return errors.New(fmt.Sprintln("add default flow error", err))
	}
	return nil
}

func Calculate(data map[string]map[string]int) error {

	ip2mac := make(map[string]string)
	var g dijkstra.Graph
	g = data
	re := djstl.Compute(g, App.Wanip)

	k, err := utils.SqlSelect()
	if err != nil {
		return err
	}
	mjson, err := json.Marshal(re)
	if err != nil {
		return err
	}
	mString := string(mjson)
	if k != mString {
		err := utils.SqlUpdate(mString)
		if err != nil {
			return err
		}
		ovsflows.Status = false
	}
	var (
		ips []string
	)
	tunnels := make(map[string]string)
	for k, v := range App.Hosts {
		if k != App.Wanip {
			if v.tunnelip != "" && v.mac != "" {
				ips = append(ips, k)
				tunnels[v.tunnelip] = v.mac
			}
		}
	}
	//add vxlan 接口
	openflow.AddVxlan(ovsflows.Client, ips, "br0")
	//生成arp

	fmt.Println(tunnels)
	openflow.Neigh_arp(tunnels)

	for k, v := range re {
		h := App.Hosts[k]
		key := h.mac
		if key != "" && v != "" {
			ip2mac[key] = v
		}
	}
	ovsflows.NewPath = ip2mac
	return nil
}

func Commit() {

	if !ovsflows.Status {
		//下发流表
		//ovs-ofctl add-flow br0  table=0,dl_dst=46:d0:28:82:9b:49,actions=output:3 -O OpenFlow13
		utils.Logger.Debugf("flow changed")

		if ovsflows.CurTable == 0 {
			ovsflows.CurTable = 1
		} else if ovsflows.CurTable == 1 {
			ovsflows.CurTable = 2
		} else if ovsflows.CurTable == 2 {
			ovsflows.CurTable = 1
		}

		for k, v := range ovsflows.NewPath {
			utils.Logger.Debugf("commit flow %s %s", k, v)
			in_pot := utils.InetAtoN64(v)
			out, _ := openflow.GetPorts(ovsflows.Client, strconv.Itoa(int(in_pot)))
			//openflow.Addflow(c,"br0",v,out,cur.Current_table)

			t0 := openflow.NewTable(ovsflows.CurTable, k)

			t0.Add(ovsflows.Client, out)
			// //table 0 指向 cur.table
			// t0 = openflow.NewTable(0, v)

			// t0.Change(ovsflows.Client, ovsflows.CurTable)
			//清空旧的流表
			openflow.Delflow(ovsflows.CurTable)
		}
		t0 := openflow.NewTable0(0)
		fmt.Printf("1111111111111 tid %v", ovsflows.CurTable)
		t0.Change(ovsflows.Client, ovsflows.CurTable)
		ovsflows.Status = true

	}
}
