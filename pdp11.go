// pdp11 emulator.
package main

import "github.com/alecthomas/kong"

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
	cpu := KB11{}
	cpu.unibus.rk11.unibus = &cpu.unibus
	cpu.Reset()
	if err := cpu.unibus.rk11.Mount(0, r.RK0); err != nil {
		return err
	}
	cpu.Load(0002000, bootrom[:]...)
	cpu.R[7] = r.StartAddr
	return cpu.Run()
}
