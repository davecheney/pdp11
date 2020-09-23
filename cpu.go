package main

type KB11 struct {
	unibus UNIBUS

	PC uint16
}

func (kb *KB11) Reset() {

}

func (kb *KB11) Run() error {
	return nil
}
