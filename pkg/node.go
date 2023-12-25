package node

import (
	"github.com/gravitational/trace"
	"github.com/solidDoWant/nodectl/internal/gpio"
)

const NodeCount = uint(4)

type INode interface {
	// Node funcs
	Number() uint

	// Power funcs
	PowerOn() error
	PowerOff() error
	Reboot() error

	// // Status funcs
	// IsNodeAttached() bool
	// IsNodePoweredOn() bool
	// IsNodePoweredOff() bool

	// Console funcs
	BaudRate() uint
	TTYDevicePath() string
}

type Node struct {
	number         uint
	gpioController gpio.IGPIOController
	outputPins     [4]gpio.IOutputPin
	inputPins      [1]gpio.IInputPin
}

func GetNodes() ([NodeCount]*Node, error) {
	controller := &gpio.GPIOController{}

	node1, err := newNode(1, [3]uint{0x1FC, 0x1F8, 0x1F3}, [1]uint{0x1F7}, controller)
	if err != nil {
		return [4]*Node{}, trace.Wrap(err, "failed to setup node 1")
	}

	node2, err := newNode(2, [3]uint{0x1FD, 0x1F9, 0x1F2}, [1]uint{0x1F6}, controller)
	if err != nil {
		return [4]*Node{}, trace.Wrap(err, "failed to setup node 2")
	}

	node3, err := newNode(3, [3]uint{0x1FF, 0x1FB, 0x1F0}, [1]uint{0x1F4}, controller)
	if err != nil {
		return [4]*Node{}, trace.Wrap(err, "failed to setup node 3")
	}

	node4, err := newNode(4, [3]uint{0x1FE, 0x1FA, 0x1F1}, [1]uint{0x1F5}, controller)
	if err != nil {
		return [4]*Node{}, trace.Wrap(err, "failed to setup node 4")
	}

	return [4]*Node{
		node1,
		node2,
		node3,
		node4,
	}, nil
}

func newNode(number uint, ouputPinNumbers [3]uint, inputPinNumbers [1]uint, gpioController gpio.IGPIOController) (*Node, error) {
	if number < 1 || number > NodeCount {
		return nil, trace.BadParameter("attempted to create a new node for out of range number %d, valid range 1-4", number)
	}

	n := &Node{
		number: number,
	}

	err := n.setupGpio(ouputPinNumbers, inputPinNumbers, gpioController)
	if err != nil {
		return nil, trace.Wrap(err, "failed to setup GPIO for node %d", number)
	}

	return n, nil
}

func (n *Node) setupGpio(ouputPinNumbers [3]uint, inputPinNumbers [1]uint, gpioController gpio.IGPIOController) error {
	for i, outputPinNumber := range ouputPinNumbers {
		outputPin, err := gpioController.NewOutputPin(outputPinNumber)
		if err != nil {
			return trace.Wrap(err, "failed to configure output pin %d (GPIO pin number %d)", i, outputPinNumber)
		}

		n.outputPins[i] = outputPin
	}

	for i, inputPinNumber := range inputPinNumbers {
		inputPin, err := gpioController.NewInputPin(inputPinNumber)
		if err != nil {
			return trace.Wrap(err, "failed to configure input pin %d (GPIO pin number %d)", i, inputPinNumber)
		}

		n.inputPins[i] = inputPin
	}

	return nil
}

func (n *Node) Number() uint {
	return n.number
}
