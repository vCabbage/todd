/*
    Host Resources - Network

	Copyright 2016 Matt Oswalt. Use or modification of this
	source code is governed by the license provided here:
	https://github.com/toddproject/todd/blob/master/LICENSE
*/

package hostresources

import (
	"net"

	"github.com/pkg/errors"
)

// GetIPOfInt will iterate over all addresses for the given network interface, but will return only
// the first one it finds. TODO(mierdin): This has obvious drawbacks, particularly with IPv6. Need to figure out a better way.
func GetIPOfInt(ifname string) (net.IP, error) {
	iface, err := net.InterfaceByName(ifname)
	if err != nil {
		return nil, errors.Wrap(err, "getting interface")
	}

	addrs, err := iface.Addrs()
	if err != nil {
		return nil, errors.Wrap(err, "listing IP addresses")
	}

	for _, addr := range addrs {
		ipnet, ok := addr.(*net.IPNet)
		if !ok || ipnet.IP.To4() == nil {
			continue
		}
		return ipnet.IP, nil
	}

	return nil, errors.New("no matching IP")
}
