package main

import (
	"fmt"
	"os"
)

type KB11 struct {
	unibus UNIBUS

	pc uint16    // holds R[7] during instruction execution
	R  [8]uint16 // R0-R7

	psw          uint16    // processor status word
	stackpointer [4]uint16 // Alternate R6 (kernel, super, illegal, user)
}

func (kb *KB11) Reset() {

}

func (kb *KB11) Run() error {
	for {
		kb.step()
	}
	return nil
}

func (kb *KB11) step() {
	kb.pc = kb.R[7]
	instr := kb.fetch16()

	kb.printstate()

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
					// sched_yield();
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
			kb.branch(instr & 0xff)
			return
		case 2: // BNE 0010 offset
			if !kb.z() {
				kb.branch(instr & 0xff)
			}
			return
		case 3: // BEQ 0014 offset
			if kb.z() {
				kb.branch(instr & 0xff)
			}
			return
		case 4: // BGE 0020 offset
			if !(kb.n() != kb.v()) {
				kb.branch(instr & 0xFF)
			}
			return
		case 5: // BLT 0024 offset
			if kb.n() != kb.v() {
				kb.branch(instr & 0xFF)
			}
			return
		case 6: // BGT 0030 offset
			if (!(kb.n() != kb.v())) && (!kb.z()) {
				kb.branch(instr & 0xFF)
			}
			return
		case 7: // BLE 0034 offset
			if (kb.n() != kb.v()) || kb.z() {
				kb.branch(instr & 0xFF)
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
		BIC < 2 > (instr)
		return
	case 5: // BIS 05SSDD
		BIS < 2 > (instr)
		return
	case 6: // ADD 06SSDD
		ADD(instr)
		return
	case 7: // 07xRSS instructions
		switch (instr >> 9) & 7 { // 07xRSS
		case 0: // MUL 070RSS
			MUL(instr)
			return
		case 1: // DIV 071RSS
			DIV(instr)
			return
		case 2: // ASH 072RSS
			ASH(instr)
			return
		case 3: // ASHC 073RSS
			ASHC(instr)
			return
		case 4: // XOR 074RSS
			XOR(instr)
			return
		case 7: // SOB 077Rnn
			SOB(instr)
			return
		default: // We don't know this 07xRSS instruction
			printf("unknown 07xRSS instruction\n")
			printstate()
			trapat(INTINVAL)
			return
		}
	case 8: // 10xxxx instructions
		switch (instr >> 8) & 0xf { // 10xxxx 8 bit instructions first
		case 0: // BPL 1000 offset
			if !N() {
				branch(instr & 0xFF)
			}
			return
		case 1: // BMI 1004 offset
			if N() {
				branch(instr & 0xFF)
			}
			return
		case 2: // BHI 1010 offset
			if (!C()) && (!Z()) {
				branch(instr & 0xFF)
			}
			return
		case 3: // BLOS 1014 offset
			if C() || Z() {
				branch(instr & 0xFF)
			}
			return
		case 4: // BVC 1020 offset
			if !V() {
				branch(instr & 0xFF)
			}
			return
		case 5: // BVS 1024 offset
			if V() {
				branch(instr & 0xFF)
			}
			return
		case 6: // BCC 1030 offset
			if !C() {
				branch(instr & 0xFF)
			}
			return
		case 7: // BCS 1034 offset
			if C() {
				branch(instr & 0xFF)
			}
			return
		case 8: // EMT 1040 operand
			trapat(030) // Trap 30 - EMT instruction
			return
		case 9: // TRAP 1044 operand
			trapat(034) // Trap 34 - TRAP instruction
			return
		default: // Remaining 10xxxx instructions where xxxx >= 05000
			switch (instr >> 6) & 077 { // 10xxDD group
			case 050: // CLRB 1050DD
				CLR < 1 > (instr)
				return
			case 051: // COMB 1051DD
				COM < 1 > (instr)
				return
			case 052: // INCB 1052DD
				INC < 1 > (instr)
				return
			case 053: // DECB 1053DD
				_DEC < 1 > (instr)
				return
			case 054: // NEGB 1054DD
				NEG < 1 > (instr)
				return
			case 055: // ADCB 01055DD
				_ADC < 1 > (instr)
				return
			case 056: // SBCB 01056DD
				SBC < 1 > (instr)
				return
			case 057: // TSTB 1057DD
				TST < 1 > (instr)
				return
			case 060: // RORB 1060DD
				ROR < 1 > (instr)
				return
			case 061: // ROLB 1061DD
				ROL < 1 > (instr)
				return
			case 062: // ASRB 1062DD
				ASR < 1 > (instr)
				return
			case 063: // ASLB 1063DD
				ASL < 1 > (instr)
				return
			// case 0o64: // MTPS 1064SS
			// case 0o65: // MFPD 1065DD
			// case 0o66: // MTPD 1066DD
			// case 0o67: // MTFS 1064SS
			default: // We don't know this 0o10xxDD instruction
				printf("unknown 0o10xxDD instruction\n")
				printstate()
				trapat(INTINVAL)
				return
			}
		}
	case 9: // MOVB 11SSDD
		MOV < 1 > (instr)
		return
	case 10: // CMPB 12SSDD
		CMP < 1 > (instr)
		return
	case 11: // BITB 13SSDD
		BIT < 1 > (instr)
		return
	case 12: // BICB 14SSDD
		BIC < 1 > (instr)
		return
	case 13: // BISB 15SSDD
		BIS < 1 > (instr)
		return
	case 14: // SUB 16SSDD
		SUB(instr)
		return
	case 15:
		if instr == 0170011 {
			// SETD ; not needed by UNIX, but used; therefore ignored
			return
		}
	default: // 15  17xxxx FPP instructions
		printf("invalid 17xxxx FPP instruction\n")
		printstate()
		trapat(INTINVAL)
	}
}

