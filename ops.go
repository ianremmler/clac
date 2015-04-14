package clac

import "robpike.io/ivy/value"

const (
	variadic = -1
)

// x is the last stack value.
// y is the penultimate stack value.

// Undo undoes the last operation.
func (c *Clac) Undo() error {
	if !c.hist.undo() {
		return errNoMoreChanges
	}
	return errNoHistUpdate
}

// Redo redoes the last undone operation.
func (c *Clac) Redo() error {
	if !c.hist.redo() {
		return errNoMoreChanges
	}
	return errNoHistUpdate
}

// y is the penultimate stack value.

// Clear clears the stack.
func (c *Clac) Clear() error {
	if len(c.working) == 0 {
		return errNoHistUpdate
	}
	c.working = Stack{}
	return nil
}

// Push pushes a value on the stack.
func (c *Clac) Push(a value.Value) error {
	return c.push(a)
}

// Drop drops the last stack value.
func (c *Clac) Drop() error {
	return c.drop(0, 1)
}

// Dropn drops the last x stack values.
func (c *Clac) DropN() error {
	num, err := c.popCount()
	if err != nil {
		return err
	}
	return c.drop(0, num)
}

// Dropr drops a range of x stack values, starting at index y.
func (c *Clac) DropR() error {
	num, err := c.popCount()
	if err != nil {
		return err
	}
	pos, err := c.popIndex()
	if err != nil {
		return err
	}
	return c.drop(pos, num)
}

// Dup duplicates the last stack value.
func (c *Clac) Dup() error {
	return c.dup(0, 1)
}

// DupN duplicates the last x stack values.
func (c *Clac) DupN() error {
	num, err := c.popCount()
	if err != nil {
		return err
	}
	return c.dup(0, num)
}

// DupR duplicates a range of x stack values, starting at index y.
func (c *Clac) DupR() error {
	num, err := c.popCount()
	if err != nil {
		return err
	}
	pos, err := c.popIndex()
	if err != nil {
		return err
	}
	return c.dup(pos, num)
}

// Pick duplicates the stack value at index x.
func (c *Clac) Pick() error {
	pos, err := c.popIndex()
	if err != nil {
		return err
	}
	return c.dup(pos, 1)
}

// Rot rotates the stack value at index x down.
func (c *Clac) Rot() error {
	return c.rot(true)
}

// Unrot rotates the stack value at index x up.
func (c *Clac) Unrot() error {
	return c.rot(false)
}

func (c *Clac) rot(isDown bool) error {
	pos, err := c.popIndex()
	if err != nil {
		return err
	}
	return c.rotate(pos, 1, isDown)
}

// Rotr rotates a range of x stack values, starting at index y, down.
func (c *Clac) RotR() error {
	return c.rotR(true)
}

// Unrotr rotates a range of x stack values, starting at index y, up.
func (c *Clac) UnrotR() error {
	return c.rotR(false)
}

func (c *Clac) rotR(isDown bool) error {
	num, err := c.popCount()
	if err != nil {
		return err
	}
	pos, err := c.popIndex()
	if err != nil {
		return err
	}
	return c.rotate(pos, num, isDown)
}

// Swap swaps the last two stack values.
func (c *Clac) Swap() error {
	return c.rotate(1, 1, true)
}

// Depth returns the number of stack values
func (c *Clac) Depth() error {
	return c.push(value.Int(len(c.Stack())))
}

type floatFunc func(vals []value.Value) (value.Value, error)
type binFloatFunc func(a, b value.Value) (value.Value, error)

func (c *Clac) applyFloat(arity int, f floatFunc) error {
	if arity < 0 {
		num, err := c.popCount()
		if err != nil {
			return err
		}
		arity = num
	}
	vals, err := c.remove(0, arity)
	if err != nil {
		return err
	}
	res, err := f(vals)
	if err != nil {
		return err
	}
	return c.push(res)
}

func reduceFloat(initVal value.Value, vals []value.Value, f binFloatFunc) (value.Value, error) {
	var err error
	val := initVal
	for _, v := range vals {
		val, err = f(val, v)
		if err != nil {
			return zero, err
		}
	}
	return val, nil
}

type intFunc func(vals []value.Value) (value.Value, error)
type binIntFunc func(a, b value.Value) (value.Value, error)

