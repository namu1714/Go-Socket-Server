package main

import (
	"strconv"
	"strings"

	. "gohipernetFake"

	"main/connectedSessions"
	"main/protocol"
	"main/roomPkg"
)

type configAppServer struct {
	GameName string

	RoomMaxCount     int32
	RoomStartNum     int32
	RoomMaxUserCount int32
}

type ChatServer struct {
	ServerIndex int
	IP          string
	Port        int

	PacketChan chan protocol.Packet

	RoomMgr *roomPkg.RoomManager
}

func createAnsStartServer(netConfig NetworkConfig, appConfig configAppServer) {
	OutPutLog(LOG_LEVEL_INFO, "", 0, "CreateServer !!!")

	var server ChatServer

	if server.setIPAddress(netConfig.BindAddress) == false {
		OutPutLog(LOG_LEVEL_INFO, "", 0, "fail. server address")
		return
	}

	protocol.Init_packet()

	maxUserCount := appConfig.RoomMaxCount * appConfig.RoomMaxUserCount
	connectedSessions.Init(netConfig.MaxSessionCount, maxUserCount)

	server.PacketChan = make(chan protocol.Packet, 256)

	roomConfig := roomPkg.RoomConfig{
		appConfig.RoomStartNum,
		appConfig.RoomMaxCount,
		appConfig.RoomMaxUserCount,
	}
	server.RoomMgr = roomPkg.NewRoomManager(roomConfig)

	go server.PacketProcess_goroutine()

	networkFuncter := SessionNetworkFunctors{}
	networkFuncter.OnConnect = server.OnConnect
	networkFuncter.OnReceive = server.OnReceive
	networkFuncter.OnReceiveBufferedData = nil
	networkFuncter.OnClose = server.OnClose
	networkFuncter.PacketTotalSizeFunc = PacketTotalSize
	networkFuncter.PacketHeaderSize = PACKET_HEADER_SIZE
	networkFuncter.IsClientSession = true

	NetLibStartNetwork(&netConfig, networkFuncter)
}

func (server *ChatServer) setIPAddress(ipAddress string) bool {
	results := strings.Split(ipAddress, ":")
	if len(results) != 2 {
		return false
	}

	server.IP = results[0]
	server.Port, _ = strconv.Atoi(results[1])

	return true
}

func (server *ChatServer) OnConnect(sessionIndex int32, sessionUniqueID uint64) {
	connectedSessions.AddSession(sessionIndex, sessionUniqueID)
}

func (server *ChatServer) OnReceive(sessionIndex int32, sessionUniqueID uint64, data []byte) bool {
	server.DistributePacket(sessionIndex, sessionUniqueID, data)
	return true
}

func (server *ChatServer) OnClose(sessionIndex int32, sessionUniqueID uint64) {
	server.disConnectClient(sessionIndex, sessionUniqueID)
}

func (server *ChatServer) disConnectClient(sessionIndex int32, sessionUniqueId uint64) {
	//로그인 안한 유저 처리
	if connectedSessions.IsLoginUser(sessionIndex) == false {
		connectedSessions.RemoveSession(sessionIndex, false)
		return
	}

	packet := protocol.Packet{
		sessionIndex,
		sessionUniqueId,
		protocol.PACKET_ID_SESSION_CLOSE_SYS,
		0,
		nil,
	}

	server.PacketChan <- packet
}
