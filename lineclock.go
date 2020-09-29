package main

import (
	"fmt"
	"time"
)

type KW11 struct {
	csr   uint16
	ticks <-chan time.Time
}

func (kw *KW11) write16(addr addr18, v uint16) {
	switch addr {
	case 0777546:
		//		fmt.Printf("kw11:write16: %06o %06o\n", addr, v)
		kw.csr = v
	default:
		fmt.Printf("kw11: write to invalid address %06o\n", addr)
		panic(trap{INTBUS})
	}
}

func (kw *KW11) read16(addr addr18) uint16 {
	switch addr {
	case 0777546:
		//		fmt.Printf("kw11:read16: %06o\n", addr)
		return kw.csr
	default:
		fmt.Printf("kw11: read from invalid address %06o\n", addr)
		panic(trap{INTBUS})
	}
}

func (kw *KW11) tick() {
	select {
	case <-kw.ticks:
		kw.csr |= (1 << 7)
		if kw.csr&(1<<6) > 0 {
			panic(interrupt{INTCLOCK, 6})
		}
	default:
	}
}