func (c *Clac) applyInt(arity int, f intFunc) error {
	if arity < 0 {
		num, err := c.popCount()
		if err != nil {
			return err
		}
		arity = num
	}
	vals, err := c.remove(0, arity)
	if err != nil {
		return err
	}
	ivals := make([]value.Value, arity)
	for i, v := range vals {
		ivals[i], err = Trunc(v)
		if err != nil {
			return err
		}
	}
	res, err := f(ivals)
	if err != nil {
		return err
	}
	return c.push(res)
}

func reduceInt(initVal value.Value, vals []value.Value, f binIntFunc) (value.Value, error) {
	var err error
	val := initVal
	for _, v := range vals {
		val, err = f(val, v)
		if err != nil {
			return zero, err
		}
	}
	return val, nil
}

// Neg returns the negation of x.
func (c *Clac) Neg() error {
	return c.applyFloat(1, func(vals []value.Value) (value.Value, error) {
		return unary("-", vals[0])
	})
}

// Abs returns the absolute value of x.
func (c *Clac) Abs() error {
	return c.applyFloat(1, func(vals []value.Value) (value.Value, error) {
		return unary("abs", vals[0])
	})
}

// Inv returns the inverse of x.
func (c *Clac) Inv() error {
	return c.applyFloat(1, func(vals []value.Value) (value.Value, error) {
		return unary("/", vals[0])
	})
}

// Add returns the sum of y and x.
func (c *Clac) Add() error {
	return c.applyFloat(2, func(vals []value.Value) (value.Value, error) {
		return binary(vals[1], "+", vals[0])
	})
}

// Sub returns the difference of y and x.
func (c *Clac) Sub() error {
	return c.applyFloat(2, func(vals []value.Value) (value.Value, error) {
		return binary(vals[1], "-", vals[0])
	})
}

// Mul returns the product of y and x.
func (c *Clac) Mul() error {
	return c.applyFloat(2, func(vals []value.Value) (value.Value, error) {
		return binary(vals[1], "*", vals[0])
	})
}

// Div returns the quotient of y divided by x.
func (c *Clac) Div() error {
	return c.applyFloat(2, func(vals []value.Value) (value.Value, error) {
		return binary(vals[1], "/", vals[0])
	})
}

// IntDiv returns the quotient of y divided by x.
func (c *Clac) IntDiv() error {
	return c.applyInt(2, func(vals []value.Value) (value.Value, error) {
		return binary(vals[1], "div", vals[0])
	})
}

// Mod returns the remainder of y divided by x.
func (c *Clac) Mod() error {
	return c.applyFloat(2, func(vals []value.Value) (value.Value, error) {
		return binary(vals[1], "mod", vals[0])
	})
}

// Pow returns y to the x power.
func (c *Clac) Pow() error {
	return c.applyFloat(2, func(vals []value.Value) (value.Value, error) {
		return binary(vals[1], "**", vals[0])
	})
}

// Sqrt returns the square root of x.
func (c *Clac) Sqrt() error {
	return c.applyFloat(1, func(vals []value.Value) (value.Value, error) {
		return unary("sqrt", vals[0])
	})
}

// Exp returns e to the power of x.
func (c *Clac) Exp() error {
	return c.applyFloat(1, func(vals []value.Value) (value.Value, error) {
		return binary(E, "**", vals[0])
	})
}

// Pow2 returns 2 to the power of x.
func (c *Clac) Pow2() error {
	return c.applyFloat(1, func(vals []value.Value) (value.Value, error) {
		return binary(value.Int(2), "**", vals[0])
	})
}

// Pow10 returns 10 to the power of x.
func (c *Clac) Pow10() error {
	return c.applyFloat(1, func(vals []value.Value) (value.Value, error) {
		return binary(value.Int(10), "**", vals[0])
	})
}

// Ln returns the natural log of x.
func (c *Clac) Ln() error {
	return c.applyFloat(1, func(vals []value.Value) (value.Value, error) {
		return unary("log", vals[0])
	})
}

// Lg returns the base 2 logarithm of x.
func (c *Clac) Lg() error {
	return c.applyFloat(1, func(vals []value.Value) (value.Value, error) {
		e := eval{}
		lnx := e.unary("log", vals[0])
		ln2 := e.unary("log", value.Int(2))
		lg := e.binary(lnx, "/", ln2)
		return lg, e.err
	})
}

