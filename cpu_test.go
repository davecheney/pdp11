package main

import (
	"testing"

	"github.com/matryer/is"
)

func TestADD(t *testing.T) {
	is := is.New(t)
	var cpu KB11
	for s := 0; s < 16; s++ {
		for d := 0; d < 16; d++ {
			src, dst := uint16(1)<<s, uint16(1)<<d
			cpu.R[0] = src
			cpu.R[1] = dst
			cpu.Load(002000, 0060001) // ADD R0, R1
			cpu.R[7] = 002000
			cpu.step()
			t.Logf("R0: %06o, R1: %06o", src, dst)
			is.Equal(cpu.R[1], src+dst)
			is.Equal(cpu.n(), (src+dst)&0x8000 > 0)
			is.Equal(cpu.z(), src+dst == 0)
			is.Equal(cpu.v(), (!((src^dst)&0x8000 > 0) && ((dst^(src+dst))&0x8000 > 0)))
			is.Equal(cpu.c(), uint32(src)+uint32(dst) > 0xffff)
		}
	}
}

func TestADC(t *testing.T) {
	is := is.New(t)
	var cpu KB11
	cpu.Load(002000, 0005500) // ADC R0
	for d := 0; d < 16; d++ {
		dst := uint16(1) << d
		cpu.R[0] = dst
		cpu.R[7] = 002000
		cpu.psw &= ^uint16(FLAGC)
		cpu.step()
		t.Logf("R0: %06o", dst)
		is.Equal(cpu.R[0], dst)
		is.Equal(cpu.n(), dst&0x8000 > 0)
		is.Equal(cpu.z(), dst == 0)
		is.Equal(cpu.v(), false)
		is.Equal(cpu.c(), false)
	}

	for d := 0; d < 16; d++ {
		dst := uint16(1) << d
		cpu.R[0] = dst
		cpu.R[7] = 002000
		cpu.psw |= FLAGC
		cpu.step()
		t.Logf("R0: %06o", dst)
		is.Equal(cpu.R[0], dst+1)
		is.Equal(cpu.n(), (dst+1)&0x8000 > 0)
		is.Equal(cpu.z(), dst+1 == 0)
		is.Equal(cpu.v(), dst == 0077777)
		is.Equal(cpu.c(), dst == 0177777)
	}
}

func TestADCB(t *testing.T) {
	is := is.New(t)

	var cpu KB11
	cpu.Load(002000, 0105500) // ADCB R0
	for d := 0; d < 16; d++ {
		dst := uint16(1) << d
		cpu.R[0] = dst
		cpu.R[7] = 002000
		cpu.psw &= ^uint16(FLAGC)
		cpu.step()
		t.Logf("R0: %06o", dst)
		is.Equal(cpu.R[0]&0xff, dst&0xff)
		is.Equal(cpu.n(), dst&0x80 > 0)
		is.Equal(cpu.z(), dst&0xff == 0)
		is.Equal(cpu.v(), false)
		is.Equal(cpu.c(), false)
	}

	for d := 0; d < 16; d++ {
		dst := uint16(1) << d
		cpu.R[0] = dst
		cpu.R[7] = 002000
		cpu.psw |= FLAGC
		cpu.step()
		t.Logf("R0: %06o", dst)
		is.Equal(cpu.R[0]&0xff, (dst+1)&0xff)
		is.Equal(cpu.n(), (dst+1)&0x80 > 0)
		is.Equal(cpu.z(), (dst+1)&0xff == 0)
		is.Equal(cpu.v(), (dst+1)&0xff == 0200)
		is.Equal(cpu.c(), (dst+1)&0xff == 0)
	}
}

func TestSUB(t *testing.T) {
	is := is.New(t)

	var cpu KB11
	cpu.Load(002000, 0160001) // SUB R0, R1
	for s := 0; s < 16; s++ {
		for d := 0; d < 16; d++ {
			src, dst := uint16(1)<<s, uint16(1)<<d
			cpu.R[0] = src
			cpu.R[1] = dst
			cpu.R[7] = 002000
			cpu.step()
			t.Logf("R0: %06o, R1: %06o", src, dst)
			is.Equal(cpu.R[1], dst-src)
			is.Equal(cpu.n(), (dst-src)&0x8000 > 0)
			is.Equal(cpu.z(), dst-src == 0)
			is.Equal(cpu.v(), (((src^dst)&0x8000 > 0) && (!((dst^(dst-src))&0x8000 > 0))))
			is.Equal(cpu.c(), src > dst)
		}
	}
}

