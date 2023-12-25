package node

import (
	"time"

	"github.com/gravitational/trace"
)

type IPowerControl interface {
	PowerOn() bool
	PowerOff() bool
	Reboot() bool
}

func (n *Node) PowerOn() error {
	for _, outputPin := range n.outputPins {
		err := outputPin.SetHigh()
		if err != nil {
			return trace.Wrap(err, "failed to setup output pin %d high", outputPin.Number())
		}
	}

	return nil
}

func (n *Node) PowerOff() error {
	for _, outputPin := range n.outputPins {
		err := outputPin.SetLow()
		if err != nil {
			return trace.Wrap(err, "failed to setup output pin %d low", outputPin.Number())
		}
	}

	return nil
}

func (n *Node) Reboot() error {
	err := n.PowerOff()
	if err != nil {
		return trace.Wrap(err, "failed to power off node")
	}

	// Wait 1 second, blocking the context, but allowing other gofuncs to continue
	select {
	case <-time.After(time.Second):
	}

	err = n.PowerOn()
	if err != nil {
		return trace.Wrap(err, "failed to power on node")
	}

	return nil
}
