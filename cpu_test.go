package main

import "testing"

func TestADD(t *testing.T) {
	expect := func(got, want interface{}) {
		if got != want {
			t.Helper()
			t.Error("got:", got, "want:", want)
		}
	}

	var cpu KB11
	for s := 0; s < 16; s++ {
		for d := 0; d < 16; d++ {
			src, dst := uint16(1)<<s, uint16(1)<<d
			cpu.R[0] = src
			cpu.R[1] = dst
			cpu.ADD(0060001) // ADD R0, R1
			t.Logf("R0: %06o, R1: %06o", src, dst)
			expect(cpu.R[1], src+dst)
			expect(cpu.n(), (src+dst)&0x8000 > 0)
			expect(cpu.z(), src+dst == 0)
			expect(cpu.v(), (!((src^dst)&0x8000 > 0) && ((dst^(src+dst))&0x8000 > 0)))
			expect(cpu.c(), uint32(src)+uint32(dst) > 0xffff)
		}
	}
}

func TestADC(t *testing.T) {
	expect := func(got, want interface{}) {
		if got != want {
			t.Helper()
			t.Error("got:", got, "want:", want)
		}
	}

	var cpu KB11
	for d := 0; d < 16; d++ {
		dst := uint16(1) << d
		cpu.R[0] = dst
		cpu.psw &= ^uint16(FLAGC)
		cpu.ADC(2, 0005500) // ADC R0
		t.Logf("R0: %06o", dst)
		expect(cpu.R[0], dst)
		expect(cpu.n(), dst&0x8000 > 0)
		expect(cpu.z(), dst == 0)
		expect(cpu.v(), false)
		expect(cpu.c(), false)
	}

	for d := 0; d < 16; d++ {
		dst := uint16(1) << d
		cpu.R[0] = dst
		cpu.psw |= FLAGC
		cpu.ADC(2, 0005500) // ADC R0
		t.Logf("R0: %06o", dst)
		expect(cpu.R[0], dst+1)
		expect(cpu.n(), (dst+1)&0x8000 > 0)
		expect(cpu.z(), dst+1 == 0)
		expect(cpu.v(), dst == 0077777)
		expect(cpu.c(), dst == 0177777)
	}
}

func TestADCB(t *testing.T) {
	expect := func(got, want interface{}) {
		if got != want {
			t.Helper()
			t.Error("got:", got, "want:", want)
		}
	}

	var cpu KB11
	for d := 0; d < 16; d++ {
		dst := uint16(1) << d
		cpu.R[0] = dst
		cpu.psw &= ^uint16(FLAGC)
		cpu.ADC(1, 0105500) // ADCB R0
		t.Logf("R0: %06o", dst)
		expect(cpu.R[0]&0xff, dst&0xff)
		expect(cpu.n(), dst&0x80 > 0)
		expect(cpu.z(), dst&0xff == 0)
		expect(cpu.v(), false)
		expect(cpu.c(), false)
	}

	for d := 0; d < 16; d++ {
		dst := uint16(1) << d
		cpu.R[0] = dst
		cpu.psw |= FLAGC
		cpu.ADC(1, 0105500) // ADCB R0
		t.Logf("R0: %06o", dst)
		expect(cpu.R[0]&0xff, (dst+1)&0xff)
		expect(cpu.n(), (dst+1)&0x80 > 0)
		expect(cpu.z(), (dst+1)&0xff == 0)
		expect(cpu.v(), (dst+1)&0xff == 0200)
		expect(cpu.c(), (dst+1)&0xff == 0)
	}
}

func TestSUB(t *testing.T) {
	expect := func(got, want interface{}) {
		if got != want {
			t.Helper()
			t.Error("got:", got, "want:", want)
		}
	}

	var cpu KB11
	for s := 0; s < 16; s++ {
		for d := 0; d < 16; d++ {
			src, dst := uint16(1)<<s, uint16(1)<<d
			cpu.R[0] = src
			cpu.R[1] = dst
			cpu.SUB(0160001) // SUB R0, R1
			t.Logf("R0: %06o, R1: %06o", src, dst)
			expect(cpu.R[1], dst-src)
			expect(cpu.n(), (dst-src)&0x8000 > 0)
			expect(cpu.z(), dst-src == 0)
			expect(cpu.v(), (((src^dst)&0x8000 > 0) && (!((dst^(dst-src))&0x8000 > 0))))
			expect(cpu.c(), !(dst+(^src)+1 < 0xffff))
		}
	}
}

