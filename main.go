// Package gobuild is a wrapper for the go build app
package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/fatih/color"
)

var platforms = map[string]map[string]string{
	"Current":         map[string]string{"GOOS": "darwin", "GOARCH": "amd64"},
	"Aix-PPC64":       map[string]string{"GOOS": "aix", "GOARCH": "ppc64"},
	"Android-386":     map[string]string{"GOOS": "android", "GOARCH": "386"},
	"Android-AMD64":   map[string]string{"GOOS": "android", "GOARCH": "amd64"},
	"Android-ARM":     map[string]string{"GOOS": "android", "GOARCH": "arm"},
	"Android-ARM64":   map[string]string{"GOOS": "android", "GOARCH": "arm64"},
	"Darwin-386":      map[string]string{"GOOS": "darwin", "GOARCH": "386"},
	"Darwin-AMD64":    map[string]string{"GOOS": "darwin", "GOARCH": "amd64"},
	"Darwin-ARM":      map[string]string{"GOOS": "darwin", "GOARCH": "arm"},
	"Darwin-ARM64":    map[string]string{"GOOS": "darwin", "GOARCH": "arm64"},
	"Dragonfly-AMD64": map[string]string{"GOOS": "dragonfly", "GOARCH": "amd64"},
	"Freebsd-386":     map[string]string{"GOOS": "freebsd", "GOARCH": "386"},
	"Freebsd-AMD64":   map[string]string{"GOOS": "freebsd", "GOARCH": "amd64"},
	"Freebsd-ARM":     map[string]string{"GOOS": "freebsd", "GOARCH": "arm"},
	"Illumos-AMD64":   map[string]string{"GOOS": "illumos", "GOARCH": "amd64"},
	"Js-WASM":         map[string]string{"GOOS": "js", "GOARCH": "wasm"},
	"Linux-386":       map[string]string{"GOOS": "linux", "GOARCH": "386"},
	"Linux-AMD64":     map[string]string{"GOOS": "linux", "GOARCH": "amd64"},
	"Linux-ARM":       map[string]string{"GOOS": "linux", "GOARCH": "arm"},
	"Linux-ARM64":     map[string]string{"GOOS": "linux", "GOARCH": "arm64"},
	"Linux-MIPS":      map[string]string{"GOOS": "linux", "GOARCH": "mips"},
	"Linux-MIPS64":    map[string]string{"GOOS": "linux", "GOARCH": "mips64"},
	"Linux-MIPS64LE":  map[string]string{"GOOS": "linux", "GOARCH": "mips64le"},
	"Linux-MIPSLE":    map[string]string{"GOOS": "linux", "GOARCH": "mipsle"},
	"Linux-PPC64":     map[string]string{"GOOS": "linux", "GOARCH": "ppc64"},
	"Linux-PPC64LE":   map[string]string{"GOOS": "linux", "GOARCH": "ppc64le"},
	"Linux-S390X":     map[string]string{"GOOS": "linux", "GOARCH": "s390x"},
	"Nacl-386":        map[string]string{"GOOS": "nacl", "GOARCH": "386"},
	"Nacl-AMD64P32":   map[string]string{"GOOS": "nacl", "GOARCH": "amd64p32"},
	"Nacl-ARM":        map[string]string{"GOOS": "nacl", "GOARCH": "arm"},
	"Netbsd-386":      map[string]string{"GOOS": "netbsd", "GOARCH": "386"},
	"Netbsd-AMD64":    map[string]string{"GOOS": "netbsd", "GOARCH": "amd64"},
	"Netbsd-ARM":      map[string]string{"GOOS": "netbsd", "GOARCH": "arm"},
	"Netbsd-ARM64":    map[string]string{"GOOS": "netbsd", "GOARCH": "arm64"},
	"Openbsd-386":     map[string]string{"GOOS": "openbsd", "GOARCH": "386"},
	"Openbsd-AMD64":   map[string]string{"GOOS": "openbsd", "GOARCH": "amd64"},
	"Openbsd-ARM":     map[string]string{"GOOS": "openbsd", "GOARCH": "arm"},
	"Openbsd-ARM64":   map[string]string{"GOOS": "openbsd", "GOARCH": "arm64"},
	"Plan9-386":       map[string]string{"GOOS": "plan9", "GOARCH": "386"},
	"Plan9-AMD64":     map[string]string{"GOOS": "plan9", "GOARCH": "amd64"},
	"Plan9-ARM":       map[string]string{"GOOS": "plan9", "GOARCH": "arm"},
	"Solaris-AMD64":   map[string]string{"GOOS": "solaris", "GOARCH": "amd64"},
	"Windows-386":     map[string]string{"GOOS": "windows", "GOARCH": "386"},
	"Windows-AMD64":   map[string]string{"GOOS": "windows", "GOARCH": "amd64"},
	"Windows-ARM":     map[string]string{"GOOS": "windows", "GOARCH": "arm"},
}

