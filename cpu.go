package main

import (
	"fmt"
	"os"
	"runtime"
)

type KB11 struct {
	unibus UNIBUS
	mmu    KT11

	pc uint16    // holds R[7] during instruction execution
	R  [8]uint16 // R0-R7

	psw                                         uint16    // processor status word
	stackpointer                                [4]uint16 // Alternate R6 (kernel, super, illegal, user)
	stacklimit, switchregister, displayregister uint16

	interrupts [8]struct{ vec, pri uint16 }

	print bool
}

func (kb *KB11) Reset() {
	kb.unibus.reset()
}

func (kb *KB11) Run() error {
	for {
		kb.run()
	}
	return nil
}

func (kb *KB11) run() {
	defer func() {
		t := recover()
		switch t := t.(type) {
		case trap:
			kb.trapat(t.vec)
		case interrupt:
			defer fmt.Printf("interrupt queued: vec: %03o pri: %03o\n", t.vec, t.pri)
			if t.vec&1 == 1 {
				panic("Thou darst calling interrupt() with an odd vector number?")
			}
			// fast path
			if kb.interrupts[0].vec == 0 {
				kb.interrupts[0].vec = t.vec
				kb.interrupts[0].pri = t.pri
				return
			}
			var i int
			for ; i < len(kb.interrupts); i++ {
				if kb.interrupts[i].vec == 0 || kb.interrupts[i].pri < t.pri {
					break
				}
			}
			for ; i < len(kb.interrupts); i++ {
				if kb.interrupts[i].vec == 0 || kb.interrupts[i].vec >= t.vec {
					break
				}
			}
			if i == len(kb.interrupts) {
				panic("interrupt table full")
			}
			for j := i + 1; j < len(kb.interrupts); j++ {
				kb.interrupts[j] = kb.interrupts[j-1]
			}
			kb.interrupts[i].vec = t.vec
			kb.interrupts[i].pri = t.pri
		default:
			panic(t)
		}
	}()

	for {
		if kb.interrupts[0].vec > 0 && kb.interrupts[0].pri >= kb.priority() {
			fmt.Printf("interrupt: vec: %03o pri: %03o\n", kb.interrupts[0].vec, kb.interrupts[0].pri)
			kb.trapat(kb.interrupts[0].vec)
			copy(kb.interrupts[:8], kb.interrupts[1:])
			kb.interrupts[7].vec = 0
			kb.interrupts[7].pri = 0
		}

		kb.step()
		kb.unibus.rk11.step()
		kb.unibus.cons.step()
		kb.unibus.lineclock.tick()
	}
}

// Load loads words into memory starting at offset bypassing the mmu.
func (kb *KB11) Load(offset addr18, words ...uint16) {
	for i, w := range words {
		kb.unibus.write16(offset+addr18(i*2), w)
	}
}

