// pdp11 emulator.
package main

import (
	"log"
	"os"
	"time"

	"github.com/alecthomas/kong"
	"golang.org/x/sys/unix"
)

func main() {
	var cli struct {
		Run runCmd `cmd default:"1" help:"help yourself to a PDP11"`
	}

	ctx := kong.Parse(&cli)
	err := ctx.Run(&kong.Context{})
	ctx.FatalIfErrorf(err)
}

type runCmd struct {
	StartAddr uint16 `name:"startaddr" default:"1026" help:"pc start address in decimal"`
	RK0       string `name:"rk0" type:"existingfile" help:"path to rk0 image"`
}

func (r *runCmd) Run(ctx *kong.Context) error {
	fd := os.Stdin.Fd()
	oldattr, err := tcget(fd)
	if err != nil {
		return err
	}
	defer tcset(fd, oldattr)

	attr := *oldattr
	// disable canonical mode processing in the line discipline driver
	attr.Iflag &^= unix.INLCR | unix.ICRNL
	attr.Iflag |= unix.ISTRIP | unix.INLCR
	attr.Lflag &^= unix.ECHO | unix.ICANON
	check(tcset(fd, &attr))

	cpu := KB11{
		switchregister: 0173030,
	}

	cpu.unibus.rk11.unibus = &cpu.unibus
	cpu.unibus.mmu = &cpu.mmu
	cpu.unibus.cons.Input = make(chan byte, 0)
	cpu.unibus.lineclock.ticks = time.Tick(999 * time.Millisecond)
	cpu.Reset()
	if err := cpu.unibus.rk11.Mount(0, r.RK0); err != nil {
		return err
	}
	cpu.Load(0002000, bootrom[:]...)
	cpu.R[7] = r.StartAddr
	go stdin(cpu.unibus.cons.Input)
	return cpu.Run()
}

func stdin(c chan uint8) {
	// for _, v := range "rpunix\n" {
	// 	c <- byte(v)
	// 	time.Sleep(200 * time.Millisecond)
	// }
	var b [1]byte
	for {
		n, _ := os.Stdin.Read(b[:])
		if n > 0 {
			c <- b[0]
		}
	}
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
