package utils

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

func ParsePorts(portStr string) ([]int, error) {
	var ports []int
	parts := strings.SplitSeq(portStr, ",")

	for p := range parts {
		p = strings.TrimSpace(p)
		
		// Check for range format (e.g., "80-100")
		if strings.Contains(p, "-") {
			rangeParts := strings.Split(p, "-")
			if len(rangeParts) != 2 {
				return nil, fmt.Errorf("invalid port range format: %s", p)
			}

			start, err := strconv.Atoi(strings.TrimSpace(rangeParts[0]))
			if err != nil {
				return nil, fmt.Errorf("invalid start port number: %s", rangeParts[0])
			}

			end, err := strconv.Atoi(strings.TrimSpace(rangeParts[1]))
			if err != nil {
				return nil, fmt.Errorf("invalid end port number: %s", rangeParts[1])
			}

			if start > end {
				return nil, fmt.Errorf("invalid port range: start port greater than end port")
			}

			if start <= 0 || end > 65535 {
				return nil, fmt.Errorf("port numbers must be between 1 and 65535")
			}

			for port := start; port <= end; port++ {
				ports = append(ports, port)
			}
		} else {
			// Single port
			port, err := strconv.Atoi(p)
			if err != nil {
				return nil, fmt.Errorf("invalid port number: %s", p)
			}
			if port <= 0 || port > 65535 {
				return nil, fmt.Errorf("port number out of range: %d", port)
			}
			ports = append(ports, port)
		}
	}

	if len(ports) == 0 {
		return nil, fmt.Errorf("no valid ports specified")
	}

	return ports, nil
}

func GetLocalNetwork() *net.IPNet {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil
	}

	for _, iface := range interfaces {
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			ipnet, ok := addr.(*net.IPNet)
			if !ok {
				continue
			}

			ip := ipnet.IP.To4()
			if ip == nil {
				continue
			}

			return ipnet
		}
	}
	return nil
}

func IPToInt(ip net.IP) uint32 {
	ip = ip.To4()
	return uint32(ip[0])<<24 | uint32(ip[1])<<16 | uint32(ip[2])<<8 | uint32(ip[3])
}

func IntToIP(n uint32) string {
	return fmt.Sprintf("%d.%d.%d.%d",
		byte(n>>24), byte(n>>16), byte(n>>8), byte(n))
}