func TestTST(t *testing.T) {
	expect := func(got, want interface{}) {
		if got != want {
			t.Helper()
			t.Error("got:", got, "want:", want)
		}
	}

	var cpu KB11
	for d := 0; d < 16; d++ {
		dst := uint16(1) << d
		cpu.R[0] = dst
		cpu.TST(2, 0005700) // TST R0
		t.Logf("R0: %06o", dst)
		expect(cpu.n(), (dst)&0x8000 > 0)
		expect(cpu.z(), dst == 0)
	}
}

func TestTSTB(t *testing.T) {
	expect := func(got, want interface{}) {
		if got != want {
			t.Helper()
			t.Error("got:", got, "want:", want)
		}
	}

	for d := 0; d < 16; d++ {
		var cpu KB11
		dst := uint16(1) << d
		cpu.R[0] = dst
		cpu.TST(1, 0105700) // TSTB R0
		t.Logf("R0: %06o, psw: %06o", dst, cpu.psw)
		expect(cpu.n(), dst&0x80 > 0)
		expect(cpu.z(), (dst&0xff) == 0)
	}
}

func TestNEG(t *testing.T) {
	expect := func(got, want interface{}) {
		if got != want {
			t.Helper()
			t.Error("got:", got, "want:", want)
		}
	}

	for d := 0; d < 16; d++ {
		var cpu KB11
		dst := uint16(1) << d
		cpu.R[0] = dst
		cpu.NEG(2, 0005400) // NEG R0
		t.Logf("R0: %06o, psw: %06o", dst, cpu.psw)
		expect(cpu.R[0], -dst)
		expect(cpu.n(), -dst&0x8000 > 0)
		expect(cpu.z(), -dst == 0)
		expect(cpu.v(), -dst == 0x8000)
		expect(cpu.c(), -dst != 0)
	}
}

func TestNEGB(t *testing.T) {
	expect := func(got, want interface{}) {
		if got != want {
			t.Helper()
			t.Error("got:", got, "want:", want)
		}
	}

	for d := 0; d < 16; d++ {
		var cpu KB11
		dst := uint16(1) << d
		cpu.R[0] = dst
		cpu.NEG(1, 0105400) // NEGB R0
		t.Logf("R0: %06o, psw: %06o", dst, cpu.psw)
		result := (-dst) & 0xff
		expect(cpu.R[0]&0xff, result)
		expect(cpu.n(), result&0x80 > 0)
		expect(cpu.z(), result&0xff == 0)
		expect(cpu.v(), result == 0x80)
		expect(cpu.c(), result != 0)
	}
}

func TestDEC(t *testing.T) {
	expect := func(got, want interface{}) {
		if got != want {
			t.Helper()
			t.Error("got:", got, "want:", want)
		}
	}

	for d := 0; d < 16; d++ {
		var cpu KB11
		dst := uint16(1) << d
		cpu.R[0] = dst
		cpu.DEC(2, 0005300) // DEC R0
		t.Logf("R0: %06o, psw: %06o", dst, cpu.psw)
		result := dst - 1
		expect(cpu.R[0], result)
		expect(cpu.n(), result&0x8000 > 0)
		expect(cpu.z(), result == 0)
		expect(cpu.v(), result == 0x8000)
	}
}

func TestDECB(t *testing.T) {
	expect := func(got, want interface{}) {
		if got != want {
			t.Helper()
			t.Error("got:", got, "want:", want)
		}
	}

	for d := 0; d < 16; d++ {
		var cpu KB11
		dst := uint16(1) << d
		cpu.R[0] = dst
		cpu.DEC(1, 0105300) // DECB R0
		t.Logf("R0: %06o, psw: %06o", dst, cpu.psw)
		result := (dst - 1) & 0xff
		expect(cpu.R[0]&0xff, result)
		expect(cpu.n(), result&0x80 > 0)
		expect(cpu.z(), result == 0)
		expect(cpu.v(), result == 0x80)
	}
}

