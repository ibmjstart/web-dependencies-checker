package main

import (
	"fmt"
	"strings"
	"sync"

	"gopkg.in/yaml.v2"
)

type urlInfo struct {
	status      string
	isAvailable bool
	retries     int
}

type service struct {
	Name        string
	Sites       []string
	isAvailable bool
}

type serviceList struct {
	Services []service
	statuses map[string]*urlInfo
	output   chan string
	quiet    bool
	sync.RWMutex
}

func ServiceList(sources [][]byte, quiet bool) (*serviceList, error) {
	var services, temp serviceList

	services.statuses = make(map[string]*urlInfo)
	services.output = make(chan string)
	services.quiet = quiet

	for _, source := range sources {
		err := yaml.Unmarshal(source, &temp)
		if err != nil {
			return nil, err
		}

		services.Services = append(services.Services, temp.Services...)
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

func (s *serviceList) safeLookup(key string) (*urlInfo, bool) {
	s.RLock()
	value, found := s.statuses[key]
	s.RUnlock()

	return value, found
}

func (s *serviceList) safeWrite(key string, value *urlInfo) {
	s.Lock()
	s.statuses[key] = value
	s.Unlock()
}

func (s *serviceList) testUrl(url string, available chan bool) {
	info, found := s.safeLookup(url)
	if !found {
		info = &urlInfo{
			status:      "",
			isAvailable: false,
			retries:     0,
		}

		for proceed := true; proceed; proceed = (!info.isAvailable && info.retries < client.maxRetries) {
			response, err := client.Head(formatUrl(url))
			if err != nil {
				if strings.Contains(err.Error(), "Timeout exceeded") {
					info.status = "Request timeout exceeded"
				} else if strings.Contains(err.Error(), "no such host") {
					info.status = "No such host"
				} else {
					info.status = err.Error()
				}
			} else {
				info.status = response.Status
			}

			isAvailable, _ := formatStatus(url, info.status)
			info.isAvailable = isAvailable

			info.retries++
		}

		s.safeWrite(url, info)
	}

	_, formattedStatus := formatStatus(url, info.status)

	if !s.quiet || !info.isAvailable {
		s.output <- fmt.Sprintf("\t %s %s %s\n", "URL:", cyan(url), formattedStatus)
	}
	available <- info.isAvailable
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
