package clac

import "math"

// x is the last value on the stack.
// y is the penultimate value on the stack.

// Clear clears the stack.
func (c *Clac) Clear() error {
	if len(c.hist.stack()) > 0 {
		c.hist.push(Stack{})
		c.updateWorking()
	}
	return nil
}

// Push pushes a value on the stack.
func (c *Clac) Push(x float64) error {
	return c.push(x)
}

// Drop drops the last value from the stack.
func (c *Clac) Drop() error {
	return c.drop(0, 1)
}

// Dropn drops the last x values from the stack.
func (c *Clac) Dropn() error {
	num, err := c.pop()
	if err != nil {
		return err
	}
	return c.drop(0, int(num))
}

// Dropr drops a range of x values from the stack, starting at index y.
func (c *Clac) Dropr() error {
	num, err := c.pop()
	if err != nil {
		return err
	}
	pos, err := c.pop()
	if err != nil {
		return err
	}
	return c.drop(int(pos), int(num))
}

// Dup duplicates the last value on the stack.
func (c *Clac) Dup() error {
	return c.dup(0, 1)
}

// Dupn duplicates the last x values on the stack.
func (c *Clac) Dupn() error {
	num, err := c.pop()
	if err != nil {
		return err
	}
	return c.dup(0, int(num))
}

// Dupr duplicates a range of x values on the stack, starting at index y.
func (c *Clac) Dupr() error {
	num, err := c.pop()
	if err != nil {
		return err
	}
	pos, err := c.pop()
	if err != nil {
		return err
	}
	return c.dup(int(pos), int(num))
}

// Pick duplicates the value on the stack at index x.
func (c *Clac) Pick() error {
	pos, err := c.pop()
	if err != nil {
		return err
	}
	return c.dup(int(pos), 1)
}

// Rot rotates the value on the stack at index x up or down.
func (c *Clac) Rot(isDown bool) error {
	pos, err := c.pop()
	if err != nil {
		return err
	}
	return c.rot(int(pos), 1, isDown)
}

// Rotr rotates a range of x values on the stack, starting at index y, up or down.
func (c *Clac) Rotr(isDown bool) error {
	num, err := c.pop()
	if err != nil {
		return err
	}
	pos, err := c.pop()
	if err != nil {
		return err
	}
	return c.rot(int(pos), int(num), isDown)
}

// Swap swaps the last two values on the stack.
func (c *Clac) Swap() error {
	return c.rot(1, 1, true)
}

// Depth returns the number of values on the stack
func (c *Clac) Depth() error {
	return c.push(float64(len(c.Stack())))
}

// FloatFunc represents a floating point function.
type FloatFunc func(x []float64) (float64, error)

func (c *Clac) applyFloat(arity int, f FloatFunc) error {
	vals, err := c.remove(0, arity)
	if err != nil {
		return tooFewArgsErr
	}
	res, err := f(vals)
	if err != nil {
		return err
	}
	if math.IsNaN(res) {
		return invalidArgErr
	}
	return c.push(res)
}

// IntFunc represents an integer function
type IntFunc func(x []int64) (int64, error)

func (c *Clac) applyInt(arity int, f IntFunc) error {
	vals, err := c.remove(0, arity)
	if err != nil {
		return tooFewArgsErr
	}
	ivals := make([]int64, arity)
	for i, v := range vals {
		if math.Abs(v) > math.MaxInt64 {
			return outOfRangeErr
		}
		ivals[i] = int64(v)
	}
	res, err := f(ivals)
	if err != nil {
		return err
	}
	return c.push(float64(res))
}

// Neg returns the negation of x.
func (c *Clac) Neg() error {
	return c.applyFloat(1, func(x []float64) (float64, error) {
		return -x[0], nil
	})
}

// Abs returns the absolute value of x.
func (c *Clac) Abs() error {
	return c.applyFloat(1, func(x []float64) (float64, error) {
		return math.Abs(x[0]), nil
	})
}

// Inv returns the inverse of x.
func (c *Clac) Inv() error {
	return c.applyFloat(1, func(x []float64) (float64, error) {
		if x[0] == 0 {
			return 0, invalidArgErr
		}
		return 1 / x[0], nil
	})
}

