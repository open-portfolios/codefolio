//go:build windows

package platform

import (
	"os"

	"golang.org/x/sys/windows"
)

type State struct {
	inMode  uint32
	outMode uint32
}

func EnterRawMode() (*State, error) {
	in := windows.Handle(os.Stdin.Fd())
	out := windows.Handle(os.Stdout.Fd())

	var inMode uint32
	if err := windows.GetConsoleMode(in, &inMode); err != nil {
		return nil, err
	}

	var outMode uint32
	if err := windows.GetConsoleMode(out, &outMode); err != nil {
		return nil, err
	}

	rawIn := inMode
	rawIn &^= windows.ENABLE_ECHO_INPUT | windows.ENABLE_LINE_INPUT | windows.ENABLE_PROCESSED_INPUT
	rawIn |= windows.ENABLE_VIRTUAL_TERMINAL_INPUT
	if err := windows.SetConsoleMode(in, rawIn); err != nil {
		return nil, err
	}

	rawOut := outMode | windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING
	if err := windows.SetConsoleMode(out, rawOut); err != nil {
		_ = windows.SetConsoleMode(in, inMode)
		return nil, err
	}

	return &State{inMode: inMode, outMode: outMode}, nil
}

func ExitRawMode(state *State) error {
	if state == nil {
		return nil
	}
	in := windows.Handle(os.Stdin.Fd())
	out := windows.Handle(os.Stdout.Fd())
	if err := windows.SetConsoleMode(in, state.inMode); err != nil {
		return err
	}
	return windows.SetConsoleMode(out, state.outMode)
}
