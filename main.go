package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/fatih/color"
	"gopkg.in/yaml.v2"
)

var data = `
services:
 - name: firstService
   sites:
    - http://github.com
    - http://google.com
 - name: secondService
   sites:
    - http://google.com
    - https://w3-03.sso.ibm.com/
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
	Services []service
	statuses map[string]string
}

func ServiceList(source []byte) (*serviceList, error) {
	var services serviceList

	services.statuses = make(map[string]string)
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

func (s *serviceList) testUrl(url string, available bool) bool {
	if _, found := s.statuses[url]; !found {
		response, err := http.Head(url)
		if err != nil {
			available = false
			s.statuses[url] = err.Error()
		} else {
			if strings.HasPrefix(response.Status, "4") || strings.HasPrefix(response.Status, "5") {
				available = false
			}
			s.statuses[url] = response.Status
		}
	}

	fmt.Printf("\t%s %s %s\n", white("URL:"), url, formatStatus(s.statuses[url]))
	return available
}

func (s *serviceList) testService(service *service) {
	available := true

	fmt.Printf("%s %s\n", white("Service:"), service.Name)

	for _, url := range service.Sites {
		available = s.testUrl(url, available)
	}

	if available {
		fmt.Printf("\t%s\n", green("Available"))
	} else {
		fmt.Printf("\t%s\n", red("Unavailable"))
	}
}

func main() {
	source := []byte(data)
	services, err := ServiceList(source)
	if err != nil {
		os.Exit(1)
	}

	fmt.Printf("%v\n", services)
	for _, cur := range services.Services {
		services.testService(&cur)
	}
}
