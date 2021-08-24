package roomPkg

import "main/protocol"

type RoomManager struct {
	_roomStartNum  int32
	_maxRoomCount  int32
	_roomCountList []int16
	_roomList      []baseRoom
}

func NewRoomManager(config RoomConfig) *RoomManager {
	roomManager := new(RoomManager)
	roomManager._initialize(config)
	return roomManager
}

func (roomMgr *RoomManager) _initialize(config RoomConfig) {
	roomMgr._roomStartNum = config.StartRoomNumber
	roomMgr._maxRoomCount = config.MaxRoomCount
	roomMgr._roomCountList = make([]int16, config.MaxRoomCount)
	roomMgr._roomList = make([]baseRoom, config.MaxRoomCount)

	for i := int32(0); i < roomMgr._maxRoomCount; i++ {
		roomMgr._roomList[i].initialize(i, config)
		roomMgr._roomList[i].settingPacketFunction()
	}
}

func (roomMgr *RoomManager) GetAllChannelUserCount() []int16 {
	maxRoomCount := roomMgr._maxRoomCount
	for i := int32(0); i < maxRoomCount; i++ {
		roomMgr._roomCountList[i] = (int16)(roomMgr._getRoomUserCount(i))
	}

	return roomMgr._roomCountList
}

func (roomMgr *RoomManager) getRoomByNumber(roomNumber int32) *baseRoom {
	roomIndex := roomNumber - roomMgr._roomStartNum

	if roomNumber < 0 || roomIndex >= roomMgr._maxRoomCount {
		return nil
	}

	return &roomMgr._roomList[roomIndex]
}

func (roomMgr *RoomManager) _getRoomUserCount(roomId int32) int32 {
	return roomMgr._roomList[roomId].getCurUserCount()
}

func (roomMgr *RoomManager) PacketProcess(roomNumber int32, packet protocol.Packet) {
	isRoomEnterReq := false

	if roomNumber == -1 && packet.Id == protocol.PACKET_ID_ROOM_ENTER_REQ {
		isRoomEnterReq = true

		var requestPacket protocol.RoomEnterReqPacket
		(&requestPacket).Decoding(packet.Data)

		roomNumber = requestPacket.RoomNumber
	}

	room := roomMgr.getRoomByNumber(roomNumber)
	if room == nil {
		protocol.NotifyErrorPacket(packet.UserSessionIndex, packet.UserSessionUniqueId,
			protocol.ERROR_CODE_ROOM_INVALIDE_NUMBER)
		return
	}

	user := room.getUser(packet.UserSessionUniqueId)
	if user == nil && isRoomEnterReq == false {
		protocol.NotifyErrorPacket(packet.UserSessionIndex, packet.UserSessionUniqueId,
			protocol.ERROR_CODE_ROOM_NOT_IN_USER)
		return
	}

	funcCount := len(room._funcPacketIdlist)
	for i := 0; i < funcCount; i++ {
		if room._funcPacketIdlist[i] != packet.Id {
			continue
		}

		room._funclist[i](user, packet)
		return
	}
}
