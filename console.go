// +build none

package main

package pdp11

import (
	"fmt"
	"os"
)

type KL11 struct {
	TKS, TKB, TPS, TPB int

	Input chan uint8
	count uint8 // step delay
	ready bool

	unibus *unibus
}

func (kl *KL11) clearterminal() {
	kl.TKS = 0
	kl.TPS = 1 << 7
	kl.TKB = 0
	kl.TPB = 0
	kl.ready = true
}

func (kl *KL11) writeterminal(char int) {
	var outb [1]byte
	switch char {
	case 13:
		// skip
	default:
		outb[0] = byte(char)
		os.Stdout.Write(outb[:])
	}
}

func (kl *KL11) addchar(char int) {
	switch char {
	case 42:
		kl.TKB = 4
	case 19:
		kl.TKB = 034
	case '\n':
		kl.TKB = '\r'
	default:
		kl.TKB = char
	}
	kl.TKS |= 0x80
	kl.ready = false
	if kl.TKS&(1<<6) != 0 {
		kl.cpu.interrupt(intTTYIN, 4)
	}
}

func (c *Console) getchar() int {
	if c.TKS&0x80 == 0x80 {
		c.TKS &= 0xff7e
		c.ready = true
		return c.TKB
	}
	return 0
}

func (c *Console) Step() {
	if c.ready {
		select {
		case v, ok := <-c.Input:
			if ok {
				c.addchar(int(v))
			}
		default:
		}
	}
	c.count++
	if c.count%32 != 0 {
		return
	}
	if c.TPS&0x80 == 0 {
		c.writeterminal(c.TPB & 0x7f)
		c.TPS |= 0x80
		if c.TPS&(1<<6) != 0 {
			c.unibus.cpu.interrupt(intTTYOUT, 4)
		}
	}
}

func (c *Console) consread16(a uint18) int {
	switch a {
	case 0777560:
		return c.TKS
	case 0777562:
		return c.getchar()
	case 0777564:
		return c.TPS
	case 0777566:
		return 0
	default:
		panic(fmt.Sprintf("read from invalid address %06o", a))
	}
}

func (c *Console) conswrite16(a uint18, v int) {
	switch a {
	case 0777560:
		if v&(1<<6) != 0 {
			c.TKS |= 1 << 6
		} else {
			c.TKS &= ^(1 << 6)
		}
	case 0777564:
		if v&(1<<6) != 0 {
			c.TPS |= 1 << 6
		} else {
			c.TPS &= ^(1 << 6)
		}
	case 0777566:
		c.TPB = v & 0xff
		c.TPS &= 0xff7f
	default:
		panic(fmt.Sprintf("write to invalid address %06o", a))
	}
}
