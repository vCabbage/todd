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

// GetDefaultInterfaceIP determines the appropriate IP address to use for either the server or agent
// based on configuration parameters passed in as arguments
//
// The server uses this address to inform the agents of the URL they should use to download assets
//
// The agents use this address so that the server knows how to orchestrate tests.
// (i.e. This agent publishes it's default address, and the server instructs other agents to target it in tests)
func GetDefaultInterfaceIP(ifname, ipAddrOverride string) (string, error) {
	if ipAddrOverride != "" {
		return ipAddrOverride, nil
	}
	return getIPOfInt(ifname)
}

// getIPOfInt will get the first usable IP address on a network interface
//
// TODO(mierdin): Need to handle IPv6 here
func getIPOfInt(ifname string) (string, error) {

	iface, err := net.InterfaceByName(ifname)
	if err != nil {
		return "", errors.Wrap(err, "Specified network interface not found")
	}

	addrs, err := iface.Addrs()
	if err != nil {
		return "", errors.Wrap(err, "Failed to retrieve addresses from network interface")
	}

	// Iterate over all the addresses and return the first one we find
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.To4().String(), nil
			}
		}
	}

	return "", errors.New("No DefaultInterface address found")
}

// getIPOfInt will iterate over all addresses for the given network interface, but will return only
// the first one it finds. TODO(mierdin): This has obvious drawbacks, particularly with IPv6. Need to figure out a better way.
func oldgetIPOfInt(ifname string) (string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	for _, iface := range interfaces {
		if iface.Name == ifname {

			addrs, err := iface.Addrs()
			if err != nil {
				return "", err
			}
			for _, addr := range addrs {
				if ipnet, ok := addr.(*net.IPNet); ok {
					if ipnet.IP.To4() != nil {
						return ipnet.IP.To4().String(), nil
					}

				}
			}
		}
	}
	return "", errors.New("No DefaultInterface address found")
}
