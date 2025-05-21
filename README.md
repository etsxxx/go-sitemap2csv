# go-sitemap2csv

A simple CLI tool to fetch, parse, and convert website sitemaps to CSV.

[![Test](https://github.com/etsxxx/go-sitemap2csv/actions/workflows/test.yml/badge.svg)](https://github.com/etsxxx/go-sitemap2csv/actions/workflows/test.yml)


## Concept
Simple. No setup. No dependencies.

## How to use

### Install

Download a binary and put it to PATH.

For example, on Linux.

```bash
sudo curl -L -o /usr/local/bin/sitemap2csv $(curl --silent "https://api.github.com/repos/etsxxx/go-sitemap2csv/releases/latest" | jq --arg PLATFORM_ARCH "$(echo `uname -s`-`uname -m` | tr A-Z a-z)" -r '.assets[] | select(.name | endswith($PLATFORM_ARCH)) | .browser_download_url')
sudo chmod 755 /usr/local/bin/sitemap2csv
```

A full list of binaries are [here](https://github.com/etsxxx/go-sitemap2csv/releases/latest).


### Run

```bash
sitemap2csv <sitemap_url> <output_csv_file>
```


## Hack and Develop

### Build

First, fork this repo, and get your clone locally.

1. Install [go](http://golang.org)
2. Install `make`
3. Install [golangci-lint](https://golangci-lint.run/usage/install/#local-installation)

Write code and remove unused modules.

```
make tidy
```

To test, run

```
make lint
make test
```

To build, run

```
make build
```

## AUTHORS

* [etsxxx](https://github.com/etsxxx)