func (kb *KB11) step() {
	kb.pc = kb.R[7]
	instr := kb.fetch16()

	if kb.print {
		kb.printstate()
	}

	switch instr >> 12 { // xxSSDD Mostly double operand instructions
	case 0: // 00xxxx mixed group
		switch instr >> 8 { // 00xxxx 8 bit instructions first (branch & JSR)
		case 0: // 000xXX Misc zero group
			switch instr >> 6 { // 000xDD group (4 case full decode)
			case 0: // 0000xx group
				switch instr {
				case 0: // HALT 000000
					println("HALT")
					kb.printstate()
					os.Exit(1)
				case 1: // WAIT 000001
					kb.WAIT()
					return
				case 3: // BPT  000003
					kb.trapat(014) // Trap 14 - BPT
					return
				case 4: // IOT  000004
					kb.trapat(020)
					return
				case 5: // RESET 000005
					kb.RESET()
					return
				case 2, 6: // RTI 000002, RTT 000006
					kb.RTT()
					return
				case 7: // MFPT
					kb.MFPT()
					return
				default: // We don't know this 0000xx instruction
					fmt.Printf("unknown 0000xx instruction\n")
					kb.printstate()
					kb.trapat(INTINVAL)
					return
				}
			case 1: // JMP 0001DD
				kb.JMP(instr)
				return
			case 2: // 00002xR single register group
				switch (instr >> 3) & 7 { // 00002xR register or CC
				case 0: // RTS 00020R
					kb.RTS(instr)
					return
				case 3: // SPL 00023N
					kb.writePSW((kb.psw & 0xf81f) | ((instr & 7) << 5))
					return
				case 4, 5: // CLR CC 00024C Part 1 without N, CLR CC 00025C Part 2 with N
					kb.writePSW(kb.psw & (^instr & 017))
					return
				case 6, 7: // SET CC 00026C Part 1 without N, SET CC 00027C Part 2 with N
					kb.writePSW(kb.psw | (instr & 017))
					return
				default: // We don't know this 00002xR instruction
					fmt.Printf("unknown 0002xR instruction\n")
					kb.printstate()
					panic(trap{INTINVAL})

				}
			case 3: // SWAB 0003DD
				kb.SWAB(instr)
				return
			default:
				fmt.Printf("unknown 000xDD instruction\n")
				kb.printstate()
				panic(trap{INTINVAL})
			}
		case 1: // BR 0004 offset
			kb.branch(instr)
			return
		case 2: // BNE 0010 offset
			if !kb.z() {
				kb.branch(instr)
			}
			return
		case 3: // BEQ 0014 offset
			if kb.z() {
				kb.branch(instr)
			}
			return
		case 4: // BGE 0020 offset
			if !(kb.n() != kb.v()) {
				kb.branch(instr)
			}
			return
		case 5: // BLT 0024 offset
			if kb.n() != kb.v() {
				kb.branch(instr)
			}
			return
		case 6: // BGT 0030 offset
			if (!(kb.n() != kb.v())) && (!kb.z()) {
				kb.branch(instr)
			}
			return
		case 7: // BLE 0034 offset
			if (kb.n() != kb.v()) || kb.z() {
				kb.branch(instr)
			}
			return
		case 8, 9: // JSR 004RDD In two parts, JSR 004RDD continued (9 bit instruction so use 2 x 8 bit
			kb.JSR(instr)
			return
		default: // Remaining 0o00xxxx instructions where xxxx >= 05000
			switch instr >> 6 { // 00xxDD
			case 050: // CLR 0050DD
				kb.CLR(2, instr)
				return
			case 051: // COM 0051DD
				kb.COM(2, instr)
				return
			case 052: // INC 0052DD
				kb.INC(2, instr)
				return
			case 053: // DEC 0053DD
				kb.DEC(2, instr)
				return
			case 054: // NEG 0054DD
				kb.NEG(2, instr)
				return
			case 055: // ADC 0055DD
				kb.ADC(2, instr)
				return
			case 056: // SBC 0056DD
				kb.SBC(2, instr)
				return
			case 057: // TST 0057DD
				kb.TST(2, instr)
				return
			case 060: // ROR 0060DD
				kb.ROR(2, instr)
				return
			case 061: // ROL 0061DD
				kb.ROL(2, instr)
				return
			case 062: // ASR 0062DD
				kb.ASR(2, instr)
				return
			case 063: // ASL 0063DD
				kb.ASL(2, instr)
				return
			case 064: // MARK 0064nn
				kb.MARK(instr)
				return
			case 065: // MFPI 0065SS
				kb.MFPI(instr)
				return
			case 066: // MTPI 0066DD
				kb.MTPI(instr)
				return
			case 067: // SXT 0067DD
				kb.SXT(instr)
				return
			default: // We don't know this 0o00xxDD instruction
				fmt.Printf("unknown 00xxDD instruction\n")
				kb.printstate()
				panic(trap{INTINVAL})
			}
		}
	case 1: // MOV  01SSDD
		kb.MOV(2, instr)
		return
	case 2: // CMP 02SSDD
		kb.CMP(2, instr)
		return
	case 3: // BIT 03SSDD
		kb.BIT(2, instr)
		return
	case 4: // BIC 04SSDD
		kb.BIC(2, instr)
		return
	case 5: // BIS 05SSDD
		kb.BIS(2, instr)
		return
	case 6: // ADD 06SSDD
		kb.ADD(instr)
		return
	case 7: // 07xRSS instructions
		switch (instr >> 9) & 7 { // 07xRSS
		case 0: // MUL 070RSS
			kb.MUL(instr)
			return
		case 1: // DIV 071RSS
			kb.DIV(instr)
			return
		case 2: // ASH 072RSS
			kb.ASH(instr)
			return
		case 3: // ASHC 073RSS
			kb.ASHC(instr)
			return
		case 4: // XOR 074RSS
			kb.XOR(instr)
			return
		case 7: // SOB 077Rnn
			kb.SOB(instr)
			return
		default: // We don't know this 07xRSS instruction
			fmt.Printf("unknown 07xRSS instruction\n")
			kb.printstate()
			panic(trap{INTINVAL})
		}
	case 8: // 10xxxx instructions
		switch (instr >> 8) & 0xf { // 10xxxx 8 bit instructions first
		case 0: // BPL 1000 offset
			if !kb.n() {
				kb.branch(instr)
			}
			return
		case 1: // BMI 1004 offset
			if kb.n() {
				kb.branch(instr)
			}
			return
		case 2: // BHI 1010 offset
			if (!kb.c()) && (!kb.z()) {
				kb.branch(instr)
			}
			return
		case 3: // BLOS 1014 offset
			if kb.c() || kb.z() {
				kb.branch(instr)
			}
			return
		case 4: // BVC 1020 offset
			if !kb.v() {
				kb.branch(instr)
			}
			return
		case 5: // BVS 1024 offset
			if kb.v() {
				kb.branch(instr)
			}
			return
		case 6: // BCC 1030 offset
			if !kb.c() {
				kb.branch(instr)
			}
			return
		case 7: // BCS 1034 offset
			if kb.c() {
				kb.branch(instr)
			}
			return
		case 8: // EMT 1040 operand
			kb.trapat(030) // Trap 30 - EMT instruction
			return
		case 9: // TRAP 1044 operand
			kb.trapat(034) // Trap 34 - TRAP instruction
			return
		default: // Remaining 10xxxx instructions where xxxx >= 05000
			switch (instr >> 6) & 077 { // 10xxDD group
			case 050: // CLRB 1050DD
				kb.CLR(1, instr)
				return
			case 051: // COMB 1051DD
				kb.COM(1, instr)
				return
			case 052: // INCB 1052DD
				kb.INC(1, instr)
				return
			case 053: // DECB 1053DD
				kb.DEC(1, instr)
				return
			case 054: // NEGB 1054DD
				kb.NEG(1, instr)
				return
			case 055: // ADCB 01055DD
				kb.ADC(1, instr)
				return
			case 056: // SBCB 01056DD
				kb.SBC(1, instr)
				return
			case 057: // TSTB 1057DD
				kb.TST(1, instr)
				return
			case 060: // RORB 1060DD
				kb.ROR(1, instr)
				return
			case 061: // ROLB 1061DD
				kb.ROL(1, instr)
				return
			case 062: // ASRB 1062DD
				kb.ASR(1, instr)
				return
			case 063: // ASLB 1063DD
				kb.ASL(1, instr)
				return
			// case 0o64: // MTPS 1064SS
			// case 0o65: // MFPD 1065DD
			// case 0o66: // MTPD 1066DD
			// case 0o67: // MTFS 1064SS
			default: // We don't know this 0o10xxDD instruction
				fmt.Printf("unknown 0o10xxDD instruction\n")
				kb.printstate()
				panic(trap{INTINVAL})
			}
		}
	case 9: // MOVB 11SSDD
		kb.MOV(1, instr)
		return
	case 10: // CMPB 12SSDD
		kb.CMP(1, instr)
		return
	case 11: // BITB 13SSDD
		kb.BIT(1, instr)
		return
	case 12: // BICB 14SSDD
		kb.BIC(1, instr)
		return
	case 13: // BISB 15SSDD
		kb.BIS(1, instr)
		return
	case 14: // SUB 16SSDD
		kb.SUB(instr)
		return
	case 15:
		if instr == 0170011 {
			// SETD ; not needed by UNIX, but used; therefore ignored
			return
		}
	default: // 15  17xxxx FPP instructions
		fmt.Printf("invalid 17xxxx FPP instruction\n")
		kb.printstate()
		panic(trap{INTINVAL})
	}
}

