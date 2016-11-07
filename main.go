package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

var data = `
services:
 - name: firstService
   sites:
    - google.com
    - github.com
 - name: secondService
   sites:
    - google.com
`

type service struct {
	Name  string
	Sites []string
}

type serviceList struct {
	Services []service
}

func main() {
	var services serviceList

	source := []byte(data)

	err := yaml.Unmarshal(source, &services)
	if err != nil {
		os.Exit(1)
	}

	fmt.Printf("%v\n", services)
}
