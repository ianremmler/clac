package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/ianremmler/clac"
	"golang.org/x/crypto/ssh/terminal"
	"robpike.io/ivy/value"
)

const usageStr = `usage:

Interactive:  clac [-i <input>]
Command line: clac [-x | -p <precision>] <input>

Command line mode requires input from arguments (without -i) and/or stdin.
`

var (
	// flags
	doInitStack      = false
	doHexOut         = false
	cliPrec     uint = 12

	trm         *terminal.Terminal
	oldTrmState *terminal.State
	lastErr     error
	cl          = clac.New()
)

var cmdMap = map[string]func() error{
	"neg":    cl.Neg,
	"n":      cl.Neg,
	"abs":    cl.Abs,
	"a":      cl.Abs,
	"inv":    cl.Inv,
	"i":      cl.Inv,
	"+":      cl.Add,
	"-":      cl.Sub,
	"*":      cl.Mul,
	"x":      cl.Mul,
	"/":      cl.Div,
	"div":    cl.IntDiv,
	"%":      cl.Mod,
	"exp":    cl.Exp,
	"^":      cl.Pow,
	"2^":     cl.Pow2,
	"10^":    cl.Pow10,
	"logn":   cl.LogN,
	"ln":     cl.Ln,
	"log":    cl.Log,
	"lg":     cl.Lg,
	"sqrt":   cl.Sqrt,
	"!":      cl.Factorial,
	"comb":   cl.Comb,
	"perm":   cl.Perm,
	"sin":    cl.Sin,
	"cos":    cl.Cos,
	"tan":    cl.Tan,
	"asin":   cl.Asin,
	"acos":   cl.Acos,
	"atan":   cl.Atan,
	"atan2":  cl.Atan2,
	"dtor":   cl.DegToRad,
	"rtod":   cl.RadToDeg,
	"rtop":   cl.RectToPolar,
	"ptor":   cl.PolarToRect,
	"floor":  cl.Floor,
	"ceil":   cl.Ceil,
	"trunc":  cl.Trunc,
	"and":    cl.And,
	"or":     cl.Or,
	"xor":    cl.Xor,
	"not":    cl.Not,
	"andn":   cl.AndN,
	"orn":    cl.OrN,
	"xorn":   cl.XorN,
	"sum":    cl.Sum,
	"avg":    cl.Avg,
	"drop":   cl.Drop,
	"k":      cl.Drop,
	"dropn":  cl.DropN,
	"dropr":  cl.DropR,
	"dup":    cl.Dup,
	"d":      cl.Dup,
	"dupn":   cl.DupN,
	"dupr":   cl.DupR,
	"pick":   cl.Pick,
	"p":      cl.Pick,
	"swap":   cl.Swap,
	"s":      cl.Swap,
	"depth":  cl.Depth,
	"min":    cl.Min,
	"max":    cl.Max,
	"minn":   cl.MinN,
	"maxn":   cl.MaxN,
	"rot":    cl.Rot,
	"rotr":   cl.RotR,
	"unrot":  cl.Unrot,
	"unrotr": cl.UnrotR,
	"mag":    cl.Mag,
	"hyp":    cl.Hypot,
	"dot":    cl.Dot,
	"dot3":   cl.Dot3,
	"cross":  cl.Cross,
	"pi":     constant(clac.Pi),
	"e":      constant(clac.E),
	"phi":    constant(clac.Phi),
}

var interactiveCmdMap = map[string]func() error{
	"undo":  cl.Undo,
	"u":     cl.Undo,
	"redo":  cl.Redo,
	"r":     cl.Redo,
	"clear": cl.Clear,
	"c":     cl.Clear,
	"reset": cl.Reset,
	"quit":  exit,
	"q":     exit,
}

type term struct {
	io.Reader
	io.Writer
}

func init() {
	log.SetFlags(0)
	log.SetPrefix("clac: ")
	flag.BoolVar(&doHexOut, "x", doHexOut,
		"Command line mode: hexidecimal output")
	flag.UintVar(&cliPrec, "p", cliPrec,
		"Command line mode: output precision")
	flag.BoolVar(&doInitStack, "i", doInitStack,
		"Initialize with input from command line arguments")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, usageStr)
		flag.PrintDefaults()
	}
}

func main() {
	flag.Parse()
	if !processCmdLine() {
		printCmdLineStack(cl.Stack())
		os.Exit(0)
	}

	interactiveSetup()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	go func() {
		<-sigChan
		exit()
	}()

	repl()
}

func constant(v value.Value) func() error {
	return func() error { return cl.Push(v) }
}

