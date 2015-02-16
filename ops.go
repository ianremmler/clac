package clac

import "math"

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
func (c *Clac) Push(a float64) error {
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
	return c.push(float64(len(c.Stack())))
}

type floatFunc func(vals []float64) (float64, error)
type binFloatFunc func(a, b float64) (float64, error)

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
	if math.IsNaN(res) {
		return errInvalidArg
	}
	return c.push(res)
}

func reduceFloat(initVal float64, vals []float64, f binFloatFunc) (float64, error) {
	var err error
	val := initVal
	for _, v := range vals {
		val, err = f(val, v)
		if err != nil {
			return 0, err
		}
	}
	return val, nil
}

type intFunc func(vals []int64) (int64, error)
type binIntFunc func(a, b int64) (int64, error)

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
	ivals := make([]int64, arity)
	for i, v := range vals {
		if math.Abs(v) > math.MaxInt64 {
			return errOutOfRange
		}
		ivals[i] = int64(v)
	}
	res, err := f(ivals)
	if err != nil {
		return err
	}
	return c.push(float64(res))
}

func reduceInt(initVal int64, vals []int64, f binIntFunc) (int64, error) {
	var err error
	val := initVal
	for _, v := range vals {
		val, err = f(val, v)
		if err != nil {
			return 0, err
		}
	}
	return val, nil
}

// Neg returns the negation of x.
func (c *Clac) Neg() error {
	return c.applyFloat(1, func(vals []float64) (float64, error) {
		return -vals[0], nil
	})
}

// Abs returns the absolute value of x.
func (c *Clac) Abs() error {
	return c.applyFloat(1, func(vals []float64) (float64, error) {
		return math.Abs(vals[0]), nil
	})
}

// Inv returns the inverse of x.
func (c *Clac) Inv() error {
	return c.applyFloat(1, func(vals []float64) (float64, error) {
		if vals[0] == 0 {
			return 0, errInvalidArg
		}
		return 1 / vals[0], nil
	})
}

// Add returns the sum of y and x.
func (c *Clac) Add() error {
	return c.applyFloat(2, func(vals []float64) (float64, error) {
		return vals[1] + vals[0], nil
	})
}

// Sub returns the difference of y and x.
func (c *Clac) Sub() error {
	return c.applyFloat(2, func(vals []float64) (float64, error) {
		return vals[1] - vals[0], nil
	})
}

// Mul returns the product of y and x.
func (c *Clac) Mul() error {
	return c.applyFloat(2, func(vals []float64) (float64, error) {
		return vals[1] * vals[0], nil
	})
}

// Div returns the quotient of y divided by x.
func (c *Clac) Div() error {
	return c.applyFloat(2, func(vals []float64) (float64, error) {
		if vals[0] == 0 {
			return 0, errInvalidArg
		}
		return vals[1] / vals[0], nil
	})
}

// Mod returns the remainder of y divided by x.
func (c *Clac) Mod() error {
	return c.applyFloat(2, func(vals []float64) (float64, error) {
		return math.Mod(vals[1], vals[0]), nil
	})
}

// Pow returns y to the x power.
func (c *Clac) Pow() error {
	return c.applyFloat(2, func(vals []float64) (float64, error) {
		return math.Pow(vals[1], vals[0]), nil
	})
}

// Sqrt returns the square root of x.
func (c *Clac) Sqrt() error {
	return c.applyFloat(1, func(vals []float64) (float64, error) {
		return math.Sqrt(vals[0]), nil
	})
}

// Hypot returns the square root of x squared + y squared.
func (c *Clac) Hypot() error {
	return c.applyFloat(2, func(vals []float64) (float64, error) {
		return math.Hypot(vals[1], vals[0]), nil
	})
}

// Exp returns e to the power of x.
func (c *Clac) Exp() error {
	return c.applyFloat(1, func(vals []float64) (float64, error) {
		return math.Exp(vals[0]), nil
	})
}