func TestINC(t *testing.T) {
	expect := func(got, want interface{}) {
		if got != want {
			t.Helper()
			t.Error("got:", got, "want:", want)
		}
	}

	for d := 0; d < 16; d++ {
		var cpu KB11
		dst := uint16(1) << d
		cpu.R[0] = dst
		cpu.INC(2, 0005200) // INC R0
		t.Logf("R0: %06o, psw: %06o", dst, cpu.psw)
		result := dst + 1
		expect(cpu.R[0], result)
		expect(cpu.n(), result&0x8000 > 0)
		expect(cpu.z(), result == 0)
		expect(cpu.v(), result == 0x8000)
	}
}

func TestINCB(t *testing.T) {
	expect := func(got, want interface{}) {
		if got != want {
			t.Helper()
			t.Error("got:", got, "want:", want)
		}
	}

	for d := 0; d < 16; d++ {
		var cpu KB11
		dst := uint16(1) << d
		cpu.R[0] = dst
		cpu.INC(1, 0105200) // INCB R0
		t.Logf("R0: %06o, psw: %06o", dst, cpu.psw)
		result := (dst + 1) & 0xff
		expect(cpu.R[0]&0xff, result)
		expect(cpu.n(), result&0x80 > 0)
		expect(cpu.z(), result == 0)
		expect(cpu.v(), result == 0x80)
	}
}

func TestCOM(t *testing.T) {
	expect := func(got, want interface{}) {
		if got != want {
			t.Helper()
			t.Error("got:", got, "want:", want)
		}
	}

	for d := 0; d < 16; d++ {
		var cpu KB11
		dst := uint16(1) << d
		cpu.R[0] = dst
		cpu.COM(2, 0005100) // COM R0
		t.Logf("R0: %06o, psw: %06o", dst, cpu.psw)
		result := ^dst
		expect(cpu.R[0], result)
		expect(cpu.n(), result&0x8000 == 0)
		expect(cpu.z(), result == 0)
		expect(cpu.v(), false)
		expect(cpu.c(), true)
	}
}

func TestCOMB(t *testing.T) {
	expect := func(got, want interface{}) {
		if got != want {
			t.Helper()
			t.Error("got:", got, "want:", want)
		}
	}

	for d := 0; d < 16; d++ {
		var cpu KB11
		dst := uint16(1) << d
		cpu.R[0] = dst
		cpu.COM(1, 0105100) // COMB R0
		t.Logf("R0: %06o, psw: %06o", dst, cpu.psw)
		result := (^dst) & 0xff
		expect(cpu.R[0]&0xff, result)
		expect(cpu.n(), result&0x80 == 0)
		expect(cpu.z(), result == 0)
		expect(cpu.v(), false)
		expect(cpu.c(), true)
	}
}

func TestCLR(t *testing.T) {
	expect := func(got, want interface{}) {
		if got != want {
			t.Helper()
			t.Error("got:", got, "want:", want)
		}
	}

	for d := 0; d < 16; d++ {
		var cpu KB11
		dst := uint16(1) << d
		cpu.R[0] = dst
		cpu.CLR(2, 0005000) // CLR R0
		t.Logf("R0: %06o, psw: %06o", dst, cpu.psw)
		expect(cpu.R[0], uint16(0))
		expect(cpu.n(), false)
		expect(cpu.z(), true)
		expect(cpu.v(), false)
		expect(cpu.c(), false)
	}
}

func TestCLRB(t *testing.T) {
	expect := func(got, want interface{}) {
		if got != want {
			t.Helper()
			t.Error("got:", got, "want:", want)
		}
	}

	for d := 0; d < 16; d++ {
		var cpu KB11
		dst := uint16(1) << d
		cpu.R[0] = dst
		cpu.CLR(1, 0105000) // CLRB R0
		t.Logf("R0: %06o, psw: %06o", dst, cpu.psw)
		expect(cpu.R[0]&0xff, uint16(0))
		expect(cpu.n(), false)
		expect(cpu.z(), true)
		expect(cpu.v(), false)
		expect(cpu.c(), false)
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
