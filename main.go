package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/fatih/color"
)

const jStartUrl string = "www.ibm.com/jstart"

var white (func(string, ...interface{}) string) = color.New(color.FgHiWhite, color.Bold).SprintfFunc()
var green (func(string, ...interface{}) string) = color.New(color.FgGreen, color.Bold).SprintfFunc()
var red (func(string, ...interface{}) string) = color.New(color.FgRed, color.Bold).SprintfFunc()
var cyan (func(string, ...interface{}) string) = color.New(color.FgCyan, color.Bold).SprintfFunc()

func readLocalSource(filepath string) ([]byte, error) {
	source, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	return source, nil
}

func checkProtocol(url string) string {
	if !strings.HasPrefix(url, "http") {
		return "http://" + url
	}

	return url
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

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Invalid number of arguments")
		fmt.Println("USAGE: ./bx-availability local_YAML_file")
		os.Exit(1)
	}

	source, err := readLocalSource(os.Args[1])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	services, err := ServiceList(source)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	go services.write()
	defer close(services.output)

	for _, cur := range services.Services {
		services.testService(&cur)
	}

	services.displayResults()
	fmt.Printf("\nCourtesy of %s - %s\n", cyan("IBM jStart"), jStartUrl)
}