// Pow2 returns 2 to the power of x.
func (c *Clac) Pow2() error {
	return c.applyFloat(1, func(vals []float64) (float64, error) {
		return math.Exp2(vals[0]), nil
	})
}

// Pow10 returns 10 to the power of x.
func (c *Clac) Pow10() error {
	return c.applyFloat(1, func(vals []float64) (float64, error) {
		return math.Pow(10, vals[0]), nil
	})
}

// Ln returns the natural log of x.
func (c *Clac) Ln() error {
	return c.applyFloat(1, func(vals []float64) (float64, error) {
		return math.Log(vals[0]), nil
	})
}

// Lg returns the base 2 logarithm of x.
func (c *Clac) Lg() error {
	return c.applyFloat(1, func(vals []float64) (float64, error) {
		return math.Log2(vals[0]), nil
	})
}

// Log returns the base 10 logarithm of x.
func (c *Clac) Log() error {
	return c.applyFloat(1, func(vals []float64) (float64, error) {
		return math.Log10(vals[0]), nil
	})
}

// Sin returns the sine of x.
func (c *Clac) Sin() error {
	return c.applyFloat(1, func(vals []float64) (float64, error) {
		return math.Sin(vals[0]), nil
	})
}

// Cos returns the cosine of x.
func (c *Clac) Cos() error {
	return c.applyFloat(1, func(vals []float64) (float64, error) {
		return math.Cos(vals[0]), nil
	})
}

// Tan returns the tangent of x.
func (c *Clac) Tan() error {
	return c.applyFloat(1, func(vals []float64) (float64, error) {
		return math.Tan(vals[0]), nil
	})
}

// Sinh returns the hyperbolic sine of x.
func (c *Clac) Sinh() error {
	return c.applyFloat(1, func(vals []float64) (float64, error) {
		return math.Sinh(vals[0]), nil
	})
}

// Cosh returns the hyperbolic cosine of x.
func (c *Clac) Cosh() error {
	return c.applyFloat(1, func(vals []float64) (float64, error) {
		return math.Cosh(vals[0]), nil
	})
}

// Tanh returns the hyperbolic tangent of x.
func (c *Clac) Tanh() error {
	return c.applyFloat(1, func(vals []float64) (float64, error) {
		return math.Tanh(vals[0]), nil
	})
}

// Asin returns the arcsine of x.
func (c *Clac) Asin() error {
	return c.applyFloat(1, func(vals []float64) (float64, error) {
		return math.Asin(vals[0]), nil
	})
}

// Acos returns the arccosine of x.
func (c *Clac) Acos() error {
	return c.applyFloat(1, func(vals []float64) (float64, error) {
		return math.Acos(vals[0]), nil
	})
}

// Atan returns the arctangent of x.
func (c *Clac) Atan() error {
	return c.applyFloat(1, func(vals []float64) (float64, error) {
		return math.Atan(vals[0]), nil
	})
}

// Asinh returns the hyperbolic arcsine of x.
func (c *Clac) Asinh() error {
	return c.applyFloat(1, func(vals []float64) (float64, error) {
		return math.Asinh(vals[0]), nil
	})
}

// Acosh returns the hyperbolic arccosine of x.
func (c *Clac) Acosh() error {
	return c.applyFloat(1, func(vals []float64) (float64, error) {
		return math.Acosh(vals[0]), nil
	})
}

// Atanh returns the hyperbolic arctangent of x.
func (c *Clac) Atanh() error {
	return c.applyFloat(1, func(vals []float64) (float64, error) {
		return math.Atanh(vals[0]), nil
	})
}

// Atan2 returns the arctangent of y / x
func (c *Clac) Atan2() error {
	return c.applyFloat(2, func(vals []float64) (float64, error) {
		return math.Atan2(vals[1], vals[0]), nil
	})
}

// DegToRad converts a value in degrees to radians.
func (c *Clac) DegToRad() error {
	return c.applyFloat(1, func(vals []float64) (float64, error) {
		return vals[0] * math.Pi / 180, nil
	})
}