func (kb *KB11) RESET() {
	if kb.currentmode() > 0 {
		// RESET is ignored outside of kernel mode
		return
	}
	kb.unibus.reset()
}

func (kb *KB11) WAIT() {
	fmt.Printf("WAIT\n")
	kb.printstate()
	runtime.Breakpoint()
}

// RTI 000004, RTT 000006
func (kb *KB11) RTT() {
	kb.R[7] = kb.pop()
	psw := kb.pop()
	psw &= 0xf8ff
	if kb.currentmode() > 0 { // user / super restrictions
		// keep SPL and allow lower only for modes and register set
		psw = (psw & 0xf81f) | (psw & 0xf8e0)
	}
	kb.writePSW(psw)
}

// MFPT 000007
func (kb *KB11) MFPT() {
	kb.trapat(010) // not a PDP11/44
}

// JMP 0001DD
func (kb *KB11) JMP(instr uint16) {
	if ((instr >> 3) & 7) == 0 {
		// Registers don't have a virtual address so trap!
		fmt.Printf("JMP called on register\n")
		kb.printstate()
		os.Exit(1)
	}
	kb.R[7] = kb.DA(instr)
}

// RTS 00020R
func (kb *KB11) RTS(instr uint16) {
	reg := instr & 7
	kb.R[7] = kb.R[reg]
	kb.R[reg] = kb.pop()
}

// SWAB 0003DD
func (kb *KB11) SWAB(instr uint16) {
	da := kb.DA(instr)
	dst := kb.read(2, da)
	dst = (dst << 8) | (dst >> 8)
	kb.write(2, da, dst)
	kb.psw &= 0xFFF0
	if dst&0xff00 == 0 {
		kb.psw |= FLAGZ
	}
	if dst&0x80 == 0x80 {
		kb.psw |= FLAGN
	}
}

func (kb *KB11) branch(instr uint16) {
	switch instr & 0x80 {
	case 0x80:
		kb.R[7] += (instr | 0xff00) << 1
	default:
		kb.R[7] += (instr & 0xff) << 1
	}
}

