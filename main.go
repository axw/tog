package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

func main() {
	if len(os.Args) <= 1 {
		fmt.Fprintf(os.Stderr, "Usage: %s command [args...]\n", filepath.Base(os.Args[0]))
		os.Exit(1)
	}

	closeLog := func() {}
	log.SetFlags(0)
	log.SetPrefix("[tog] ")
	log.SetOutput(ioutil.Discard)
	if togDebugFilename := os.Getenv("TOG_DEBUG"); togDebugFilename != "" {
		if togDebugFilename == "stderr" {
			log.SetOutput(os.Stderr)
		} else {
			f, err := os.OpenFile(togDebugFilename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
			if err == nil {
				log.SetOutput(f)
				closeLog = func() {
					log.SetOutput(ioutil.Discard)
					f.Close()
				}
			}
		}
	}

	args := os.Args[1:]
	tool := filepath.Base(args[0])
	switch tool {
	case "compile":
		args = transformCompile(args)
	}

	log.Printf("executing %s", args)
	closeLog()
	execv(args)
}

func transformCompile(args []string) []string {
	var (
		buildid    string
		importpath string
		gofiles    []string
	)
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-buildid":
			buildid = args[i+1]
			i++
		case "-p":
			importpath = args[i+1]
			i++
		}
	}

	// TODO(axw) short-circuit below if we have already
	// transformed for buildid previously with the same
	// transformation modules.

	for i := len(args) - 1; i > 0; i-- {
		if !strings.HasSuffix(args[i], ".go") {
			gofiles = args[i+1:]
			break
		}
	}

	_ = buildid
	_ = importpath
	_ = gofiles
	return args
}

func execv(args []string) {
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		switch err := err.(type) {
		case *exec.ExitError:
			// TODO(axw) Windows exit status
			switch sys := err.Sys().(type) {
			case syscall.WaitStatus:
				os.Exit(sys.ExitStatus())
			}
		default:
			fmt.Fprintln(os.Stderr, err)
		}
		os.Exit(254)
	}
}
