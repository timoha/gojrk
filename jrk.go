package gojrk

import (
	"io"
	"os"
	"syscall"

	"github.com/pkg/term/termios"
)

// JRK is used communicate with jrk USB motor controller
type JRK struct {
	io.ReadWriteCloser
}

func (jrk *JRK) variable(cmd byte) (int, error) {
	if _, err := jrk.Write([]byte{cmd}); err != nil {
		return -1, err
	}

	var response [2]byte
	if _, err := jrk.Read(response[:]); err != nil {
		return -1, err
	}

	return int(response[0]) + 256*int(response[1]), nil
}

// Feedback returns current value of the jrk's Feeback variable (0-4095).
func (jrk *JRK) Feedback() (int, error) {
	return jrk.variable(0xA5)
}

// Feedback returns current value of the jrk's Target variable (0-4095).
func (jrk *JRK) Target() (int, error) {
	return jrk.variable(0xA3)
}

// SetTarget sets new Target value (0-4095).
func (jrk *JRK) SetTarget(v int) error {
	_, err := jrk.Write([]byte{byte(0xC0 + v&0x1F), byte((v >> 5) & 0x7F)})
	// TODO: read again for errors
	return err
}

// NewJRK connects to jrk USB motor contoller.
func NewJRK(path string) (*JRK, error) {
	f, err := os.OpenFile(path, os.O_RDWR|syscall.O_NOCTTY, 0)
	if err != nil {
		return nil, err
	}

	var o syscall.Termios
	if err := termios.Tcgetattr(f.Fd(), &o); err != nil {
		return nil, err
	}

	o.Lflag &^= syscall.ECHO | syscall.ECHONL | syscall.ICANON | syscall.ISIG | syscall.IEXTEN
	o.Oflag &^= syscall.ONLCR | syscall.OCRNL

	if err := termios.Tcsetattr(f.Fd(), termios.TCSANOW, &o); err != nil {
		return nil, err
	}

	return &JRK{ReadWriteCloser: f}, nil
}