// JSR 004RDD
func (kb *KB11) JSR(instr uint16) {
	if ((instr >> 3) & 7) == 0 {
		fmt.Printf("JSR called on register\n")
		kb.printstate()
		os.Exit(1)
	}
	dst := kb.DA(instr)
	reg := (instr >> 6) & 7
	kb.push(kb.R[reg])
	kb.R[reg] = kb.R[7]
	kb.R[7] = dst
}

// CLR 0050DD, CLRB 1050DD
func (kb *KB11) CLR(l int, instr uint16) {
	kb.write(l, kb.DA(instr), 0)
	kb.psw &= 0xFFF0
	kb.psw |= FLAGZ
}

// COM 0051DD, COMB 1051DD
func (kb *KB11) COM(l int, instr uint16) {
	da := kb.DA(instr)
	dst := kb.read(l, da) ^ max(l)
	kb.write(l, da, dst)
	kb.psw &= 0xFFF0
	kb.psw |= FLAGC
	if dst&msb(l) == 0 {
		kb.psw |= FLAGN
	}
	if dst == 0 {
		kb.psw |= FLAGZ
	}
}

// INC 0052DD, INCB 1052DD
func (kb *KB11) INC(l int, instr uint16) {
	da := kb.DA(instr)
	dst := kb.read(l, da) + 1
	kb.write(l, da, dst)
	kb.psw &= 0xFFF1
	kb.setNZV(l, dst)
}

// DEC 0053DD, DECB 1053DD
func (kb *KB11) DEC(l int, instr uint16) {
	da := kb.DA(instr)
	dst := kb.read(l, da) - 1
	kb.write(l, da, dst)
	kb.setNZV(l, dst)
}

// NEG 0054DD, NEGB 1054DD
func (kb *KB11) NEG(l int, instr uint16) {
	da := kb.DA(instr)
	dst := kb.read(l, da)
	dst = (-dst) & max(l)
	kb.write(l, da, dst)
	kb.psw &= 0xFFF0
	if dst&msb(l) > 0 {
		kb.psw |= FLAGN
	}
	if dst == 0 {
		kb.psw |= FLAGZ
	} else {
		kb.psw |= FLAGC
	}
	if dst == msb(l) {
		kb.psw |= FLAGV
	}
}

func (kb *KB11) ADC(l int, instr uint16) {
	da := kb.DA(instr)
	dst := kb.read(l, da)
	switch kb.psw & FLAGC {
	case FLAGC:
		kb.psw &= 0xFFF0
		if (dst+1)&msb(l) > 0 {
			kb.psw |= FLAGN
		}
		if dst == max(l) {
			kb.psw |= FLAGZ
		}
		if dst == 0077777 {
			kb.psw |= FLAGV
		}
		if dst == 0177777 {
			kb.psw |= FLAGC
		}
		kb.write(l, da, (dst+1)&max(l))
	default:
		kb.psw &= 0xFFF0
		if dst&msb(l) > 0 {
			kb.psw |= FLAGN
		}
		if dst == 0 {
			kb.psw |= FLAGZ
		}
	}
}

func (kb *KB11) SBC(l int, instr uint16) {
	da := kb.DA(instr)
	dst := kb.read(l, da)
	switch kb.psw & FLAGC {
	case FLAGC:
		kb.psw &= 0xFFF0
		if (dst-1)&msb(l) > 0 {
			kb.psw |= FLAGN
		}
		if dst-1 == 0 {
			kb.psw |= FLAGZ
		}
		if dst&msb(l) > 0 {
			kb.psw |= FLAGV
		}
		if dst == 0 {
			kb.psw |= FLAGC
		}
		kb.write(l, da, (dst-1)&max(l))
	default:
		kb.psw &= 0xFFF0 | FLAGC
		if dst&msb(l) > 0 {
			kb.psw |= FLAGN
		}
		if dst == 0 {
			kb.psw |= FLAGZ
		}
		if dst&max(l) == 0 {
			kb.psw |= FLAGZ
		}
	}
}

// TST 0057DD, TSTB 1057DD
func (kb *KB11) TST(l int, instr uint16) {
	dst := kb.read(l, kb.DA(instr))
	kb.psw &= 0xFFF0
	if dst&msb(l) > 0 {
		kb.psw |= FLAGN
	}
	if dst == 0 {
		kb.psw |= FLAGZ
	}
}

func (kb *KB11) ROR(l int, instr uint16) {
	da := kb.DA(instr)
	dst := kb.read(l, da)
	result := dst >> 1
	if kb.c() {
		// shift carry in from the left
		result |= msb(l)
	}
	kb.write(l, da, result)
	kb.psw &= 0xFFF0
	if dst&1 > 0 {
		// shift lsb into carry
		kb.psw |= FLAGC
	}

	if result&msb(l) > 0 {
		kb.psw |= FLAGN
	}
	if result&max(l) == 0 {
		kb.psw |= FLAGZ
	}
	if (result&msb(l) > 0) != (dst&1 > 0) {
		kb.psw |= FLAGV
	}
}

