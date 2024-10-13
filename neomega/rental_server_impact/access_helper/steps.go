package access_helper

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/OmineDev/neomega-core/i18n"
	"github.com/OmineDev/neomega-core/minecraft/protocol/packet"
	"github.com/OmineDev/neomega-core/minecraft_neo/cascade_conn/base_net"
	"github.com/OmineDev/neomega-core/minecraft_neo/cascade_conn/byte_frame_conn"
	"github.com/OmineDev/neomega-core/minecraft_neo/cascade_conn/packet_conn"
	"github.com/OmineDev/neomega-core/minecraft_neo/login_and_spawn_core"
	"github.com/OmineDev/neomega-core/minecraft_neo/login_and_spawn_core/options"
	"github.com/OmineDev/neomega-core/neomega"
	"github.com/OmineDev/neomega-core/neomega/fbauth"
	"github.com/OmineDev/neomega-core/neomega/minecraft_conn"
	"github.com/OmineDev/neomega-core/neomega/rental_server_impact/challenges"
	"github.com/pterm/pterm"
)

// Copied from phoenixbuilder/core/core
// func initializeMinecraftConnection(ctx context.Context, authenticator minecraft.Authenticator) (conn *minecraft.Conn, err error) {
// 	dialer := minecraft.Dialer{
// 		Authenticator: authenticator,
// 	}
// 	conn, err = dialer.DialContext(ctx, "raknet")
// 	if err != nil {
// 		return
// 	}

//		return
//	}

func loginAuthServer(ctx context.Context, authenticator Authenticator) (privateKey *ecdsa.PrivateKey, authResp map[string]any, err error) {
	fmt.Println(i18n.T(i18n.S_generating_client_key_pair))
	privateKey, _ = ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	publicKey, _ := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	fmt.Println(i18n.T(i18n.S_retrieving_client_information_from_auth_server))
	authResp, err = authenticator.GetAccess(ctx, publicKey)
	if err != nil {
		return nil, nil, err
	}
	serverMessage, _ := authResp["server_msg"].(string)
	if len(serverMessage) > 0 {
		fmt.Println(i18n.T(i18n.S_message_from_auth_server))
		fmt.Println(pterm.LightGreen(strings.ReplaceAll(fmt.Sprintf("    %s", serverMessage), "\n", "\n    ")))
	}
	return privateKey, authResp, nil
}

