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

func (kl *KL11) read16(a addr18) uint16 {
	// fmt.Printf("kl11:read16: %06o\n", a)
	switch a {
	case 0777560:
		// 777560 Receive Control and Status register
		return kl.rcsr
	case 0777562:
		// 777562 Receive Buffer
		if (kl.rcsr & 0x80) > 0 {
			kl.rcsr &= 0xff7e
			return kl.rbuf
		}
		return 0
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
	fmt.Printf("kl11:write16: %06o %06o\n", a, v)
	switch a {
	case 0777560:
		// 777560 Receive Control and Status register
		if v&(1<<6) > 0 {
			kl.rcsr |= 1 << 6
		} else {
			kl.rcsr &= ^uint16(1 << 6)
		}
	// case 0777562:
	// 777562 Receive Buffer
	// kl.rcsr &= ^uint16(1 << 7)
	// read only, write reset rcsr.done
	case 0777564:
		// 777564 Transmit Control and Status register
		if v&1<<6 > 0 {
			kl.xcsr |= 1 << 6
		} else {
			kl.xcsr &= ^uint16(1 << 6)
		}
	case 0777566:
		// 777566 Transmit Buffer
		kl.xbuf = v & 0x7f
		kl.xcsr &= 0xff7f
		kl.xmitcount = 0
	default:
		fmt.Printf("KL11: write to invalid address %06o\n", a)
		panic(trap{INTBUS})
	}
}

func (kl *KL11) step() {
	if kl.rcsr&0x80 == 0 {
		// receiver not busy, poll for character
		select {
		case c := <-kl.Input:
			kl.rbuf = uint16(c & 0x7f)
			kl.rcsr |= 0x80
			if kl.rcsr&0x40 > 0 {
				panic(interrupt{INTTTYIN, 4})
			}
		default:
		}
	}
	if kl.xcsr&0x80 == 0 {
		kl.xmitcount++
		if kl.xmitcount > 32 {
			os.Stderr.Write([]byte{byte(kl.xbuf & 0x7f)})
			kl.xcsr |= 0x80
			if kl.xcsr&(1<<6) > 0 {
				panic(interrupt{INTTTYOUT, 4})
			}
		}
	}
}