func (kb *KB11) ROL(l int, instr uint16) {
	da := kb.DA(instr)
	dst := kb.read(l, da)
	result := dst<<1 | kb.psw&1
	kb.write(l, da, result)
	kb.psw &= 0xFFF0
	if dst&msb(l) > 0 {
		kb.psw |= FLAGC
	}
	if result&msb(l) > 0 {
		kb.psw |= FLAGN
	}
	if result&max(l) == 0 {
		kb.psw |= FLAGZ
	}
	if kb.c() != kb.n() {
		kb.psw |= FLAGV
	}
}

func (kb *KB11) ASR(l int, instr uint16) {
	da := kb.DA(instr)
	dst := kb.read(l, da)
	result := (dst&msb(l) | dst>>1) & max(l)
	kb.write(l, da, result)
	kb.psw &= 0xFFF0
	if result&msb(l) > 0 {
		kb.psw |= FLAGN
	}
	if result&max(l) == 0 {
		kb.psw |= FLAGZ
	}
	if dst&1 == 1 {
		kb.psw |= FLAGC
	}
	if kb.n() != kb.c() {
		kb.psw |= FLAGV
	}
}

func (kb *KB11) ASL(l int, instr uint16) {
	da := kb.DA(instr)
	dst := kb.read(l, da)
	result := (dst << 1) & max(l)
	kb.write(l, da, result)
	kb.psw &= 0xFFF0
	if result&msb(l) > 0 {
		kb.psw |= FLAGN
	}
	if result&max(l) == 0 {
		kb.psw |= FLAGZ
	}
	if result&msb(l) > 0 {
		kb.psw |= FLAGC
	}
	if kb.c() != kb.n() {
		kb.psw |= FLAGV
	}
}

// MARK 0064NN
func (kb *KB11) MARK(instr uint16) {
	kb.R[6] = kb.R[7] + ((instr & 077) << 1)
	kb.R[7] = kb.R[5]
	kb.R[5] = kb.pop()
}

// MFPI 0065SS
func (kb *KB11) MFPI(instr uint16) {
	var uval uint16
	if !(instr&0x38 > 0) {
		reg := instr & 7
		if (reg != 6) || (kb.currentmode() == kb.previousmode()) {
			uval = kb.R[reg]
		} else {
			uval = kb.stackpointer[kb.previousmode()]
		}
	} else {
		da := kb.DA(instr)
		addr := kb.mmu.decode(false, da, kb.previousmode())
		uval = kb.unibus.read16(addr)
	}
	kb.push(uval)
	kb.setNZ(2, uval)
}

// MTPI 0066DD
func (kb *KB11) MTPI(instr uint16) {
	uval := kb.pop()
	if !(instr&0x38 > 0) {
		reg := instr & 7
		if (reg != 6) || (kb.currentmode() == kb.previousmode()) {
			kb.R[reg] = uval
		} else {
			kb.stackpointer[kb.previousmode()] = uval
		}
		da := kb.DA(instr)
		addr := kb.mmu.decode(true, da, kb.previousmode())
		kb.unibus.write16(addr, uval)
	}
	kb.setNZ(2, uval)
}

// SXT 0067DD
func (kb *KB11) SXT(instr uint16) {
	if kb.n() {
		kb.write(2, kb.DA(instr), 0xffff)
		kb.psw &^= FLAGZ
	} else {
		kb.write(2, kb.DA(instr), 0)
		kb.psw |= FLAGZ
	}
	kb.psw &^= FLAGV
}

// MOV 01SSDD, MOVB 11SSDD
func (kb *KB11) MOV(len int, instr uint16) {
	src := kb.read(len, kb.SA(instr))
	if !(instr&0x38 > 0) && (len == 1) {
		if src&0200 > 0 {
			// Special case: movb sign extends register to word size)
			src |= 0xFF00
		}
		kb.R[instr&7] = src
		kb.setNZ(len, src)
		return
	}
	kb.write(len, kb.DA(instr), src)
	kb.setNZ(len, src)
}

// CMP 02SSDD, CMPB 12SSDD
func (kb *KB11) CMP(l int, instr uint16) {
	src := kb.read(l, kb.SA(instr))
	da := kb.DA(instr)
	dst := kb.read(l, da)
	result := src + (^(dst) + 1)
	kb.psw &= 0xFFF0
	if result == 0 {
		kb.psw |= FLAGZ
	}
	if result&msb(l) > 0 {
		kb.psw |= FLAGN
	}
	if (src^dst)&msb(l) > 0 && !((dst^result)&msb(l) > 0) {
		kb.psw |= FLAGV
	}
	if src < dst {
		kb.psw |= FLAGC
	}

}