// Add returns the sum of y and x.
func (c *Clac) Add() error {
	return c.applyFloat(2, func(x []float64) (float64, error) {
		return x[1] + x[0], nil
	})
}

// Sub returns the difference of y and x.
func (c *Clac) Sub() error {
	return c.applyFloat(2, func(x []float64) (float64, error) {
		return x[1] - x[0], nil
	})
}

// Mul returns the product of y and x.
func (c *Clac) Mul() error {
	return c.applyFloat(2, func(x []float64) (float64, error) {
		return x[1] * x[0], nil
	})
}

// Div returns the quotient of y divided by x.
func (c *Clac) Div() error {
	return c.applyFloat(2, func(x []float64) (float64, error) {
		if x[0] == 0 {
			return 0, invalidArgErr
		}
		return x[1] / x[0], nil
	})
}

// Mod returns the remainder of y divided by x.
func (c *Clac) Mod() error {
	return c.applyFloat(2, func(x []float64) (float64, error) {
		return math.Mod(x[1], x[0]), nil
	})
}

// Pow returns y to the x power.
func (c *Clac) Pow() error {
	return c.applyFloat(2, func(x []float64) (float64, error) {
		return math.Pow(x[1], x[0]), nil
	})
}

// Sqrt returns the square root of x.
func (c *Clac) Sqrt() error {
	return c.applyFloat(1, func(x []float64) (float64, error) {
		return math.Sqrt(x[0]), nil
	})
}

// Hypot returns the square root of x squared + y squared.
func (c *Clac) Hypot() error {
	return c.applyFloat(2, func(x []float64) (float64, error) {
		return math.Hypot(x[1], x[0]), nil
	})
}

// Exp returns e to the power of x.
func (c *Clac) Exp() error {
	return c.applyFloat(1, func(x []float64) (float64, error) {
		return math.Exp(x[0]), nil
	})
}

// Pow2 returns 2 to the power of x.
func (c *Clac) Pow2() error {
	return c.applyFloat(1, func(x []float64) (float64, error) {
		return math.Exp2(x[0]), nil
	})
}

// Pow10 returns 10 to the power of x.
func (c *Clac) Pow10() error {
	return c.applyFloat(1, func(x []float64) (float64, error) {
		return math.Pow(10, x[0]), nil
	})
}

// Ln returns the natural log of x.
func (c *Clac) Ln() error {
	return c.applyFloat(1, func(x []float64) (float64, error) {
		return math.Log(x[0]), nil
	})
}

// Lg returns the base 2 logarithm of x.
func (c *Clac) Lg() error {
	return c.applyFloat(1, func(x []float64) (float64, error) {
		return math.Log2(x[0]), nil
	})
}

// Log returns the base 10 logarithm of x.
func (c *Clac) Log() error {
	return c.applyFloat(1, func(x []float64) (float64, error) {
		return math.Log10(x[0]), nil
	})
}

// Sin returns the sine of x.
func (c *Clac) Sin() error {
	return c.applyFloat(1, func(x []float64) (float64, error) {
		return math.Sin(x[0]), nil
	})
}

// Cos returns the cosine of x.
func (c *Clac) Cos() error {
	return c.applyFloat(1, func(x []float64) (float64, error) {
		return math.Cos(x[0]), nil
	})
}

// Tan returns the tangent of x.
func (c *Clac) Tan() error {
	return c.applyFloat(1, func(x []float64) (float64, error) {
		return math.Tan(x[0]), nil
	})
}

// Sinh returns the hyperbolic sine of x.
func (c *Clac) Sinh() error {
	return c.applyFloat(1, func(x []float64) (float64, error) {
		return math.Sinh(x[0]), nil
	})
}

// Cosh returns the hyperbolic cosine of x.
func (c *Clac) Cosh() error {
	return c.applyFloat(1, func(x []float64) (float64, error) {
		return math.Cosh(x[0]), nil
	})
}

// Tanh returns the hyperbolic tangent of x.
func (c *Clac) Tanh() error {
	return c.applyFloat(1, func(x []float64) (float64, error) {
		return math.Tanh(x[0]), nil
	})
}

