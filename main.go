package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/fatih/color"
	"gopkg.in/yaml.v2"
)

var data = `
services:
 - name: firstService
   sites:
    - gmail.com
    - http://google.com
    - stackoverflow.com
    - http://github.com/reSoley/asdf
    - http://youtu.be
    - https://w3-03.sso.ibm.com/
    - http://facebook.com
 - name: secondService
   sites:
    - http://google.com
    - https://w3-03.sso.ibm.com/
    - gmail.com
 - name: thirdService
   sites:
    - http://psu.edu
    - http://facebook.com
`

var white (func(string, ...interface{}) string) = color.New(color.FgHiWhite, color.Bold).SprintfFunc()
var green (func(string, ...interface{}) string) = color.New(color.FgGreen, color.Bold).SprintfFunc()
var red (func(string, ...interface{}) string) = color.New(color.FgRed, color.Bold).SprintfFunc()
var cyan (func(string, ...interface{}) string) = color.New(color.FgCyan, color.Bold).SprintfFunc()

type service struct {
	Name  string
	Sites []string
}

type serviceList struct {
	Services    []service
	isAvailable map[string]bool
	statuses    map[string]string
	output      chan string
	sync.RWMutex
}

func ServiceList(source []byte) (*serviceList, error) {
	var services serviceList

	services.isAvailable = make(map[string]bool)
	services.statuses = make(map[string]string)
	services.output = make(chan string)

	err := yaml.Unmarshal(source, &services)
	if err != nil {
		return nil, err
	}

	return &services, nil
}

func formatStatus(status string) string {
	if strings.HasPrefix(status, "2") {
		return fmt.Sprintf("%s", green(status))
	} else if strings.HasPrefix(status, "3") {
		return fmt.Sprintf("%s", cyan(status))
	} else if strings.HasPrefix(status, "4") || strings.HasPrefix(status, "5") {
		return fmt.Sprintf("%s", red(status))
	} else {
		return fmt.Sprintf("%s %s", red("FAILED:"), status)
	}
}

func checkProtocol(url string) string {
	if !strings.HasPrefix(url, "http") {
		return "http://" + url
	}

	return url
}

func (s *serviceList) write() {
	for x := range s.output {
		fmt.Print(x)
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
		response, err := http.Head(checkProtocol(url))
		if err != nil {
			s.safeWrite(url, err.Error())
		} else {
			s.safeWrite(url, response.Status)
		}
	}

	status, _ := s.safeLookup(url)

	s.output <- fmt.Sprintf("\t%s %s %s\n", white("URL:"), url, formatStatus(status))

	if strings.HasPrefix(status, "2") {
		available <- true
	} else {
		available <- false
	}
}

func (s *serviceList) testService(service *service) {
	available := make(chan bool)

	s.output <- fmt.Sprintf("%s %s\n", white("Service:"), service.Name)

	for _, url := range service.Sites {
		go s.testUrl(url, available)
	}

	isAvailable := true
	for i := 0; i < len(service.Sites); i++ {
		isAvailable = <-available && isAvailable
	}

	if isAvailable {
		s.output <- fmt.Sprintf("\t%s\n", green("Available"))
	} else {
		s.output <- fmt.Sprintf("\t%s\n", red("Unavailable"))
	}

	s.isAvailable[service.Name] = isAvailable
}

func (s *serviceList) displayResults() {
	fmt.Printf("\n%s\n", green("Available Services"))
	for service, isAvailable := range s.isAvailable {
		if isAvailable {
			fmt.Printf("%s\n", service)
		} else {
			defer fmt.Printf("%s\n", service)
		}
	}
	fmt.Printf("\n%s\n", red("Unavailable Services"))
}

func main() {
	source := []byte(data)
	services, err := ServiceList(source)
	if err != nil {
		os.Exit(1)
	}

	go services.write()
	defer close(services.output)

	for _, cur := range services.Services {
		services.testService(&cur)
	}

	services.displayResults()
}