// BIT 03SSDD, BITB 13SSDD
func (kb *KB11) BIT(l int, instr uint16) {
	src := kb.read(l, kb.SA(instr))
	dst := kb.read(l, kb.DA(instr))
	result := src & dst
	kb.setNZ(l, result)
}

// BIC 04SSDD, BICB 14SSDD
func (kb *KB11) BIC(l int, instr uint16) {
	src := kb.read(l, kb.SA(instr))
	da := kb.DA(instr)
	dst := kb.read(l, da)
	dst = (max(l) ^ src) & dst
	kb.write(l, da, dst)
	kb.setNZ(l, dst)
}

// BIS 05SSDD, BISB 15SSDD
func (kb *KB11) BIS(l int, instr uint16) {
	src := kb.read(l, kb.SA(instr))
	da := kb.DA(instr)
	dst := src | kb.read(l, da)
	kb.write(l, da, dst)
	kb.setNZ(l, dst)
}

// ADD 06SSDD
func (kb *KB11) ADD(instr uint16) {
	src := kb.read(2, kb.SA(instr))
	da := kb.DA(instr)
	dst := kb.read(2, da)
	sum := src + dst
	kb.write(2, da, sum)
	kb.psw &= 0xFFF0
	if sum == 0 {
		kb.psw |= FLAGZ
	}
	if sum&msb(2) > 0 {
		kb.psw |= FLAGN
	}
	if (!((src^dst)&msb(2) > 0)) && ((dst^sum)&msb(2) > 0) {
		kb.psw |= FLAGV
	}
	if int(src)+int(dst) > 0xffff {
		kb.psw |= FLAGC
	}
}

// SUB 16SSDD
func (kb *KB11) SUB(instr uint16) {
	// mask off top bit of instr so SA computes L=2
	src := kb.read(2, kb.SA(instr&0077777))
	da := kb.DA(instr)
	dst := kb.read(2, da)
	result := dst - src
	kb.write(2, da, result)
	kb.psw &= 0xFFF0
	if result == 0 {
		kb.psw |= FLAGZ
	}
	if result&msb(2) > 0 {
		kb.psw |= FLAGN
	}
	if ((src^dst)&msb(2) > 0) && (!((dst^result)&msb(2) > 0)) {
		kb.psw |= FLAGV
	}
	if src > dst {
		kb.psw |= FLAGC
	}
}

func (kb *KB11) MUL(instr uint16) {
	reg := (instr >> 6) & 7
	val1 := int32(kb.R[reg])
	if val1 < 0 {
		val1 = -val1
	}
	da := kb.DA(instr)
	val2 := int32(kb.read(2, da))
	if val2 < 0 {
		val2 = -val2
	}
	sval := val1 * val2
	kb.R[(instr>>6)&7] = uint16(sval >> 16)
	kb.R[((instr>>6)&7)|1] = uint16(sval & 0xFFFF)
	kb.psw &= 0xFFF0
	if sval < 0 {
		kb.psw |= FLAGN
	}
	if sval == 0 {
		kb.psw |= FLAGZ
	}
	if (sval < (1 << 15)) || (sval >= ((1 << 15) - 1)) {
		kb.psw |= FLAGC
	}
	// kb.printstate()
}

func (kb *KB11) DIV(instr uint16) {
	reg := (instr >> 6) & 7
	val1 := int32(kb.R[reg])<<16 | int32(kb.R[reg|1])
	val2 := int32(kb.read(2, kb.DA(instr)))
	kb.psw &= 0xFFF0
	if val2 == 0 {
		kb.psw |= FLAGC
		return
	}
	if (val1 / val2) >= 0x10000 {
		kb.psw |= FLAGV
		return
	}
	kb.R[reg] = uint16(val1 / val2)
	kb.R[reg|1] = uint16(val1 % val2)
	if kb.R[reg] == 0 {
		kb.psw |= FLAGZ
	}
	if kb.R[reg]&0100000 > 0 {
		kb.psw |= FLAGN
	}
	if val1 == 0 {
		kb.psw |= FLAGV
	}
	// kb.printstate()
}

// ASH 072RSS
func (kb *KB11) ASH(instr uint16) {
	reg := kb.R[(instr>>6)&7]
	shift := kb.read(2, kb.DA(instr)) & 077
	kb.psw &= 0xFFF0
	var sval uint16
	if shift&040 > 0 {
		shift = (077 ^ shift) + 1
		if reg&0100000 > 0 {
			sval = 0xFFFF ^ (0xFFFF >> shift)
			sval |= reg >> shift
		} else {
			sval = reg >> shift
		}
		if reg&(1<<(shift-1)) > 0 {
			kb.psw |= FLAGC
		}
	} else {
		sval = reg << shift
		if reg&(1<<(16-shift)) > 0 {
			kb.psw |= FLAGC
		}
	}
	kb.R[(instr>>6)&7] = sval
	if sval == 0 {
		kb.psw |= FLAGZ
	}
	if sval&msb(2) > 0 {
		kb.psw |= FLAGN
	}
	if (sval&msb(2))^(reg&msb(2)) > 0 {
		kb.psw |= FLAGV
	}
}

