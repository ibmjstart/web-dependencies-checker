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

type retryClient struct {
	maxRetries int
	userAgent  string
	http.Client
}

var client = &retryClient{}

func Client(timeout, maxRetries int) {
	client.Timeout = time.Duration(timeout) * time.Second
	client.maxRetries = maxRetries
	client.userAgent = "Mozilla/5.0 (Windows NT 6.1) AppleWebKit/537.36 " +
		"(KHTML, like Gecko) Chrome/41.0.2228.0 Safari/537.36"
}

func (c *retryClient) newRequest(url string) (*http.Request, error) {
	request, err := http.NewRequest("HEAD", formatUrl(url), nil)
	if err != nil {
		return nil, err
	}

	request.Header.Set("User-Agent", c.userAgent)
	return request, nil
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

func getAvailability(status string) bool {
	if strings.HasPrefix(status, "2") {
		return true
	}

	return false
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

func formatStatus(url, status string) string {
	formattedStatus := ""

	if strings.HasPrefix(status, "2") {
		formattedStatus += fmt.Sprintf("%s", green(status))
	} else if strings.HasPrefix(status, "3") {
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

	return formattedStatus
}

func parseArgs() ([][]byte, bool, error) {
	timeout := flag.Int("t", 60, "http request timeout (in seconds)")
	retries := flag.Int("r", 0, "number of http request retries")
	quiet := flag.Bool("q", false, "only display status for failed requests")
	noColor := flag.Bool("c", false, "disable color output")
	flag.Parse()

	if *noColor {
		color.NoColor = true
	}

	Client(*timeout, *retries)
	locations := flag.Args()

	if len(locations) == 0 {
		return nil, false, fmt.Errorf("No YAML file(s) provided")
	}

	sourceData, err := getData(locations)
	if err != nil {
		return nil, false, err
	}

	return sourceData, *quiet, nil
}

func printUsage(err error) {
	usage := red("Invalid arguments: ") + err.Error() + "\n\n" +
		cyan("USAGE:") + " ./bx-availability [-t seconds] [-r retries] [-q] [-c] " +
		"YAML_file_location [YAML_file_location...]\n\n" +
		cyan("OPTIONS:") + " t - http request timeout (in seconds) (default 60)\n" +
		"         r - number of http request retries (default 0)\n" +
		"         q - only display status for failed requests\n" +
		"         c - disable color output\n"

	fmt.Print(usage)
	os.Exit(1)
}

func main() {
	sourceData, quiet, err := parseArgs()
	if err != nil {
		printUsage(err)
	}

	services, err := ServiceList(sourceData, quiet)
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

	if !quiet {
		services.displayResults()
	}
	fmt.Printf("\nCourtesy of %s - %s\n", cyan("IBM jStart"), jStartUrl)
}
