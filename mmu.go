package main

import (
	"fmt"
)

type page struct {
	par, pdr uint16
}

func (p *page) addr() addr18 { return addr18(p.par & 07777) }
func (p *page) len() uint16  { return (p.pdr >> 8) & 0x7f }
func (p *page) read() bool   { return p.pdr&2 == 2 }
func (p *page) write() bool  { return p.pdr&6 == 6 }
func (p *page) ed() bool     { return p.pdr&8 == 8 }

type KT11 struct {
	SR0, SR1, SR2 uint16
	pages         [16]page
}

func (kt *KT11) decode(wr bool, a, mode uint16) addr18 {
	if kt.SR0&01 == 0 {
		addr := addr18(a)
		if addr > 0167777 {
			return addr + 0600000
		}
		//fmt.Printf("decode: fast %06o -> %06o\n", a, addr)
		return addr
	}
	i := a >> 13
	if mode > 0 {
		i += 8
	}
	if wr && !kt.pages[i].write() {
		kt.SR0 = (1 << 13) | 1
		kt.SR0 |= a >> 12 & ^uint16(1)
		if mode > 0 {
			kt.SR0 |= (1 << 5) | (1 << 6)
		}
		// SR2 = cpu.PC;
		fmt.Printf("mmu::decode write to read-only page %06o\n", a)
		panic(trap{INTFAULT})
	}
	if !kt.pages[i].read() {
		kt.SR0 = (1 << 15) | 1
		kt.SR0 |= a >> 12 & ^uint16(1)
		if mode > 0 {
			kt.SR0 |= (1 << 5) | (1 << 6)
		}
		// SR2 = cpu.PC;
		fmt.Printf("mmu::decode read from no-access page %06o\n", a)
		panic(trap{INTFAULT})
	}
	block := (a >> 6) & 0177
	disp := addr18(a & 077)
	if (kt.pages[i].ed() && (block < kt.pages[i].len())) || (!kt.pages[i].ed() && (block > kt.pages[i].len())) {
		kt.SR0 = (1 << 14) | 1
		kt.SR0 |= a >> 12 & ^uint16(1)
		if mode > 0 {
			kt.SR0 |= (1 << 5) | (1 << 6)
		}
		// SR2 = cpu.PC;
		fmt.Printf("page length exceeded, address %06o (block %03o) is beyond length %03o\n",
			a, block, kt.pages[i].len())
		panic(trap{INTFAULT})
	}
	if wr {
		kt.pages[i].pdr |= 1 << 6
	}
	aa := ((addr18(block) + kt.pages[i].addr()) << 6) + disp
	// fmt.Printf("decode: slow %06o -> %06o\n", a, aa)
	return aa
}

func (kt *KT11) write16(addr addr18, v uint16) {
	// fmt.Printf("kt11:write16: %06o %06o\n", addr, v)
	i := (addr & 017) >> 1
	switch addr & ^addr18(037) {
	case 0772300:
		kt.pages[i].pdr = v
	case 0772340:
		kt.pages[i].par = v
	case 0777600:
		kt.pages[i+8].pdr = v
	case 0777640:
		kt.pages[i+8].par = v
	default:
		fmt.Printf("mmu:write16: %06o %06o\n", addr, v)
		panic(trap{INTBUS})
	}
}

func (kt *KT11) read16(addr addr18) uint16 {
	// fmt.Printf("kt11:read16: %06o\n", addr)
	i := (addr & 017) >> 1
	switch addr & ^addr18(037) {
	case 0772300:
		return kt.pages[i].pdr
	case 0772340:
		return kt.pages[i].par
	case 0777600:
		return kt.pages[i+8].pdr
	case 0777640:
		return kt.pages[i+8].par
	default:
		fmt.Printf("mmu:write16: %06o\n", addr)
		panic(trap{INTBUS})
	}
}