func (kb *KB11) ASHC(instr uint16) {
	val1 := uint32(kb.R[(instr>>6)&7])<<16 | uint32(kb.R[((instr>>6)&7)|1])
	da := kb.DA(instr)
	val2 := kb.read(2, da) & 077
	kb.psw &= 0xFFF0
	var sval uint32
	if val2&040 > 0 {
		val2 = (077 ^ val2) + 1
		if val1&0x80000000 > 0 {
			sval = 0xFFFFFFFF ^ (0xFFFFFFFF >> val2)
			sval |= uint32(val1 >> val2)
		} else {
			sval = val1 >> val2
		}
		if val1&(1<<(val2-1)) > 0 {
			kb.psw |= FLAGC
		}
	} else {
		sval = (val1 << val2) & 0xFFFFFFFF
		if val1&(1<<(32-val2)) > 0 {
			kb.psw |= FLAGC
		}
	}
	kb.R[(instr>>6)&7] = uint16(sval >> 16)
	kb.R[((instr>>6)&7)|1] = uint16(sval)
	if sval == 0 {
		kb.psw |= FLAGZ
	}
	if sval&0x80000000 > 0 {
		kb.psw |= FLAGN
	}
	if (sval&0x80000000)^(val1&0x80000000) > 0 {
		kb.psw |= FLAGV
	}
}

// XOR 074RDD
func (kb *KB11) XOR(instr uint16) {
	reg := kb.R[(instr>>6)&7]
	da := kb.DA(instr)
	dst := reg ^ kb.read(2, da)
	kb.write(2, da, dst)
	kb.setNZ(2, dst)
}

// SOB 077RNN
func (kb *KB11) SOB(instr uint16) {
	reg := (instr >> 6) & 7
	kb.R[reg]--
	if kb.R[reg] > 0 {
		kb.R[7] -= (instr & 077) << 1
	}
}

func (kb *KB11) trapat(vec uint16) {
	if vec&1 > 0 {
		fmt.Printf("Thou darst calling trapat() with an odd vector number?\n")
		os.Exit(1)
	}

	fmt.Printf("trap: vec: %03o\n", vec)
	if vec == 0220 {
		//	kb.print = true
	}

	psw := kb.psw
	kb.kernelmode()
	kb.push(psw)
	kb.push(kb.R[7])

	kb.R[7] = kb.read16(vec)
	kb.writePSW(kb.read16(vec+2) | (kb.previousmode() << 12))
}

func (kb *KB11) fetch16() uint16 {
	val := kb.read16(kb.R[7])
	kb.R[7] += 2
	return val
}

func (kb *KB11) push(v uint16) {
	kb.R[6] -= 2
	kb.write16(kb.R[6], v)
}

func (kb *KB11) pop() uint16 {
	val := kb.read16(kb.R[6])
	kb.R[6] += 2
	return val
}

func (kb *KB11) SA(instr uint16) uint16 {
	// reconstruct L00SSDD as L0000SS
	instr = (instr & (1 << 15)) | ((instr >> 6) & 077)
	return kb.DA(instr)
}

func (kb *KB11) DA(instr uint16) uint16 {
	v := instr & 077
	l := (2 - (instr >> 15))

	if (v & 070) == 000 {
		return 0170000 | (v & 7)
	}
	if ((v & 7) >= 6) || (v&010) > 0 {
		l = 2
	}
	var addr uint16
	switch v & 060 {
	case 000:
		v &= 7
		addr = kb.R[v&7]
	case 020:
		addr = kb.R[v&7]
		kb.R[v&7] += l
	case 040:
		kb.R[v&7] -= l
		addr = kb.R[v&7]
	case 060:
		addr = kb.fetch16()
		addr += kb.R[v&7]
	}
	if v&010 > 0 {
		addr = kb.read16(addr)
	}
	return addr
}

func (kb *KB11) read(l int, a uint16) uint16 {
	if (a & 0177770) == 0170000 {
		return kb.R[a&7] & max(l)
	}
	if l == 2 {
		return kb.read16(a)
	}
	switch a & 1 {
	case 1:
		return kb.read16(a&^1) >> 8
	default:
		return kb.read16(a) & 0xFF
	}
}

func (kb *KB11) read16(va uint16) uint16 {
	a := kb.mmu.decode(false, va, kb.currentmode())
	switch a {
	case 0777776:
		return kb.psw
	case 0777774:
		return kb.stacklimit
	case 0777570:
		return 0173030 // kb.switchregister
	default:
		if a == 0140000 {
			kb.printstate()
		}
		return kb.unibus.read16(a)
	}
}

