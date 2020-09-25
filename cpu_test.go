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
