package main

import "fmt"

type page struct {
	par, pdr uint16
}

func (p *page) addr() uint16 { return p.par & 07777 }
func (p *page) len() uint16  { return (p.pdr >> 8) & 0x7f }
func (p *page) read() bool   { return p.pdr&2 > 0 }
func (p *page) write() bool  { return p.pdr&6 > 0 }
func (p *page) ed() bool     { return p.pdr&8 > 0 }

type KT11 struct {
	SR0, SR1, SR2 uint16
	pages         [16]page
}

func (kt *KT11) decode(wr bool, a, mode uint16) addr18 {
	if (kt.SR0 & 1) == 0 {
		if a > 0167777 {
			return addr18(a) + 0600000
		}
		return addr18(a)
	}
	i := a >> 13
	if mode > 0 {
		i += 8
	}

	if wr && !kt.pages[i].write() {
		kt.SR0 = (1 << 13) | 1
		kt.SR0 |= (a >> 12) &^ 1
		if mode > 0 {
			kt.SR0 |= (1 << 5) | (1 << 6)
		}
		// SR2 = cpu.PC;
		fmt.Printf("mmu::decode write to read-only page %06o\n", a)
		panic(trap{INTFAULT})
	}
	if !kt.pages[i].read() {
		kt.SR0 = (1 << 15) | 1
		kt.SR0 |= (a >> 12) &^ 1
		if mode > 0 {
			kt.SR0 |= (1 << 5) | (1 << 6)
		}
		// SR2 = cpu.PC;
		fmt.Printf("mmu::decode read from no-access page %06o\n", a)
		panic(trap{INTFAULT})
	}
	block := (a >> 6) & 0177
	disp := a & 077
	if (kt.pages[i].ed() && (block < kt.pages[i].len())) || (!kt.pages[i].ed() && (block > kt.pages[i].len())) {
		kt.SR0 = (1 << 14) | 1
		kt.SR0 |= (a >> 12) &^ 1
		if mode > 0 {
			kt.SR0 |= (1 << 5) | (1 << 6)
		}
		// SR2 = cpu.PC;
		fmt.Printf("page length exceeded, address %06o (block %03o) is beyond length %03o\r\n",
			a, block, kt.pages[i].len())
		panic(trap{INTFAULT})
	}
	if wr {
		kt.pages[i].pdr |= 1 << 6
	}
	aa := addr18(kt.pages[i].par & 07777)
	aa += addr18(block)
	aa <<= 6
	aa += addr18(disp)

	//    printf("decode: slow %06o -> %06o\n", a, aa);

	return aa
}

func (kt *KT11) Read16(a addr18) uint16 {
	if (a >= 0772300) && (a < 0772320) {
		return kt.pages[((a & 017) >> 1)].pdr
	}
	if (a >= 0772340) && (a < 0772360) {
		return kt.pages[((a & 017) >> 1)].par
	}
	if (a >= 0777600) && (a < 0777620) {
		return kt.pages[((a&017)>>1)+8].pdr
	}
	if (a >= 0777640) && (a < 0777660) {
		return kt.pages[((a&017)>>1)+8].par
	}
	fmt.Printf("mmu::read16 invalid read from %06o\n", a)
	panic(trap{INTBUS})
}
func (kt *KT11) Write16(a addr18, v uint16) {
	i := (a & 017) >> 1
	if (a >= 0772300) && (a < 0772320) {
		kt.pages[i].pdr = v
		return
	}
	if (a >= 0772340) && (a < 0772360) {
		kt.pages[i].par = v
		return
	}
	if (a >= 0777600) && (a < 0777620) {
		kt.pages[i+8].pdr = v
		return
	}
	if (a >= 0777640) && (a < 0777660) {
		kt.pages[i+8].par = v
		return
	}
	fmt.Printf("mmu::write16 write to invalid address %06o\n", a)
	panic(trap{INTBUS})
}