func TestCMP(t *testing.T) {
	is := is.New(t)

	var cpu KB11
	cpu.Load(002000, 0020001) // CMP R0, R1
	for s := 0; s < 16; s++ {
		for d := 0; d < 16; d++ {
			src, dst := uint16(1)<<s, uint16(1)<<d
			cpu.R[0] = src
			cpu.R[1] = dst
			cpu.R[7] = 002000
			cpu.step()
			result := uint32(src) + 0x10000 - uint32(dst)
			t.Logf("R0: %06o, R1: %06o", src, dst)
			is.Equal(cpu.R[0], src)
			is.Equal(cpu.R[1], dst)
			is.Equal(cpu.n(), result&0x8000 > 0)
			is.Equal(cpu.z(), src-dst == 0)
			is.Equal(cpu.v(), (src^dst)&0x8000 > 0 && !((dst^(src+(^dst)+1))&0x8000 > 0))
			is.Equal(cpu.c(), src < dst)
		}
	}

	// bug found in early v6 unix boot
	cpu.R[1] = 0137000
	cpu.R[7] = 0000032
	cpu.Load(addr18(cpu.R[7]), 0020701) // CMP PC, R1
	cpu.step()

	is.Equal(cpu.R[1], uint16(0137000))
	is.Equal(cpu.R[7], uint16(0000034))
	is.Equal(cpu.n(), false)
	is.Equal(cpu.z(), false)
	is.Equal(cpu.v(), false)
	is.Equal(cpu.c(), true)
}

func TestCMPB(t *testing.T) {
	is := is.New(t)

	var cpu KB11
	cpu.Load(002000, 0120001) // CMPB R0, R1
	for s := 0; s < 16; s++ {
		for d := 0; d < 16; d++ {
			src, dst := uint16(1)<<s, uint16(1)<<d
			cpu.R[0] = src
			cpu.R[1] = dst
			cpu.R[7] = 002000
			cpu.step()
			result := src + 0x100 - dst
			t.Logf("R0: %06o, R1: %06o", src, dst)
			is.Equal(cpu.R[0], src)
			is.Equal(cpu.R[1], dst)
			is.Equal(cpu.n(), result&0x80 > 0)
			is.Equal(cpu.z(), result&0xff == 0)
			is.Equal(cpu.v(), (src^dst)&0x80 > 0 && !((dst^result)&0x80 > 0))
			is.Equal(cpu.c(), src&0xff < dst&0xff)
		}
	}
}

func TestBIT(t *testing.T) {
	is := is.New(t)

	var cpu KB11
	cpu.Load(002000, 0030001) // BIT R0, R1
	for s := 0; s < 16; s++ {
		for d := 0; d < 16; d++ {
			src, dst := uint16(1)<<s, uint16(1)<<d
			cpu.R[0] = src
			cpu.R[1] = dst
			cpu.R[7] = 002000
			cpu.step()
			t.Logf("R0: %06o, R1: %06o", src, dst)
			is.Equal(cpu.n(), (src&dst)&0x8000 > 0)
			is.Equal(cpu.z(), src&dst == 0)
			is.Equal(cpu.v(), false)
		}
	}
}

func TestBITB(t *testing.T) {
	is := is.New(t)

	var cpu KB11
	cpu.Load(002000, 0130001) // BITB R0, R1
	for s := 0; s < 16; s++ {
		for d := 0; d < 16; d++ {
			src, dst := uint16(1)<<s, uint16(1)<<d
			cpu.R[0] = src
			cpu.R[1] = dst
			cpu.R[7] = 002000
			cpu.step()
			t.Logf("R0: %06o, R1: %06o", src, dst)
			is.Equal(cpu.n(), (src&dst)&0x80 > 0)
			is.Equal(cpu.z(), src&dst&0xff == 0)
			is.Equal(cpu.v(), false)
		}
	}
}

func TestBIC(t *testing.T) {
	is := is.New(t)

	var cpu KB11
	cpu.Load(002000, 0040001) // BIC R0, R1
	for s := 0; s < 16; s++ {
		for d := 0; d < 16; d++ {
			src, dst := uint16(1)<<s, uint16(1)<<d
			cpu.R[0] = src
			cpu.R[1] = dst
			cpu.R[7] = 002000
			cpu.step()
			t.Logf("R0: %06o, R1: %06o", src, dst)
			is.Equal(cpu.R[1], (^src)&dst)
			is.Equal(cpu.n(), (^src)&dst&0x8000 > 0)
			is.Equal(cpu.z(), (^src)&dst == 0)
			is.Equal(cpu.v(), false)
		}
	}
}

