package main

import "fmt"

const (
	INTBUS    = 0004
	INTINVAL  = 0010
	INTDEBUG  = 0014
	INTIOT    = 0020
	INTTTYIN  = 0060
	INTTTYOUT = 0064
	INTFAULT  = 0250
	INTCLOCK  = 0100
	INTRK     = 0220
)

type interrupt struct {
	vec uint16
	pri uint16
}

func (i interrupt) String() string {
	return fmt.Sprintf("interrupt: %06o, pri: %03o", i.vec, i.pri)
}

// Trap is a PDP11 trap.
type trap struct {
	vec uint16
}

func (t trap) String() string {
	return fmt.Sprintf("trap: %06o", t.vec)
}
