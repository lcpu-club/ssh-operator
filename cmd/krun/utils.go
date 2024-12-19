package main

import (
	"errors"
	"net"
	"os"
	"strings"
)

func getFastestIP() (string, error) {
	// Get the fastest IP address
	netdev := os.Getenv("KRUN_NETDEV")
	if netdev == "" {
		// Get system netdevs
		ifs, err := net.Interfaces()
		if err != nil {
			return "", err
		}
		for _, i := range ifs {
			if strings.HasPrefix(i.Name, "net") {
				netdev = i.Name
			}
		}
		if netdev == "" {
			if len(ifs) < 2 {
				return "", errors.New("no network interface found")
			}
			// ifs[0] is lo
			netdev = ifs[1].Name
		}
	}
	i, err := net.InterfaceByName(netdev)
	if err != nil {
		return "", err
	}
	addrs, err := i.Addrs()
	if err != nil {
		return "", err
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			return ipnet.IP.String(), nil
		}
	}
	return "", errors.New("no IP address found on interface " + netdev)
}

func getShortHostname() (string, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return "", err
	}
	return strings.Split(hostname, ".")[0], nil
}
