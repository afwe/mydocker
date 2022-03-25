package network

import (
	"encoding/json"
	"fmt"
	"mydocker/container"
	"net"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"

	"github.com/sirupsen/logrus"
)

var (
	defaultNetworkPath = "/var/run/mydocker/network/network/"
	drivers            = map[string]NetworkDriver{}
	networks           = map[string]*Network{}
)

type Network struct {
	Name    string
	IpRange *net.IPNet
	Driver  string
}
type Endpoint struct {
	ID          string           `json:"id"`
	Device      netlink.Veth     `json:"dev"`
	IPAddress   net.IP           `json:"ip"`
	MacAddress  net.HardwareAddr `json:"mac"`
	PortMapping []string         `json:"poetmapping"`
	Network     *Network
}
type NetworkDriver interface {
	Name() string
	Create(subnet, name string) (*Network, error)
	Delete(network Network) error
	Connect(network *Network, endpoint *Endpoint) error
	Disconnect(network Network, endpoint *Endpoint) error
}

func Init() error {
	var bridgeDriver = BridgeNetworkDriver{}
	drivers[bridgeDriver.Name()] = &bridgeDriver
	if _, err := os.Stat(defaultNetworkPath); err != nil {
		if os.IsNotExist(err) {
			os.MkdirAll(defaultNetworkPath, 0777)
		} else {
			return err
		}
	}
	filepath.Walk(defaultNetworkPath, func(nwPath string, info os.FileInfo, err error) error {
		if strings.HasSuffix(nwPath, "/") {
			return nil
		}
		_, nwName := path.Split(nwPath)
		nw := &Network{
			Name: nwName,
		}
		if err := nw.load(nwPath); err != nil {
			logrus.Errorf("error load network: %s", err)
		}
		networks[nwName] = nw
		return nil
	})
	return nil
}
func (nw *Network) dump(dumpPath string) error {
	if _, err := os.Stat(dumpPath); err != nil {
		if os.IsNotExist(err) {
			os.MkdirAll(dumpPath, 0777)
		} else {
			return err
		}
	}
	nwPath := path.Join(dumpPath, nw.Name)
	nwFile, err := os.OpenFile(nwPath, os.O_TRUNC|os.O_WRONLY|
		os.O_CREATE, 0777)
	if err != nil {
		logrus.Errorf("error: %v", err)
		return err
	}
	defer nwFile.Close()
	nwJson, err := json.Marshal(nw)
	if err != nil {
		logrus.Errorf("error : %v", err)
		return err
	}
	_, err = nwFile.Write(nwJson)
	if err != nil {
		logrus.Errorf("error: %v", err)
		return err
	}
	return nil
}
func (nw *Network) load(dumpPath string) error {
	nwConfigFile, err := os.Open(dumpPath)
	defer nwConfigFile.Close()
	if err != nil {
		return err
	}
	nwJson := make([]byte, 2000)
	n, err := nwConfigFile.Read(nwJson)
	if err != nil {
		return err
	}
	err = json.Unmarshal(nwJson[:n], nw)
	if err != nil {
		logrus.Errorf(" Error load nw info %v", err)
		return err
	}
	return nil
}
func (nw *Network) remove(dumpPath string) error {
	if _, err := os.Stat(path.Join(dumpPath, nw.Name)); err != nil {
		if os.IsNotExist(err) {
			return nil
		} else {
			return err
		}
	} else {
		return os.Remove(path.Join(dumpPath, nw.Name))
	}
}
func configEndpointIpAddressAndRoute(ep *Endpoint, cinfo *container.ContainerInfo) error {
	peerLink, err := netlink.LinkByName(ep.Device.PeerName)
	if err != nil {
		return fmt.Errorf("fail config endpoint: %v", err)
	}
	defer enterContainerNetns(&peerLink, cinfo)()
	interfaceIP := *ep.Network.IpRange
	interfaceIP.IP = ep.IPAddress
	if err = setInterfaceIP(ep.Device.PeerName, interfaceIP.String()); err != nil {
		return fmt.Errorf("%v,%s", ep.Network, err)
	}
	if err = setInterfaceUP(ep.Device.PeerName); err != nil {
		return err
	}
	if err = setInterfaceUP("lo"); err != nil {
		return err
	}
	_, cidr, _ := net.ParseCIDR("0.0.0.0/0")
	defaultRoute := &netlink.Route{
		LinkIndex: peerLink.Attrs().Index,
		Gw:        ep.Network.IpRange.IP,
		Dst:       cidr,
	}
	if err = netlink.RouteAdd(defaultRoute); err != nil {
		return err
	}
	return nil
}
func enterContainerNetns(enLink *netlink.Link,
	cinfo *container.ContainerInfo) func() {
	f, err := os.OpenFile(fmt.Sprintf("/proc/%s/ns/net", cinfo.Pid), os.O_RDONLY, 0)
	if err != nil {
		logrus.Errorf("error get container net namespace %v", err)
	}
	nsFD := f.Fd()
	runtime.LockOSThread()
	if err = netlink.LinkSetNsFd(*enLink, int(nsFD)); err != nil {
		logrus.Errorf("err set link netns,%v", err)
	}
	origins, err := netns.Get()
	if err != nil {
		logrus.Errorf("error get cur netns,%v", err)
	}
	if err = netns.Set(netns.NsHandle(nsFD)); err != nil {
		logrus.Errorf("err set netns, %v", err)
	}
	return func() {
		netns.Set(origins)
		origins.Close()
		runtime.UnlockOSThread()
		f.Close()
	}
}

func configPortMapping(ep *Endpoint, cinfo *container.ContainerInfo) error {
	for _, pm := range ep.PortMapping {
		portMapping := strings.Split(pm, ":")
		if len(portMapping) != 2 {
			logrus.Errorf("port mapping format err %v", pm)
			continue
		}
		iptablesCmd := fmt.Sprintf("-t nat -A PREROUTING -p tcp --dport %s -j DNAT --to-destination %s:%s",
			portMapping[0], ep.IPAddress.String(), portMapping[1])
		cmd := exec.Command("iptables", strings.Split(iptablesCmd, " ")...)
		output, err := cmd.Output()
		if err != nil {
			logrus.Errorf("iptables output %v", output)
			continue
		}
	}
	return nil
}
