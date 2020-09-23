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
				printf("unknown 000xDD instruction\n")
				printstate()
				trapat(INTINVAL)
				return
			}
		case 1: // BR 0004 offset
			branch(instr & 0xff)
			return
		case 2: // BNE 0010 offset
			if !kb.z() {
				branch(instr & 0xff)
			}
			return
		case 3: // BEQ 0014 offset
			if kb.z() {
				branch(instr & 0xff)
			}
			return
		case 4: // BGE 0020 offset
			if !(kb.n() ^ kb.v()) {
				branch(instr & 0xFF)
			}
			return
		case 5: // BLT 0024 offset
			if kb.n() ^ kb.v() {
				branch(instr & 0xFF)
			}
			return
		case 6: // BGT 0030 offset
			if (!(kb.n() ^ kb.v())) && (!kb.z()) {
				branch(instr & 0xFF)
			}
			return
		case 7: // BLE 0034 offset
			if (kb.n() ^ kb.v()) || kb.z() {
				branch(instr & 0xFF)
			}
			return
		case 8: // JSR 004RDD In two parts
		case 9: // JSR 004RDD continued (9 bit instruction so use 2 x 8 bit
			JSR(instr)
			return
		default: // Remaining 0o00xxxx instructions where xxxx >= 05000
			switch instr >> 6 { // 00xxDD
			case 050: // CLR 0050DD
				CLR < 2 > (instr)
				return
			case 051: // COM 0051DD
				COM < 2 > (instr)
				return
			case 052: // INC 0052DD
				INC < 2 > (instr)
				return
			case 053: // DEC 0053DD
				_DEC < 2 > (instr)
				return
			case 054: // NEG 0054DD
				NEG < 2 > (instr)
				return
			case 055: // ADC 0055DD
				_ADC < 2 > (instr)
				return
			case 056: // SBC 0056DD
				SBC < 2 > (instr)
				return
			case 057: // TST 0057DD
				TST < 2 > (instr)
				return
			case 060: // ROR 0060DD
				ROR < 2 > (instr)
				return
			case 061: // ROL 0061DD
				ROL < 2 > (instr)
				return
			case 062: // ASR 0062DD
				ASR < 2 > (instr)
				return
			case 063: // ASL 0063DD
				ASL < 2 > (instr)
				return
			case 064: // MARK 0064nn
				MARK(instr)
				return
			case 065: // MFPI 0065SS
				MFPI(instr)
				return
			case 066: // MTPI 0066DD
				MTPI(instr)
				return
			case 067: // SXT 0067DD
				SXT(instr)
				return
			default: // We don't know this 0o00xxDD instruction
				printf("unknown 00xxDD instruction\n")
				printstate()
				trapat(INTINVAL)
			}
		}
	case 1: // MOV  01SSDD
		MOV < 2 > (instr)
		return
	case 2: // CMP 02SSDD
		CMP < 2 > (instr)
		return
	case 3: // BIT 03SSDD
		BIT < 2 > (instr)
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

func (kb *KB11) disasm(pc uint16) {}
