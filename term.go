package main

import (
	"fmt"
	"image/color"
	"os"
	"syscall"
	"unsafe"
)

const TERM_COLOR_RESET = "\x1b[0m"

func termWidth() (int, error) {
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
		return 0, errno
	}

	return int(ws.Col), nil
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