func (kb *KB11) RESET() {
	if kb.currentmode() > 0 {
		// RESET is ignored outside of kernel mode
		return
	}
	kb.unibus.reset()
}

// RTI 000004, RTT 000006
func (kb *KB11) RTT() {
	kb.R[7] = kb.pop()
	psw := pop()
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
	dst := kb.memread(2, da)
	dst = (dst << 8) | (dst >> 8)
	kb.memwrite(2, da, dst)
	kb.psw &= 0xFFF0
	if dst&0xff00 == 0 {
		psw |= FLAGZ
	}
	if dst&0x80 == 0x80 {
		psw |= FLAGN
	}
}

func (kb *KB11) branch(i uint16) {
	if o & 0x80 {
		o = -(((^o) + 1) & 0xFF)
	}
	o <<= 1
	R[7] += o
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
	kb.psw &= 0xFFF0
	kb.psw |= FLAGZ
	kb.memwrite(l, kb.DA(instr), 0)
}

// COM 0051DD, COMB 1051DD
func (kb *KB11) COM(l int, instr uint16) {
	da := kb.DA(instr)
	dst := ^kb.memread(l, da)
	kb.memwrite(l, da, dst)
	setNZC(2, dst)
}

// INC 0052DD, INCB 1052DD
func (kb *KB11) INC(l int, instr uint16) {
	da := kb.DA(instr)
	dst := kb.memread(l, da)
	result := dst + 1
	kb.memwrite(l, da, result)
	kb.setNZV(l, result)
}

// DEC 0053DD, DECB 1053DD
func (kb *KB11) DEC(l int, instr uint16) {
	da := kb.DA(instr)
	uval := (kb.memread(l, da) - 1) & max(l)
	kb.memwrite(l, da, uval)
	kb.setNZV(l, uval)
}

// NEG 0054DD, NEGB 1054DD
func (kb *KB11) NEG(l int, instr uint16) {
	da := kb.DA(instr)
	sval := (-kb.memread(l, da)) & max(l)
	kb.memwrite(l, da, sval)
	kb.psw &= 0xFFF0
	if sval & msb(l) {
		kb.psw |= FLAGN
	}
	if sval == 0 {
		kb.psw |= FLAGZ
	} else {
		kb.psw |= FLAGC
	}
	if sval == 0x8000 {
		kg.psw |= FLAGV
	}
}

func (kb *KB11) ADC(l int, instr uint16) {
	da := kb.DA(instr)
	uval := kb.memread(l, da)
	if kb.c() {
		kb.memwrite(l, da, (uval+1)&max(l))
		kb.psw &= 0xFFF0
		if (uval + 1) & msb {
			kb.psw |= FLAGN
		}
		if uval == max(l) {
			kb.psw |= FLAGZ
		}
		if uval == 0077777 {
			kb.psw |= FLAGV
		}
		if uval == 0177777 {
			kb.psw |= FLAGC
		}
	} else {
		kb.psw &= 0xFFF0
		if uval & msb(l) {
			kb.psw |= FLAGN
		}
		if uval == 0 {
			kb.psw |= FLAGN
		}
	}
}