// Log returns the base 10 logarithm of x.
func (c *Clac) Log() error {
	return c.applyFloat(1, func(vals []value.Value) (value.Value, error) {
		e := &eval{}
		lnx := e.unary("log", vals[0])
		ln10 := e.unary("log", value.Int(10))
		log := e.binary(lnx, "/", ln10)
		return log, e.err
	})
}

// Sin returns the sine of x.
func (c *Clac) Sin() error {
	return c.applyFloat(1, func(vals []value.Value) (value.Value, error) {
		return unary("sin", vals[0])
	})
}

// Cos returns the cosine of x.
func (c *Clac) Cos() error {
	return c.applyFloat(1, func(vals []value.Value) (value.Value, error) {
		return unary("cos", vals[0])
	})
}

// Tan returns the tangent of x.
func (c *Clac) Tan() error {
	return c.applyFloat(1, func(vals []value.Value) (value.Value, error) {
		return unary("tan", vals[0])
	})
}

// Asin returns the arcsine of x.
func (c *Clac) Asin() error {
	return c.applyFloat(1, func(vals []value.Value) (value.Value, error) {
		return unary("asin", vals[0])
	})
}

// Acos returns the arccosine of x.
func (c *Clac) Acos() error {
	return c.applyFloat(1, func(vals []value.Value) (value.Value, error) {
		return unary("acos", vals[0])
	})
}

// Atan returns the arctangent of x.
func (c *Clac) Atan() error {
	return c.applyFloat(1, func(vals []value.Value) (value.Value, error) {
		return unary("atan", vals[0])
	})
}

// Atan2 returns the arctangent of y / x
func (c *Clac) Atan2() error {
	return c.applyFloat(2, func(vals []value.Value) (value.Value, error) {
		return atan2(vals[1], vals[0])
	})
}

func atan2(x, y value.Value) (value.Value, error) {
	e := &eval{}

	// special cases
	tan := value.Value(zero)
	if isTrue(e.binary(y, "==", zero)) {
		if isTrue(e.binary(x, "<", zero)) {
			tan = Pi
		}
		return tan, e.err
	}
	if isTrue(e.binary(x, "==", zero)) {
		ySgn := e.unary("sgn", y)
		tan = e.binary(Pi, "/", value.Int(2))
		tan = e.binary(tan, "*", ySgn)
		return tan, e.err
	}

	tan = e.binary(y, "/", x)
	angle := e.unary("atan", tan)
	if isTrue(e.binary(x, "<", zero)) {
		if isTrue(e.binary(tan, "<=", zero)) {
			angle = e.binary(angle, "+", Pi)
		} else {
			angle = e.binary(angle, "-", Pi)
		}
	}
	return angle, e.err
}

// DegToRad converts a value in degrees to radians.
func (c *Clac) DegToRad() error {
	return c.applyFloat(1, func(vals []value.Value) (value.Value, error) {
		e := &eval{}
		radPerDeg := e.binary(Pi, "/", value.Int(180))
		rad := e.binary(vals[0], "*", radPerDeg)
		return rad, e.err
	})
}

// RadToDeg converts a value in radians to degrees.
func (c *Clac) RadToDeg() error {
	return c.applyFloat(1, func(vals []value.Value) (value.Value, error) {
		e := &eval{}
		degPerRad := e.binary(value.Int(180), "/", Pi)
		deg := e.binary(vals[0], "*", degPerRad)
		return deg, e.err
	})
}

// Hypot calculates the 2D hypotenuse of a right triangles with legs x and y
func (c *Clac) Hypot() error {
	return c.applyFloat(1, func(vals []value.Value) (value.Value, error) {
		return hypot(vals[1], vals[0])
	})
}

func hypot(x, y value.Value) (value.Value, error) {
	e := &eval{}
	hyp := e.unary("sqrt", e.binary(e.binary(x, "*", x), "+", e.binary(y, "*", y)))
	return hyp, e.err
}

// RectToPolar converts 2D rectangular coordinates y,x to polar coordinates.
func (c *Clac) RectToPolar() error {
	e := &eval{}
	y := e.e(func() (value.Value, error) { return c.pop() })
	x := e.e(func() (value.Value, error) { return c.pop() })
	radius := e.e(func() (value.Value, error) { return hypot(x, y) })
	e.e(func() (value.Value, error) { return zero, c.push(radius) })
	angle := e.e(func() (value.Value, error) { return atan2(x, y) })
	e.e(func() (value.Value, error) { return zero, c.push(angle) })
	return e.err
}

