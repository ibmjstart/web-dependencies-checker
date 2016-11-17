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
	quiet    bool
	sync.RWMutex
}

func ServiceList(sources [][]byte, quiet bool) (*serviceList, error) {
	var services, temp serviceList

	services.statuses = make(map[string]string)
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

	isAvailable, formattedStatus := formatStatus(url, status)

	if !s.quiet || !isAvailable {
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