func TestBICB(t *testing.T) {
	is := is.New(t)

	var cpu KB11
	cpu.Load(002000, 0140001) // BICB R0, R1
	for s := 0; s < 16; s++ {
		for d := 0; d < 16; d++ {
			src, dst := uint16(1)<<s, uint16(1)<<d
			cpu.R[0] = src
			cpu.R[1] = dst
			cpu.R[7] = 002000
			cpu.step()
			t.Logf("R0: %06o, R1: %06o", src, dst)
			is.Equal(cpu.R[1]&0xff, (^src)&dst&0xff)
			is.Equal(cpu.n(), (^src)&dst&0x80 > 0)
			is.Equal(cpu.z(), (^src)&dst&0xff == 0)
			is.Equal(cpu.v(), false)
		}
	}
}

func TestBIS(t *testing.T) {
	is := is.New(t)

	var cpu KB11
	cpu.Load(002000, 0050001) // BIS R0, R1
	for s := 0; s < 16; s++ {
		for d := 0; d < 16; d++ {
			src, dst := uint16(1)<<s, uint16(1)<<d
			cpu.R[0] = src
			cpu.R[1] = dst
			cpu.R[7] = 002000
			cpu.step()
			t.Logf("R0: %06o, R1: %06o", src, dst)
			is.Equal(cpu.R[1], src|dst)
			is.Equal(cpu.n(), (src|dst)&0x8000 > 0)
			is.Equal(cpu.z(), src|dst == 0)
			is.Equal(cpu.v(), false)
		}
	}
}

func TestBISB(t *testing.T) {
	is := is.New(t)

	var cpu KB11
	cpu.Load(002000, 0150001) // BISB R0, R1
	for s := 0; s < 16; s++ {
		for d := 0; d < 16; d++ {
			src, dst := uint16(1)<<s, uint16(1)<<d
			cpu.R[0] = src
			cpu.R[1] = dst
			cpu.R[7] = 002000
			cpu.step()
			t.Logf("R0: %06o, R1: %06o", src, dst)
			is.Equal(cpu.R[1]&0xff, (src|dst)&0xff)
			is.Equal(cpu.n(), (src|dst)&0x80 > 0)
			is.Equal(cpu.z(), (src|dst)&0xff == 0)
			is.Equal(cpu.v(), false)
		}
	}
}

func TestSBC(t *testing.T) {
	is := is.New(t)
	var cpu KB11
	cpu.Load(002000, 0005600) // SBC R0
	for d := 0; d < 16; d++ {
		dst := uint16(1) << d
		cpu.R[0] = dst
		cpu.R[7] = 002000
		cpu.psw &= ^uint16(FLAGC)
		cpu.step()
		t.Logf("R0: %06o", dst)
		is.Equal(cpu.R[0], dst)
		is.Equal(cpu.n(), dst&0x8000 > 0)
		is.Equal(cpu.z(), dst == 0)
		is.Equal(cpu.v(), false)
		is.Equal(cpu.c(), false)
	}

	for d := 0; d < 16; d++ {
		dst := uint16(1) << d
		cpu.R[0] = dst
		cpu.R[7] = 002000
		cpu.psw |= FLAGC
		cpu.step()
		t.Logf("R0: %06o", dst)
		is.Equal(cpu.R[0], dst-1)
		is.Equal(cpu.n(), (dst-1)&0x8000 > 0)
		is.Equal(cpu.z(), dst-1 == 0)
		is.Equal(cpu.v(), dst == 0100000)
		is.Equal(cpu.c(), dst == 0)
	}
}

func TestSBCB(t *testing.T) {
	is := is.New(t)
	var cpu KB11
	cpu.Load(002000, 0105600) // SBCB R0
	for d := 0; d < 16; d++ {
		dst := uint16(1) << d
		cpu.R[0] = dst
		cpu.R[7] = 002000
		cpu.psw &= ^uint16(FLAGC)
		cpu.step()
		t.Logf("R0: %06o", dst)
		is.Equal(cpu.R[0], dst)
		is.Equal(cpu.n(), dst&0x80 > 0)
		is.Equal(cpu.z(), dst&0xff == 0)
		is.Equal(cpu.v(), false)
		is.Equal(cpu.c(), false)
	}

	for d := 0; d < 16; d++ {
		dst := uint16(1) << d
		cpu.R[0] = dst
		cpu.R[7] = 002000
		cpu.psw |= FLAGC
		cpu.step()
		t.Logf("R0: %06o", dst)
		is.Equal(cpu.R[0]&0xff, (dst-1)&0xff)
		is.Equal(cpu.n(), (dst-1)&0x80 > 0)
		is.Equal(cpu.z(), (dst-1)&0xff == 0)
		is.Equal(cpu.v(), dst&0xff == 0200)
		is.Equal(cpu.c(), dst&0xff == 0)
	}
}