// RadToDeg converts a value in radians to degrees.
func (c *Clac) RadToDeg() error {
	return c.applyFloat(1, func(vals []float64) (float64, error) {
		return vals[0] * 180 / math.Pi, nil
	})
}

// RectToPolar converts 2D rectangular coordinates y,x to polar coordinates.
func (c *Clac) RectToPolar() error {
	y, err := c.pop()
	if err != nil {
		return err
	}
	x, err := c.pop()
	if err != nil {
		return err
	}
	if c.push(math.Hypot(x, y)) != nil {
		return err
	}
	return c.push(math.Atan2(y, x))
}

// PolarToRect converts 2D polar coordinates y<x to rectangular coordinates.
func (c *Clac) PolarToRect() error {
	theta, err := c.pop()
	if err != nil {
		return err
	}
	r, err := c.pop()
	if err != nil {
		return err
	}
	if c.push(r*math.Cos(theta)) != nil {
		return err
	}
	return c.push(r * math.Sin(theta))
}

// Floor returns largest integer not greater than x.
func (c *Clac) Floor() error {
	return c.applyFloat(1, func(vals []float64) (float64, error) {
		return math.Floor(vals[0]), nil
	})
}

// Ceil returns smallest integer not less than x.
func (c *Clac) Ceil() error {
	return c.applyFloat(1, func(vals []float64) (float64, error) {
		return math.Ceil(vals[0]), nil
	})
}

// Trunc returns x truncated to the nearest integer toward 0.
func (c *Clac) Trunc() error {
	return c.applyFloat(1, func(vals []float64) (float64, error) {
		return math.Trunc(vals[0]), nil
	})
}

// And returns the bitwise and of the integer portions of y and x.
func (c *Clac) And() error {
	return c.applyInt(2, func(vals []int64) (int64, error) {
		return vals[1] & vals[0], nil
	})
}

// Or returns the bitwise or of the integer portions of y and x.
func (c *Clac) Or() error {
	return c.applyInt(2, func(vals []int64) (int64, error) {
		return vals[1] | vals[0], nil
	})
}

// Xor returns the bitwise exclusive or of the integer portions of y and x.
func (c *Clac) Xor() error {
	return c.applyInt(2, func(vals []int64) (int64, error) {
		return vals[1] ^ vals[0], nil
	})
}

// Not returns the bitwise not of the integer portion x.
func (c *Clac) Not() error {
	return c.applyInt(1, func(vals []int64) (int64, error) {
		return ^vals[0], nil
	})
}

// Andn returns the bitwise and of the integer portions of the last x stack values.
func (c *Clac) AndN() error {
	return c.applyInt(variadic, func(vals []int64) (int64, error) {
		return reduceInt(^0, vals, func(a, b int64) (int64, error) {
			return a & b, nil
		})
	})
}

// Orn returns the bitwise or of the integer portions of the last x stack values.
func (c *Clac) OrN() error {
	return c.applyInt(variadic, func(vals []int64) (int64, error) {
		return reduceInt(0, vals, func(a, b int64) (int64, error) {
			return a | b, nil
		})
	})
}

// Xorn returns the bitwise exclusive or of the integer portions of the last x stack values.
func (c *Clac) XorN() error {
	return c.applyInt(variadic, func(vals []int64) (int64, error) {
		return reduceInt(0, vals, func(a, b int64) (int64, error) {
			return a ^ b, nil
		})
	})
}

// Sum returns the sum of the last x stack values
func (c *Clac) Sum() error {
	return c.applyFloat(variadic, func(vals []float64) (float64, error) {
		return reduceFloat(0, vals, func(a, b float64) (float64, error) {
			return a + b, nil
		})
	})
}

// Avg returns the mean of the last x stack values
func (c *Clac) Avg() error {
	return c.applyFloat(variadic, func(vals []float64) (float64, error) {
		sum, _ := reduceFloat(0, vals, func(a, b float64) (float64, error) {
			return a + b, nil
		})
		return sum / float64(len(vals)), nil
	})
}