// Asin returns the arcsine of x.
func (c *Clac) Asin() error {
	return c.applyFloat(1, func(x []float64) (float64, error) {
		return math.Asin(x[0]), nil
	})
}

// Acos returns the arccosine of x.
func (c *Clac) Acos() error {
	return c.applyFloat(1, func(x []float64) (float64, error) {
		return math.Acos(x[0]), nil
	})
}

// Atan returns the arctangent of x.
func (c *Clac) Atan() error {
	return c.applyFloat(1, func(x []float64) (float64, error) {
		return math.Atan(x[0]), nil
	})
}

// Asinh returns the hyperbolic arcsine of x.
func (c *Clac) Asinh() error {
	return c.applyFloat(1, func(x []float64) (float64, error) {
		return math.Asinh(x[0]), nil
	})
}

// Acosh returns the hyperbolic arccosine of x.
func (c *Clac) Acosh() error {
	return c.applyFloat(1, func(x []float64) (float64, error) {
		return math.Acosh(x[0]), nil
	})
}

// Atanh returns the hyperbolic arctangent of x.
func (c *Clac) Atanh() error {
	return c.applyFloat(1, func(x []float64) (float64, error) {
		return math.Atanh(x[0]), nil
	})
}

// Atan2 returns the arctangent of y / x
func (c *Clac) Atan2() error {
	return c.applyFloat(2, func(x []float64) (float64, error) {
		return math.Atan2(x[1], x[0]), nil
	})
}

// DegToRad converts a value in degrees to radians.
func (c *Clac) DegToRad() error {
	return c.applyFloat(1, func(x []float64) (float64, error) {
		return x[0] * math.Pi / 180, nil
	})
}

// RadToDeg converts a value in radians to degrees.
func (c *Clac) RadToDeg() error {
	return c.applyFloat(1, func(x []float64) (float64, error) {
		return x[0] * 180 / math.Pi, nil
	})
}

// Floor returns largest integer not greater than x.
func (c *Clac) Floor() error {
	return c.applyFloat(1, func(x []float64) (float64, error) {
		return math.Floor(x[0]), nil
	})
}

// Ceil returns smallest integer not less than x.
func (c *Clac) Ceil() error {
	return c.applyFloat(1, func(x []float64) (float64, error) {
		return math.Ceil(x[0]), nil
	})
}

// Trunc returns x truncated to the nearest integer toward 0.
func (c *Clac) Trunc() error {
	return c.applyFloat(1, func(x []float64) (float64, error) {
		return math.Trunc(x[0]), nil
	})
}

// And returns the bitwise and of the integer portions of y and x.
func (c *Clac) And() error {
	return c.applyInt(2, func(x []int64) (int64, error) {
		return x[1] & x[0], nil
	})
}

// Or returns the bitwise or of the integer portions of y and x.
func (c *Clac) Or() error {
	return c.applyInt(2, func(x []int64) (int64, error) {
		return x[1] | x[0], nil
	})
}

// Xor returns the bitwise exclusive or of the integer portions of y and x.
func (c *Clac) Xor() error {
	return c.applyInt(2, func(x []int64) (int64, error) {
		return x[1] ^ x[0], nil
	})
}

// Not returns the bitwise not of the integer portion x.
func (c *Clac) Not() error {
	return c.applyInt(1, func(x []int64) (int64, error) {
		return ^x[0], nil
	})
}

func sum(vals []float64) float64 {
	sum := 0.0
	for _, x := range vals {
		sum += x
	}
	return sum
}

// Sum returns the sum of the last x values on the stack
func (c *Clac) Sum() error {
	num, err := c.pop()
	if err != nil {
		return err
	}
	if num < 1 {
		return outOfRangeErr
	}
	return c.applyFloat(int(num), func(x []float64) (float64, error) {
		return sum(x), nil
	})
}

// Avg returns the mean of the last x values on the stack
func (c *Clac) Avg() error {
	num, err := c.pop()
	if err != nil {
		return err
	}
	if num < 1 {
		return outOfRangeErr
	}
	return c.applyFloat(int(num), func(x []float64) (float64, error) {
		return sum(x) / math.Floor(num), nil
	})
}