func loginMCServer(ctx context.Context, privateKey *ecdsa.PrivateKey, authResp map[string]any) (conn minecraft_conn.Conn, err error) {
	address, _ := authResp["ip_address"].(string)

	fmt.Println(i18n.T(i18n.S_establishing_raknet_connection))
	rakNetConn, err := base_net.RakNet.DialContext(ctx, address)
	if err != nil {
		return nil, err
	}

	fmt.Println(i18n.T(i18n.S_establishing_byte_frame_connection))
	byteFrameConn := byte_frame_conn.NewConnectionFromNet(rakNetConn)

	fmt.Println(i18n.T(i18n.S_generating_packet_connection))
	packetConn := packet_conn.NewPacketConn(byteFrameConn, false)

	fmt.Println(i18n.T(i18n.S_generating_key_login_request))
	opt := options.NewDefaultOptions(address, authResp, privateKey)

	readQueue := NewInfinityQueue()
	loginAndSpawnCore := login_and_spawn_core.NewLoginAndSpawnCore(packetConn, opt)
	go packetConn.ListenRoutine(func(pk packet.Packet, raw []byte) {
		// fmt.Println("read:", pk.ID())
		loginAndSpawnCore.Receive(pk)
		readQueue.PutPacket(pk, raw)
	})
	err = loginAndSpawnCore.Login(ctx)
	if err != nil {
		return nil, err
	}
	fmt.Println(i18n.T(i18n.S_login_accomplished))

	fmt.Println(i18n.T(i18n.S_sending_additional_information))
	packetConn.WritePacket(&packet.ClientCacheStatus{
		Enabled: false,
	})
	packetConn.WritePacket(&packet.NeteaseJson{
		Data: []byte(fmt.Sprintf(`{"eventName":"LOGIN_UID","resid":"","uid":"%s"}`, strconv.FormatInt(opt.IdentityData.Uid, 10))),
	})
	// conn.WritePacket(&packet.PyRpc{
	// 	Value: py_rpc.FromGo([]any{
	// 		"e",
	// 		[]any{},
	// 		nil,
	// 	}),
	// })
	outfitInfo, _ := authResp["outfit_info"].(map[string]any)
	usingModList := []string{}
	for uuid, level := range outfitInfo {
		usingModList = append(usingModList, uuid)
		if level == nil {
			delete(outfitInfo, uuid)
		}
	}
	packetConn.WritePacket(&packet.PyRpc{
		Value: []any{
			"SyncUsingMod",
			[]any{
				usingModList,
				opt.ClientData.SkinID,
				opt.ClientData.SkinItemID,
				true,
				outfitInfo,
			},
			nil,
		},
		OperationType: packet.PyRpcOperationTypeSend,
	})

	// Only this packet is necessary
	packetConn.WritePacket(&packet.PyRpc{
		Value: []any{
			"ClientLoadAddonsFinishedFromGac",
			[]any{},
			nil,
		},
		OperationType: packet.PyRpcOperationTypeSend,
	})

	// Generally, following packets are sent after "SetStartType"
	packetConn.WritePacket(&packet.PyRpc{
		Value: []any{
			"arenaGamePlayerFinishLoad",
			[]any{},
			nil,
		},
		OperationType: packet.PyRpcOperationTypeSend,
	})
	packetConn.WritePacket(&packet.PyRpc{
		Value: []any{
			"ModEventC2S",
			[]any{
				"Minecraft",
				"vipEventSystem",
				"PlayerUiInit",
				fmt.Sprintf("%d", loginAndSpawnCore.GameData().EntityUniqueID),
			},
			nil,
		},
		OperationType: packet.PyRpcOperationTypeSend,
	})
	packetConn.WritePacket(&packet.PyRpc{
		Value: []any{
			"ClientInitUIFinishedEventFromGac",
			[]any{},
			nil,
		},
		OperationType: packet.PyRpcOperationTypeSend,
	})
	packetConn.Flush()
	fmt.Println(i18n.T(i18n.S_packing_core))
	return &shallowWrap{
		byteFrameConn:  byteFrameConn,
		PacketConnBase: packetConn,
		Core:           loginAndSpawnCore,
		InfinityQueue:  readQueue,
		identityData:   loginAndSpawnCore.IdentityData,
	}, nil
}

func loginMCServerWithRetry(ctx context.Context, authenticator Authenticator, retryTimesRemains int) (conn minecraft_conn.Conn, err error) {
	privateKey, authResp, err := loginAuthServer(ctx, authenticator)
	if err != nil {
		return nil, err
	}
	// chain info will be vaild in a short time, so it can be used to re-login
	retryTimes := 0
	for {
		conn, err = loginMCServer(ctx, privateKey, authResp)
		if err == nil {
			break
		} else {
			fmt.Println(err)
		}
		if retryTimesRemains <= 0 {
			break
		}
		retryTimes++
		fmt.Printf(i18n.T(i18n.S_fail_connecting_to_mc_server_retrying), retryTimes)
		// wait for 1s
		time.Sleep(time.Second)
		retryTimesRemains--
	}
	if err != nil {
		return nil, err
	}
	fmt.Println(i18n.T(i18n.S_done_connecting_to_mc_server))
	return conn, nil
}

func makeAuthenticatorAndChallengeSolver(options *ImpactOption, writeBackFBToken bool) (authenticator Authenticator, challengeSolver challenges.CanSolveChallenge, err error) {
	clientOptions := fbauth.MakeDefaultClientOptions()
	clientOptions.AuthServer = options.AuthServer
	fmt.Printf(i18n.T(i18n.S_connecting_to_auth_server)+": %v\n", options.AuthServer)
	fbClient, err := fbauth.CreateClient(clientOptions)
	if err != nil {
		return nil, nil, err
	}
	challengeSolver = fbClient
	fmt.Println(i18n.T(i18n.S_done_connecting_to_auth_server))
	hashedPassword := ""
	if options.UserToken == "" {
		psw_sum := sha256.Sum256([]byte(options.UserPassword))
		hashedPassword = hex.EncodeToString(psw_sum[:])
	}
	authenticator = fbauth.NewAccessWrapper(fbClient, options.ServerCode, options.ServerPassword, options.UserToken, options.UserName, hashedPassword, writeBackFBToken)
	return authenticator, challengeSolver, nil
}

