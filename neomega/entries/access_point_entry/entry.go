package access_point

import (
	"context"
	"crypto/md5"
	"fmt"

	"github.com/OmineDev/neomega-core/i18n"
	"github.com/OmineDev/neomega-core/neomega"
	"github.com/OmineDev/neomega-core/neomega/rental_server_impact/access_helper"
	"github.com/OmineDev/neomega-core/neomega/rental_server_impact/info_collect_utils"
	"github.com/OmineDev/neomega-core/nodes"
	"github.com/OmineDev/neomega-core/nodes/defines"
	"github.com/OmineDev/neomega-core/nodes/underlay_conn"
)

const ENTRY_NAME = "omega_access_point"

func Entry(args *Args) {
	fmt.Println(i18n.T(i18n.S_neomega_access_point_starting))
	impactOption := args.ImpactOption
	var err error

	if err := info_collect_utils.ReadUserInfoAndUpdateImpactOptions(impactOption); err != nil {
		panic(err)
	}

	accessOption := access_helper.DefaultOptions()
	accessOption.ImpactOption = args.ImpactOption
	accessOption.MakeBotCreative = true
	accessOption.DisableCommandBlock = false
	accessOption.ReasonWithPrivilegeStuff = true

	var omegaCore neomega.MicroOmega
	var node defines.Node
	ctx := context.Background()
	{
		server, err := underlay_conn.NewServerFromBasicNet(args.AccessArgs.AccessPointAddr)
		if err != nil {
			panic(err)
		}
		master := nodes.NewMasterNode(server)
		node = nodes.NewGroup("neomega", master, false)
	}
	omegaCore, err = access_helper.ImpactServer(ctx, node, accessOption)
	if err != nil {
		panic(err)
	}
	huid := defines.Empty
	if args.UserToken != "" {
		huid = defines.FromString(StrMD5Str(args.UserToken))
	}
	node.SetValue("HashedUserID", huid)
	hserverCode := defines.Empty
	if args.ServerCode != "" {
		hserverCode = defines.FromString(StrMD5Str(args.ServerCode))
	}
	node.SetValue("HashedServerCode", hserverCode)
	node.SetTags("access-point-ready")
	node.PublishMessage("reboot", defines.FromString("reboot to refresh data"))
	fmt.Println(i18n.T(i18n.S_neomega_access_point_ready))

	panic(<-omegaCore.Dead())
}

func StrMD5Str(data string) string {
	return BytesMD5Str([]byte(data))
}

func BytesMD5Str(data []byte) string {
	return fmt.Sprintf("%x", md5.Sum(data))
}
