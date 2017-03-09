package tui

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

//Prompt asks the user for string variable, showing 'ask' and assigning 'dft' if user answered nothing
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

// SelectorItem is an interface for anything that can be named
type SelectorItem interface {
	Name() string
}

// NewItem inserts SelectorItem to Selector
/*func (sel Selector) NewItem(i SelectorItem) {
	sel = append(sel, struct {
		num  int
		name string
		item SelectorItem
	}{len(sel), i.Name(), i})
}*/

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

/*func slicepop(in []interface{}, i int) (out []interface{}) {
	for d:=0; d<len(in); d++ {
		if d!=i {
			append(out,in[d])
		}
	}
}*/

// SingleChoice shows Selector and returns one selected item
/*func (sel Selector) SingleChoice() SelectorItem {
	for _, s := range sel {
		fmt.Println(strconv.Itoa(s.num) + " : " + s.name)
	}
	n:= func() int {
		for {
			check := Prompt("Your choice : ", "")
			choice, err := strconv.Atoi(check)
			if err != nil {
				fmt.Println("Bad choice : " + check)
			} else {
				return choice	
			}
		}
	}()
	return sel[n].item
}*/
