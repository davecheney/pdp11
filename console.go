package main

import (
	"fmt"
	"os"
)

type KL11 struct {
	rcsr, rbuf, xcsr, xbuf uint16
	xmitcount, recvcount   int
	Input                  chan byte
}

func (kl *KL11) reset() {
	kl.rcsr = 0
	kl.rbuf = 0
	kl.xcsr = 0x80
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
	switch a {
	case 0777560:
		// 777560 Receive Control and Status register
		return kl.rcsr
	case 0777562:
		// 777562 Receive Buffer
		kl.rcsr &= ^uint16(1 << 7)
		return kl.rbuf & 0xff
	case 0777564:
		// 777564 Transmit Control and Status register
		return kl.xcsr
	case 0777566:
		// 777566 Transmit Buffer
		return 0 // write only
	default:
		fmt.Printf("KL11: read from invalid address %06o\n", a)
		panic(trap{INTBUS})
	}
}

func (kl *KL11) write16(a addr18, v uint16) {
	switch a {
	case 0777560:
		// 777560 Receive Control and Status register
		kl.rcsr = v
	case 0777562:
		// 777562 Receive Buffer
		kl.rcsr &= ^uint16(1 << 7)
		// read only, write reset rcsr.done
	case 0777564:
		// 777564 Transmit Control and Status register
		kl.xcsr = v
	case 0777566:
		// 777566 Transmit Buffer
		fmt.Printf("kl11:write16: %06o %06o\n", a, v)
		kl.xbuf = v & 0xff
		kl.xcsr &= ^uint16(1 << 7)
	default:
		fmt.Printf("KL11: write to invalid address %06o\n", a)
		panic(trap{INTBUS})
	}
}

func (kl *KL11) step() {
	if kl.xbuf > 0 {
		kl.writeterminal(kl.xbuf)
		kl.xmitcount = 29
		kl.xbuf = 0
	}
	if kl.xmitcount > 0 {
		kl.xmitcount--
	}
	if kl.xmitcount == 0 {
		kl.xcsr |= (1 << 7)
		if kl.xcsr&1<<6 > 0 {
			panic(interrupt{INTTTYOUT, 4})
		}
	}
	if kl.rcsr&1<<7 == 0 {
		select {
		case c := <-kl.Input:
			kl.rbuf = uint16(c)
			kl.rcsr |= (1 << 7)
			if kl.rcsr&1<<6 > 0 {
				panic(interrupt{INTTTYIN, 4})
			}
		default:
		}
	}
}
