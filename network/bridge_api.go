package network

import (
	"fmt"
	"net"

	"github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
)

func (d *BridgeNetworkDriver) Name() string {
	return "bridge"
}
func (d *BridgeNetworkDriver) Create(subnet string, name string) (*Network, error) {
	ip, ipRange, _ := net.ParseCIDR(subnet)
	ipRange.IP = ip
	n := &Network{
		Name:    name,
		IpRange: ipRange,
		Driver:  d.Name(),
	}
	fmt.Println(n)
	err := d.initBridge(n)
	if err != nil {
		logrus.Errorf("error init bridge: %v", err)
	}

	return n, err
}
func (d *BridgeNetworkDriver) Delete(network Network) error {
	bridgeName := network.Name
	br, err := netlink.LinkByName(bridgeName)
	if err != nil {
		return err
	}
	return netlink.LinkDel(br)
}
func (d *BridgeNetworkDriver) Connect(network *Network, endpoint *Endpoint) error {
	fmt.Println(network)
	bridgeName := network.Name
	br, err := netlink.LinkByName(bridgeName)
	if err != nil {
		return err
	}
	la := netlink.NewLinkAttrs()
	la.Name = endpoint.ID[:5]
	la.MasterIndex = br.Attrs().Index
	endpoint.Device = netlink.Veth{
		LinkAttrs: la,
		PeerName:  "cif-" + endpoint.ID[:5],
	}
	if err = netlink.LinkAdd(&endpoint.Device); err != nil {
		return fmt.Errorf("Error add endpoint device %v", err)
	}
	if err = netlink.LinkSetUp(&endpoint.Device); err != nil {
		return fmt.Errorf("err set endpoint device up:%v", err)
	}
	return nil
}
func (d *BridgeNetworkDriver) Disconnect(network Network, endpoint *Endpoint) error {
	return nil
}