func TestROR(t *testing.T) {
	is := is.New(t)

	var cpu KB11
	cpu.Load(002000, 0006000) // ROR R0
	for d := 0; d < 16; d++ {
		dst := uint16(1) << d
		cpu.R[0] = dst
		cpu.R[7] = 002000
		cpu.psw &= ^uint16(FLAGC)
		cpu.step()
		t.Logf("R0: %06o", dst)
		is.Equal(cpu.R[0], dst>>1)
		is.Equal(cpu.n(), false)
		is.Equal(cpu.z(), dst>>1 == 0)
		is.Equal(cpu.v(), cpu.n() != cpu.c())
		is.Equal(cpu.c(), dst&1 > 0)
	}

	for d := 0; d < 16; d++ {
		dst := uint16(1) << d
		cpu.R[0] = dst
		cpu.R[7] = 002000
		cpu.psw |= FLAGC
		cpu.step()
		t.Logf("R0: %06o", dst)
		result := 0x8000 | dst>>1
		is.Equal(cpu.R[0], result)
		is.Equal(cpu.n(), true)
		is.Equal(cpu.z(), false)
		is.Equal(cpu.v(), cpu.n() != cpu.c())
		is.Equal(cpu.c(), dst&1 > 0)
	}
}

func TestRORB(t *testing.T) {
	is := is.New(t)

	var cpu KB11
	cpu.Load(002000, 0106000) // RORB R0
	for d := 0; d < 16; d++ {
		dst := uint16(1) << d
		cpu.R[0] = dst
		cpu.R[7] = 002000
		cpu.psw &= ^uint16(FLAGC)
		cpu.step()
		result := (dst & 0xff) >> 1
		t.Logf("R0: %06o", dst)
		is.Equal(cpu.R[0]&0xff, result)
		is.Equal(cpu.n(), false)
		is.Equal(cpu.z(), result == 0)
		is.Equal(cpu.v(), cpu.n() != cpu.c())
		is.Equal(cpu.c(), dst&1 > 0)
	}

	for d := 0; d < 16; d++ {
		dst := uint16(1) << d
		cpu.R[0] = dst
		cpu.R[7] = 002000
		cpu.psw |= FLAGC
		cpu.step()
		t.Logf("R0: %06o", dst)
		result := 0x80 | (dst>>1)&0xff
		is.Equal(cpu.R[0]&0xff, result)
		is.Equal(cpu.n(), true)
		is.Equal(cpu.z(), false)
		is.Equal(cpu.v(), cpu.n() != cpu.c())
		is.Equal(cpu.c(), dst&1 > 0)
	}
}

func TestROL(t *testing.T) {
	is := is.New(t)

	var cpu KB11
	cpu.Load(002000, 0006100) // ROL R0
	for d := 0; d < 16; d++ {
		dst := uint16(1) << d
		cpu.R[0] = dst
		cpu.R[7] = 002000
		cpu.psw &= ^uint16(FLAGC)
		cpu.step()
		t.Logf("R0: %06o", dst)
		result := dst << 1
		is.Equal(cpu.R[0], result)
		is.Equal(cpu.n(), result&0x8000 > 0)
		is.Equal(cpu.z(), result == 0)
		is.Equal(cpu.v(), cpu.n() != cpu.c())
		is.Equal(cpu.c(), dst&0x8000 > 0)
	}

	for d := 0; d < 16; d++ {
		dst := uint16(1) << d
		cpu.R[0] = dst
		cpu.R[7] = 002000
		cpu.psw |= FLAGC
		cpu.step()
		t.Logf("R0: %06o", dst)
		result := dst<<1 | 1
		is.Equal(cpu.R[0], result)
		is.Equal(cpu.n(), result&0x8000 > 0)
		is.Equal(cpu.z(), result == 0)
		is.Equal(cpu.v(), cpu.n() != cpu.c())
		is.Equal(cpu.c(), dst&0x8000 > 0)
	}
}

