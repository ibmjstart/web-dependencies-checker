package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
)

const jStartUrl string = "www.ibm.com/jstart"

var green (func(string, ...interface{}) string) = color.New(color.FgGreen, color.Bold).SprintfFunc()
var yellow (func(string, ...interface{}) string) = color.New(color.FgYellow, color.Bold).SprintfFunc()
var red (func(string, ...interface{}) string) = color.New(color.FgRed, color.Bold).SprintfFunc()
var cyan (func(string, ...interface{}) string) = color.New(color.FgCyan).SprintfFunc()

var client = &http.Client{}

func setTimeout(seconds int) {
	client.Timeout = time.Duration(seconds) * time.Second
}

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

	if !strings.HasPrefix(response.Status, "2") {
		return nil, fmt.Errorf("%s at %s", response.Status, sourceUrl)
	}

	source, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return source, nil
}

func getData(locations []string) ([][]byte, error) {
	var tempData []byte
	var dataList [][]byte
	var err error

	for _, location := range locations {
		if strings.HasPrefix(location, "http") {
			tempData, err = readWebSource(location)
		} else {
			tempData, err = readLocalSource(location)
		}

		if err != nil {
			return nil, err
		}

		dataList = append(dataList, tempData)
	}

	return dataList, nil
}

func formatUrl(url string) string {
	if strings.Contains(url, "*.") {
		url = strings.Replace(url, "*.", "", -1)
	}

	if !strings.HasPrefix(url, "http") {
		return "http://" + url
	}

	return url
}

func formatStatus(url, status string) (bool, string) {
	isAvailable := false
	formattedStatus := ""

	if strings.HasPrefix(status, "2") {
		isAvailable = true
		formattedStatus += fmt.Sprintf("%s", green(status))
	} else if strings.HasPrefix(status, "3") {
		isAvailable = true
		formattedStatus += fmt.Sprintf("%s", yellow(status))
	} else if strings.HasPrefix(status, "4") || strings.HasPrefix(status, "5") {
		formattedStatus += fmt.Sprintf("%s", red(status))
	} else {
		formattedStatus += fmt.Sprintf("%s %s", red("FAILED:"), status)
	}

	if strings.Contains(url, "*.") {
		formattedStatus += fmt.Sprintf("\n\t      %s Wildcards unsupported, reporting for %s ",
			yellow("WARNING:"), yellow(strings.Replace(url, "*.", "", -1)))
	}

	return isAvailable, formattedStatus
}

func printUsage(err error) {
	usage := red("Invalid arguments: ") + err.Error() + "\n\n" +
		cyan("USAGE:") + " ./bx-availability [-t seconds] [-v] [-c] " +
		"YAML_file_location [YAML_file_location...]\n\n" +
		cyan("OPTIONS:") + " t - http request timeout (in seconds)\n" +
		"         v - verbose\n" +
		"         c - disable color output\n"

	fmt.Print(usage)
	os.Exit(1)
}

func parseArgs() ([][]byte, bool, error) {
	timeout := flag.Int("t", 60, "http request timeout (in seconds)")
	verbose := flag.Bool("v", false, "display status response for all URLs")
	noColor := flag.Bool("c", false, "disable color")
	flag.Parse()

	if *noColor {
		color.NoColor = true
	}

	setTimeout(*timeout)
	locations := flag.Args()

	if len(locations) == 0 {
		return nil, false, fmt.Errorf("No YAML file(s) provided")
	}

	sourceData, err := getData(locations)
	if err != nil {
		return nil, false, err
	}

	return sourceData, *verbose, nil
}

func main() {
	sourceData, verbose, err := parseArgs()
	if err != nil {
		printUsage(err)
	}

	services, err := ServiceList(sourceData, verbose)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	done := make(chan int)
	go services.write(done)

	for i, _ := range services.Services {
		services.testService(i)
		_ = <-done
	}

	close(services.output)
	close(done)

	if verbose {
		services.displayResults()
	}
	fmt.Printf("\nCourtesy of %s - %s\n", cyan("IBM jStart"), jStartUrl)
}
