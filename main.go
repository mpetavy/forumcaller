// +build windows

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// forumlauncher.exe -username czmadmin -password czmAdmin2008 -sopInstanceUid 1.2.276.0.75.2.2.70.0.3.9210271872519.20170801150225000.133221
//
// Windows compile:
// Docu: https://github.com/josephspurrier/goversioninfo
// go get github.com/josephspurrier/goversioninfo/cmd/goversioninfo
// go generate
// go install -ldflags -H=windowsgui
//
// MacOS compile:
// Docu: https://medium.com/@mattholt/packaging-a-go-application-for-macos-f7084b00f6b5

const (
	timeout = 2000
)

var (
	viewerpath string
)

func init() {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	files, err := ioutil.ReadDir(wd)
	if err != nil {
		panic(err)
	}

	for _, f := range files {
		name := strings.ToLower(f.Name())

		if strings.Index(name, "viewer") != -1 && (strings.HasSuffix(name, ".exe") || strings.HasSuffix(name, ".dmg")) {
			viewerpath = wd + string(filepath.Separator) + name
			break
		}

		if strings.Index(name, "launchforum") != -1 && (strings.HasSuffix(name, ".cmd") || strings.HasSuffix(name, ".sh")) {
			viewerpath = wd + string(filepath.Separator) + name
			break
		}
	}
}

func isWindows() bool {
	return strings.ToLower(runtime.GOOS) == "windows"
}

func fileExists(filename string) bool {
	var b bool
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		b = false
	}

	if _, err := os.Stat(filename); !os.IsNotExist(err) {
		b = true
	}

	return b
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	app, err := os.Executable()
	if err != nil {
		panic(err)
	}

	fmt.Printf("\n")
	fmt.Printf("%s - Launcher for FORUM Viewer\n", strings.ToUpper(app))
	fmt.Printf("\n")

	usr, err := user.Current()
	if err != nil {
		panic(err)
	}

	log.Printf("user home dir: %s", usr.HomeDir)

	sessionName := ""

	if isWindows() {
		v,b := os.LookupEnv("SESSIONNAME")

		if b {
			sessionName = "-" + strings.ToUpper(v)
		}
	}

	filename := filepath.Join(usr.HomeDir, fmt.Sprintf(".forumlauncher%s.properties",sessionName))

	log.Printf("launcher file: %s", filename)

	file, err := os.Create(filename)
	if err != nil {
		panic(err)
	}

	var cmdLine string

	if len(os.Args) > 1 {
		cmdLine = strings.Join(os.Args[1:], " ")
	} else {
		cmdLine = "-show"
	}

	fmt.Fprintf(file, "%s", cmdLine)

	log.Printf("launcher parameter: %s", cmdLine)

	err = file.Close()
	if err != nil {
		panic(err)
	}

	log.Printf("launcher file written")

	var fileTaken bool

	log.Printf("loop on timeout %d msec or file taken ...", timeout)

	start := time.Now()

	for {
		time.Sleep(time.Millisecond * 250)

		fileTaken = !fileExists(filename)

		if fileTaken || time.Now().Sub(start) > (time.Duration(timeout)*time.Millisecond) {
			break
		}
	}

	if fileTaken {
		log.Printf("launcher file was successfully taken by running forum viewer instance")
		os.Exit(0)
	}

	log.Printf("launcher file still exists after timeout -> start a new FORUM instance ...")

	err = os.Remove(filename)
	if err != nil {
		panic(err)
	}

	var args []string

	if len(viewerpath) == 0 || !fileExists(viewerpath) {
		panic(fmt.Sprintf("no viewer executable found or viewerpath does not exist: %s", viewerpath))
	}

	if strings.HasSuffix(strings.ToLower(viewerpath), ".cmd") || strings.HasSuffix(strings.ToLower(viewerpath), ".bat") {
		args = append(args, "cmd.exe")
		args = append(args, "/c")
	}

	args = append(args, viewerpath)
	args = append(args, cmdLine)

	pargs := make([]string, len(args))

	copy(pargs, args)

	for i := range pargs {
		if isWindows() {
			pargs[i] = "\"" + pargs[i] + "\""
		} else {
			pargs[i] = "'" + pargs[i] + "'"
		}
	}

	log.Printf("exec command: %s %s", pargs[0], strings.Join(pargs[1:], " "))

	cmd := exec.Command(args[0], args[1:]...)

	err = cmd.Start()
	if err != nil {
		panic(err)
	}
}