func TestROLB(t *testing.T) {
	is := is.New(t)

	var cpu KB11
	cpu.Load(002000, 0106100) // ROLB R0
	for d := 0; d < 16; d++ {
		dst := uint16(1) << d
		cpu.R[0] = dst
		cpu.R[7] = 002000
		cpu.psw &= ^uint16(FLAGC)
		cpu.step()
		t.Logf("R0: %06o", dst)
		result := dst << 1 & 0xff
		is.Equal(cpu.R[0]&0xff, result)
		is.Equal(cpu.n(), result&0x80 > 0)
		is.Equal(cpu.z(), result&0xff == 0)
		is.Equal(cpu.v(), cpu.n() != cpu.c())
		is.Equal(cpu.c(), dst&0x80 > 0)
	}

	for d := 0; d < 16; d++ {
		dst := uint16(1) << d
		cpu.R[0] = dst
		cpu.R[7] = 002000
		cpu.psw |= FLAGC
		cpu.step()
		t.Logf("R0: %06o", dst)
		result := (dst<<1 | 1) & 0xff
		is.Equal(cpu.R[0]&0xff, result)
		is.Equal(cpu.n(), result&0x80 > 0)
		is.Equal(cpu.z(), result&0xff == 0)
		is.Equal(cpu.v(), cpu.n() != cpu.c())
		is.Equal(cpu.c(), dst&0x80 > 0)
	}
}

func TestSXT(t *testing.T) {
	is := is.New(t)

	var cpu KB11
	cpu.Load(002000, 0006700) // SXT R0
	for d := 0; d < 16; d++ {
		dst := uint16(1) << d
		cpu.R[0] = dst
		cpu.R[7] = 002000
		cpu.psw &= ^uint16(FLAGN)
		cpu.step()
		t.Logf("R0: %06o", dst)
		is.Equal(cpu.R[0], uint16(0))
		is.Equal(cpu.n(), false)
		is.Equal(cpu.z(), !cpu.n())
		is.Equal(cpu.v(), false)
	}

	for d := 0; d < 16; d++ {
		dst := uint16(1) << d
		cpu.R[0] = dst
		cpu.R[7] = 002000
		cpu.psw |= uint16(FLAGN)
		cpu.step()
		t.Logf("R0: %06o", dst)
		is.Equal(cpu.R[0], uint16(0xffff))
		is.Equal(cpu.n(), true)
		is.Equal(cpu.z(), !cpu.n())
		is.Equal(cpu.v(), false)
	}
}

func TestTST(t *testing.T) {
	is := is.New(t)

	var cpu KB11
	cpu.Load(002000, 0005700) // TST R0
	for d := 0; d < 16; d++ {
		dst := uint16(1) << d
		cpu.R[0] = dst
		cpu.R[7] = 002000
		cpu.step()
		t.Logf("R0: %06o", dst)
		is.Equal(cpu.n(), (dst)&0x8000 > 0)
		is.Equal(cpu.z(), dst == 0)
	}
}

func TestTSTB(t *testing.T) {
	is := is.New(t)

	var cpu KB11
	cpu.Load(002000, 0105700) // TSTB R0
	for d := 0; d < 16; d++ {
		dst := uint16(1) << d
		cpu.R[0] = dst
		cpu.R[7] = 002000
		cpu.step()
		t.Logf("R0: %06o, psw: %06o", dst, cpu.psw)
		is.Equal(cpu.n(), dst&0x80 > 0)
		is.Equal(cpu.z(), (dst&0xff) == 0)
	}
}

func TestNEG(t *testing.T) {
	is := is.New(t)

	var cpu KB11
	cpu.Load(002000, 0005400) // NEG R0
	for d := 0; d < 16; d++ {
		dst := uint16(1) << d
		cpu.R[0] = dst
		cpu.R[7] = 002000
		cpu.step()
		t.Logf("R0: %06o, psw: %06o", dst, cpu.psw)
		is.Equal(cpu.R[0], -dst)
		is.Equal(cpu.n(), -dst&0x8000 > 0)
		is.Equal(cpu.z(), -dst == 0)
		is.Equal(cpu.v(), -dst == 0x8000)
		is.Equal(cpu.c(), -dst != 0)
	}
}

func TestNEGB(t *testing.T) {
	is := is.New(t)

	var cpu KB11
	cpu.Load(002000, 0105400) // NEG R0
	for d := 0; d < 16; d++ {
		dst := uint16(1) << d
		cpu.R[0] = dst
		cpu.R[7] = 002000
		cpu.step()
		t.Logf("R0: %06o, psw: %06o", dst, cpu.psw)
		result := (-dst) & 0xff
		is.Equal(cpu.R[0]&0xff, result)
		is.Equal(cpu.n(), result&0x80 > 0)
		is.Equal(cpu.z(), result&0xff == 0)
		is.Equal(cpu.v(), result == 0x80)
		is.Equal(cpu.c(), result != 0)
	}
}

func TestDEC(t *testing.T) {
	is := is.New(t)

	var cpu KB11
	cpu.Load(002000, 0005300) // DEC R0
	for d := 0; d < 16; d++ {
		dst := uint16(1) << d
		cpu.R[0] = dst
		cpu.R[7] = 002000
		cpu.step()
		t.Logf("R0: %06o, psw: %06o", dst, cpu.psw)
		result := dst - 1
		is.Equal(cpu.R[0], result)
		is.Equal(cpu.n(), result&0x8000 > 0)
		is.Equal(cpu.z(), result == 0)
		is.Equal(cpu.v(), result == msb(2)-1)
	}
}

