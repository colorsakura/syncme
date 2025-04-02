package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/colorsakura/syncme/scanner"
	"github.com/colorsakura/syncme/utils"
)

func main() {
	// Parse command line flags
	portsStr := flag.String("ports", "", "Ports to scan (comma-separated)")
	timeout := flag.Int("timeout", 1, "Timeout in seconds")
	concurrency := flag.Int("concurrency", 5120, "Number of concurrent scanners")
	flag.Parse()

	if *portsStr == "" {
		flag.Usage()
		os.Exit(1)
	}

	// Parse ports
	ports, err := utils.ParsePorts(*portsStr)
	if err != nil {
		log.Fatalf("Error parsing ports: %v", err)
	}

	// Create scanner
	s, err := scanner.NewScanner(scanner.ScanConfig{
		Ports:       ports,
		Timeout:     *timeout,
		Concurrency: *concurrency,
	})
	if err != nil {
		log.Fatalf("Error creating scanner: %v", err)
	}

	// Start scanning
	startTime := time.Now()
	results := s.Start()

	// Process results
	openPorts := 0
	for result := range results {
		if result.Error != nil {
			continue
		}
		if result.Open {
			openPorts++
			fmt.Printf("[+] %s is open\n", result.Target)
		}
	}

	// Print summary
	duration := time.Since(startTime)
	fmt.Printf("\nScan completed in %v\n", duration)
	fmt.Printf("Found %d open ports\n", openPorts)
}
