// Package clac implements an RPN calculator.
package clac

import (
	"errors"
	"math/big"

	"robpike.io/ivy/config"
	"robpike.io/ivy/value"
)

var (
	Pi, E value.BigFloat

	errTooFewArgs    = errors.New("too few arguments")
	errInvalidArg    = errors.New("invalid argument")
	errOutOfRange    = errors.New("argument out of range")
	errNoMoreChanges = errors.New("no more changes")
	errNoHistUpdate  = errors.New("") // for cmds that don't add to history

	zero   = value.Int(0)
	ivyCfg = &config.Config{}
)

func init() {
	value.SetConfig(ivyCfg)
	E, Pi = value.Consts()
}

func SetFormat(format string) {
	ivyCfg.SetFormat(format)
}

// IsNum returns whether the string represents a number
func IsNum(in string) bool {
	_, ok := big.NewFloat(0).SetString(in)
	return ok
}

// Stack represents a stack of floating point numbers.
type Stack []value.Value

type stackHist struct {
	cur  int
	hist []Stack
}

func newStackHist() *stackHist {
	return &stackHist{hist: []Stack{Stack{}}}
}

func (s *stackHist) undo() bool {
	if s.cur <= 0 {
		return false
	}
	s.cur--
	return true
}

func (s *stackHist) redo() bool {
	if s.cur >= len(s.hist)-1 {
		return false
	}
	s.cur++
	return true
}

func (s *stackHist) push(stack Stack) {
	s.hist = append(s.hist[:s.cur+1], stack)
	s.cur++
}

func (s *stackHist) stack() Stack {
	return s.hist[s.cur]
}

// Clac represents an RPN calculator.
type Clac struct {
	working Stack
	hist    *stackHist
}

// New returns an initialized Clac instance.
func New() *Clac {
	c := &Clac{}
	c.Reset()
	return c
}

// Reset resets clac to its initial state
func (c *Clac) Reset() {
	c.working = Stack{}
	c.hist = newStackHist()
}

// Stack returns the current stack.
func (c *Clac) Stack() Stack {
	return c.working
}

// Exec executes a clac command, along with necessary bookkeeping
func (c *Clac) Exec(f func() error) error {
	c.updateWorking()
	err := f()
	if err == nil {
		c.hist.push(c.working)
	}
	c.updateWorking()
	if err == errNoHistUpdate {
		return nil
	}
	return err
}

func (c *Clac) updateWorking() {
	c.working = append(Stack{}, c.hist.stack()...)
}

func (c *Clac) checkRange(pos, num int, isEndOK bool) (int, int, error) {
	max := len(c.working)
	if isEndOK {
		max++
	}
	start, end := pos, pos+num-1
	if start < 0 || start > end {
		return 0, 0, errInvalidArg
	}
	if start >= max || end >= max {
		return 0, 0, errTooFewArgs
	}
	return start, end, nil
}

func (c *Clac) insert(vals []value.Value, pos int) error {
	idx, _, err := c.checkRange(pos, 1, true)
	if err != nil {
		return err
	}
	c.working = append(c.working[:idx], append(vals, c.working[idx:]...)...)
	return nil
}

func (c *Clac) push(x value.Value) error {
	return c.insert([]value.Value{x}, 0)
}

func (c *Clac) remove(pos, num int) ([]value.Value, error) {
	start, end, err := c.checkRange(pos, num, false)
	if err != nil {
		return nil, err
	}
	vals := append([]value.Value{}, c.working[start:end+1]...)
	c.working = append(c.working[:start], c.working[end+1:]...)
	return vals, nil
}

func (c *Clac) pop() (value.Value, error) {
	x, err := c.remove(0, 1)
	if err != nil {
		return zero, errTooFewArgs
	}
	return x[0], err
}

func (c *Clac) popIntMin(min int) (int, error) {
	x, err := c.pop()
	if err != nil {
		return 0, err
	}
	xi, ok := x.(value.Int)
	if !ok {
		return 0, errInvalidArg
	}
	n := int(xi)
	if n < min {
		return 0, errInvalidArg
	}
	return n, err
}

func (c *Clac) popIndex() (int, error) {
	return c.popIntMin(0)
}

func (c *Clac) popCount() (int, error) {
	return c.popIntMin(1)
}

func (c *Clac) vals(pos, num int) (Stack, error) {
	start, end, err := c.checkRange(pos, num, false)
	if err != nil {
		return nil, err
	}
	return c.working[start : end+1], nil
}

func (c *Clac) dup(pos, num int) error {
	vals, err := c.vals(pos, num)
	if err != nil {
		return err
	}
	return c.insert(vals, 0)
}

func (c *Clac) drop(pos, num int) error {
	_, err := c.remove(pos, num)
	return err
}

func (c *Clac) rotate(pos, num int, isDown bool) error {
	from, to := pos, 0
	if !isDown {
		from, to = 0, pos-num+1
	}
	vals, err := c.remove(from, num)
	if err != nil {
		return err
	}
	return c.insert(vals, to)
}

func Unary(op string, a value.Value) (val value.Value, err error) {
	defer func() { err = errVal(recover()) }()
	val = value.Unary(op, a)
	return val, err
}

func Binary(a value.Value, op string, b value.Value) (val value.Value, err error) {
	defer func() { err = errVal(recover()) }()
	val = value.Binary(a, op, b)
	return val, err
}

func errVal(val interface{}) error {
	if val == nil {
		return nil
	}
	if err, ok := val.(error); ok {
		return err
	}
	return errors.New("unknown error")
}
