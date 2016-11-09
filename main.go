package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
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

func readWebSource(sourceUrl string) ([]byte, error) {
	response, err := http.Get(sourceUrl)
	if err != nil {
		return nil, err
	}

	source, err := ioutil.ReadAll(response.Body)
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

func printUsage() {
	fmt.Println("Invalid arguments")
	fmt.Println("USAGE: ./bx-availability [-r|-l] YAML_file")
	fmt.Println("OPTIONS: r - read from remote web source")
	fmt.Println("         l - read from local file")
	os.Exit(1)
}

func main() {
	var source []byte
	var err error

	if len(os.Args) < 3 {
		printUsage()
	}

	if os.Args[1] == "-r" {
		source, err = readWebSource(os.Args[2])
	} else if os.Args[1] == "-l" {
		source, err = readLocalSource(os.Args[2])
	} else {
		printUsage()
	}

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