func TestDECB(t *testing.T) {
	is := is.New(t)

	var cpu KB11
	cpu.Load(002000, 0105300) // DECB R0
	for d := 0; d < 16; d++ {
		dst := uint16(1) << d
		cpu.R[0] = dst
		cpu.R[7] = 002000
		cpu.step()
		t.Logf("R0: %06o, psw: %06o", dst, cpu.psw)
		result := (dst - 1) & 0xff
		is.Equal(cpu.R[0]&0xff, result)
		is.Equal(cpu.n(), result&0x80 > 0)
		is.Equal(cpu.z(), result == 0)
		is.Equal(cpu.v(), result == msb(1)-1)
	}
}

func TestINC(t *testing.T) {
	is := is.New(t)

	var cpu KB11
	cpu.Load(002000, 0005200) // INC R0
	for d := 0; d < 16; d++ {
		dst := uint16(1) << d
		cpu.R[0] = dst
		cpu.R[7] = 002000
		cpu.step()
		t.Logf("R0: %06o, psw: %06o", dst, cpu.psw)
		result := dst + 1
		is.Equal(cpu.R[0], result)
		is.Equal(cpu.n(), result&msb(2) > 0)
		is.Equal(cpu.z(), result == 0)
		is.Equal(cpu.v(), result == msb(2)-1)
	}
}

func TestINCB(t *testing.T) {
	is := is.New(t)

	var cpu KB11
	cpu.Load(002000, 0105200) // INCB R0
	for d := 0; d < 16; d++ {
		dst := uint16(1) << d
		cpu.R[0] = dst
		cpu.R[7] = 002000
		cpu.step()
		t.Logf("R0: %06o, psw: %06o", dst, cpu.psw)
		result := (dst + 1) & 0xff
		is.Equal(cpu.R[0]&0xff, result)
		is.Equal(cpu.n(), result&msb(1) > 0)
		is.Equal(cpu.z(), result == 0)
		is.Equal(cpu.v(), result == msb(1)-1)
	}
}

func TestCOM(t *testing.T) {
	is := is.New(t)

	var cpu KB11
	cpu.Load(002000, 0005100) // COM R0
	for d := 0; d < 16; d++ {
		dst := uint16(1) << d
		cpu.R[0] = dst
		cpu.R[7] = 002000
		cpu.step()
		t.Logf("R0: %06o, psw: %06o", dst, cpu.psw)
		result := ^dst
		is.Equal(cpu.R[0], result)
		is.Equal(cpu.n(), result&0x8000 == 0)
		is.Equal(cpu.z(), result == 0)
		is.Equal(cpu.v(), false)
		is.Equal(cpu.c(), true)
	}
}

func TestCOMB(t *testing.T) {
	is := is.New(t)

	var cpu KB11
	cpu.Load(002000, 0105100) // COMB R0
	for d := 0; d < 16; d++ {
		dst := uint16(1) << d
		cpu.R[0] = dst
		cpu.R[7] = 002000
		cpu.step()
		t.Logf("R0: %06o, psw: %06o", dst, cpu.psw)
		result := (^dst) & 0xff
		is.Equal(cpu.R[0]&0xff, result)
		is.Equal(cpu.n(), result&0x80 == 0)
		is.Equal(cpu.z(), result == 0)
		is.Equal(cpu.v(), false)
		is.Equal(cpu.c(), true)
	}
}

func TestCLR(t *testing.T) {
	is := is.New(t)

	var cpu KB11
	cpu.Load(002000, 0005000) // CLR R0
	for d := 0; d < 16; d++ {
		dst := uint16(1) << d
		cpu.R[0] = dst
		cpu.R[7] = 002000
		cpu.step()
		t.Logf("R0: %06o, psw: %06o", dst, cpu.psw)
		is.Equal(cpu.R[0], uint16(0))
		is.Equal(cpu.n(), false)
		is.Equal(cpu.z(), true)
		is.Equal(cpu.v(), false)
		is.Equal(cpu.c(), false)
	}
}

func TestCLRB(t *testing.T) {
	is := is.New(t)

	var cpu KB11
	cpu.Load(002000, 0105000) // CLRB R0
	for d := 0; d < 16; d++ {
		dst := uint16(1) << d
		cpu.R[0] = dst
		cpu.R[7] = 002000
		cpu.step()
		t.Logf("R0: %06o, psw: %06o", dst, cpu.psw)
		is.Equal(cpu.R[0]&0xff, uint16(0))
		is.Equal(cpu.n(), false)
		is.Equal(cpu.z(), true)
		is.Equal(cpu.v(), false)
		is.Equal(cpu.c(), false)
	}
}