// Min returns the minimum of x and y
func (c *Clac) Min() error {
	return c.applyFloat(2, func(vals []float64) (float64, error) {
		return math.Min(vals[0], vals[1]), nil
	})
}

// Max returns the maximum of x and y
func (c *Clac) Max() error {
	return c.applyFloat(2, func(vals []float64) (float64, error) {
		return math.Max(vals[0], vals[1]), nil
	})
}

// Minn returns the minimum of the last x stack values.
func (c *Clac) MinN() error {
	return c.applyFloat(variadic, func(vals []float64) (float64, error) {
		return reduceFloat(math.MaxFloat64, vals, func(a, b float64) (float64, error) {
			return math.Min(a, b), nil
		})
	})
}

// Maxn returns the maximum of the last x stack values.
func (c *Clac) MaxN() error {
	return c.applyFloat(variadic, func(vals []float64) (float64, error) {
		return reduceFloat(-math.MaxFloat64, vals, func(a, b float64) (float64, error) {
			return math.Max(a, b), nil
		})
	})
}

// Gamma returns the gamma function of x
func (c *Clac) Gamma() error {
	return c.applyFloat(1, func(vals []float64) (float64, error) {
		return math.Gamma(vals[0]), nil
	})
}

// Factorial returns the factorial of x
func (c *Clac) Factorial() error {
	return c.applyInt(1, func(vals []int64) (int64, error) {
		if vals[0] < 0 {
			return 0, errInvalidArg
		}
		n := int64(math.Gamma(float64(vals[0]+1)) + 0.5)
		if n < 0 {
			return 0, errInvalidArg
		}
		return n, nil
	})
}

// Comb returns the number of combinations of x taken from y
func (c *Clac) Comb() error {
	return c.applyInt(2, func(vals []int64) (int64, error) {
		if vals[0] < 0 || vals[1] < 0 || vals[1] < vals[0] {
			return 0, errInvalidArg
		}
		nf := math.Gamma(float64(vals[1] + 1))
		rf := math.Gamma(float64(vals[0] + 1))
		nrf := math.Gamma(float64(vals[1] - vals[0] + 1))
		n := int64(nf/(nrf*rf) + 0.5)
		if n < 0 {
			return 0, errInvalidArg
		}
		return n, nil
	})
}

// Perm returns the number of permutations of x taken from y
func (c *Clac) Perm() error {
	return c.applyInt(2, func(vals []int64) (int64, error) {
		if vals[0] < 0 || vals[1] < 0 || vals[1] < vals[0] {
			return 0, errInvalidArg
		}
		nf := math.Gamma(float64(vals[1] + 1))
		nrf := math.Gamma(float64(vals[1] - vals[0] + 1))
		n := int64(nf/nrf + 0.5)
		if n < 0 {
			return 0, errInvalidArg
		}
		return n, nil
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
	if num < 1 {
		return errInvalidArg
	}
	vals, err := c.remove(0, 2*num)
	if err != nil {
		return err
	}
	a, b := vals[:num], vals[num:]
	dot := 0.0
	for i := range a {
		dot += a[i] * b[i]
	}
	return c.push(dot)
}

// Cross returns the cross product of two 3D vectors
// The vectors are composed of the last 6 items on the stack
func (c *Clac) Cross() error {
	vals, err := c.remove(0, 6)
	if err != nil {
		return err
	}
	a, b := vals[:3], vals[3:]
	cross := []float64{
		a[1]*b[2] - a[2]*b[1],
		a[2]*b[0] - a[0]*b[2],
		a[0]*b[1] - a[1]*b[0],
	}
	return c.insert(cross, 0)
}

// Mag returns the magnitude of the vector represented by the last x stack values
func (c *Clac) Mag() error {
	return c.applyFloat(variadic, func(vals []float64) (float64, error) {
		magSq, _ := reduceFloat(0, vals, func(a, b float64) (float64, error) {
			return a + b*b, nil
		})
		return math.Sqrt(magSq), nil
	})
}
