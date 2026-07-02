//go:build !windows

package tux

import (
	"os"

	"golang.org/x/sys/unix"
)

type terminalState struct {
	termios unix.Termios
}

func enterRawMode() (*terminalState, error) {
	fd := int(os.Stdin.Fd())
	termios, err := unix.IoctlGetTermios(fd, unix.TCGETS)
	if err != nil {
		return nil, err
	}

	raw := *termios
	raw.Iflag &^= unix.BRKINT | unix.ICRNL | unix.INPCK | unix.ISTRIP | unix.IXON
	raw.Oflag &^= unix.OPOST
	raw.Cflag |= unix.CS8
	raw.Lflag &^= unix.ECHO | unix.ICANON | unix.IEXTEN | unix.ISIG
	raw.Cc[unix.VMIN] = 1
	raw.Cc[unix.VTIME] = 0

	if err := unix.IoctlSetTermios(fd, unix.TCSETS, &raw); err != nil {
		return nil, err
	}
	return &terminalState{termios: *termios}, nil
}

func exitRawMode(state *terminalState) error {
	if state == nil {
		return nil
	}
	return unix.IoctlSetTermios(int(os.Stdin.Fd()), unix.TCSETS, &state.termios)
}
