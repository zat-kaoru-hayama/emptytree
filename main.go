package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Agent interface {
	DoFile(string) error
	DoDir(string) error
	Close() error
}

type DryRun struct{}

func (DryRun) DoFile(string) error { return nil }
func (DryRun) DoDir(string) error  { return nil }
func (DryRun) Close() error        { return nil }

type NormalRun struct{}

func (NormalRun) DoDir(s string) error {
	err := os.Mkdir(s, 0777)
	if errors.Is(err, os.ErrExist) {
		return nil
	}
	return err
}

func (NormalRun) DoFile(s string) error {
	fd, err := os.OpenFile(s,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if errors.Is(err, os.ErrExist) {
		return nil
	}
	if err != nil {
		return err
	}
	return fd.Close()
}

func (NormalRun) Close() error { return nil }

type Undo struct {
	list []string
}

func (*Undo) DoFile(fname string) error {
	stat, err := os.Stat(fname)
	if err != nil {
		return nil
	}
	if stat.Size() > 0 {
		return fmt.Errorf("%s: not zero size", fname)
	}
	return os.Remove(fname)
}

func (u *Undo) DoDir(fname string) error {
	u.list = append(u.list, fname)
	return nil
}

func (u *Undo) Close() error {
	for i := len(u.list) - 1; i >= 0; i-- {
		os.Remove(u.list[i])
	}
	return nil
}

var flagDryRun = flag.Bool("n", false, "dry run")
var flagUndo = flag.Bool("u", false, "undo")

func mains(args []string) error {
	var agent Agent
	if *flagDryRun {
		agent = DryRun{}
	} else if *flagUndo {
		agent = &Undo{}
	} else {
		agent = NormalRun{}
	}
	defer agent.Close()

	for _, root := range args {
		err := filepath.Walk(root, func(path string, f os.FileInfo, _ error) (err error) {
			relativePath := path
			if strings.HasPrefix(relativePath, root) {
				relativePath = relativePath[len(root):]
			}
			if len(relativePath) <= 0 {
				return nil
			}
			if relativePath[0] == filepath.Separator {
				relativePath = relativePath[1:]
			}
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
