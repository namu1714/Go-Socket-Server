package main

import (
	"bytes"
	. "gohipernetFake"
	"time"

	"main/connectedSessions"
	"main/protocol"
)

func (server *ChatServer) DistributePacket(sessionIndex int32, sessionUniqueID uint64, packetData []byte) {
	packetID := protocol.PeekPacketID(packetData)
	bodySize, bodyData := protocol.PeekPacketBody(packetData)

	packet := protocol.Packet{Id: packetID}
	packet.UserSessionIndex = sessionIndex
	packet.UserSessionUniqueId = sessionUniqueID
	packet.Id = packetID
	packet.DataSize = bodySize
	packet.Data = make([]byte, packet.DataSize)
	copy(packet.Data, bodyData)

	server.PacketChan <- packet
}

func (server *ChatServer) PacketProcess_goroutine() {
	for {
		if server.PacketProcess_goroutine_Impl() {
			OutPutLog(LOG_LEVEL_INFO, "", 0, "Wanted Stop PacketProcess goroutine")
			break
		}
	}

	OutPutLog(LOG_LEVEL_INFO, "", 0,"Stop rooms PacketProcess goroutine")
}

func (server *ChatServer) PacketProcess_goroutine_Impl() bool {
	IsWantedTermination := false
	defer PrintPanicStack()

	for {
		packet := <-server.PacketChan
		sessionIndex := packet.UserSessionIndex
		sessionUniqueId := packet.UserSessionUniqueId
		bodySize := packet.DataSize
		bodyData := packet.Data

		if packet.Id == protocol.PACKET_ID_LOGIN_REQ {
			ProcessPacketLogin(sessionIndex, sessionUniqueId, bodySize, bodyData)
		} else if packet.Id == protocol.PACKET_ID_SESSION_CLOSE_SYS {
			ProcessPacketSessionClosed(server, sessionIndex, sessionUniqueId)
		} else {
			roomNumber, _ := connectedSessions.GetRoomNumber(sessionIndex)
			server.RoomMgr.PacketProcess(roomNumber, packet)
		}
	}

	return IsWantedTermination
}

func ProcessPacketLogin(sessionIndex int32, sessionUniqueId uint64, bodySize int16, bodyData []byte) {
	//중복 로그인만 아니면 성공
	var request protocol.LoginReqPacket
	if (&request).Decoding(bodyData) == false {
		_sendLoginResult(sessionIndex, sessionUniqueId, protocol.ERROR_CODE_PACKET_DECODING_FAIL)
		return
	}

	userID := bytes.Trim(request.UserID[:], "\x00")

	if len(userID) <= 0 {
		_sendLoginResult(sessionIndex, sessionUniqueId, protocol.ERROR_CODE_LOGIN_USER_INVALID_ID)
		return
	}

	curTime := time.Now().Unix()

	if connectedSessions.SetLogin(sessionIndex, sessionUniqueId, userID, curTime) == false {
		_sendLoginResult(sessionIndex, sessionUniqueId, protocol.ERROR_CODE_LOGIN_USER_DUPLICATION)
		return
	}

	_sendLoginResult(sessionIndex, sessionUniqueId, protocol.ERROR_CODE_NONE)
}

func _sendLoginResult(sessionIndex int32, sessionUniqueID uint64, result int16) {
	var response protocol.LoginResPacket
	response.Result = result
	sendPacket, _ := response.EncodingPacket()

	NetLibIPostSendToClient(sessionIndex, sessionUniqueID, sendPacket)
}

func ProcessPacketSessionClosed(server *ChatServer, sessionIndex int32, sessionUniqueID uint64) {
	roomNumber, _ := connectedSessions.GetRoomNumber(sessionIndex)

	if roomNumber > -1 {
		packet := protocol.Packet{
			sessionIndex,
			sessionUniqueID,
			protocol.PACKET_ID_ROOM_LEAVE_REQ,
			0,
			nil,
		}

		server.RoomMgr.PacketProcess(roomNumber, packet)
	}

	connectedSessions.RemoveSession(sessionIndex, true)
}
