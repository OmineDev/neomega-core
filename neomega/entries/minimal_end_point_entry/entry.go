package minimal_end_point_entry

import (
	"fmt"
	"os"
	"time"

	"github.com/OmineDev/neomega-core/i18n"
	"github.com/OmineDev/neomega-core/neomega"
	"github.com/OmineDev/neomega-core/neomega/bundle"
	"github.com/OmineDev/neomega-core/neomega/chunks/define"
	"github.com/OmineDev/neomega-core/nodes"
	"github.com/OmineDev/neomega-core/nodes/defines"
	"github.com/OmineDev/neomega-core/nodes/underlay_conn"
)

const ENTRY_NAME = "omega_minimal_end_point"

func Entry(args *Args) {
	var node defines.Node
	// ctx := context.Background()
	{
		client, err := underlay_conn.NewClientFromBasicNet(args.AccessPointAddr, time.Second)
		if err != nil {
			panic(err)
		}
		slave, err := nodes.NewSlaveNode(client)
		if err != nil {
			panic(err)
		}
		node = nodes.NewGroup("neomega", slave, false)
		node.ListenMessage("reboot", func(msg defines.Values) {
			reason, _ := msg.ToString()
			fmt.Println(reason)
			os.Exit(3)
		}, false)
		if !node.CheckNetTag("access-point") {
			panic(i18n.T(i18n.S_no_access_point_in_network))
		}
		for {
			if node.CheckNetTag("access-point-ready") {
				break
			}
			time.Sleep(time.Second)
		}
	}

	omegaCore, err := bundle.NewEndPointMicroOmega(node)
	if err != nil {
		panic(err)
	}
	_, _ = omegaCore.GetGameControl().SendWebSocketCmdNeedResponse("execute in overworld run tp @s 1024 200 1024").BlockGetResult()
	omegaCore.GetLowLevelAreaRequester().AttachSubChunkResultListener(func(scr neomega.SubChunkResult) {
		// fmt.Printf("%v %v\n", scr.SubCunkPos(), scr.Error())
	})
	ret, err := omegaCore.GetLowLevelAreaRequester().
		LowLevelRequestChunk(define.ChunkPos{1024 >> 4, 1024 >> 4}).
		AutoDimension().
		FullY().
		X(0).
		ZRange(0, 3).
		GetResult().
		SetTimeout(time.Second * 3).
		BlockGetResult()
	if err != nil {
		panic(err)
	}
	fmt.Println(ret.AllOk())
	fmt.Println(ret.AllErrors())
	chunks := ret.ToChunks(nil)
	fmt.Println(chunks)

	panic(<-node.WaitClosed())
}
