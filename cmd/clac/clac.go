package main

import (
	"bufio"
	"errors"
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
	"robpike.io/ivy/config"
	"robpike.io/ivy/value"
)

const usageStr = `usage:

Interactive:  clac [-i <input>]
Command line: [... |] clac [-x] [<input>]

Command line mode requires input from arguments (without -i) and/or stdin.
`

var (
	cfg         = &config.Config{}
	trm         *terminal.Terminal
	oldTrmState *terminal.State
	lastErr     error
	cl          = clac.New()
	doHexOut    = false
	doInitStack = false
	cmdList     = []string{}
	cmdMap      = map[string]func() error{
		"neg":  cl.Neg,
		"abs":  cl.Abs,
		"inv":  cl.Inv,
		"+":    cl.Add,
		"-":    cl.Sub,
		"*":    cl.Mul,
		"/":    cl.Div,
		"div":  cl.IntDiv,
		"mod":  cl.Mod,
		"exp":  cl.Exp,
		"^":    cl.Pow,
		"2^":   cl.Pow2,
		"10^":  cl.Pow10,
		"ln":   cl.Ln,
		"log":  cl.Log,
		"lg":   cl.Lg,
		"sqrt": cl.Sqrt,
		// 		"gamma":  cl.Gamma,
		"!":    cl.Factorial,
		"comb": cl.Comb,
		"perm": cl.Perm,
		"sin":  cl.Sin,
		"cos":  cl.Cos,
		"tan":  cl.Tan,
		"asin": cl.Asin,
		"acos": cl.Acos,
		"atan": cl.Atan,
		// 		"atan2":  cl.Atan2,
		"dtor": cl.DegToRad,
		"rtod": cl.RadToDeg,
		// 		"rtop":  cl.RectToPolar,
		// 		"ptor":  cl.PolarToRect,
		"floor": cl.Floor,
		"ceil":  cl.Ceil,
		// 		"trunc":  cl.Trunc,
		"and":  cl.And,
		"or":   cl.Or,
		"xor":  cl.Xor,
		"not":  cl.Not,
		"andn": cl.AndN,
		"orn":  cl.OrN,
		"xorn": cl.XorN,
		"sum":  cl.Sum,
		"avg":  cl.Avg,
		// 		"dot":    cl.Dot,
		// 		"dot3":   cl.Dot3,
		// 		"cross":  cl.Cross,
		// 		"mag":    cl.Mag,
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
		"pi": func() error {
			_, pi := value.Consts()
			return cl.Push(pi)
		},
		"e": func() error {
			e, _ := value.Consts()
			return cl.Push(e)
		},
		// 		"phi":    func() error { return cl.Push(math.Phi) },
		"quit": func() error { exit(); return nil },
		"help": func() error { help(); return nil },
	}
)

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
		"In command line mode, output stack in hexidecimal format")
	flag.BoolVar(&doInitStack, "i", doInitStack,
		"Initialize with input from command line arguments")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, usageStr)
		flag.PrintDefaults()
	}
	value.SetConfig(cfg)
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
		// 		input, err := lnr.Prompt(" ")
		input, err := trm.ReadLine()
		lastErr = nil
		if err == io.EOF {
			exit()
		}
		if err != nil {
			continue
		}
		if strings.TrimSpace(input) != "" {
			// 			lnr.AppendHistory(input)
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
	for i := range stack {
		// 		if doHexOut {
		// 			fmt.Printf("%#x", int64(stack[len(stack)-i-1]))
		// 		} else {
		fmt.Print(stack[len(stack)-i-1])
		// 		}
		if i < len(stack)-1 {
			fmt.Print(" ")
		}
	}
	fmt.Println()
}

func exit() {
	fmt.Println()
	// 	lnr.Close()
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
		if clac.IsNum(tok) {
			num, err := value.Parse(tok)
			if err != nil {
				errorHandler(err)
				continue
			}
			if err = cl.Exec(func() error { return cl.Push(num) }); err != nil {
				errorHandler(errors.New("push: " + err.Error()))
			}
			continue
		}
		if cmd, ok := cmdMap[tok]; ok {
			if err := cl.Exec(cmd); err != nil {
				errorHandler(errors.New(tok + ": " + err.Error()))
			}
			continue
		}
		errorHandler(errors.New(tok + ": invalid input"))
	}
}

func printStack(stack clac.Stack) {
	_, numRows, err := terminal.GetSize(syscall.Stdout)
	if err != nil {
		numRows = len(stack) + 1
	}
	clearScreen()

	for i := numRows - 3; i >= 0; i-- {
		fmt.Printf("%2d:", i)
		if i < len(stack) {
			cfg.SetFormat("%#16.10g")
			fmt.Printf(" %s", stack[i])

			if val, err := clac.Unary("floor", stack[i]); err == nil {
				cfg.SetFormat(" %#19x")
				fmt.Print(val)
			}
		}
		fmt.Println()
	}
	if lastErr == nil {
		fmt.Println(strings.Repeat("-", 40))
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
