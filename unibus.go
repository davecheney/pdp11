package main

// addr18 is an 18 bit unibus address.
type addr18 uint32

// UNIBUS is a PDP11 UNIBUS 18 bus.
type UNIBUS struct {

	// 128 KW of core memory minus 4 KW of io page.
	// [000000, 177777)
	core [((256 << 10) - (8 << 10)) >> 1]uint16
}

// Read16 reads addr from the UNIBUS.
func (u *UNIBUS) Read16(addr addr18) uint16 {
	if int(addr) < len(u.core)<<1 {
		return u.core[addr>>1]
	}
	switch addr {

	default:
		panic(trap{INTBUS})
	}
}

// Write16 writes v to addr on the UNIBUS.
func (u *UNIBUS) Write16(addr addr18, v uint16) {
	if int(addr) < len(u.core)<<1 {
		u.core[addr>>1] = v
		return
	}
	switch addr {

	default:
		panic(trap{INTBUS})
	}
}

func (u *UNIBUS) reset() {
	kb.cons.clearterminal()
	kb.rk11.reset()
	kb.kw11.write16(0777546, 0x00) // disable line clock INTR
}