func interactiveSetup() {
	if !terminal.IsTerminal(syscall.Stdin) {
		log.Fatalln("this doesn't look like an interactive terminal")
	}
	var err error
	oldTrmState, err = terminal.MakeRaw(syscall.Stdin)
	if err != nil {
		log.Fatalln(err)
	}
	trm = terminal.NewTerminal(term{os.Stdin, os.Stdout}, "")

	for cmd, fn := range interactiveCmdMap {
		cmdMap[cmd] = fn
	}
	interactiveCmdMap = nil
}

func repl() {
	for {
		printStack(cl.Stack())
		input, err := trm.ReadLine()
		lastErr = nil
		if err == io.EOF {
			exit()
		}
		if err != nil {
			continue
		}
		processInput(input, true)
	}
}

func processCmdLine() bool {
	input := ""
	if stat, err := os.Stdin.Stat(); err == nil && stat.Mode()&os.ModeNamedPipe != 0 {
		if pipeInput, err := ioutil.ReadAll(os.Stdin); err == nil {
			input = string(pipeInput)
		}
	}
	if len(flag.Args()) > 0 {
		input += " " + strings.Join(flag.Args(), " ")
	}
	isInteractive := doInitStack || (input == "")
	cl.EnableHistory(isInteractive)
	processInput(string(input), false)
	return isInteractive
}

func printCmdLineStack(stack clac.Stack) {
	if doHexOut {
		clac.SetFormat("%#x")
	} else {
		clac.SetFormat(fmt.Sprintf("%%.%dg", cliPrec))
	}
	for i := range stack {
		val := stack[len(stack)-i-1]
		var err error
		if doHexOut {
			val, err = clac.Trunc(val)
		}
		if err != nil {
			fmt.Print("error")
		} else {
			fmt.Print(clac.Sprint(val))
		}
		if i < len(stack)-1 {
			fmt.Print(" ")
		}
	}
	fmt.Println()
}

func exit() error {
	terminal.Restore(syscall.Stdin, oldTrmState)
	fmt.Println()
	os.Exit(0)
	return nil
}

func processInput(input string, isInteractive bool) {
	errorHandler := func(err error) { lastErr = err }
	if !isInteractive {
		errorHandler = func(err error) { log.Println(err) }
	}
	scanner := bufio.NewScanner(strings.NewReader(input))
	scanner.Split(bufio.ScanWords)
	for scanner.Scan() {
		tok := scanner.Text()
		if num, err := clac.ParseNum(tok); err == nil {
			if err = cl.Exec(func() error { return cl.Push(num) }); err != nil {
				errorHandler(fmt.Errorf("push: %s", err))
			}
			continue
		}
		if cmd, ok := cmdMap[tok]; ok {
			if err := cl.Exec(cmd); err != nil {
				errorHandler(fmt.Errorf("%s: %s", tok, err))
			}
			continue
		}
		errorHandler(fmt.Errorf("%s: invalid input", tok))
	}
	if err := scanner.Err(); err != nil {
		errorHandler(err)
	}
}

func printStack(stack clac.Stack) {
	cols, rows, err := terminal.GetSize(syscall.Stdout)
	if err != nil {
		rows = len(stack) + 1
	}
	// ensure sane width
	if cols < 20 {
		cols = 20
	}
	clearScreen()

	dataCols := cols - 4
	hexCols := dataCols / 2
	floatCols := dataCols - hexCols
	floatFmt := fmt.Sprintf("%%%d.%dg", floatCols-1, floatCols-8)
	hexFmt := fmt.Sprintf("%%#%dx", hexCols-3)
	for i := rows - 3; i >= 0; i-- {
		line := fmt.Sprintf("%02d:", i)
		if i < len(stack) {
			clac.SetFormat(floatFmt)
			line += fmt.Sprintf(fmt.Sprintf(" %%%ds", floatCols), clac.Sprint(stack[i]))
			if val, err := clac.Trunc(stack[i]); err == nil {
				clac.SetFormat(hexFmt)
				hexStr := fmt.Sprintf(fmt.Sprintf(" %%%ds", hexCols-1), clac.Sprint(val))
				if len(hexStr) > hexCols {
					hexStr = hexStr[:hexCols-1] + "…"
				}
				line += hexStr
			}
		}
		fmt.Println(line + "\r")
	}
	if lastErr == nil {
		fmt.Println(strings.Repeat("-", cols))
	} else {
		fmt.Println("Error:", lastErr)
	}
	fmt.Print("\r")
}

func clearScreen() {
	fmt.Print("\033[2J\033[H")
}

func waitKey() {
	bufio.NewReader(os.Stdin).ReadByte()
}
