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
	"sort"
	"strings"
	"syscall"

	"github.com/ianremmler/clac"
	"golang.org/x/crypto/ssh/terminal"
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
	cliPrec     uint = 5

	trm         *terminal.Terminal
	oldTrmState *terminal.State
	lastErr     error
	cl          = clac.New()
	cmdList     = []string{}
)

var cmdMap = map[string]func() error{
	"neg":    cl.Neg,
	"abs":    cl.Abs,
	"inv":    cl.Inv,
	"+":      cl.Add,
	"-":      cl.Sub,
	"*":      cl.Mul,
	"/":      cl.Div,
	"div":    cl.IntDiv,
	"mod":    cl.Mod,
	"exp":    cl.Exp,
	"^":      cl.Pow,
	"2^":     cl.Pow2,
	"10^":    cl.Pow10,
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
	"clear":  cl.Clear,
	"drop":   cl.Drop,
	"dropn":  cl.DropN,
	"dropr":  cl.DropR,
	"dup":    cl.Dup,
	"dupn":   cl.DupN,
	"dupr":   cl.DupR,
	"pick":   cl.Pick,
	"swap":   cl.Swap,
	"depth":  cl.Depth,
	"undo":   cl.Undo,
	"redo":   cl.Redo,
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
	"pi":     func() error { return cl.Push(clac.Pi) },
	"e":      func() error { return cl.Push(clac.E) },
	"reset":  func() error { return cl.Reset() },
	"quit":   func() error { exit(); return nil },
	"help":   func() error { help(); return nil },
}

type term struct {
	io.Reader
	io.Writer
}

func init() {
	log.SetFlags(0)
	log.SetPrefix("clac: ")
	for cmd := range cmdMap {
		cmdList = append(cmdList, cmd)
	}
	sort.Strings(cmdList)
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
	if processCmdLine() {
		printCmdLineStack(cl.Stack())
		os.Exit(0)
	}
	if !terminal.IsTerminal(syscall.Stdin) {
		log.Fatalln("this doesn't look like an interactive terminal")
	}
	var err error
	oldTrmState, err = terminal.MakeRaw(syscall.Stdin)
	if err != nil {
		log.Fatalln(err)
	}
	trm = terminal.NewTerminal(term{os.Stdin, os.Stdout}, "")
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	go func() {
		<-sigChan
		exit()
	}()

	repl()
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
		parseInput(input, func(err error) { lastErr = err })
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
	if input != "" {
		parseInput(string(input), func(err error) { log.Println(err) })
		return !doInitStack
	}
	return false
}

func printCmdLineStack(stack clac.Stack) {
	if doHexOut {
		clac.SetFormat("%#x")
	} else {
		clac.SetFormat(fmt.Sprintf("%%.%dg", cliPrec))
	}
	for i := range stack {
		val := stack[len(stack)-i-1]
		if doHexOut {
			var err error
			if val, err = clac.Trunc(val); err != nil {
				fmt.Print("error")
				continue
			}
		}
		fmt.Print(val)
		if i < len(stack)-1 {
			fmt.Print(" ")
		}
	}
	fmt.Println()
}

func exit() {
	fmt.Println()
	terminal.Restore(syscall.Stdin, oldTrmState)
	os.Exit(0)
}

func help() {
	clearScreen()
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
	waitKey()
}

func parseInput(input string, errorHandler func(err error)) {
	cmdReader := strings.NewReader(input)
	for {
		tok := ""
		if _, err := fmt.Fscan(cmdReader, &tok); err != nil {
			if err != io.EOF {
				errorHandler(err)
			}
			break
		}
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
			line += fmt.Sprintf(fmt.Sprintf(" %%%ds", floatCols), stack[i])
			if val, err := clac.Trunc(stack[i]); err == nil {
				clac.SetFormat(hexFmt)
				hexStr := fmt.Sprintf(fmt.Sprintf(" %%%ds", hexCols-1), val)
				if len(hexStr) > hexCols {
					hexStr = hexStr[:hexCols-1] + "…"
				}
				line += hexStr
			}
		}
		fmt.Println(line)
	}
	if lastErr == nil {
		fmt.Println(strings.Repeat("-", cols))
	} else {
		fmt.Println("Error:", lastErr)
	}
}

func clearScreen() {
	fmt.Print("\033[2J\033[H")
}

func waitKey() {
	bufio.NewReader(os.Stdin).ReadByte()
}
