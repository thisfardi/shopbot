package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func input(prompt string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(INP(prompt))
	text, _ := reader.ReadString('\n')
	return strings.TrimSpace(text)
}

func iinput(prompt, errmsg string) int {
	for {
		inp := input(prompt)
		res, err := strconv.Atoi(inp)
		if err != nil {
			fmt.Println(E(errmsg))
		} else {
			return res
		}
	}
}
