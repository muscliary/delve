// Generates .travis.yml configuration using pkg/goversion/compat.go
// Usage go run scripts/gen-travis.go > .travis.yml

package main

import (
	"bufio"
	"fmt"
	"os"
	"text/template"

	"github.com/go-delve/delve/pkg/goversion"
)

type arguments struct {
	GoVersions []goVersion
}

type goVersion struct {
	Major, Minor int
}

var maxVersion = goVersion{Major: goversion.MaxSupportedVersionOfGoMajor, Minor: goversion.MaxSupportedVersionOfGoMinor}
var minVersion = goVersion{Major: goversion.MinSupportedVersionOfGoMajor, Minor: goversion.MinSupportedVersionOfGoMinor}

func (v goVersion) dec() goVersion {
	v.Minor--
	if v.Minor < 0 {
		panic("TODO: fill the maximum minor version number for v.Maxjor here")
	}
	return v
}

func (v goVersion) MaxVersion() bool {
	return v == maxVersion
}

func (v goVersion) DotX() string {
	return fmt.Sprintf("%d.%d.x", v.Major, v.Minor)
}

func (v goVersion) String() string {
	return fmt.Sprintf("%d.%d", v.Major, v.Minor)
}

func main() {
	var args arguments

	args.GoVersions = append(args.GoVersions, maxVersion)
	for {
		v := args.GoVersions[len(args.GoVersions)-1].dec()
		args.GoVersions = append(args.GoVersions, v)
		if v == minVersion {
			break
		}
	}

	travisfh, err := os.Create(".travis.yml")
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not create .travis.yml: %v")
		os.Exit(1)
	}

	out := bufio.NewWriter(travisfh)
	err = template.Must(template.New("travis.yml").Parse(`language: go
sudo: required
go_import_path: github.com/go-delve/delve

os:
  - linux

arch:
  - amd64
  - arm64

go:
  - {{index .GoVersions 0}}
  - tip

matrix:
  allow_failures:
    - go: tip
  exclude:
    - arch: arm64
      go: tip

before_install:
  - export GOFLAGS=-mod=vendor
  - if [ $TRAVIS_OS_NAME = "linux" ]; then sudo apt-get -qq update; sudo apt-get install -y dwz; echo "dwz version $(dwz --version)"; fi


# 386 linux
jobs:
  include:
    -  os: linux
       services: docker
       env: go_32_version={{index .GoVersions 0}}.2 # Linux/i386 tests will fail on go1.15 prior to 1.15.2 (see issue #2134)

script: >-
    if [ $TRAVIS_OS_NAME = "linux" ] && [ $go_32_version ]; then
      docker pull i386/centos:7;
      docker run \
      -v $(pwd):/delve \
      --env TRAVIS=true \
      --env CI=true \
      --privileged i386/centos:7 \
      /bin/bash -c "set -x && \
           cd delve && \
           yum -y update && yum -y upgrade && \
           yum -y install wget make git gcc && \
           wget -q https://dl.google.com/go/go${go_32_version}.linux-386.tar.gz && \
           tar -C /usr/local -xzf go${go_32_version}.linux-386.tar.gz && \
           export PATH=$PATH:/usr/local/go/bin && \
           go version && \
           uname -a && \
           make test";
    else
      make test;
    fi
  
cache:
  directories:
    - $HOME/AppData/Local/Temp/chocolatey
`)).Execute(out, args)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error executing template: %v", err)
		os.Exit(1)
	}
	_ = out.Flush()
	_ = travisfh.Close()

	githubfh, err := os.Create(".github/workflows/test.yml")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not create .github/test.yml: %v", err)
		os.Exit(1)
	}
	out = bufio.NewWriter(githubfh)
	err = template.Must(template.New(".github/workflows/test.yml").Parse(`name: Delve CI

on: [push, pull_request]

jobs:
  build:
    runs-on: ${{"{{"}}matrix.os{{"}}"}}
    strategy:
      matrix:
        include:
          - go: {{index .GoVersions 0}}
            os: macos-latest
          - go: {{index .GoVersions 1}}
            os: ubuntu-latest
          - go: {{index .GoVersions 2}}
            os: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - uses: actions/setup-go@v1
        with:
          go-version: ${{"{{"}}matrix.go{{"}}"}}
      - run: go run _scripts/make.go test
`)).Execute(out, args)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error executing template: %v", err)
		os.Exit(1)
	}

	_ = out.Flush()
	_ = githubfh.Close()
}
