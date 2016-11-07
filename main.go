package main

import (
	"fmt"
	"net/http"
	"os"

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
var blue (func(string, ...interface{}) string) = color.New(color.FgBlue, color.Bold).SprintfFunc()

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

func (s *serviceList) testService(service *service) {
	for _, url := range service.Sites {
		if _, found := s.statuses[url]; !found {
			response, err := http.Head(url)
			if err != nil {
				s.statuses[url] = err.Error()
			} else {
				s.statuses[url] = response.Status
			}
		}
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

	fmt.Printf("%v\n", services.statuses)
}
