package main

import "testing"

func TestADD(t *testing.T) {
	expect := func(got, want interface{}) {
		if got != want {
			t.Helper()
			t.Fatal("got:", want, "want:", got)
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
			t.Fatal("got:", want, "want:", got)
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
			expect(cpu.R[1], src-dst)
			expect(cpu.n(), (src-dst)&0x8000 > 0)
			expect(cpu.z(), src-dst == 0)
			expect(cpu.v(), (((src^dst)&0x8000 > 0) && (!((dst^(src+dst))&0x8000 > 0))))
			expect(cpu.c(), (dst)+(^src)+1 < 0xffff)
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
