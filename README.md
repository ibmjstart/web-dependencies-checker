# web-dependencies-checker

[![standard-readme compliant](https://img.shields.io/badge/readme%20style-standard-brightgreen.svg?style=flat-square)](https://github.com/RichardLitt/standard-readme)

> A tool for checking the availability of web dependencies

## Table of Contents

- [Background](#background)
- [Install](#install)
- [Usage](#usage)
- [Contribute](#contribute)
- [License](#license)

## Background

Numerous programs and services require the ability to talk to the web in order to get their jobs done. Sometimes, firewall
settings or other environmental constraints can prevent this communication from happening. This tool can be used to see
if a service's web dependencies are reachable from your machine. Anyone maintaining a project with web dependencies can
create a simple yaml file that the web-dependencies-checker can run against.

## Install

You can download the `web-dependencies-checker` executable for your system from the
[releases page](https://github.com/ibmjstart/web-dependencies-checker/releases). Once downloaded, navigate to the executable's
destination and run it as detailed in the following section.

## Usage

This tool is invoked as follows:

`./wdc [-t seconds] [-r retries] [-q] [-c] YAML_file_location [YAML_file_location...]`

#### YAML File

The `YAML_file_location` argument can be either a path to a local YAML file or the url of a remote one. Urls must begin
with `http(s)`. One or more YAML files can be provided as arguments and they can be a combination of local paths and urls.

The format of the YAML file must be as follows:

```yaml
---
services:
 - name: firstService
   sites:
    - site1
    - site2
 - name: secondService
   sites:
    - site3
    - site4
```

It can include any number of services, and each service can contain any number of urls. Multiple services can contain
the same url. All urls will use TCP by default. If a url requires http(s) be sure to prepend the url with `http(s)://`.
Wildcard subdomains are not supported. Any url with wildcards will be stripped of them before testing. IP addresses are
supported.

The included YAML file `test_yaml.yml` is a sample list of services and urls. The following command can be used to check
the services listed in that file.

```
./wdc yaml/test_yaml.yml
```

#### Options

Name | Flag | Description
---  | ---  | ---
Timeout | -t | HTTP request timeout in seconds (defaults to 60)
Retries | -r | Number of HTTP request retries (defaults to 0)
Quiet | -q | Only display status for failed requests
No Color | -c | Disable color output

## Contribute

PRs accepted.

Small note: If editing the README, please conform to the [standard-readme](https://github.com/RichardLitt/standard-readme) specification.

## License
Apache 2.0
 © IBM jStart
