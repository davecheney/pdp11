package main

import (
	"encoding/binary"
	"fmt"
	"io/ioutil"
)

const (
	RKOVR = (1 << 14)
	RKNXD = (1 << 7)
	RKNXC = (1 << 6)
	RKNXS = (1 << 5)
)

type RK05 struct {
	buf []byte
	pos uint32
}

func (rk *RK05) write16(v uint16) {
	binary.LittleEndian.PutUint16(rk.buf[rk.pos:], v)
	rk.pos += 2
}

func (rk *RK05) read16() uint16 {
	v := binary.LittleEndian.Uint16(rk.buf[rk.pos:])
	rk.pos += 2
	return v
}

type RK11 struct {
	rkds, rker, rkcs, rkwc, rkba, rkda uint16
	drive, sector, surface, cylinder   uint32

	units [8]RK05

	unibus *UNIBUS
}

func (rk *RK11) Mount(unit int, path string) error {
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	rk.units[unit].buf = buf
	return nil
}

func (rk *RK11) read16(a addr18) uint16 {
	//fmt.Printf("rk11:read16: %06o\n", a)
	switch a {
	case 0777400:
		// 777400 Drive Status
		return rk.rkds
	case 0777402:
		// 777402 Error Register
		return rk.rker
	case 0777404:
		// 777404 Control Status
		return rk.rkcs & 0xfffe // go bit is write only
	// case 0777406:
	// 	// 777406 Word Count
	// 	return rk.rkwc
	case 0777412:
		return rk.rkda
	default:
		fmt.Printf("rk11::read16 invalid read %06o\n", a)
		panic(trap{INTBUS})
	}
}

func (rk *RK11) write16(a addr18, v uint16) {
	// fmt.Printf("rk11:write16: %06o %06o\n", a, v)
	switch a {
	case 0777404:
		// RKCS
		rk.rkcs &= 0xf080
		rk.rkcs |= v & ^uint16(0xf080) // Bits 7 and 12 - 15 are read only
	case 0777406:
		// RKWC
		rk.rkwc = v
	case 0777410:
		// RKBA
		rk.rkba = v
	case 0777412:
		// RKDA
		rk.rkda = v
		rk.drive = uint32(v >> 13)
		rk.cylinder = uint32(v>>5) & 0377
		rk.surface = uint32(v>>4) & 1
		rk.sector = uint32(v & 15)
	default:
		fmt.Printf("rk11::write16 invalid write %06o: %06o\n", a, v)
		panic(trap{INTBUS})
	}
}

func (rk *RK11) rknotready() {
	// fmt.Println("rk11: not ready")
	rk.rkds &^= 1 << 6
	rk.rkcs &^= 1 << 7
}

func (rk *RK11) rkready() {
	// fmt.Println("rk11: ready")
	rk.rkds |= 1 << 6
	rk.rkcs |= 1 << 7
	rk.rkcs &^= 1 // no go
}

func (rk *RK11) _go() bool {
	return rk.rkcs&1 == 1
}

func (rk *RK11) step() {
	if !rk._go() {
		// no GO bit
		return
	}

	switch (rk.rkcs >> 1) & 7 {
	case 0:
		// controller reset
		rk.reset()
	case 1, 2, 3: // write, read, check
		if rk.drive != 0 {
			rk.rker |= 0x8080 // NXD
			break
		}
		if rk.cylinder > 0312 {
			rk.rker |= 0x8040 // NXC
			break
		}
		if rk.sector > 013 {
			rk.rker |= 0x8020 // NXS
			break
		}
		rk.rknotready()
		rk.seek()
		rk.readwrite()
	case 6: // Drive Reset - falls through to be finished as a seek
		rk.rker = 0
		fallthrough
	case 4: // Seek (and drive reset) - complete immediately
		rk.seek()
		rk.rkcs &= ^uint16(0x2000) // Clear search complete - reset by rk11_seekEnd
		rk.rkcs |= 0x80            // set done - ready to accept new command
		panic(interrupt{INTRK, 5})
	case 5: // Read Check
		break
	case 7: // Write Lock - not implemented :-(
		break
	default:
		panic(fmt.Sprintf("unimplemented RK05 operation %06o\n", ((rk.rkcs & 017) >> 1)))
	}
}

func (rk *RK11) readwrite() {
	if rk.rkwc == 0 {
		rk.rkready()
		if rk.rkcs&(1<<6) > 0 {
			panic(interrupt{INTRK, 5})
		}
		return
	}

	w := ((rk.rkcs >> 1) & 7) == 1
	// fmt.Printf("rk11: step: RKCS: %06o RKBA: %06o RKWC: %06o cylinder: %03o surface: %03o sector: %03o write: %v rker: %06o\n", rk.rkcs, rk.rkba, rk.rkwc, rk.cylinder, rk.surface, rk.sector, w, rk.rker)

	for i := 0; i < 256 && rk.rkwc != 0; i++ {
		if w {
			val := rk.unibus.read16(addr18(rk.rkba))
			rk.units[rk.drive].write16(val)
		} else {
			val := rk.units[rk.drive].read16()
			rk.unibus.write16(addr18(rk.rkba), val)
		}
		rk.rkba += 2
		rk.rkwc++
	}
	rk.sector++
	if rk.sector > 013 {
		rk.sector = 0
		rk.surface++
		if rk.surface > 1 {
			rk.surface = 0
			rk.cylinder++
			if rk.cylinder > 0312 {
				rk.rker |= RKOVR
				return
			}
		}
	}
}

func (rk *RK11) seek() {
	rk.units[rk.drive].pos = (rk.cylinder*24 + rk.surface*12 + rk.sector) * 512
	if rk.units[rk.drive].pos > uint32(len(rk.units[rk.drive].buf)) {
		panic(fmt.Sprintf("rkstep: failed to seek\n"))
	}
}

func (rk *RK11) reset() {
	fmt.Println("rk11: reset")
	rk.rkds = 04700 // Set bits 6, 7, 8, 11
	rk.rker = 0
	rk.rkcs = 0200
	rk.rkwc = 0
	rk.rkba = 0
	rk.rkda = 0
	rk.drive = 0
	rk.cylinder = 0
	rk.surface = 0
	rk.sector = 0
}
