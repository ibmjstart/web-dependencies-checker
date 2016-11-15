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

	source, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return source, nil
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

func printUsage() {
	fmt.Println(red("Invalid arguments"))
	fmt.Printf("\n%s ./bx-availability -r remote_YAML_file [-t seconds] [-v] [-c]\n", cyan("USAGE:"))
	fmt.Println("          -or-")
	fmt.Println("       ./bx-availability -l local_YAML_file [-t seconds] [-v] [-c]")
	fmt.Printf("\n%s r - read YAML from remote web source\n", cyan("OPTIONS:"))
	fmt.Println("         l - read YAML from local file")
	fmt.Println("         t - http request timeout (in seconds)")
	fmt.Println("         v - verbose")
	fmt.Println("         c - disable color output")
	os.Exit(1)
}

func parseArgs() (func(string) ([]byte, error), string, bool, error) {
	remote := flag.String("r", "", "read YAML from remote web source")
	local := flag.String("l", "", "read YAML from local file")
	timeout := flag.Int("t", 60, "http request timeout (in seconds)")
	verbose := flag.Bool("v", false, "display status response for all URLs")
	noColor := flag.Bool("c", false, "disable color")
	flag.Parse()

	if *noColor {
		color.NoColor = true
	}

	setTimeout(*timeout)

	if (*remote == "") == (*local == "") {
		return nil, "", false, fmt.Errorf("")
	}

	if *remote != "" {
		return readWebSource, *remote, *verbose, nil
	}

	return readLocalSource, *local, *verbose, nil
}

func main() {
	read, sourceFile, verbose, err := parseArgs()
	if err != nil {
		printUsage()
	}

	sourceData, err := read(sourceFile)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
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