const supportedPlatforms string = `SUPPORTED PLATFORMS:
  Aix-PPC64, Android-386, Android-AMD64, Android-ARM, Android-ARM64, Darwin-386, Darwin-AMD64
  Darwin-ARM, Darwin-ARM64, Dragonfly-AMD64, Freebsd-386, Freebsd-AMD64, Freebsd-ARM, Illumos-AMD64
  Js-WASM, Linux-386, Linux-AMD64, Linux-ARM, Linux-ARM64, Linux-MIPS, Linux-MIPS64, Linux-MIPS64LE
  Linux-MIPSLE, Linux-PPC64, Linux-PPC64LE, Linux-S390X, Nacl-386, Nacl-AMD64P32, Nacl-ARM, Netbsd-386
  Netbsd-AMD64, Netbsd-ARM, Netbsd-ARM64, Openbsd-386, Openbsd-AMD64, Openbsd-ARM, Openbsd-ARM64
  Plan9-386, Plan9-AMD64, Plan9-ARM, Solaris-AMD64, Windows-386, Windows-AMD64, Windows-ARM
`

type arrayFlagString []string

func (i *arrayFlagString) String() string {
	return fmt.Sprintf("%s", *i)
}

func (i *arrayFlagString) Set(value string) error {
	if len(*i) > 0 {
		return errors.New("interval flag already set")
	}
	for _, t := range strings.Split(value, ",") {
		t = strings.Trim(t, " ")
		if _, ok := platforms[t]; ok {
			*i = append(*i, t)
		} else {
			return errors.New("invalid platform")
		}
	}
	return nil
}

func main() {
	var platform arrayFlagString

	// Current environment variable values
	GOOS := os.Getenv("GOOS")
	GOARCH := os.Getenv("GOARCH")

	// Create objects to colorize stdout
	green := color.New(color.FgGreen)
	red := color.New(color.FgRed)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, supportedPlatforms)
		fmt.Fprintf(os.Stderr, "\nEXAMPLES:\n")
		fmt.Fprintf(os.Stderr, "  gobuild -platform Windows-AMD64,Darwin-AMD64,Linux-AMD64\n")
		fmt.Fprintf(os.Stderr, "\nOPTIONS:\n")
		flag.PrintDefaults()
	}
	flag.Var(&platform, "platform", "Comma-separated list of platforms to build for")
	flag.Parse()

	// Ensure at least one platform is defined, otherwise display usage and exit
	if len(platform) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	for _, p := range platform {
		// Create output folder
		path := fmt.Sprintf("output/%v-%v", platforms[p]["GOOS"], platforms[p]["GOARCH"])
		fmt.Printf("Compiling for %v ... ", p)
		err := os.MkdirAll(path, os.ModePerm)
		if err != nil {
			log.Fatal("unable to create output folder")
		}

		// Set environment variables
		os.Setenv("GOOS", platforms[p]["GOOS"])
		os.Setenv("GOARCH", platforms[p]["GOARCH"])

		// Compile
		cmd := exec.Command("go", "build", "-o", path)
		err = cmd.Run()
		if err != nil {
			red.Printf("fail\nCommand finished with error: %v\n", err)
			continue
		}
		green.Println("success")
	}

	// Reset environment variables
	os.Setenv("GOOS", GOOS)
	os.Setenv("GOARCH", GOARCH)
}
