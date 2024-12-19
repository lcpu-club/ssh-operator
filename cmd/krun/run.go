package main

import (
	"errors"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/urfave/cli/v2"
)

func stringRangeToSlice(s string) []int {
	s = strings.Trim(s, " \t\r\n")
	rl, rr, ok := strings.Cut(s, "-")
	if !ok {
		return []int{}
	}
	l, err := strconv.Atoi(rl)
	if err != nil {
		return []int{}
	}
	r, err := strconv.Atoi(rr)
	if err != nil {
		return []int{}
	}
	rslt := []int{}
	for i := l; i <= r; i++ {
		rslt = append(rslt, i)
	}
	return rslt
}

func stringToStringSlice(s []string) []string {
	rslt := []string{}
	for _, v := range s {
		v1 := strings.Split(v, ",")
		for _, v2 := range v1 {
			r1 := strings.Trim(v2, " \t\r\n")
			if strings.HasSuffix(r1, "]") {
				rPre, rRange, ok := strings.Cut(r1, "[")
				if !ok {
					rslt = append(rslt, r1)
					continue
				}
				rangeSlice := stringRangeToSlice(rRange[:len(rRange)-1])
				for _, r := range rangeSlice {
					rslt = append(rslt, rPre+strconv.Itoa(r))
				}
			} else if strings.Contains(r1, "-") {
				rs := stringRangeToSlice(r1)
				for _, r := range rs {
					rslt = append(rslt, strconv.Itoa(r))
				}
			} else {
				rslt = append(rslt, r1)
			}
		}
	}
	return rslt
}

type lockedWriter struct {
	lock   *sync.Mutex
	writer io.Writer
}

func (lw *lockedWriter) Write(p []byte) (n int, err error) {
	lw.lock.Lock()
	defer lw.lock.Unlock()
	return lw.writer.Write(p)
}

type PrefixWriter struct {
	Prefix string
	Writer io.Writer
}

func (pw *PrefixWriter) Write(p []byte) (n int, err error) {
	// for lines in p
	//   write prefix + line
	//   if line does not end with '\n'
	//     write '\n'
	// return len(p), err
	origLen := len(p)
	newOut := ""
	if p[len(p)-1] == '\n' {
		p = p[:len(p)-1]
	}
	for _, c := range strings.Split(string(p), "\n") {
		newOut += pw.Prefix + c + "\n"
	}
	_, err = pw.Writer.Write([]byte(newOut))
	return origLen, err
}

func runCommandHandler(c *cli.Context) error {
	state, err := loadState(c)
	if err != nil {
		return err
	}

	runner := sshRunner

	hosts := c.StringSlice("hosts")
	nProcs := c.Int("n-procs")
	nHosts := c.Int("n-hosts")
	wd := c.String("wd")

	commands := c.Args().Slice()

	hosts = stringToStringSlice(hosts)

	if wd == "" {
		wd, err = os.Getwd()
		if err != nil {
			return err
		}
	}

	if len(hosts) == 0 {
		if nHosts > len(state) {
			return errors.New("not enough hosts")
		}
		if nHosts < 1 {
			nHosts = len(state)
		}

		for i := 0; i < nHosts; i++ {
			hosts = append(hosts, state[i].IP)
		}
	} else {
		for i, host := range hosts {
			if n, err := strconv.ParseInt(host, 10, 64); err == nil {
				if n < 0 || int(n) >= len(state) {
					return errors.New("invalid host number")
				}
				hosts[i] = state[n].IP
			}
		}
	}

	wg := &sync.WaitGroup{}
	outPipe := &lockedWriter{
		lock:   &sync.Mutex{},
		writer: os.Stdout,
	}
	errPipe := &lockedWriter{
		lock:   &sync.Mutex{},
		writer: os.Stderr,
	}

	if len(commands) == 0 {
		return errors.New("no command specified")
	}

	for _, host := range hosts {
		for range nProcs {
			wg.Add(1)
			go func(host string) {
				defer wg.Done()
				cmd := runner.Command(host, commands)
				cmd.Stdout = &PrefixWriter{
					Prefix: host + ": ",
					Writer: outPipe,
				}
				cmd.Stderr = &PrefixWriter{
					Prefix: host + ": ",
					Writer: errPipe,
				}
				cmd.Dir = wd
				err := cmd.Run()
				if err != nil {
					errPipe.Write([]byte("krun: " + err.Error() + "\n"))
				}
			}(host)
		}
	}

	wg.Wait()

	return nil
}
