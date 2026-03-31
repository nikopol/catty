package main

import (
	"fmt"
	"image/color"
	"os"
	"syscall"
	"unsafe"
)

const TERM_COLOR_RESET = "\x1b[0m"

func termSize() (int, int, error) {
	type winsize struct {
		Row    uint16
		Col    uint16
		Xpixel uint16
		Ypixel uint16
	}

	ws := &winsize{}
	_, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL,
		os.Stdout.Fd(),
		uintptr(syscall.TIOCGWINSZ),
		uintptr(unsafe.Pointer(ws)),
	)
	if errno != 0 {
		return 0, 0, errno
	}

	return int(ws.Col), int(ws.Row), nil
}

func termFgColor(c color.Color) string {
	r, g, b, _ := c.RGBA()
	return fmt.Sprintf("\x1b[38;2;%d;%d;%dm", uint8(r>>8), uint8(g>>8), uint8(b>>8))
}

func termFgBgColor(fg color.Color, bg color.Color) string {
	fr, fgc, fb, _ := fg.RGBA()
	br, bgc, bb, _ := bg.RGBA()
	return fmt.Sprintf(
		"\x1b[38;2;%d;%d;%dm\x1b[48;2;%d;%d;%dm",
		uint8(fr>>8),
		uint8(fgc>>8),
		uint8(fb>>8),
		uint8(br>>8),
		uint8(bgc>>8),
		uint8(bb>>8),
	)
}

func enableRawInput() (func(), error) {
	fd := int(os.Stdin.Fd())

	termios, err := getTermios(fd)
	if err != nil {
		return nil, err
	}
	original := *termios
	raw := original
	raw.Lflag &^= syscall.ICANON | syscall.ECHO
	raw.Iflag &^= syscall.ICRNL | syscall.IXON
	raw.Cc[syscall.VMIN] = 1
	raw.Cc[syscall.VTIME] = 0

	if err := setTermios(fd, &raw); err != nil {
		return nil, err
	}

	return func() {
		_ = setTermios(fd, &original)
		fmt.Fprint(os.Stdout, "\x1b[0m\x1b[2J\x1b[H")
	}, nil
}

func getTermios(fd int) (*syscall.Termios, error) {
	termios := &syscall.Termios{}
	_, _, errno := syscall.Syscall6(
		syscall.SYS_IOCTL,
		uintptr(fd),
		uintptr(syscall.TCGETS),
		uintptr(unsafe.Pointer(termios)),
		0,
		0,
		0,
	)
	if errno != 0 {
		return nil, errno
	}
	return termios, nil
}

func setTermios(fd int, termios *syscall.Termios) error {
	_, _, errno := syscall.Syscall6(
		syscall.SYS_IOCTL,
		uintptr(fd),
		uintptr(syscall.TCSETS),
		uintptr(unsafe.Pointer(termios)),
		0,
		0,
		0,
	)
	if errno != 0 {
		return errno
	}
	return nil
}
