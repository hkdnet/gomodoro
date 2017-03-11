package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"time"

	"gopkg.in/yaml.v2"

	"github.com/mitchellh/go-homedir"
)

func NewSpan(sec int) Span {
	ret := Span{
		TotalSeconds: sec,
		RestSeconds:  sec,
	}
	ret.Minutes = minutesOf(ret)
	ret.Seconds = secondsOf(ret)
	return ret
}

type Span struct {
	TotalSeconds int
	RestSeconds  int
	Minutes      int
	Seconds      int
}
type config struct {
	PomodoroTime int    `yaml:"pomodoro"`
	BreakTime    int    `yaml:"break"`
	Pre          string `yaml:"pre"`
	Post         string `yaml:"post"`
}

func NewConfig(path string) (c config, err error) {
	c = config{}
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return
	}
	err = yaml.Unmarshal(data, &c)
	if err != nil {
		return
	}
	return
}

func minutesOf(s Span) int {
	return s.RestSeconds / 60
}

func secondsOf(s Span) int {
	return s.RestSeconds % 60
}

func (s Span) Tick() Span {
	return NewSpan(s.RestSeconds - 1)
}

func (s Span) String() string {
	return fmt.Sprintf("%02d:%02d", s.Minutes, s.Seconds)
}

func tmuxStr(s string) string {
	return "#[fg=mycolor,bg=mycolor]#[fg=default]" + s + "#[fg=mycolor,bg=mycolor]"
}

func main() {
	os.Exit(run())
}

func run() int {
	fs := flag.NewFlagSet("gomodoro", flag.ContinueOnError)
	home, err := homedir.Dir()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	var isBreak bool
	var isDaemon bool
	var filePath string
	var configPath string
	fs.BoolVar(&isBreak, "b", false, "where break time or not")
	fs.BoolVar(&isDaemon, "d", false, "where run as daemon or not")
	fs.StringVar(&filePath, "f", path.Join(home, ".gomodoro.tmux"), "rest time")
	fs.StringVar(&configPath, "c", path.Join(home, ".gomodoro.yml"), "path to config file")
	c, err := NewConfig(configPath)
	fs.Parse(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	var s Span
	if isBreak {
		s = NewSpan(c.BreakTime * 60)
	} else {
		s = NewSpan(c.PomodoroTime * 60)
	}
	if len(c.Pre) != 0 {
		err = exec.Command("bash", "-lc", c.Pre).Run()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
	}
	tickCh := make(chan struct{})
	errCh := make(chan error)
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	go func() {
		<-sigCh
		errCh <- errors.New("Interrupted")
	}()
	go func() {
		tmp := time.Now()
		threshold := 1.0
		for {
			dur := time.Now().Sub(tmp)
			if dur.Seconds() >= threshold {
				tickCh <- struct{}{}
				threshold++
			}
		}
	}()
	defer func() {
		err := os.Remove(filePath)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}()
loop:
	for {
		select {
		case <-tickCh:
			s = s.Tick()
			if s.RestSeconds < 0 {
				break loop
			}
			t := tmuxStr(s.String())
			go func(s string) {
				err := ioutil.WriteFile(filePath, []byte(s), 0644)
				if err != nil {
					errCh <- err
				}
			}(t)
		case err := <-errCh:
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
	}
	if len(c.Post) != 0 {
		err = exec.Command("bash", "-lc", c.Post).Run()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
	}
	return 0
}