// PolarToRect converts 2D polar coordinates y<x to rectangular coordinates.
func (c *Clac) PolarToRect() error {
	e := &eval{}
	angle := e.e(func() (value.Value, error) { return c.pop() })
	radius := e.e(func() (value.Value, error) { return c.pop() })
	x := e.binary(radius, "*", e.unary("cos", angle))
	e.e(func() (value.Value, error) { return zero, c.push(x) })
	y := e.binary(radius, "*", e.unary("sin", angle))
	e.e(func() (value.Value, error) { return zero, c.push(y) })
	return e.err
}

// Floor returns largest integer not greater than x.
func (c *Clac) Floor() error {
	return c.applyFloat(1, func(vals []value.Value) (value.Value, error) {
		return unary("floor", vals[0])
	})
}

// Ceil returns smallest integer not less than x.
func (c *Clac) Ceil() error {
	return c.applyFloat(1, func(vals []value.Value) (value.Value, error) {
		return unary("ceil", vals[0])
	})
}

// Trunc returns x truncated to the nearest integer toward 0.
func (c *Clac) Trunc() error {
	return c.applyFloat(1, func(vals []value.Value) (value.Value, error) {
		return Trunc(vals[0])
	})
}

// And returns the bitwise and of the integer portions of y and x.
func (c *Clac) And() error {
	return c.applyInt(2, func(vals []value.Value) (value.Value, error) {
		return binary(vals[1], "&", vals[0])
	})
}

// Or returns the bitwise or of the integer portions of y and x.
func (c *Clac) Or() error {
	return c.applyInt(2, func(vals []value.Value) (value.Value, error) {
		return binary(vals[1], "|", vals[0])
	})
}

// Xor returns the bitwise exclusive or of the integer portions of y and x.
func (c *Clac) Xor() error {
	return c.applyInt(2, func(vals []value.Value) (value.Value, error) {
		return binary(vals[1], "^", vals[0])
	})
}

// Not returns the bitwise not of the integer portion x.
func (c *Clac) Not() error {
	return c.applyInt(1, func(vals []value.Value) (value.Value, error) {
		return unary("^", vals[0])
	})
}

// AndN returns the bitwise and of the integer portions of the last x stack values.
func (c *Clac) AndN() error {
	return c.applyInt(variadic, func(vals []value.Value) (value.Value, error) {
		return reduceInt(value.Int(-1), vals, func(a, b value.Value) (value.Value, error) {
			return binary(a, "&", b)
		})
	})
}

// OrN returns the bitwise or of the integer portions of the last x stack values.
func (c *Clac) OrN() error {
	return c.applyInt(variadic, func(vals []value.Value) (value.Value, error) {
		return reduceInt(zero, vals, func(a, b value.Value) (value.Value, error) {
			return binary(a, "|", b)
		})
	})
}

// XorN returns the bitwise exclusive or of the integer portions of the last x stack values.
func (c *Clac) XorN() error {
	return c.applyInt(variadic, func(vals []value.Value) (value.Value, error) {
		return reduceInt(zero, vals, func(a, b value.Value) (value.Value, error) {
			return binary(a, "^", b)
		})
	})
}

// Sum returns the sum of the last x stack values
func (c *Clac) Sum() error {
	return c.applyFloat(variadic, func(vals []value.Value) (value.Value, error) {
		return reduceFloat(zero, vals, func(a, b value.Value) (value.Value, error) {
			return binary(a, "+", b)
		})
	})
}

// Avg returns the mean of the last x stack values
func (c *Clac) Avg() error {
	return c.applyFloat(variadic, func(vals []value.Value) (value.Value, error) {
		sum, _ := reduceFloat(zero, vals, func(a, b value.Value) (value.Value, error) {
			return binary(a, "+", b)
		})
		return binary(sum, "/", value.Int(len(vals)))
	})
}

// Min returns the minimum of x and y
func (c *Clac) Min() error {
	return c.applyFloat(2, func(vals []value.Value) (value.Value, error) {
		return binary(vals[1], "min", vals[0])
	})
}

// Max returns the maximum of x and y
func (c *Clac) Max() error {
	return c.applyFloat(2, func(vals []value.Value) (value.Value, error) {
		return binary(vals[1], "max", vals[0])
	})
}

