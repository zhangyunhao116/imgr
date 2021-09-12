package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/zhangyunhao116/skipset"
)

var debug = flag.Bool("v", false, "debug mode")

func main() {
	flag.Parse()

	if *debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	var removeAll bool
	if len(flag.Args()) != 0 && flag.Args()[len(flag.Args())-1] == "all" {
		removeAll = true
	}

	dirs, err := os.ReadDir("./")
	if err != nil {
		logrus.Fatalln(err.Error())
	}

	var (
		jpgs = skipset.NewString()
		raws = skipset.NewString()
	)
	for _, v := range dirs {
		name := v.Name()
		lowername := strings.ToLower(v.Name())
		if strings.HasSuffix(lowername, ".arw") {
			raws.Add(name)
			continue
		}
		if strings.HasSuffix(lowername, ".jpg") {
			jpgs.Add(name)
			continue
		}
	}

	// DEBUG jpgs and raws.
	i := 0
	jpgs.Range(func(value string) bool {
		i++
		logrus.Debugln("JPG:", fmt.Sprintf("[%d/%d]", i, jpgs.Len()), value)
		return true
	})

	i = 0
	raws.Range(func(value string) bool {
		i++
		logrus.Debugln("RAW:", fmt.Sprintf("[%d/%d]", i, raws.Len()), value)
		return true
	})

	// Generate removes.
	var removes = skipset.NewString()
	raws.Range(func(value1 string) bool {
		needRemove := true
		name1 := value1[:len(value1)-4]
		jpgs.Range(func(value2 string) bool {
			name2 := value2[:len(value2)-4]
			if name1 == name2 {
				needRemove = false
				return false
			}
			return true
		})
		if needRemove || removeAll {
			removes.Add(value1)
		}
		return true
	})

	// Remove.
	i = 0
	removes.Range(func(value string) bool {
		i++
		logrus.Infoln("REMOVED:", fmt.Sprintf("[%d/%d]", i, removes.Len()), value)
		_, err := execCommandPrintOnlyFailed("REMOVE:", "rm "+value)
		if err != nil {
			logrus.Fatalln(err.Error())
		}
		return true
	})
}

func execCommand(prefix, cmd string) (string, error) {
	var stderr bytes.Buffer
	command := exec.Command("bash", "-c", cmd)
	command.Stderr = &stderr
	out, err := command.Output()
	if err != nil {
		return stderr.String(), errors.New(prefix + ": " + err.Error())
	}
	return string(out), nil
}

func execCommandPrint(prefix, cmd string) (string, error) {
	out, err := execCommand(prefix, cmd)
	if err != nil {
		logrus.Warningf(`exec "%s" error: `, prefix)
	}
	print(string(out))
	return out, err
}

func execCommandPrintOnlyFailed(prefix, cmd string) (string, error) {
	out, err := execCommand(prefix, cmd)
	if err != nil {
		logrus.Warningf(`exec "%s" error: `, prefix)
		print(string(out))
	}
	return out, err
}
