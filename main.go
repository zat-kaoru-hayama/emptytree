package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

type Agent interface {
	DoFile(string) error
	DoDir(string) error
}

type DryRun struct{}

func (DryRun) DoFile(string) error { return nil }
func (DryRun) DoDir(string) error  { return nil }

type NormalRun struct{}

func (NormalRun) DoDir(s string) error {
	return os.Mkdir(s, 0777)
}

func (NormalRun) DoFile(s string) error {
	fd, err := os.OpenFile(s,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	return fd.Close()
}

var flagDryRun = flag.Bool("n", false, "dry run")

func mains(args []string) error {
	var agent Agent
	if *flagDryRun {
		agent = DryRun{}
	} else {
		agent = NormalRun{}
	}
	for _, root := range args {
		baseLen := len(root) + 1
		err := filepath.Walk(root, func(path string, f os.FileInfo, _ error) (err error) {
			if len(path) < baseLen {
				return
			}
			relativePath := path[baseLen:]
			if f.IsDir() {
				fmt.Printf("mkdir      \"%s\"\n", relativePath)
				err = agent.DoDir(relativePath)
			} else {
				fmt.Printf("type nul > \"%s\"\n", relativePath)
				err = agent.DoFile(relativePath)
			}
			return
		})
		if err != nil {
			return fmt.Errorf("%s: %w", root, err)
		}
	}
	return nil
}

func main() {
	flag.Parse()
	if err := mains(flag.Args()); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
