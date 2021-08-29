package roomPkg

import (
	. "gohipernetFake"

	"main/connectedSessions"
	"main/protocol"
)

func (room *baseRoom) _packetProcess_Invite(user *roomUser, packet protocol.Packet) int16 {
	sessionIndex := packet.UserSessionIndex
	sessionUniqueId := packet.UserSessionUniqueId

	var invitePacket protocol.RoomInviteReqPacket
	if invitePacket.Decoding(packet.Data) == false {
		_sendRoomInviteResult(sessionIndex, sessionUniqueId, protocol.ERROR_CODE_PACKET_DECODING_FAIL)
		return protocol.ERROR_CODE_PACKET_DECODING_FAIL
	}

	sendUserID, _ := connectedSessions.GetUserID(sessionIndex)
	roomNum, _ := connectedSessions.GetRoomNumber(sessionIndex)

	inviteUserIndex, inviteUserUniqueId := connectedSessions.GetNetWorkInfoByUserID(invitePacket.ID)
	if(inviteUserIndex == -1){
		_sendRoomInviteResult(sessionIndex, sessionUniqueId, protocol.ERROR_CODE_INVITE_ROOM_INVALID_USER_ID)
		return protocol.ERROR_CODE_INVITE_ROOM_INVALID_USER_ID
	}

	_sendRoomInvitation(inviteUserIndex, inviteUserUniqueId, roomNum, sendUserID)

	_sendRoomChatResult(sessionIndex, sessionUniqueId, protocol.ERROR_CODE_NONE)

	return protocol.ERROR_CODE_NONE
}

func _sendRoomInvitation(sessionIndex int32, sessionUniqueId uint64, roomNum int32, sendUserId []byte) {
	var invitation protocol.RoomInviteNtfPacket
	invitation.RoomNumber = roomNum
	invitation.IDLen = int8(len(sendUserId))
	invitation.ID = sendUserId[:]

	sendBuf, _ := invitation.EncodingPacket()
	NetLibIPostSendToClient(sessionIndex, sessionUniqueId, sendBuf)
}

func _sendRoomInviteResult(sessionIndex int32, sessionUniqueId uint64, result int16) {
	response := protocol.RoomInviteResPacket{result}
	sendPacket, _ := response.EncodingPacket()
	NetLibIPostSendToClient(sessionIndex, sessionUniqueId, sendPacket)
}