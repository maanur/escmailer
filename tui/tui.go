package tui

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

// Prompt asks the user for string variable, showing 'ask' and assigning 'dft' if user answered nothing
func Prompt(ask string, dft string) (output string) {
	// в библитеку бы
	consolereader := bufio.NewReader(os.Stdin)
	fmt.Println(ask)
	rn, err := consolereader.ReadBytes('\r') // this will prompt the user for input
	if err != nil {
		log.Fatal(err)
	}
	output = string(rn[:len(rn)-1])
	if output == "" {
		return dft
	}
	return output
}

// PromptYN asks the user for bool variable
func PromptYN(ask string, dft bool) bool {
	l:="n"
	if dft {
		l="y"
	}
	for {
		r:= Prompt(ask+" (y/n, default:"+l+")", l)
		if r=="y" {
			return true
		}
		if r=="n" {
			return false
		}
		fmt.Println("Некорректный ответ: "+r)
	}
}

// SelectorItem is an interface for anything that can be named
type SelectorItem interface {
	Name() string
}

// NewItem inserts SelectorItem to Selector

//MultiChoice returns a slice of SelectorItem, picked by user from original slice
func MultiChoice(sel []SelectorItem) []SelectorItem {
	for i:=0; i<len(sel); i++ {
		fmt.Println(strconv.Itoa(i) + " : " + sel[i].Name())
	}
	var choice []int
	for {
		repeat := false
		for _, check := range strings.Fields(Prompt("Your choice (ex.: 1 2 3) :", "")) {
			repeat = false
			o, err := strconv.Atoi(check)
			if err != nil {
				fmt.Println("Bad choice : " + check)
				repeat = true
			} else {
				if o >= len(sel) || o < 0 {
					fmt.Println("Out of range : " + check)
					repeat = true
				} else {
					choice = append(choice, o)
				}
			}
		}
		if !repeat {
			break
		}
	}
	b := sel[:0]
	for x:=0 ; x<=len(sel); x++ {
		for _, n:=range choice {
			if x==n {
				b=append(b,sel[x])
			}
		}
    }
	return b
}
