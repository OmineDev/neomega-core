package main

import (
	"fmt"

	"github.com/OmineDev/neomega-core/nodes/defines"
	"github.com/OmineDev/neomega-core/nodes/underlay_conn"
)

func Server() {
	fmt.Println("server start")
	server, err := underlay_conn.NewServerFromBasicNet("tcp://0.0.0.0:7333")
	if err != nil {
		panic(err)
	}
	server.ExposeAPI("echo").InstantAPI(func(callerAndArgs defines.ArgWithCaller) (defines.Values, error) {
		fmt.Println("server recv:", callerAndArgs.Caller, callerAndArgs.Args.ToStrings())
		go func() {
			ret, err := server.CallWithResponse(callerAndArgs.Caller, "client-echo", defines.FromStrings("server", "hi")).BlockGetResult()
			fmt.Println("server get response ", ret.ToStrings(), err)
			server.CallOmitResponse(callerAndArgs.Caller, "client-echo", defines.FromStrings("server", "hi", "no resp"))
		}()
		return defines.FromString("server echo").Extend(callerAndArgs.Args), nil
	})
	<-server.WaitClosed()
	fmt.Println("server closed")
}

func main() {
	Server()
}