func (kb *KB11) write(l int, a, v uint16) {
	if (a & 0177770) == 0170000 {
		r := a & 7
		if l == 2 {
			kb.R[r] = v
		} else {
			kb.R[r] &= 0xFF00
			kb.R[r] |= v
		}
		return
	}
	if l == 2 {
		kb.write16(a, v)
		return
	}
	switch a & 1 {
	case 1:
		mem := v<<8 | kb.read16(a&^1)&0xff
		kb.write16(a&^1, mem)
	default:
		mem := kb.read16(a)&0xff00 | v&0xff
		kb.write16(a, mem)
	}
}

func (kb *KB11) write16(va, v uint16) {
	a := kb.mmu.decode(true, va, kb.currentmode())
	switch a {
	case 0777776:
		kb.writePSW(v)
	case 0777774:
		kb.stacklimit = v
	case 0777570:
		kb.displayregister = v
	default:
		if a == 0140000 {
			kb.printstate()
		}
		kb.unibus.write16(a, v)
	}
}

// Set N & Z clearing V (C unchanged)
func (kb *KB11) setNZ(l int, v uint16) {
	kb.psw &= 0xfff1
	if v&msb(l) > 0 {
		kb.psw |= FLAGN
	}
	if v == 0 {
		kb.psw |= FLAGZ
	}
}

// Set N, Z & V (C unchanged)
func (kb *KB11) setNZV(l int, v uint16) {
	kb.setNZ(l, v)
	if v == msb(l)-1 {
		kb.psw |= FLAGV
	}
}

// Set N, Z & C clearing V
func (kb *KB11) setNZC(l int, v uint16) {
	kb.psw &= 0xFFF0
	kb.psw |= FLAGC
	if v&max(l) == 0 {
		kb.psw |= FLAGZ
	}
	if v&msb(l) > 0 {
		kb.psw |= FLAGN
	}
}

func (kb *KB11) writePSW(psw uint16) {
	kb.stackpointer[kb.currentmode()] = kb.R[6]
	kb.psw = psw
	kb.R[6] = kb.stackpointer[kb.currentmode()]
}

// currentmode returns the current cpu mode.
// 0: kernel, 1: supervisor, 2: illegal, 3: user
func (kb *KB11) currentmode() uint16 { return kb.psw >> 14 }

// previousmode returns the previous cpu mode.
// 0: kernel, 1: supervisor, 2: illegal, 3: user
func (kb *KB11) previousmode() uint16 { return (kb.psw >> 12) & 3 }

func (kb *KB11) kernelmode() { kb.writePSW((kb.psw & 0007777) | (kb.currentmode() << 12)) }

// priority returns the current CPU interrupt priority.
func (kb *KB11) priority() uint16 { return (kb.psw >> 5) & 7 }

const (
	FLAGC = 1
	FLAGV = 2
	FLAGZ = 4
	FLAGN = 8
)

func (kb *KB11) n() bool { return kb.psw&FLAGN == FLAGN }
func (kb *KB11) z() bool { return kb.psw&FLAGZ == FLAGZ }
func (kb *KB11) v() bool { return kb.psw&FLAGV == FLAGV }
func (kb *KB11) c() bool { return kb.psw&FLAGC == FLAGC }

func (kb *KB11) printstate() {
	prev := func() string {
		switch kb.previousmode() {
		case 3:
			return "u"
		default:
			return "k"
		}
	}

	curr := func() string {
		switch kb.currentmode() {
		case 3:
			return "U"
		default:
			return "K"
		}
	}

	n := func() string {
		if kb.n() {
			return "N"
		}
		return " "
	}

	z := func() string {
		if kb.z() {
			return "Z"
		}
		return " "
	}

	v := func() string {
		if kb.v() {
			return "V"
		}
		return " "
	}

	c := func() string {
		if kb.c() {
			return "C"
		}
		return " "
	}

	fmt.Printf("R0 %06o R1 %06o R2 %06o R3 %06o R4 %06o R5 %06o R6 %06o R7 %06o\n",
		kb.R[0], kb.R[1], kb.R[2], kb.R[3], kb.R[4], kb.R[5], kb.R[6], kb.R[7])
	fmt.Printf("[%s%s%s%s%s%s", prev(), curr(), n(), z(), v(), c())
	fmt.Printf("]  instr %06o: %06o\t ", kb.pc, kb.read16(kb.pc))
	kb.disasm(kb.pc)
	fmt.Println()
}

func max(l int) uint16 {
	if l == 2 {
		return 0xffff
	}
	return 0xff
}

func msb(l int) uint16 {
	if l == 2 {
		return 0x8000
	}
	return 0x80
}

func mask(l int) uint16 {
	if l == 2 {
		return 0x7fff
	}
	return 0x7f
}

func xor(a, b bool) bool { return a != b }
