package main

import (
	"os"

	"github.com/urfave/cli/v2"
)

func printError(err error) {
	if err != nil {
		os.Stderr.WriteString("krun: " + err.Error() + "\r\n")
		os.Exit(1)
	}
}

func main() {
	app := cli.NewApp()
	app.Name = "krun"
	app.Usage = "Run a command in a set of Kubernetes pod"
	app.Version = "0.1.0"

	app.Commands = []*cli.Command{
		{
			Name:    "init",
			Aliases: []string{"i"},
			Usage:   "Initialize the environment for krun",
			Flags: []cli.Flag{
				&cli.IntFlag{
					Name:        "threads",
					Usage:       "Number of threads to use for peering",
					DefaultText: "8",
					Value:       8,
					Aliases:     []string{"t"},
				},
				&cli.IntFlag{
					Name:        "wait-min",
					Usage:       "Minimum peers to wait for the peering to complete",
					DefaultText: "1",
					Value:       1,
					Aliases:     []string{"min", "m"},
					EnvVars:     []string{"KRUN_WAIT_MIN"},
				},
				&cli.BoolFlag{
					Name:        "no-hosts",
					Usage:       "Do not write /etc/hosts file",
					DefaultText: "false",
					Value:       false,
					Aliases:     []string{"nh", "n"},
				},
			},
			Action: initCommandHandler,
		},
		{
			Name:   "_init",
			Hidden: true,
			Usage:  "Print the init info for peering",
			Action: initPeerInfoHandler,
		},
		{
			Name:    "hosts",
			Aliases: []string{"ho", "host"},
			Usage:   "List the hosts of the service",
			Action:  hostsCommandHandler,
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:        "short",
					Aliases:     []string{"s", "one-line"},
					Usage:       "Short one-line output format.",
					DefaultText: "false",
					Value:       false,
				},
				&cli.BoolFlag{
					Name:        "no-headers",
					Usage:       "Do not show table headers",
					Value:       false,
					DefaultText: "false",
					Aliases:     []string{"n", "no-header"},
				},
				&cli.BoolFlag{
					Name:    "hostname",
					Usage:   "Show hostnames of the service",
					Value:   false,
					Aliases: []string{"H", "hostnames"},
				},
				&cli.BoolFlag{
					Name:    "ip",
					Usage:   "Show IP addresses of the service",
					Value:   false,
					Aliases: []string{"I", "ips"},
				},
				&cli.StringFlag{
					Name:    "suffix",
					Usage:   "Suffix to output after the lines",
					Value:   "",
					Aliases: []string{"S"},
				},
			},
		},
		{
			Name:    "run",
			Aliases: []string{"r"},
			Usage:   "Run a command in a set of pods",
			Action:  runCommandHandler,
			Flags: []cli.Flag{
				&cli.StringSliceFlag{
					Name:    "hosts",
					Usage:   "Hosts to run the command on",
					Aliases: []string{"w", "host"},
				},
				&cli.IntFlag{
					Name:    "n-hosts",
					Usage:   "Number of hosts to run the command on",
					Value:   1,
					Aliases: []string{"N", "nh"},
				},
				&cli.IntFlag{
					Name:        "n-procs",
					Usage:       "Number of processes to run in parallel in total",
					DefaultText: "1",
					Value:       1,
					Aliases:     []string{"n", "np"},
				},
				&cli.StringFlag{
					Name:    "wd",
					Usage:   "Working directory to run the command in",
					Value:   "",
					Aliases: []string{"d", "dir"},
				},
			},
		},
	}
	app.Flags = append(app.Flags, &cli.StringFlag{
		Name:        "state-file",
		Usage:       "File to store the state of the peering",
		Aliases:     []string{"s", "state"},
		DefaultText: "/tmp/krun.state",
		Value:       "/tmp/krun.state",
	})

	err := app.Run(os.Args)
	if err != nil {
		printError(err)
	}
}