func copeWithRentalServerChallenge(ctx context.Context, omegaCore neomega.MicroOmega, canSolveChallenge challenges.CanSolveChallenge) (err error) {
	fmt.Println(i18n.T(i18n.S_coping_with_rental_server_challenges))
	challengeSolver := challenges.NewPyRPCResponder(omegaCore, canSolveChallenge)

	err = challengeSolver.ChallengeCompete(ctx)
	if err != nil {
		return ErrFBChallengeSolvingTimeout
	}
	fmt.Println(i18n.T(i18n.S_done_coping_with_rental_server_challenges))
	return nil
}

func reasonWithPrivilegeStuff(ctx context.Context, omegaCore neomega.MicroOmega, options *PrivilegeStuffOptions) (err error) {
	fmt.Println(i18n.T(i18n.S_checking_bot_op_permission_and_game_cheat_mode))
	helper := challenges.NewOperatorChallenge(omegaCore, func() {
		if options.OpPrivilegeRemovedCallBack != nil {
			options.OpPrivilegeRemovedCallBack()
		}
		if options.DieOnLosingOpPrivilege {
			omegaCore.CloseWithError(ErrBotOpPrivilegeRemoved)
		}
	})
	waitErr := make(chan error)
	go func() {
		waitErr <- helper.WaitForPrivilege(ctx)
	}()
	select {
	case err = <-waitErr:
	case err = <-omegaCore.WaitClosed():
	}
	if err != nil {
		return err
	}
	fmt.Println(i18n.T(i18n.S_done_checking_bot_op_permission_and_game_cheat_mode))
	return nil
}

func makeBotCreative(omegaCoreCtrl neomega.GameCtrl) {
	waitor := make(chan struct{})
	fmt.Println(i18n.T(i18n.S_switching_bot_to_creative_mode))
	omegaCoreCtrl.SendWebSocketCmdNeedResponse("gamemode c @s").SetTimeout(time.Second * 3).AsyncGetResult(func(output *packet.CommandOutput, err error) {
		if err == nil && output != nil {
			fmt.Println(i18n.T(i18n.S_done_setting_bot_to_creative_mode))
			close(waitor)
		} else {
			panic("failed to set bot to creative mode")
		}
	})
	<-waitor
}

func disableCommandBlock(omegaCoreCtrl neomega.GameCtrl) {
	omegaCoreCtrl.SendWOCmd("gamerule commandblocksenabled false")
	//	waitor := make(chan struct{})
	//	omegaCoreCtrl.SendPlayerCmdNeedResponse("gamerule commandblocksenabled false").AsyncGetResult(func(output *packet.CommandOutput, err error) {
	fmt.Println(i18n.T(i18n.S_done_setting_commandblocksenabled_false))
	//		close(waitor)
	//	})
	//
	// <-waitor
}

func waitDead(omegaCore neomega.MicroOmega) {
	// SetTime packet will be sent by server every 256 ticks, even dodaylightcycle gamerule disabled
	threshold := time.Minute
	startTime := time.Now()
	lastReceivePacket := time.Now()
	omegaCore.GetGameListener().SetAnyPacketCallBack(func(p packet.Packet) {
		lastReceivePacket = time.Now()
	}, false)
	for {
		time.Sleep(time.Second)
		nowTime := time.Now()
		if lastReceivePacket.Add(time.Second * 5).Before(nowTime) {
			flyTime := nowTime.Sub(lastReceivePacket)
			deadTime := threshold - flyTime
			fmt.Printf(i18n.T(i18n.S_bot_no_resp_could_been_feeding_massive_data_reboot_count_down)+"\n", float32(deadTime)/float32(time.Second))
			omegaCore.GetGameControl().SendWebSocketCmdOmitResponse("errorcmd")
		}
		if lastReceivePacket.Add(threshold).Before(nowTime) {
			omegaCore.CloseWithError(fmt.Errorf(i18n.T(i18n.S_no_response_after_a_long_time_bot_is_down), threshold, time.Since(startTime).Seconds()))
			break
		}
	}
}
