package main

import "fmt"

var (
	rs = [...]string{"R0", "R1", "R2", "R3", "R4", "R5", "SP", "PC"}
)

const (
	DD = 1 << 1
	S  = 1 << 2
	RR = 1 << 3
	O  = 1 << 4
	N  = 1 << 5
)

type D struct {
	mask uint16
	ins  uint16
	msg  string
	flag uint8
	b    bool
}

var (
	disamtable = [...]D{
		{0177777, 0000001, "WAIT", 0, false},
		{0177777, 0000002, "RTI", 0, false},
		{0177777, 0000003, "BPT", 0, false},
		{0177777, 0000004, "IOT", 0, false},
		{0177777, 0000005, "RESET", 0, false},
		{0177777, 0000006, "RTT", 0, false},
		{0177777, 0000007, "MFPT", 0, false},

		{0177700, 0000100, "JMP", DD, false},
		{0177770, 0000200, "RTS", RR, false},
		{0177700, 0000300, "SWAB", DD, false},

		{0177700, 0006400, "MARK", N, false},
		{0177700, 0006500, "MFPI", DD, false},
		{0177700, 0006600, "MTPI", DD, false},
		{0177700, 0006700, "SXT", DD, false},

		{0177400, 0104000, "EMT", N, false},
		{0177400, 0104400, "TRAP", N, false},
		{0177400, 0100000, "BPL", O, false},
		{0177400, 0100400, "BMI", O, false},
		{0177400, 0101000, "BHI", O, false},
		{0177400, 0101400, "BLOS", O, false},
		{0177400, 0102000, "BVC", O, false},
		{0177400, 0102400, "BVS", O, false},
		{0177400, 0103000, "BCC", O, false},
		{0177400, 0103400, "BCS", O, false},
		{0177400, 0000400, "BR", O, false},
		{0177400, 0001000, "BNE", O, false},
		{0177400, 0001400, "BEQ", O, false},
		{0177400, 0002000, "BGE", O, false},
		{0177400, 0002400, "BLT", O, false},
		{0177400, 0003000, "BGT", O, false},
		{0177400, 0003400, "BLE", O, false},

		{0177000, 0004000, "JSR", RR | DD, false},
		{0177000, 0070000, "MUL", RR | DD, false},
		{0177000, 0071000, "DIV", RR | DD, false},
		{0177000, 0072000, "ASH", RR | DD, false},
		{0177000, 0073000, "ASHC", RR | DD, false},
		{0177000, 0077000, "SOB", RR | O, false},
		{0170000, 0060000, "ADD", S | DD, false},
		{0170000, 0160000, "SUB", S | DD, false},

		{0077700, 0005000, "CLR", DD, true},
		{0077700, 0005100, "COM", DD, true},
		{0077700, 0005200, "INC", DD, true},
		{0077700, 0005300, "DEC", DD, true},
		{0077700, 0005400, "NEG", DD, true},
		{0077700, 0005500, "ADC", DD, true},
		{0077700, 0005600, "SBC", DD, true},
		{0077700, 0005700, "TST", DD, true},
		{0077700, 0006000, "ROR", DD, true},
		{0077700, 0006100, "ROL", DD, true},
		{0077700, 0006200, "ASR", DD, true},
		{0077700, 0006300, "ASL", DD, true},

		{0070000, 0010000, "MOV", S | DD, true},
		{0070000, 0020000, "CMP", S | DD, true},
		{0070000, 0030000, "BIT", S | DD, true},
		{0070000, 0040000, "BIC", S | DD, true},
		{0070000, 0050000, "BIS", S | DD, true},
		{0000000, 0000001, "HALT", 0, false}, // fake instruction so HALT is left in l
	}
)

func (kb *KB11) disasmaddr(m, a uint16) {
	if m&7 > 0 {
		switch m {
		case 027:
			a += 2
			fmt.Printf("$%06o", kb.read16(a))
			return
		case 037:
			a += 2
			fmt.Printf("*%06o", kb.read16(a))
			return
		case 067:
			a += 2
			fmt.Printf("*%06o", (a+2+(kb.read16(a)))&0xFFFF)
			return
		case 077:
			fmt.Printf("**%06o", (a+2+(kb.read16(a)))&0xFFFF)
			return
		}
	}

	switch m & 070 {
	case 000:
		fmt.Printf("%s", rs[m&7])
	case 010:
		fmt.Printf("(%s)", rs[m&7])
	case 020:
		fmt.Printf("(%s)+", rs[m&7])
	case 030:
		fmt.Printf("*(%s)+", rs[m&7])
	case 040:
		fmt.Printf("-(%s)", rs[m&7])
	case 050:
		fmt.Printf("*-(%s)", rs[m&7])
	case 060:
		a += 2
		fmt.Printf("%06o (%s)", kb.read16(a), rs[m&7])
	case 070:
		a += 2
		fmt.Printf("*%06o (%s)", kb.read16(a), rs[m&7])
	}
}

func (kb *KB11) disasm(a uint16) {
	ins := kb.read16(a)

	var l D
	for _, l = range disamtable {
		if (ins & l.mask) == l.ins {
			break
		}
	}
	if l.ins == 0 {
		fmt.Printf("???")
		return
	}

	fmt.Printf("%s", l.msg)
	if l.b && (ins&0100000) > 0 {
		fmt.Printf("B")
	}
	s := (ins & 07700) >> 6
	d := ins & 077
	o := ins & 0377
	switch l.flag {
	case S | DD:
		fmt.Printf(" ")
		kb.disasmaddr(s, a)
		fmt.Printf(",")
		fallthrough
	case DD:
		fmt.Printf(" ")
		kb.disasmaddr(d, a)
	case RR | O:
		fmt.Printf(" %s,", rs[(ins&0700)>>6])
		o &= 077
		fallthrough
	case O:
		if o&0x80 > 0 {
			fmt.Printf(" -%03o", (2 * ((0xFF ^ o) + 1)))
		} else {
			fmt.Printf(" +%03o", (2 * o))
		}
	case RR | DD:
		fmt.Printf(" %s, ", rs[(ins&0700)>>6])
		kb.disasmaddr(d, a)
		fallthrough
	case RR:
		fmt.Printf(" %s", rs[ins&7])
	}
}