func (kb *KB11) SBC(l int, instr uint16) {
	da := kb.DA(instr)
	sval := kb.memread(l, da)
	if kb.c() {
		kb.memwrite(l, da, (sval-1)&max(l))
		kb.psw &= 0xFFF0
		if (sval - 1) & msb(l) {
			kb.psw |= FLAGN
		}
		if sval == 1 {
			kb.psw |= FLAGZ
		}
		if sval {
			kb.psw |= FLAGC
		}
		if sval == 0100000 {
			kb.psw |= FLAGV
		}
	} else {
		kb.psw &= 0xFFF0
		if sval & msb {
			kb.psw |= FLAGN
		}
		if sval == 0 {
			kv.psw |= FLAGZ
		}
		if sval == 0100000 {
			kb.psw |= FLAGV
		}
		kb.psw |= FLAGC
	}
}

// TST 0057DD, TSTB 1057DD
func (kb *KB11) TST(l int, instr uint16) {
	dst = memread(l, DA(instr))
	kb.psw &= 0xFFF0
	if dst == 0 {
		kb.psw |= FLAGZ
	}
	if dst & msb(l) {
		kb.psw |= FLAGN
	}
}

func (kb *KB11) ROR(l int, instr uint16) {
	da := kb.DA(instr)
	sval := kb.memread(l, da)
	if kb.c() {
		sval |= max(l) + 1
	}
	kb.psw &= 0xFFF0
	if sval & 1 {
		kb.psw |= FLAGC
	}
	// watch out for integer wrap around
	if sval & (max(l) + 1) {
		kb.psw |= FLAGN
	}
	if !(sval & max(l)) {
		kb.psw |= FLAGZ
	}
	if (sval&1 == 1) != (sval & (max(l)+1 == max(l)+1)) {
		kb.psw |= FLAGV
	}
	sval >>= 1
	kb.memwrite(l, da, sval)
}

func (kb *KB11) ROL(l int, instr uint16) {
	da := kb.DA(instr)
	sval := kb.memread(l, da) << 1
	if kb.c() {
		sval |= 1
	}
	kb.psw &= 0xFFF0
	if sval & (max(l) + 1) {
		kb.psw |= FLAGC
	}
	if sval & msb(l) {
		kb.psw |= FLAGN
	}
	if !(sval & max(l)) {
		kb.psw |= FLAGZ
	}
	if (sval ^ (sval >> 1)) & msb(l) {
		kb.psw |= FLAGV
	}
	sval &= max
	kb.memwrite(l, da, sval)
}

func (kb *KB11) ASR(l int, instr uint16) {
	da := kb.DA(instr)
	uval := kb.memread(l, da)
	kb.psw &= 0xFFF0
	if uval & 1 {
		kb.psw |= FLAGC
	}
	if uval & msb(l) {
		kb.psw |= FLAGN
	}
	if (uval & msb(l)) != (uval & 1) {
		kb.psw |= FLAGV
	}
	uval = (uval & msb(l)) | (uval >> 1)
	if uval == 0 {
		kb.psw |= FLAGZ
	}
	kb.memwrite(l, da, uval)
}

func (kb *KB11) ASL(l int, instr uint16) {
	da := kb.DA(instr)
	// TODO(dfc) doesn't need to be an sval
	sval := kb.memread(l, da)
	kb.psw &= 0xFFF0
	if sval & msb(l) {
		kb.psw |= FLAGC
	}
	if sval & (msb(l) >> 1) {
		kb.psw |= FLAGN
	}
	if (sval ^ (sval << 1)) & msb(l) {
		kb.psw |= FLAGV
	}
	sval = (sval << 1) & max(l)
	if sval == 0 {
		kb.psw |= FLAGZ
	}
	kb.memwrite(l, da, sval)
}

// MARK 0064NN
func (kb *KB11) MARK(instr uint16) {
	kb.R[6] = kb.R[7] + ((instr & 077) << 1)
	kb.R[7] = kb.R[5]
	kb.R[5] = kb.pop()
}

func (kb *KB11) MFPI(instr uint16) {
	da := kb.DA(instr)
	var uval uint16
	if da == 0170006 {
		if (kb.currentmode() == 3) && (kb.previousmode() == 3) {
			uval = kb.R[6]
		} else {
			uval = kb.stackpointer[previousmode()]
		}
	} else if isReg(da) {
		fmt.Printf("invalid MFPI instruction\n")
		kb.printstate()
		os.Exit(1)
	} else {
		uval = kb.unibus.read16(kb.mmu.decode(false, da, kb.previousmode()))
	}
	kb.push(uval)
	kb.setNZ(2, uval)
}

