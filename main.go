//
// Copyright 2012 Nathan Fiedler. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

// Package goswat implements the command for a Go debugger.
package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

// atExitMutex is used to modify the the list of exit functions.
var atExitMutex sync.Mutex

// atExitFuncs are functions called when the program is exiting.
var atExitFuncs []func()

// RunAtExit registers a function to be invoked when the Exit() function is
// called. There is no guarantee that these functions will be invoked if the
// run time is brought down abruptly (i.e. os.Exit() is called). The
// functions will be invoked in the order in which they are registered.
func RunAtExit(fn func()) {
	// Go currently lacks an "atexit" callback, so we have this
	// hack to provide us with the bare minimum, for now.
	atExitMutex.Lock()
	defer atExitMutex.Unlock()
	atExitFuncs = append(atExitFuncs, fn)
}

// Exit invokes the functions registered to be called prior to exiting, then
// invokes os.Exit() to exit from the program. This function should be called
// instead of os.Exit() in all but the most extreme cases.
func Exit() {
	atExitMutex.Lock()
	for _, fn := range atExitFuncs {
		fn()
	}
	os.Exit(0)
}

// main starts the debugger
func main() {
	// while not a guarantee, at least try to exit cleanly
	defer Exit()
	setupLogging()
	logSysInfo()
	welmsg := `Welcome to GoSwat! To get started, try the 'help' command.
Use 'exit' or Ctrl-c to exit the debugger.`
//Startup commands can be placed in ".goswatrc" in ~ or .`
	fmt.Println(welmsg)
	// TODO: initialize the liswat and swatcl environments
	// TODO: initialize and set up the curses-based interface
	// TODO: find and run the RC file, if any
	// TODO: process the command line arguments, if any
	repl()
}

// repl implements the read-eval-print-loop in which commands are read from
// standard input and the results are displayed to standard output.
func repl() {
	// the following will work on any system, but it is rather crude
	// alternatives include:
	// - exp/terminal -- only works on Linux, but could possibly correct that
	// - http://code.google.com/p/termbox/
	// - http://code.google.com/p/goncurses/
	// - https://github.com/jabb/gocurse
	stdin := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("(goswat) ")
		input, err := stdin.ReadString(10)
		if err != nil {
			fmt.Println(err)
		} else {
			input = strings.TrimSpace(input)
			// process the command
			if input == "exit" {
				fmt.Println("Goodbye")
				Exit()
			} else if input == "help" {
				fmt.Println("I don't really do anything right now...")
			}
		}
	}
}

// setupLogging sets the output of the standard logger to a file in the
// user's home directory, so all log messages will be directed there. If
// anything goes wrong, this function will call log.Fatal().
func setupLogging() {
	usr, err := user.Current()
	if err != nil {
		log.Fatalln(err)
	}
	goswatdir := filepath.Join(usr.HomeDir, ".goswat")
	if _, err := os.Stat(goswatdir); err != nil {
		if os.IsNotExist(err) {
			os.Mkdir(goswatdir, 0755)
		} else {
			log.Fatalln(err)
		}
	}
	logname := filepath.Join(goswatdir, "messages.log")
	logfile, err := os.OpenFile(logname, os.O_WRONLY | os.O_APPEND | os.O_CREATE, 0644)
	if err != nil {
		log.Fatalln(err)
	}

	out := bufio.NewWriter(logfile)
	log.SetOutput(out)
	closer := func() {
		out.Flush()
		logfile.Sync()
		logfile.Close()
	}
	RunAtExit(closer)
	// from this point on, everything will go to messages.log
}

// logSysInfo writes a set of information about the system to the
// log file, useful for debugging in the event of an error.
func logSysInfo() {
	header := "-------------------------------------------------------------------------------"
	now := time.Now()
	log.Println(header)
	log.Printf("Log Session: %s\n", now.Format(time.ANSIC))
	log.Printf("Product Version = %s\n", "dev") // TODO
	//log.Printf("Operating System = %s\n", "TODO")
	log.Printf("Go Version = %s\n", runtime.Version())
	//log.Printf("System Locale; Encoding = %s\n", "TODO")
	usr, err := user.Current()
	if err != nil {
		log.Println(err)
		Exit()
	}
	log.Printf("Home Directory = %s\n", usr.HomeDir)
	pwd, err := os.Getwd()
	if err != nil {
		log.Println(err)
		Exit()
	}
	log.Printf("Current Directory = %s\n", pwd)
	log.Printf("GOROOT = %s\n", runtime.GOROOT())
	// print selected entries from the environment
	keys := []string{"PATH", "LANG", "LC_ALL", "SHELL", "TERM"}
	for _, key := range keys {
		if val := os.Getenv(key); val != "" {
			log.Printf("%s = %s", key, val)
		}
	}
	log.Println(header)
}