func TestXOR(t *testing.T) {
	is := is.New(t)
	var cpu KB11
	for s := 0; s < 16; s++ {
		for d := 0; d < 16; d++ {
			src, dst := uint16(1)<<s, uint16(1)<<d
			cpu.R[0] = src
			cpu.R[1] = dst
			cpu.Load(002000, 0074001) // XOR R0, R1
			cpu.R[7] = 002000
			cpu.step()
			t.Logf("R0: %06o, R1: %06o", src, dst)
			is.Equal(cpu.R[1], src^dst)
			is.Equal(cpu.n(), (src^dst)&0x8000 > 0)
			is.Equal(cpu.z(), src^dst == 0)
			is.Equal(cpu.v(), false)
		}
	}
}

func TestASR(t *testing.T) {
	is := is.New(t)

	var cpu KB11
	cpu.Load(002000, 0006200) // ASR R0
	for d := 0; d < 16; d++ {
		dst := uint16(1) << d
		cpu.R[0] = dst
		cpu.R[7] = 002000
		cpu.step()
		t.Logf("R0: %06o", dst)
		result := uint16(int16(dst) >> 1)
		is.Equal(cpu.R[0], result)
		is.Equal(cpu.n(), (result)&0x8000 > 0)
		is.Equal(cpu.z(), (result) == 0)
		is.Equal(cpu.v(), (dst&1 == 1) != (result&0x8000 > 0))
		is.Equal(cpu.c(), (dst&1) == 1)
	}
}

func TestASRB(t *testing.T) {
	is := is.New(t)

	var cpu KB11
	cpu.Load(002000, 0106200) // ASRB R0
	for d := 0; d < 16; d++ {
		dst := uint16(1) << d
		cpu.R[0] = dst
		cpu.R[7] = 002000
		cpu.step()
		t.Logf("R0: %06o", dst)
		result := uint16(uint8(int8(dst) / 2))
		is.Equal(cpu.R[0]&0xff, result)
		is.Equal(cpu.n(), dst&0x80 > 0)
		is.Equal(cpu.z(), (result) == 0)
		is.Equal(cpu.v(), (dst&1 == 1) != (result&0x80 > 0))
		is.Equal(cpu.c(), (dst&1) == 1)
	}
}

func TestASL(t *testing.T) {
	is := is.New(t)

	var cpu KB11
	cpu.Load(002000, 0006300) // ASL R0
	for d := 0; d < 16; d++ {
		dst := uint16(1) << d
		cpu.R[0] = dst
		cpu.R[7] = 002000
		cpu.step()
		t.Logf("R0: %06o", dst)
		result := dst << 1
		is.Equal(cpu.R[0], result)
		is.Equal(cpu.n(), (result)&0x8000 > 0)
		is.Equal(cpu.z(), (result) == 0)
		is.Equal(cpu.v(), cpu.n() != (result&0x8000 > 0))
		is.Equal(cpu.c(), (result&0x8000) > 0)
	}
}

func TestASLB(t *testing.T) {
	is := is.New(t)

	var cpu KB11
	cpu.Load(002000, 0106300) // ASLB R0
	for d := 0; d < 16; d++ {
		dst := uint16(1) << d
		cpu.R[0] = dst
		cpu.R[7] = 002000
		cpu.step()
		t.Logf("R0: %06o", dst)
		result := uint16(uint8(dst) << 1)
		is.Equal(cpu.R[0]&0xff, result)
		is.Equal(cpu.n(), (result)&0x80 > 0)
		is.Equal(cpu.z(), (result)&0xff == 0)
		is.Equal(cpu.v(), cpu.n() != (result&0x80 > 0))
		is.Equal(cpu.c(), (result&0x80) > 0)
	}
}

