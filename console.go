package main

import (
	"fmt"
	"os"
)

type KL11 struct {
	rcsr, rbuf, xcsr, xbuf uint16
	count                  uint8
}

func (kl *KL11) clearterminal() {
	kl.rcsr = 0
	kl.xcsr = 0x80
	kl.rbuf = 0
	kl.xbuf = 0
}

func (kl *KL11) writeterminal(char uint16) {
	var outb [1]byte
	switch char {
	case 13:
		// skip
	default:
		outb[0] = byte(char)
		os.Stdout.Write(outb[:])
	}
}

func (kl *KL11) addchar(c uint16) {
	if (kl.rcsr & 0x80) > 0 {
		// unit not busy
		kl.rbuf = c
		kl.rcsr |= 0x80
		if kl.rcsr&0x40 > 0 {
			panic(interrupt{INTTTYIN, 4})
		}
	}
}

func (kl *KL11) read16(a addr18) uint16 {
	switch a & 06 {
	case 00:
		return kl.rcsr
	case 02:
		if kl.rcsr&0x80 > 0 {
			kl.rcsr &= 0xff7e
			return kl.rbuf
		}
		return 0
	case 04:
		return kl.xcsr
	case 06:
		return 0
	default:
		fmt.Printf("KL11: read from invalid address %06o\n", a)
		panic(trap{INTBUS})
	}
}

func (kl *KL11) write16(a addr18, v uint16) {
	switch a & 06 {
	case 00:
		if v&(1<<6) > 0 {
			kl.rcsr |= 1 << 6
		} else {
			kl.rcsr &= ^uint16(1 << 6)
		}
	case 04:
		if v&(1<<6) > 0 {
			kl.xcsr |= 1 << 6
		} else {
			kl.xcsr &= ^uint16(1 << 6)
		}
	case 06:
		kl.xbuf = v & 0x7f
		kl.xcsr &= 0xff7f
		if kl.xcsr&0x80 == 0 {
			kl.writeterminal(kl.xbuf & 0x7f)
			kl.xcsr |= 0x80
			if kl.xcsr&(1<<6) > 0 {
				panic(interrupt{INTTTYOUT, 4})
			}
		}
	default:
		fmt.Printf("KL11: write to invalid address %06o\n", a)
		panic(trap{INTBUS})
	}
}