// MinN returns the minimum of the last x stack values.
func (c *Clac) MinN() error {
	return c.applyFloat(variadic, func(vals []value.Value) (value.Value, error) {
		return reduceFloat(vals[0], vals, func(a, b value.Value) (value.Value, error) {
			return binary(a, "min", b)
		})
	})
}

// MaxN returns the maximum of the last x stack values.
func (c *Clac) MaxN() error {
	return c.applyFloat(variadic, func(vals []value.Value) (value.Value, error) {
		return reduceFloat(vals[0], vals, func(a, b value.Value) (value.Value, error) {
			return binary(a, "max", b)
		})
	})
}

// Factorial returns the factorial of x
func (c *Clac) Factorial() error {
	return c.applyInt(1, func(vals []value.Value) (value.Value, error) {
		return factorial(vals[0])
	})
}

func factorial(val value.Value) (value.Value, error) {
	e := eval{}
	n, err := valToInt(val)
	if err != nil {
		return zero, err
	}
	var fact value.Value = value.Int(1)
	for i := 2; i <= n; i++ {
		fact = e.binary(fact, "*", value.Int(i))
	}
	return fact, e.err
}

// Comb returns the number of combinations of x taken from y
func (c *Clac) Comb() error {
	return c.applyInt(2, func(vals []value.Value) (value.Value, error) {
		e := &eval{}
		nf := e.e(func() (value.Value, error) { return factorial(vals[1]) })
		rf := e.e(func() (value.Value, error) { return factorial(vals[0]) })
		nr := e.binary(vals[1], "-", vals[0])
		nrf := e.e(func() (value.Value, error) { return factorial(nr) })
		denom := e.binary(nrf, "*", rf)
		n := e.binary(nf, "/", denom)
		return n, e.err
	})
}

// Perm returns the number of permutations of x taken from y
func (c *Clac) Perm() error {
	return c.applyInt(2, func(vals []value.Value) (value.Value, error) {
		e := &eval{}
		nf := e.e(func() (value.Value, error) { return factorial(vals[1]) })
		nr := e.binary(vals[1], "-", vals[0])
		nrf := e.e(func() (value.Value, error) { return factorial(nr) })
		n := e.binary(nf, "/", nrf)
		return n, e.err
	})
}

// Dot returns the dot product of two vectors of size x
// The vectors are composed of the 2*x items on the stack above x
func (c *Clac) Dot() error {
	num, err := c.popCount()
	if err != nil {
		return err
	}
	return c.dot(num)
}

// Dot3 returns the dot product of two 3D vectors
// The vectors are composed of the last 6 items on the stack
func (c *Clac) Dot3() error {
	return c.dot(3)
}

func (c *Clac) dot(num int) error {
	e := &eval{}
	if num < 1 {
		return errInvalidArg
	}
	vals, err := c.remove(0, 2*num)
	if err != nil {
		return err
	}
	a, b := vals[:num], vals[num:]
	dot := zero
	for i := range a {
		dot = e.binary(dot, "+", e.binary(a[i], "*", b[i]))
	}
	if e.err != nil {
		return e.err
	}
	return c.push(dot)
}

// Cross returns the cross product of two 3D vectors
// The vectors are composed of the last 6 items on the stack
func (c *Clac) Cross() error {
	e := &eval{}
	vals, err := c.remove(0, 6)
	if err != nil {
		return err
	}
	a, b := vals[:3], vals[3:]
	cross := []value.Value{
		e.binary(e.binary(a[1], "*", b[2]), "-", e.binary(a[2], "*", b[1])),
		e.binary(e.binary(a[2], "*", b[0]), "-", e.binary(a[0], "*", b[2])),
		e.binary(e.binary(a[0], "*", b[1]), "-", e.binary(a[1], "*", b[0])),
	}
	if e.err != nil {
		return e.err
	}
	return c.insert(cross, 0)
}

// Mag returns the magnitude of the vector represented by the last x stack values
func (c *Clac) Mag() error {
	return c.applyFloat(variadic, func(vals []value.Value) (value.Value, error) {
		e := &eval{}
		magSq, _ := reduceFloat(zero, vals, func(a, b value.Value) (value.Value, error) {
			mag := e.binary(a, "+", e.binary(b, "*", b))
			return mag, e.err
		})
		return unary("sqrt", magSq)
	})
}

func isTrue(val value.Value) bool {
	ival, ok := val.(value.Int)
	if !ok {
		return true
	}
	return ival != 0
}
