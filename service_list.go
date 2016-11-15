package main

import (
	"fmt"
	"strings"
	"sync"

	"gopkg.in/yaml.v2"
)

type service struct {
	Name        string
	Sites       []string
	isAvailable bool
}

type serviceList struct {
	Services []service
	statuses map[string]string
	output   chan string
	verbose  bool
	sync.RWMutex
}

func ServiceList(source []byte, verbose bool) (*serviceList, error) {
	var services serviceList

	services.statuses = make(map[string]string)
	services.output = make(chan string)
	services.verbose = verbose

	err := yaml.Unmarshal(source, &services)
	if err != nil {
		return nil, err
	}

	return &services, nil
}

func (s *serviceList) write(done chan int) {
	for {
		x := <-s.output
		fmt.Print(x)

		if x == fmt.Sprintf("\t %s\n", green("Available")) || x == fmt.Sprintf("\t %s\n", red("Unavailable")) {
			done <- 0
		}
	}
}

func (s *serviceList) safeLookup(key string) (string, bool) {
	s.RLock()
	value, found := s.statuses[key]
	s.RUnlock()

	return value, found
}

func (s *serviceList) safeWrite(key, value string) {
	s.Lock()
	s.statuses[key] = value
	s.Unlock()
}

func (s *serviceList) testUrl(url string, available chan bool) {
	_, found := s.safeLookup(url)

	if !found {
		response, err := client.Head(formatUrl(url))
		if err != nil {
			if strings.Contains(err.Error(), "Timeout exceeded") {
				s.safeWrite(url, "Request timeout exceeded")
			} else if strings.Contains(err.Error(), "no such host") {
				s.safeWrite(url, "No such host")
			} else {
				s.safeWrite(url, err.Error())
			}
		} else {
			s.safeWrite(url, response.Status)
		}
	}

	status, _ := s.safeLookup(url)

	isAvailable, formattedStatus := formatStatus(status)
	if strings.Contains(url, "*.") {
		formattedStatus = fmt.Sprintf("%s - %s Wildcard subdomains unsupported. Status reported for domain.", formattedStatus, yellow("WARNING:"))
	}

	if s.verbose || !isAvailable {
		s.output <- fmt.Sprintf("\t %s %s %s\n", "URL:", cyan(url), formattedStatus)
	}
	available <- isAvailable
}

func (s *serviceList) testService(serviceIndex int) {
	available := make(chan bool)

	s.output <- fmt.Sprintf("%s %s\n", "Service:", s.Services[serviceIndex].Name)

	for _, url := range s.Services[serviceIndex].Sites {
		go s.testUrl(url, available)
	}

	isAvailable := true
	for i := 0; i < len(s.Services[serviceIndex].Sites); i++ {
		isAvailable = <-available && isAvailable
	}

	if isAvailable {
		s.output <- fmt.Sprintf("\t %s\n", green("Available"))
	} else {
		s.output <- fmt.Sprintf("\t %s\n", red("Unavailable"))
	}

	s.Services[serviceIndex].isAvailable = isAvailable
}

func (s *serviceList) displayResults() {
	fmt.Printf("\n%s\n", green("Available Services"))
	for _, service := range s.Services {
		if service.isAvailable {
			fmt.Printf("%s\n", service.Name)
		} else {
			defer fmt.Printf("%s\n", service.Name)
		}
	}
	fmt.Printf("\n%s\n", red("Unavailable Services"))
}
