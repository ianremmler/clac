package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"os"
	"os/signal"
	"sort"
	"strings"

	"github.com/ianremmler/clac"
	"github.com/peterh/liner"
)

var (
	lnr     *liner.State
	cl      = clac.New()
	cmdList = []string{}
	cmdMap  = map[string]func() error{
		"neg":    cl.Neg,
		"abs":    cl.Abs,
		"inv":    cl.Inv,
		"+":      cl.Add,
		"-":      cl.Sub,
		"*":      cl.Mul,
		"/":      cl.Div,
		"mod":    cl.Mod,
		"exp":    cl.Exp,
		"pow":    cl.Pow,
		"pow2":   cl.Pow2,
		"pow10":  cl.Pow10,
		"ln":     cl.Ln,
		"log":    cl.Log,
		"lg":     cl.Lg,
		"sqrt":   cl.Sqrt,
		"hypot":  cl.Hypot,
		"sin":    cl.Sin,
		"cos":    cl.Cos,
		"tan":    cl.Tan,
		"asin":   cl.Asin,
		"acos":   cl.Acos,
		"atan":   cl.Atan,
		"sinh":   cl.Sin,
		"cosh":   cl.Cos,
		"tanh":   cl.Tan,
		"asinh":  cl.Asin,
		"acosh":  cl.Acos,
		"atanh":  cl.Atan,
		"atan2":  cl.Atan2,
		"d2r":    cl.D2R,
		"r2d":    cl.R2D,
		"floor":  cl.Floor,
		"ceil":   cl.Ceil,
		"trunc":  cl.Trunc,
		"and":    cl.And,
		"or":     cl.Or,
		"xor":    cl.Xor,
		"not":    cl.Not,
		"clear":  cl.Clear,
		"drop":   cl.Drop,
		"dropn":  cl.Dropn,
		"dropr":  cl.Dropr,
		"dup":    cl.Dup,
		"dupn":   cl.Dupn,
		"dupr":   cl.Dupr,
		"pick":   cl.Pick,
		"swap":   cl.Swap,
		"undo":   cl.Undo,
		"redo":   cl.Redo,
		"rot":    func() error { return cl.Rot(true) },
		"rotr":   func() error { return cl.Rotr(true) },
		"unrot":  func() error { return cl.Rot(false) },
		"unrotr": func() error { return cl.Rotr(false) },
		"pi":     func() error { return cl.Push(math.Pi) },
		"e":      func() error { return cl.Push(math.E) },
		"phi":    func() error { return cl.Push(math.Phi) },
		"quit":   func() error { exit(); return nil },
		"help":   func() error { help(); return nil },
	}
)

func init() {
	log.SetFlags(0)
	log.SetPrefix("Error: ")
	for cmd := range cmdMap {
		cmdList = append(cmdList, cmd)
	}
	sort.Strings(cmdList)
}

func main() {
	if processPipe() {
		fmt.Println(strings.Trim(fmt.Sprint(cl.Stack()), "[]"))
		os.Exit(0)
	}

	lnr = liner.NewLiner()
	lnr.SetWordCompleter(complete)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	go func() {
		<-sigChan
		exit()
	}()

	for {
		printStack(cl.Stack())
		input, err := lnr.Prompt("> ")
		if err == io.EOF {
			exit()
		}
		if err != nil {
			continue
		}
		lnr.AppendHistory(input)
		parseInput(input)
		fmt.Println()
	}
}

func processPipe() bool {
	if stat, err := os.Stdin.Stat(); err == nil && stat.Mode()&os.ModeNamedPipe != 0 {
		if input, err := ioutil.ReadAll(os.Stdin); err == nil {
			parseInput(string(input))
			return true
		}
	}
	return false
}

func exit() {
	fmt.Println()
	lnr.Close()
	os.Exit(0)
}

func help() {
	fmt.Println()
	for i := range cmdList {
		fmt.Printf("%-8s", cmdList[i])
		if (i+1)%5 == 0 {
			fmt.Println()
		}
	}
	if len(cmdList)%5 != 0 {
		fmt.Println()
	}
	fmt.Print("\n[Press any key to continue]")
	bufio.NewReader(os.Stdin).ReadByte()
	fmt.Println()
}

func parseInput(input string) {
	cmdReader := strings.NewReader(input)
	for {
		tok := ""
		if _, err := fmt.Fscan(cmdReader, &tok); err != nil {
			if err != io.EOF {
				log.Println(err)
			}
			break
		}
		num, err := clac.ParseNum(tok)
		if err == nil {
			if err = cl.Push(num); err != nil {
				log.Println(tok+":", err)
			}
			continue
		}
		if cmd, ok := cmdMap[tok]; ok {
			if err = cmd(); err != nil {
				log.Println(tok+":", err)
			}
			continue
		}
		log.Println(tok + ": invalid input")
	}
}

func complete(in string, pos int) (string, []string, string) {
	start := strings.LastIndexAny(in[:pos], " \t") + 1
	end := len(in)
	if idx := strings.IndexAny(in[pos:], " \t"); idx >= 0 {
		end = pos + idx
	}
	head, word, tail := in[:start], in[start:end], in[end:]
	cmds := []string{}
	for i := range cmdList {
		if strings.HasPrefix(cmdList[i], word) {
			cmds = append(cmds, cmdList[i])
		}
	}
	return head, cmds, tail
}

func printStack(stack clac.Stack) {
	for i := len(stack) - 1; i >= 0; i-- {
		fmt.Printf("%2d: %16.10g", i, stack[i])
		if math.Abs(stack[i]) < math.MaxInt64 {
			fmt.Printf(" %#19x", int64(stack[i]))
		}
		fmt.Println()
	}
	fmt.Println(strings.Repeat("-", 40))
}