package main

import (
	"fmt"
	"math"
)

// c prefix stands for color (to avoid "x redeclared in this block")
//
// l (after c) stands for light. for example, lblue = light blue
//
// color is taken from python colorama module
const (
	cblue    = "\x1b[34m"
	cgreen   = "\x1b[32m"
	cred     = "\x1b[31m"
	cyellow  = "\x1b[33m"
	clblue   = "\x1b[94m"
	clgreen  = "\x1b[92m"
	clred    = "\x1b[91m"
	clyellow = "\x1b[93m"
	creset   = "\x1b[39m"
)

var colorEnabled = true

// info
func I(text string) string {
	return blue("[*] ") + lblue(text)
}

// ok
func OK(text string) string {
	return green("[+] ") + lgreen(text)
}

// error
func E(text string) string {
	return red("[!] ") + lred(text)
}

// warning
func W(text string) string {
	return yellow("[!] ") + lyellow(text)
}

// input
func INP(text string) string {
	return green("[?] ") + lgreen(text)
}

func blue(text string) string {
	if colorEnabled {
		return cblue + text + creset
	}
	return text
}

func green(text string) string {
	if colorEnabled {
		return cgreen + text + creset
	}
	return text
}

func red(text string) string {
	if colorEnabled {
		return cred + text + creset
	}
	return text
}

func yellow(text string) string {
	if colorEnabled {
		return cyellow + text + creset
	}
	return text
}

func lblue(text string) string {
	if colorEnabled {
		return clblue + text + creset
	}
	return text
}

func lgreen(text string) string {
	if colorEnabled {
		return clgreen + text + creset
	}
	return text
}

func lred(text string) string {
	if colorEnabled {
		return clred + text + creset
	}
	return text
}

func lyellow(text string) string {
	if colorEnabled {
		return clyellow + text + creset
	}
	return text
}

// convert rgb to ansi256
func rgbToAnsi256(r, g, b byte) byte {
	if r == g && g == b {
		if r < 8 {
			return 16
		}
		if r > 248 {
			return 231
		}
		return byte(math.Round(((float64(r)-8)/247)*24)) + 232
	}

	return byte(16+(36*math.Round(float64(r)/255*5))) +
		byte((6 * math.Round(float64(g)/255*5))) +
		byte(math.Round(float64(b)/255*5))
}

// foreground color
func fg(color byte, text string) string {
	return fmt.Sprint("\033[38;5;", color, "m", text, "\033[0;00m")
}

// background color
func bg(color byte, text string) string {
	return fmt.Sprint("\033[48;5;", color, "m", text, "\033[0;00m\n")
}

// foreground color rgb
func fgr(r, g, b byte, text string) string {
	return fg(rgbToAnsi256(r, g, b), text)
}

// background color rgb
func bgr(r, g, b byte, text string) string {
	return bg(rgbToAnsi256(r, g, b), text)
}
