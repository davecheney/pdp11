package main

import "fmt"

type KW11 struct {
	csr   uint16
	ticks uint16
}

func (kw *KW11) write16(addr addr18, v uint16) {
	fmt.Printf("kw11:write16: %06o %06o\n", addr, v)
	switch addr {
	case 0777546:
		kw.csr = v
	default:
		fmt.Printf("kw11: write to invalid address %06o\n", addr)
		panic(trap{INTBUS})
	}
}

func (kw *KW11) read16(addr addr18) uint16 {
	fmt.Printf("kw11:read16: %06o\n", addr)
	switch addr {
	case 0777546:
		return kw.csr
	default:
		fmt.Printf("kw11: read from invalid address %06o\n", addr)
		panic(trap{INTBUS})
	}
}

func (kw *KW11) tick() {
	kw.ticks++
	if kw.ticks == 0 {
		kw.csr |= (1 << 7)
		if kw.csr&(1<<6) > 0 {
			panic(interrupt{INTCLOCK, 6})
		}

	}
}