func TestASH(t *testing.T) {
	is := is.New(t)
	var cpu KB11
	for s := 0; s < 16; s++ {
		for d := 0; d < 64; d++ {
			src, dst := uint16(1)<<s, uint16(d)
			cpu.R[0] = src
			cpu.R[1] = dst
			cpu.Load(002000, 0072001) // ASH R0, R1
			cpu.R[7] = 002000
			cpu.step()
			t.Logf("R0: %06o, R1: %06o", src, dst)
			switch dst & 040 {
			case 000:
				// positive is shift left
				shift := dst & 037
				result := src << shift
				is.Equal(cpu.R[0], result)
				is.Equal(cpu.n(), result&0x8000 > 0)
				is.Equal(cpu.z(), result == 0)
				is.Equal(cpu.v(), (result&0x8000 > 0) != (src&0x8000 > 0))
				is.Equal(cpu.c(), src&(1<<(16-shift)) > 0)
			case 040:
				// negative is shift right
				shift := 077 ^ dst + 1
				result := uint16(int16(src) >> shift)
				is.Equal(cpu.R[0], result)
				is.Equal(cpu.n(), result&0x8000 > 0)
				is.Equal(cpu.z(), result == 0)
				is.Equal(cpu.v(), (result&0x8000 > 0) != (src&0x8000 > 0))
				is.Equal(cpu.c(), src>>(shift-1) == 1)
			default:
				t.Fatal("impossible shift")
			}
		}
	}
}

func TestASHC(t *testing.T) {
	is := is.New(t)
	var cpu KB11
	for s := 0; s < 32; s++ {
		for d := 0; d < 64; d++ {
			reg1, reg2, dst := uint16((uint32(1)<<s)>>16), uint16(1<<s), uint16(d)
			cpu.R[0] = reg1
			cpu.R[1] = reg2
			cpu.R[2] = dst
			cpu.Load(002000, 0073002) // ASHC R0, R1
			cpu.R[7] = 002000
			cpu.step()
			t.Logf("R0: %06o, R1: %06o, R2: %06o", reg1, reg2, dst)
			switch dst & 040 {
			case 000:
				// positive is shift left
				shift := dst & 037
				result := (uint32(reg1)<<16 | uint32(reg2)) << shift
				is.Equal(cpu.R[0], uint16(result>>16))
				is.Equal(cpu.R[1], uint16(result))
				is.Equal(cpu.n(), result&0x80000000 > 0)
				is.Equal(cpu.z(), result == 0)
				is.Equal(cpu.v(), (result&0x80000000 > 0) != (reg1&0x8000 > 0))
				is.Equal(cpu.c(), (uint32(reg1)<<16|uint32(reg2))&(1<<(32-shift)) > 0)
			case 040:
				// negative is shift right
				shift := 077 ^ dst + 1
				result := uint32(int32((uint32(reg1)<<16 | uint32(reg2))) >> shift)
				is.Equal(cpu.R[0], uint16(result>>16))
				is.Equal(cpu.R[1], uint16(result))
				is.Equal(cpu.n(), result&0x80000000 > 0)
				is.Equal(cpu.z(), result == 0)
				is.Equal(cpu.v(), (result&0x80000000 > 0) != (reg1&0x8000 > 0))
				is.Equal(cpu.c(), (uint32(reg1)<<16|uint32(reg2))>>(shift-1) == 1)
			default:
				t.Fatal("impossible shift")
			}
		}
	}
}

func BenchmarkADD(b *testing.B) {
	var cpu KB11
	cpu.Load(0002000,
		0060001, // ADD R0, R1
	)
	for i := 0; i < b.N; i++ {
		cpu.R[0] = uint16(i)
		cpu.R[1] = uint16(i)
		cpu.R[7] = 0002000
		cpu.step()
	}
}

func BenchmarkSUB(b *testing.B) {
	var cpu KB11
	cpu.Load(0002000,
		0160001, // SUB R0, R1
	)
	for i := 0; i < b.N; i++ {
		cpu.R[0] = uint16(i)
		cpu.R[1] = uint16(i)
		cpu.R[7] = 0002000
		cpu.step()
	}
}

func BenchmarkTST(b *testing.B) {
	var cpu KB11
	cpu.Load(0002000,
		0005700, // TST R0
	)
	for i := 0; i < b.N; i++ {
		cpu.R[0] = uint16(i)
		cpu.R[7] = 0002000
		cpu.step()
	}
}

func BenchmarkTSTB(b *testing.B) {
	var cpu KB11
	cpu.Load(0002000,
		0105700, // TSTB R0
	)
	for i := 0; i < b.N; i++ {
		cpu.R[0] = uint16(i)
		cpu.R[7] = 0002000
		cpu.step()
	}
}

func BenchmarkNEG(b *testing.B) {
	var cpu KB11
	cpu.Load(0002000,
		0005400, // NEG R0
	)
	for i := 0; i < b.N; i++ {
		cpu.R[0] = uint16(i)
		cpu.R[7] = 0002000
		cpu.step()
	}
}

func BenchmarkNEGB(b *testing.B) {
	var cpu KB11
	cpu.Load(0002000,
		0105400, // NEGB R0
	)
	for i := 0; i < b.N; i++ {
		cpu.R[0] = uint16(i)
		cpu.R[7] = 0002000
		cpu.step()
	}
}
