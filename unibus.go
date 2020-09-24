package main

import "fmt"

// addr18 is an 18 bit unibus address.
type addr18 uint32

// UNIBUS is a PDP11 UNIBUS 18 bus.
type UNIBUS struct {

	// 128 KW of core memory minus 4 KW of io page.
	// [000000, 177777)
	core [(128 << 10) - (4 << 10)]uint16

	rk11 RK11
}

// read16 reads addr from the UNIBUS.
func (u *UNIBUS) read16(addr addr18) uint16 {
	if int(addr>>1) < len(u.core) {
		return u.core[addr>>1]
	}
	switch addr & ^addr18(077) {
	case 0777400:
		return u.rk11.read16(addr)
	default:
		fmt.Printf("unibus: read from invalid address %06o\n", addr)
		panic(trap{INTBUS})
	}
}

// write16 writes v to addr on the UNIBUS.
func (u *UNIBUS) write16(addr addr18, v uint16) {
	if int(addr>>1) < len(u.core) {
		u.core[addr>>1] = v
		return
	}

	switch addr & ^addr18(077) {
	case 0777400:
		u.rk11.write16(addr, v)
	default:
		fmt.Printf("unibus: write to invalid address %06o\n", addr)
		panic(trap{INTBUS})
	}
}

func (u *UNIBUS) reset() {
	// u.cons.clearterminal()
	// u.rk11.reset()
	// u.kw11.write16(0777546, 0x00) // disable line clock INTR
}
