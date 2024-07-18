package main

import (
	"fmt"
	"time"

	"github.com/OmineDev/neomega-core/nodes/defines"
	"github.com/OmineDev/neomega-core/nodes/underlay_conn"
)

func Server() {
	fmt.Println("server start")
	server, err := underlay_conn.NewServerFromBasicNet("tcp://0.0.0.0:7241")
	if err != nil {
		panic(err)
	}
	server.ExposeAPI("echo").InstantAPI(func(argsAndCaller defines.ArgWithCaller) (defines.Values, error) {
		fmt.Println("server recv:", argsAndCaller.Caller, argsAndCaller.Args.ToStrings())
		go func() {
			ret, err := server.CallWithResponse(argsAndCaller.Caller, "client-echo", defines.FromStrings("server", "hi")).BlockGetResult()
			fmt.Println("server get response ", ret.ToStrings(), err)
			server.CallOmitResponse(argsAndCaller.Caller, "client-echo", defines.FromStrings("server", "hi", "no resp"))
		}()
		return defines.FromString("server echo").Extend(argsAndCaller.Args), nil
	})
	<-server.WaitClosed()
	fmt.Println("server closed")
}

func Client(id string) {
	fmt.Println("client start")
	client, err := underlay_conn.NewClientFromBasicNet("tcp://127.0.0.1:7333", time.Second)
	if err != nil {
		panic(err)
	}
	go func() {
		ret, err := client.CallWithResponse("echo", defines.FromStrings("hello", "world")).BlockGetResult()
		fmt.Println("client get response ", ret.ToStrings(), err)
		client.CallOmitResponse("echo", defines.FromStrings("hello", "world", "no resp"))
	}()
	client.ExposeAPI("client-echo").InstantAPI(func(args defines.Values) (defines.Values, error) {
		fmt.Println(fmt.Sprintf("client %v recv:", id), args.ToStrings())
		return defines.FromString(fmt.Sprintf("client %v echo", id)).Extend(args), nil
	})
	<-client.WaitClosed()
	fmt.Printf("client %v closed\n", id)
}

func main() {
	go Client("1")
	go Client("2")
	c := make(chan struct{})
	<-c
}
