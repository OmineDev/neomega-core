package main

import access_point "github.com/OmineDev/neomega-core/neomega/entries/access_point_entry"

func main() {
	args := access_point.GetArgs()
	omegaCore, _ := access_point.Entry(args)
	panic(<-omegaCore.WaitClosed())
}
