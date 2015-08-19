clac
====

Clac is a practical RPN calculator for the terminal and shell.

Features include:
- Command line completion and history
- Unlimited undo/redo
- Integer input using C-style decimal, octal, or hexidecimal syntax
- Decimal and hexidecimal display of all stack values, all the time
- Pipeline mode processes input from stdin and prints results to stdout

Clac uses Rob Pike's [Ivy](http://robpike.io/ivy) calculator for exact/high
precision calculations.  Ivy requires Go 1.5, hence so does Clac.

To get it, make sure you have [Go](http://golang.org/doc/install) installed,
then run: `go get github.com/ianremmler/clac/cmd/clac`.
