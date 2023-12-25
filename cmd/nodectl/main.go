package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"syscall"

	"github.com/gravitational/trace"
	node "github.com/solidDoWant/nodectl/pkg"
	"github.com/urfave/cli/v2"
)

func main() {
	outLogger := log.New(os.Stdout, "", 0)

	app := &cli.App{
		Name:                 "nodectl",
		Version:              "v0.0.2", // TODO update this via linker flags at some point
		EnableBashCompletion: true,
		Commands: []*cli.Command{
			listCommand(outLogger),
			poweronCommand(),
			poweroffCommand(),
			rebootCommand(),
			flashCommand(),
			consoleCommand(),
			rescanCommand(),
		},
		// TODO allow for setting log level
	}

	err := app.Run(os.Args)
	if err != nil {
		exitHandler(trace.Wrap(err, "an error occured while runnning the app"))
	}
}

func listCommand(logger *log.Logger) *cli.Command {
	activeOnlyFlagName := "active-only"

	action := func(ctx *cli.Context) error {
		entries, err := node.NewPCIeController().List(ctx.Bool(activeOnlyFlagName))
		if err != nil {
			return trace.Wrap(err, "failed to list system PCIe devices")
		}

		for _, entry := range entries {
			logger.Println(entry)
		}

		if len(entries) == 0 {
			logger.Println("No devices found. Run the rescan command then try again.")
		}

		return nil
	}

	return &cli.Command{
		Name:        "list",
		Aliases:     []string{"l"},
		Description: "lists all connected nodes",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    activeOnlyFlagName,
				Aliases: []string{"e"},
				Usage:   "set to only report nodes that are active and online",
			},
		},
		Action: action,
	}
}

func poweronCommand() *cli.Command {
	flags, providedNodeNumbers := getNodeSelectionFlags()
	action := func(ctx *cli.Context) error {
		nodes, err := node.GetNodes()
		if err != nil {
			return trace.Wrap(err, "failed to get all nodes")
		}

		for _, nodeNumber := range *providedNodeNumbers {
			err = nodes[nodeNumber].PowerOn()
			if err != nil {
				return trace.Wrap(err, "failed to power on node %d", nodeNumber)
			}
		}

		return nil
	}

	return &cli.Command{
		Name:        "poweron",
		Aliases:     []string{"u"},
		Description: "powers on nodes",
		Flags:       flags,
		Action:      action,
	}
}

func poweroffCommand() *cli.Command {
	flags, providedNodeNumbers := getNodeSelectionFlags()
	action := func(ctx *cli.Context) error {
		nodes, err := node.GetNodes()
		if err != nil {
			return trace.Wrap(err, "failed to get all nodes")
		}

		for _, nodeNumber := range *providedNodeNumbers {
			err = nodes[nodeNumber].PowerOff()
			if err != nil {
				return trace.Wrap(err, "failed to power off node %d", nodeNumber)
			}
		}

		return nil
	}

	return &cli.Command{
		Name:        "poweroff",
		Aliases:     []string{"d"},
		Description: "powers off nodes",
		Flags:       flags,
		Action:      action,
	}
}

func rebootCommand() *cli.Command {
	flags, providedNodeNumbers := getNodeSelectionFlags()
	action := func(ctx *cli.Context) error {
		nodes, err := node.GetNodes()
		if err != nil {
			return trace.Wrap(err, "failed to get all nodes")
		}

		for _, nodeNumber := range *providedNodeNumbers {
			err = nodes[nodeNumber].Reboot()
			if err != nil {
				return trace.Wrap(err, "failed to power off node %d", nodeNumber)
			}
		}

		return nil
	}

	return &cli.Command{
		Name:        "reboot",
		Aliases:     []string{"restart", "r"},
		Description: "reboots nodes",
		Flags:       flags,
		Action:      action,
	}
}

