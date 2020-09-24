package main

// addr18 is an 18 bit unibus address.
type addr18 uint32

// UNIBUS is a PDP11 UNIBUS 18 bus.
type UNIBUS struct {

	// 128 KW of core memory minus 4 KW of io page.
	// [000000, 177777)
	core [(128 << 10) - (4 << 10)]uint16
}

// read16 reads addr from the UNIBUS.
func (u *UNIBUS) read16(addr addr18) uint16 {
	if int(addr>>1) < len(u.core) {
		return u.core[addr>>1]
	}
	switch addr {

	default:
		panic(trap{INTBUS})
	}
}

// write16 writes v to addr on the UNIBUS.
func (u *UNIBUS) write16(addr addr18, v uint16) {
	if int(addr>>1) < len(u.core) {
		u.core[addr>>1] = v
		return
	}
	switch addr {

	default:
		panic(trap{INTBUS})
	}
}

func (u *UNIBUS) reset() {
	// u.cons.clearterminal()
	// u.rk11.reset()
	// u.kw11.write16(0777546, 0x00) // disable line clock INTR
}
