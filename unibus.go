package main

import "fmt"

// addr18 is an 18 bit unibus address.
type addr18 uint32

// UNIBUS is a PDP11 UNIBUS 18 bus.
type UNIBUS struct {

	// 128 KW of core memory minus 4 KW of io page.
	// [000000, 760000)
	core [(128 - 4) << 10]uint16

	rk11      RK11
	cons      KL11
	mmu       *KT11
	lineclock KW11
}

// read16 reads addr from the UNIBUS.
func (u *UNIBUS) read16(addr addr18) uint16 {
	// fmt.Printf("unibus: read16: %06o\n", addr)
	if addr < 0760000 {
		return u.core[addr>>1]
	}
	switch addr & ^addr18(077) {
	case 0777400:
		return u.rk11.read16(addr)
	case 0777500:
		switch addr {
		case 0777546:
			return u.lineclock.read16(addr)
		case 0777572:
			return u.mmu.SR0
		case 0777574:
			return u.mmu.SR1
		case 0777576:
			return u.mmu.SR2
		default:
			return u.cons.read16(addr)
		}
	case 0772200, 0772300, 0777600:
		return u.mmu.read16(addr)
	default:
		fmt.Printf("unibus: read from invalid address %06o\n", addr)
		panic(trap{INTBUS})
	}
}

// write16 writes v to addr on the UNIBUS.
func (u *UNIBUS) write16(addr addr18, v uint16) {
	if addr < 0760000 {
		u.core[addr>>1] = v
		return
	}

	switch addr & ^addr18(077) {
	case 0777400:
		u.rk11.write16(addr, v)
	case 0777500:
		switch addr {
		case 0777546:
			u.lineclock.write16(addr, v)
		case 0777572:
			u.mmu.SR0 = v
		case 0777574:
			u.mmu.SR1 = v
		case 0777576:
			u.mmu.SR2 = v
		default:
			u.cons.write16(addr, v)
		}
	case 0772200, 0772300, 0777600:
		u.mmu.write16(addr, v)
	default:
		fmt.Printf("unibus: write to invalid address %06o\n", addr)
		panic(trap{INTBUS})
	}
}

func (u *UNIBUS) reset() {
	u.cons.reset()
	u.rk11.reset()
	u.lineclock.write16(0777546, 0x00) // disable line clock INTR
}
