package gpio

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gravitational/trace"
)

const sysfsGpioRoot = "/sys/class/gpio/"

type IPin interface {
	Number() uint
}

type pin struct {
	number   uint
	gpioRoot string // This is for testing purposes
}

func (p *pin) Number() uint {
	return p.number
}

func newPin(number uint, isInput bool) (*pin, error) {
	p := &pin{
		number: number,
	}

	isSetup, err := p.isAlreadySetup()
	if err != nil {
		return nil, trace.Wrap(err, "failed to determine if pin %d is already setup", p.Number())
	}
	if isSetup {
		return p, nil
	}

	err = p.userspaceExportPin()
	if err != nil {
		return nil, trace.Wrap(err, "failed to export pin %d for userspace access", p.Number())
	}

	err = p.setPinDirection(isInput)
	if err != nil {
		return nil, trace.Wrap(err, "failed to set pin direction")
	}

	return p, nil
}

func (p *pin) isAlreadySetup() (bool, error) {
	_, err := os.Stat(p.getSysfsDirectory())
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}

	return false, trace.Wrap(err, "failed to check if pin GPIO sysfs directory exists")
}

func (p *pin) userspaceExportPin() error {
	// File mode does not matter here as the file will always exist
	err := os.WriteFile(filepath.Join(sysfsGpioRoot, "export"), []byte(fmt.Sprintf("%d\n", p.Number())), 0)
	if err != nil {
		return trace.Wrap(err, "failed to userspace export GPIO pin via the sysfs interface")
	}

	return nil
}

func (p *pin) setPinDirection(setInput bool) error {
	filePath := filepath.Join(p.getSysfsDirectory(), "direction")

	value := "in"
	if !setInput {
		value = "out"
	}

	err := os.WriteFile(filePath, []byte(value), 0)
	if err != nil {
		return trace.Wrap(err, "failed to set GPIO pin direction via sysfs interface")
	}

	return nil
}

func (p *pin) getSysfsDirectory() string {
	return filepath.Join(sysfsGpioRoot, fmt.Sprintf("gpio%d", p.Number()))
}

type IOutputPin interface {
	IPin
	SetHigh() error
	SetLow() error
}

type outputPin struct {
	*pin
}

func newOutputPin(number uint) (*outputPin, error) {
	p, err := newPin(number, false)
	if err != nil {
		return nil, trace.Wrap(err, "failed to set the output pin")
	}

	return &outputPin{
		pin: p,
	}, nil
}

func (op *outputPin) SetHigh() error {
	return op.setOutputLevel(true)
}

func (op *outputPin) SetLow() error {
	return op.setOutputLevel(false)
}

func (op *outputPin) setOutputLevel(setHigh bool) error {
	value := "1"
	if !setHigh {
		value = "0"
	}
	value += "\n"

	// File mode does not matter here as the file will always exist
	err := os.WriteFile(filepath.Join(op.getSysfsDirectory(), "value"), []byte(value), 0)
	if err != nil {
		level := "high"
		if !setHigh {
			level = "low"
		}

		return trace.Wrap(err, "failed to set GPIO pin %d output level %s via the sysfs interface", op.Number(), level)
	}

	return nil
}

type IInputPin interface {
	IPin
	GetValue() (uint, error)
}

type inputPin struct {
	*pin
}

func newInputPin(number uint) (*inputPin, error) {
	p, err := newPin(number, true)
	if err != nil {
		return nil, trace.Wrap(err, "failed to set the input pin")
	}

	return &inputPin{
		pin: p,
	}, nil
}

func (ip *inputPin) GetValue() (uint, error) {
	data, err := os.ReadFile(filepath.Join(ip.getSysfsDirectory(), "value"))
	if err != nil {
		return 0, trace.Wrap(err, "failed to read GPIO pin %d output level via the sysfs interface", ip.Number())
	}

	fileString := string(data)
	val, err := strconv.Atoi(fileString)
	if err != nil {
		return 0, trace.Wrap(err, "failed to parse file data %q into an integer", fileString)
	}

	if val != 0 && val != 1 {
		return 0, trace.Errorf("value was expected to be 0 or 1, got %d", val)
	}

	return uint(val), nil
}

type IGPIOController interface {
	NewOutputPin(number uint) (IOutputPin, error)
	NewInputPin(number uint) (IInputPin, error)
}

type GPIOController struct{}

func (gpioc *GPIOController) NewOutputPin(number uint) (IOutputPin, error) {
	return newOutputPin(number)
}

func (gpioc *GPIOController) NewInputPin(number uint) (IInputPin, error) {
	return newInputPin(number)
}
