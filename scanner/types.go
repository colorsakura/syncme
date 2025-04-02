package scanner

import "net"

type ScanConfig struct {
	Ports       []int
	Timeout     int
	Concurrency int
}

type ScanResult struct {
	Target string
	Open   bool
	Error  error
}

type Scanner struct {
	config  ScanConfig
	results chan ScanResult
	network *net.IPNet
}