func flashCommand() *cli.Command {
	selectionFlags, _ := getNodeSelectionFlags()

	return &cli.Command{
		Name:        "flash",
		Aliases:     []string{"f"},
		Description: "flashes the selected nodes",
		Flags: append(
			selectionFlags,
			&cli.PathFlag{
				Name:    "file",
				Aliases: []string{"f"},
				Usage:   "the image file to flash",
				Action: func(ctx *cli.Context, p cli.Path) error {
					stat, err := os.Stat(p)
					if err != nil {
						if errors.Is(err, os.ErrNotExist) {
							return trace.Wrap(err, "the file to flash %q does not exist", p)
						}

						return trace.Wrap(err, "failed to get filesystem info for image file %q", p)
					}

					if !stat.Mode().Type().IsRegular() {
						return trace.BadParameter("the image file path must point to a regular file")
					}

					return nil
				},
			},
		),
		Action: func(ctx *cli.Context) error {
			return trace.Errorf("this subcommand is not currently supported")
		},
	}
}

func getNodeSelectionFlags() ([]cli.Flag, *[]uint) {
	providedNodeNumbers := []uint{}

	return []cli.Flag{
		&cli.BoolFlag{
			Name:    "all",
			Aliases: []string{"a"},
			Usage:   "set to select all nodes",
			Action: func(ctx *cli.Context, b bool) error {
				if len(providedNodeNumbers) > 0 {
					// This could occur if both the "all" flag and specific node flags are set
					return nil
				}

				for nodeNumber := uint(1); nodeNumber <= node.NodeCount; nodeNumber++ {
					providedNodeNumbers = append(providedNodeNumbers, nodeNumber)
				}

				return nil
			},
		},
		&cli.UintFlag{
			Name:    "node",
			Aliases: []string{"n"},
			Usage:   "set to select specific nodes",
			Action: func(ctx *cli.Context, nodeNumber uint) error {
				if len(providedNodeNumbers) > 0 {
					// This could occur if both the "all" flag and specific node flags are set
					return nil
				}

				if nodeNumber > node.NodeCount {
					return trace.BadParameter("the node numbers must be less than %d", node.NodeCount)
				}
				if nodeNumber < 1 {
					return trace.BadParameter("the node numbers must be greater than %d", 1)
				}

				providedNodeNumbers = append(providedNodeNumbers, nodeNumber)
				return nil
			},
		},
	}, &providedNodeNumbers
}

func consoleCommand() *cli.Command {
	providedNodeNumber := uint(0)
	action := func(ctx *cli.Context) error {
		nodes, err := node.GetNodes()
		if err != nil {
			return trace.Wrap(err, "failed to get all nodes")
		}
		node := nodes[providedNodeNumber]

		picocomPath, err := exec.LookPath("picocom")
		if err != nil {
			return trace.Wrap(err, "failed to file picocom program path")
		}

		// This does not return at all upon success
		err = syscall.Exec(picocomPath, []string{"picocom", "--baud", fmt.Sprintf("%d", node.BaudRate()), node.TTYDevicePath()}, os.Environ())
		return trace.Wrap(err, "failed to invoke picocom")
	}

	return &cli.Command{
		Name:        "console",
		Aliases:     []string{"c"},
		Description: "start a serial console session with selected node",
		Flags: []cli.Flag{
			&cli.UintFlag{
				Name:    "node",
				Aliases: []string{"n"},
				Usage:   "set to select a specific node",
				Action: func(ctx *cli.Context, nodeNumber uint) error {
					if nodeNumber > node.NodeCount {
						return trace.BadParameter("the node number must be less than %d", node.NodeCount)
					}
					if nodeNumber < 1 {
						return trace.BadParameter("the node number must be greater than %d", 1)
					}

					providedNodeNumber = nodeNumber
					return nil
				},
				Required: true,
			},
		},
		Action: action,
	}
}

func rescanCommand() *cli.Command {
	return &cli.Command{
		Name:        "rescan",
		Aliases:     []string{"s"},
		Description: "triggers a PCIe rescan to look for newly attached nodes",
		Action: func(ctx *cli.Context) error {
			err := node.NewPCIeController().RescanAll()
			return trace.Wrap(err, "failed to rescan all PCIe nodes")
		},
	}
}

func exitHandler(err error) {
	if err == nil {
		os.Exit(0)
	}

	log.New(os.Stderr, "", 0).Fatal(err.Error())
}
