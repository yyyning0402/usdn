package openflow

import (
	"log"
	"syscall"

	"github.com/vishvananda/netlink"

	//"reflect"
	"fmt"
	"net"
	"strings"
)

func Veth(port1 string, port2 string) (netlink.Link, netlink.Link, error) {
	//判断veth接口是否存在 不存则创建veth pair
	//"veth0""veth1"
	// l,err :=netlink.LinkList()
	// if err != nil {
	// 	panic(err)
	// }
	// for _, v := range l {
	// 	if v.Type() == "veth" {
	// 		a := v.Attrs()
	// 		log.Println(a.Name,"xxxxxxxxc")
	// 	}
	// }
	veth0, err := netlink.LinkByName(port1)
	veth1, err := netlink.LinkByName(port2)
	if veth0 == nil || veth1 == nil {
		log.Println("Create Veth pair")
		veth0, veth1, err := CreateVeth(port1, port2)
		return veth0, veth1, err
	}
	return veth0, veth1, err
	// return "success",nil
	// result,err := Cmd("ip link add veth0 type veth peer name veth1")
	// if err != nil {
	// 	return result,err
	// }
	// result,err =  Cmd("ifconfig veth0 up")
	// if err != nil {
	// 	return result,err
	// }
	// result,err = Cmd("ifconfig veth1 up")
	// if err != nil {
	// 	return result,err
	// }
	// return "Create veth success ",nil
}

func CreateVeth(port1 string, port2 string) (netlink.Link, netlink.Link, error) {
	vethLink := &netlink.Veth{LinkAttrs: netlink.LinkAttrs{Name: port1}, PeerName: port2}
	if err := netlink.LinkAdd(vethLink); err != nil {
		return nil, nil, err
	}
	veth0, _ := netlink.LinkByName(port1)
	veth1, _ := netlink.LinkByName(port2)

	for _, i := range []string{"veth0", "veth1"} {
		cmd := fmt.Sprintf("ifconfig %s up", i)
		Cmd(cmd)
	}

	return veth0, veth1, nil
}

func Setip(ipaddr string, port string) error {
	inface, _ := netlink.LinkByName(port)
	add, _ := netlink.AddrList(inface, syscall.AF_INET)
	for _, v := range add {
		//log.Println(v.IP,reflect.TypeOf(v.IP),"xxxxx",[]byte(v.IP),reflect.TypeOf([]byte(v.IP)),string([]byte(v.IP)),"kkkk")
		ip := fmt.Sprintf("%v", v.IP)
		ipaddr = strings.Split(ipaddr, "/")[0]
		if ipaddr == ip {
			log.Println(v.IP)
			return nil
		}
	}
	addr, _ := netlink.ParseAddr(ipaddr)
	fmt.Println("xxxxxxxxxxxxxxxxxxxx", inface, addr)
	if err := netlink.AddrAdd(inface, addr); err != nil {
		return err
	}
	return nil
}

func Neigh_arp(addr map[string]string) {
	//veth0
	netlink.NeighList(327, syscall.AF_INET)
	inface, _ := netlink.LinkByName("veth0")
	index := inface.Attrs().Index
	for k, v := range addr {
		cidr := strings.Split(k, "/")
		var neigh netlink.Neigh
		neigh.LinkIndex = index
		neigh.State = netlink.NUD_PERMANENT
		neigh.IP = net.ParseIP(cidr[0])
		neigh.HardwareAddr, _ = net.ParseMAC(v)
		if err := netlink.NeighAdd(&neigh); err != nil {
			log.Println(err, neigh.IP)
		}
	}
}
