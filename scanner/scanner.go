package scanner

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/colorsakura/synco/utils"
)

func NewScanner(config ScanConfig) (*Scanner, error) {
	network := utils.GetLocalNetwork()
	if network == nil {
		return nil, fmt.Errorf("failed to get local network information")
	}

	return &Scanner{
		config:  config,
		results: make(chan ScanResult),
		network: network,
	}, nil
}

func (s *Scanner) Start() chan ScanResult {
	tasks := make(chan string)

	// Generate tasks
	go s.generateTasks(tasks)

	// Start workers
	var wg sync.WaitGroup
	for range s.config.Concurrency {
		wg.Add(1)
		go s.worker(tasks, &wg)
	}

	// Close results channel when all workers are done
	go func() {
		wg.Wait()
		close(s.results)
	}()

	return s.results
}

func (s *Scanner) worker(tasks <-chan string, wg *sync.WaitGroup) {
	defer wg.Done()
	timeout := time.Duration(s.config.Timeout) * time.Second

	for target := range tasks {
		conn, err := net.DialTimeout("tcp", target, timeout)
		if err != nil {
			s.results <- ScanResult{Target: target, Open: false, Error: err}
			continue
		}
		conn.Close()
		s.results <- ScanResult{Target: target, Open: true}
	}
}

func (s *Scanner) generateTasks(tasks chan<- string) {
	defer close(tasks)

	ip := s.network.IP.To4()
	mask := s.network.Mask

	network := ip.Mask(mask)
	broadcast := make(net.IP, len(network))
	for i := range network {
		broadcast[i] = network[i] | ^mask[i]
	}

	for i := utils.IPToInt(network) + 1; i < utils.IPToInt(broadcast); i++ {
		target := utils.IntToIP(i)
		for _, port := range s.config.Ports {
			tasks <- fmt.Sprintf("%s:%d", target, port)
		}
	}
}