func (kb *KB11) MTPI(instr uint16) {
	da := kb.DA(instr)
	uval := kb.pop()
	if da == 0170006 {
		if (kb.currentmode() == 3) && (kb.previousmode() == 3) {
			kb.R[6] = uval
		} else {
			kb.stackpointer[previousmode()] = uval
		}
	} else if isReg(da) {
		fmt.Printf("invalid MTPI instrution\n")
		kb.printstate()
		os.Exit(1)
	} else {
		kb.unibus.write16(kb.mmu.decode(true, da, kb.previousmode()), uval)
	}
	setNZ(2, uval)
}

// SXT 0067DD
func (kb *KB11) SXT(instr uint16) {
	n := func() uint16 {
		if kb.n() {
			return 0xffff
		}
		return 0
	}

	result := n()
	memwrite(2, kb.DA(instr), result)
	setNZ(2, result)
}

// MOV 01SSDD, MOVB 11SSDD
func (kb *KB11) MOV(len int, instr uint16) {
	src := memread(len, kb.SA(instr))
	if !(instr & 0x38) && (len == 1) {
		if src & 0200 {
			// Special case: movb sign extends register to word size
			src |= 0xFF00
		}
		kb.R[instr&7] = src
		kb.setNZ(len, src)
		return
	}
	memwrite(len, kb.DA(instr), src)
	setNZ(len, src)
}

// CMP 02SSDD, CMPB 12SSDD
func (kb *KB11) CMP(l int, instr uint16) {
	val1 := kb.memread(l, kb.SA(instr))
	da := kb.DA(instr)
	val2 := kb.memread(l, da)
	sval := (val1 - val2) & max(l)
	kb.psw &= 0xFFF0
	if sval == 0 {
		kb.psw |= FLAGZ
	}
	if sval & msb(l) {
		kb.psw |= FLAGN
	}
	if ((val1 ^ val2) & msb(l)) && (!((val2 ^ sval) & msb(l))) {
		kb.psw |= FLAGV
	}
	if val1 < val2 {
		kb.psw |= FLAGC
	}
}

// BIT 03SSDD, BITB 13SSDD
func (kb *KB11) BIT(l int, instr uint16) {
	src := kb.memread(l, kb.SA(instr))
	dst := kb.memread(l, kb.DA(instr))
	result := src & dst
	setNZ(l, result)
}

func (kb *KB11) trapat(vec uint16) {
	if vec&1 > 0 {
		fmt.Printf("Thou darst calling trapat() with an odd vector number?\n")
		os.Exit(1)
	}

	fmt.Printf("trap: %03o\n", vec)

	psw := kb.psw
	kb.kernelmode()
	kb.push(psw)
	kb.push(kb.R[7])

	kb.R[7] = kb.read16(vec)
	kb.writePSW(kb.read16(vec+2) | (previousmode() << 12))
}

func (kb *KB11) fetch16() uint16 {
	val := kb.read16(kb.R[7])
	kb.R[7] += 2
	return val
}

func (kb *KB11) push(v uint16) {
	kb.R[6] -= 2
	write16(kb.R[6], v)
}

func (kb *KB11) pop() uint16 {
	val := kb.read16(kb.R[6])
	lb.R[6] += 2
	return val
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

// priority returns the current CPU interrupt priority.
func (kb *KB11) priority() uint16 { return (kb.psw >> 5) & 7 }

const (
	FLAGC = 1
	FLAGV = 2
	FLAGZ = 4
	FLAGN = 8
)

func (kb *KB11) n() bool { return kb.psw&FLAGN > 0 }
func (kb *KB11) z() bool { return kb.psw&FLAGZ > 0 }
func (kb *KB11) v() bool { return kb.psw&FLAGV > 0 }
func (kb *KB11) c() bool { return kb.psw&FLAGC > 0 }

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
			"K"
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
	fmt.Printf("[%s%s%s%s%s%s", prev, curr(), n(), z(), v(), c())
	fmt.Printf("]  instr %06o: %06o\t ", kb.pc, read16(kb.pc))
	kb.disasm(PC)
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
	return 0x00
}

func (kb *KB11) disasm(pc uint16) {}
