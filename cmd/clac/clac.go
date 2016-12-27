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
	"os/exec"
	"strings"
	"syscall"

	"github.com/ianremmler/clac"
	"golang.org/x/crypto/ssh/terminal"
	"robpike.io/ivy/value"
)

type runMode int

const (
	cliMode runMode = iota
	tuiMode
	dmenuMode
)

var (
	// flags
	doDmenu          = false
	doInitStack      = false
	doHexOut         = false
	outPrec     uint = 12

	cl      = clac.New()
	lastErr error
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

var uiCmdMap = map[string]func() error{
	"undo":  cl.Undo,
	"u":     cl.Undo,
	"redo":  cl.Redo,
	"r":     cl.Redo,
	"clear": cl.Clear,
	"c":     cl.Clear,
	"reset": cl.Reset,
	"quit":  quit,
	"q":     quit,
}

type term struct {
	io.Reader
	io.Writer
}

func init() {
	log.SetFlags(0)
	log.SetPrefix("clac: ")
	flag.BoolVar(&doDmenu, "d", doDmenu, "dmenu mode")
	flag.BoolVar(&doHexOut, "x", doHexOut, "hexidecimal output")
	flag.UintVar(&outPrec, "p", outPrec, "output precision")
	flag.BoolVar(&doInitStack, "i", doInitStack, "initialize stack")
}

func main() {
	flag.Parse()
	var mode runMode
	mode, lastErr = processCmdLine()
	switch mode {
	case cliMode:
		cliRun()
	case tuiMode:
		tuiRun()
	case dmenuMode:
		dmenuRun()
	}
}

func constant(v value.Value) func() error {
	return func() error { return cl.Push(v) }
}

func cliRun() {
	fmt.Println(stackStr(cl.Stack()))
	if lastErr != nil {
		log.Fatal(lastErr)
	}
}

func uiSetup() {
	for cmd, fn := range uiCmdMap {
		cmdMap[cmd] = fn
	}
	uiCmdMap = nil
}

func tuiRun() {
	uiSetup()
	if !terminal.IsTerminal(syscall.Stdin) {
		log.Fatalln("this doesn't look like an interactive terminal")
	}
	oldTrmState, err := terminal.MakeRaw(syscall.Stdin)
	if err != nil {
		log.Fatalln(err)
	}
	trm := terminal.NewTerminal(term{os.Stdin, os.Stdout}, "")
	for lastErr != io.EOF {
		tuiPrintStack(cl.Stack())
		var input string
		input, lastErr = trm.ReadLine()
		if lastErr == nil {
			lastErr = processInput(input)
		}
	}
	terminal.Restore(syscall.Stdin, oldTrmState)
}

func dmenuSetup() {
	uiSetup()
	cmdMap["hex"] = func() error { doHexOut = true; return nil }
	cmdMap["dec"] = func() error { doHexOut = false; return nil }
	cmdMap["conv"] = dmenuConv
}

func dmenuRun() {
	dmenuSetup()
	for {
		stack := stackStr(cl.Stack())
		if len(stack) > 0 {
			stack = " " + stack
		}
		out, err := exec.Command("dmenu", "-p", "clac:"+stack).Output()
		if err != nil {
			return
		}
		if err := processInput(string(out)); err != nil {
			exec.Command("dmenu", "-p", "clac: "+err.Error()).Run()
		}
	}
}

func dmenuConv() error {
	val, err := cl.Pop()
	if err != nil {
		return err
	}
	valStr := stackStr(clac.Stack{val})
	have, err := exec.Command("dmenu", "-p", "clac: conv: have: "+valStr).Output()
	if err != nil {
		return clac.ErrNoHistUpdate
	}
	want, err := exec.Command("dmenu", "-p", "clac: conv: want:").Output()
	if err != nil {
		return clac.ErrNoHistUpdate
	}
	haveStr := valStr + " " + strings.TrimSpace(string(have))
	wantStr := strings.TrimSpace(string(want))
	out, err := exec.Command("units", "-t", haveStr, wantStr).Output()
	if err != nil {
		errStr := strings.SplitN(string(out), "\n", 2)[0]
		if errStr != "" {
			return errors.New(errStr)
		}
		return err
	}
	num, err := clac.ParseNum(strings.TrimSpace(string(out)))
	if err != nil {
		return err
	}
	return cl.Push(num)
}

func processCmdLine() (runMode, error) {
	input := ""
	if stat, err := os.Stdin.Stat(); err == nil && stat.Mode()&os.ModeNamedPipe != 0 {
		if pipeInput, err := ioutil.ReadAll(os.Stdin); err == nil {
			input = string(pipeInput)
		}
	}
	if len(flag.Args()) > 0 {
		input += " " + strings.Join(flag.Args(), " ")
	}
	mode := cliMode
	switch {
	case doDmenu:
		mode = dmenuMode
	case doInitStack || input == "":
		mode = tuiMode
	}
	cl.EnableHistory(mode != cliMode)
	err := processInput(string(input))
	return mode, err
}

func stackStr(stack clac.Stack) string {
	out := ""
	if doHexOut {
		clac.SetFormat("%#x")
	} else {
		clac.SetFormat(fmt.Sprintf("%%.%dg", outPrec))
	}
	for i := range stack {
		val := stack[len(stack)-i-1]
		var err error
		if doHexOut {
			val, err = clac.Trunc(val)
		}
		if err != nil {
			out += err.Error()
		} else {
			out += clac.Sprint(val)
		}
		if i < len(stack)-1 {
			out += " "
		}
	}
	return out
}

func quit() error {
	os.Exit(0)
	return nil
}

func processInput(input string) error {
	scanner := bufio.NewScanner(strings.NewReader(input))
	scanner.Split(bufio.ScanWords)
	for scanner.Scan() {
		tok := scanner.Text()
		if num, err := clac.ParseNum(tok); err == nil {
			if err = cl.Exec(func() error { return cl.Push(num) }); err != nil {
				return fmt.Errorf("push: %s", err)
			}
		} else if cmd, ok := cmdMap[tok]; ok {
			if err := cl.Exec(cmd); err != nil {
				return fmt.Errorf("%s: %s", tok, err)
			}
		} else {
			return fmt.Errorf("invalid input: \"%s\"", tok)
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}

func tuiPrintStack(stack clac.Stack) {
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
	info := ""
	if lastErr != nil {
		info = fmt.Sprintf("[ %s ]", lastErr)
	}
	fmt.Println(info + strings.Repeat("-", cols-len(info)))
	fmt.Print("\r")
}

func clearScreen() {
	fmt.Print("\033[2J\033[H")
}
