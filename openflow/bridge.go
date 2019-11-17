package openflow

import (
	"log"
	"strconv"
	"usdn/utils"

	"github.com/digitalocean/go-openvswitch/ovs"
	//"reflect"
)

func CreateBridge(c *ovs.Client, bridge string) {
	if err := c.VSwitch.AddBridge(bridge); err != nil {
		log.Println("failed to add bridge: ", err)
	}
}

func AddVxlan(c *ovs.Client, a []string, bridge string) {
	//var addr string
	for _, i := range a {
		addr := strconv.FormatInt(utils.InetAtoN64(i), 10)
		if err := c.VSwitch.AddPort(bridge, addr); err != nil {
			log.Fatalf("failed to add bridge: %v", err)
		}
		var op ovs.InterfaceOptions
		op.Type = ovs.InterfaceTypeVXLAN
		op.RemoteIP = i
		if err := c.VSwitch.Set.Interface(addr, op); err != nil {
			log.Fatalf("failed to add bridge: %v", err)
		}
	}
}
func AddPort(c *ovs.Client, bridge, port string) {
	if err := c.VSwitch.AddPort(bridge, port); err != nil {
		log.Fatalf("failed to add bridge: %v", err)
	}
}
