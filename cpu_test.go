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
			is.Equal(cpu.c(), !(dst+(^src)+1 < 0xffff))
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
			result := src + (^dst) + 1
			t.Logf("R0: %06o, R1: %06o", src, dst)
			is.Equal(cpu.R[0], src)
			is.Equal(cpu.R[1], dst)
			is.Equal(cpu.n(), result&0x8000 > 0)
			is.Equal(cpu.z(), src-dst == 0)
			is.Equal(cpu.v(), (src^dst)&0x8000 > 0 && !((dst^(src+(^dst)+1))&0x8000 > 0))
			is.Equal(cpu.c(), result == 0xffff)
		}
	}
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
			result := src + (^dst) + 1
			t.Logf("R0: %06o, R1: %06o", src, dst)
			is.Equal(cpu.R[0], src)
			is.Equal(cpu.R[1], dst)
			is.Equal(cpu.n(), result&0x80 > 0)
			is.Equal(cpu.z(), result&0xff == 0)
			is.Equal(cpu.v(), (src^dst)&0x80 > 0 && !((dst^result)&0x80 > 0))
			is.Equal(cpu.c(), result&0xff == 0xff)
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
		is.Equal(cpu.c(), dst != 0)
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
		is.Equal(cpu.c(), dst&0xff != 0)
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
		is.Equal(cpu.z(), true)
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
		is.Equal(cpu.z(), false)
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
		is.Equal(cpu.v(), result == 0x8000)
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
		is.Equal(cpu.v(), result == 0x80)
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
		is.Equal(cpu.n(), result&0x8000 > 0)
		is.Equal(cpu.z(), result == 0)
		is.Equal(cpu.v(), result == 0x8000)
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
		is.Equal(cpu.n(), result&0x80 > 0)
		is.Equal(cpu.z(), result == 0)
		is.Equal(cpu.v(), result == 0x80)
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
