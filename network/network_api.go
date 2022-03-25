package network

import (
	"fmt"
	"mydocker/container"
	"net"
	"os"
	"text/tabwriter"

	"github.com/sirupsen/logrus"
)
/*mydocker network create --driver bridge --subnet 192.168.10.1/24 testbridge*/
/*mydocker run -ti -net testbridge busybox sh*/
/*mydocker run -ti -p 80:80 -net testbridge busybox sh*/
func Connect(networkName string, cinfo *container.ContainerInfo) error {
	network, ok := networks[networkName]
	if !ok {
		return fmt.Errorf("no such network: %s", networkName)
	}
	ip, err := ipAllocator.Allocate(network.IpRange)
	if err != nil {
		return err
	}
	ep := &Endpoint{
		ID:          fmt.Sprintf("%s-%s", cinfo.Id, networkName),
		IPAddress:   ip,
		Network:     network,
		PortMapping: cinfo.PortMapping,
	}
	if err = drivers[network.Driver].Connect(network, ep); err != nil {
		return err
	}
	if err = configEndpointIpAddressAndRoute(ep, cinfo); err != nil {
		return err
	}
	return configPortMapping(ep, cinfo)
}
func ListNetwork() {
	w := tabwriter.NewWriter(os.Stdout, 12, 1, 3, ' ', 0)
	fmt.Fprint(w, "NAME\tIPRange\tDriver\t")
	for _, nw := range networks {
		fmt.Fprintf(w, "%s\t%s\t%s\n",
			nw.Name,
			nw.IpRange.String(),
			nw.Driver,
		)
	}
	if err := w.Flush(); err != nil {
		logrus.Errorf("Flush error %v", err)
		return
	}
}
func CreateNetwork(driver, subnet, name string) error {
	_, cidr, _ := net.ParseCIDR(subnet)
	ip, err := ipAllocator.Allocate(cidr)
	if err != nil {
		return err
	}
	cidr.IP = ip
	nw, err := drivers[driver].Create(cidr.String(), name)
	if err != nil {
		return err
	}
	return nw.dump(defaultNetworkPath)
}
func DeleteNetwork(networkName string) error {
	nw, ok := networks[networkName]
	if !ok {
		return fmt.Errorf("No such Network: %s", networkName)
	}
	if err := ipAllocator.Release(nw.IpRange, &nw.IpRange.IP); err != nil {
		return fmt.Errorf("Error remove network gateway ip :%s", err)
	}
	if err := drivers[nw.Driver].Delete(*nw); err != nil {
		return fmt.Errorf("error network drivererror:%s", err)
	}
	return nw.remove(defaultNetworkPath)
}
