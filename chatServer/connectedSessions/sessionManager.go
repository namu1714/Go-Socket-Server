package connectedSessions

import (
	"sync"
	"sync/atomic"
	"time"
)

type Manager struct {
	_UserIDsessionMap *sync.Map

	_maxSessionCount int32
	_sessionList     []*session

	_maxUserCount int32

	_currentLoginUserCount int32
}

var _manager Manager

func Init(maxSessionCount int, maxUserCount int32) bool {
	_manager._UserIDsessionMap = new(sync.Map)
	_manager._maxUserCount = maxUserCount

	_manager._maxSessionCount = int32(maxSessionCount)
	_manager._sessionList = make([]*session, maxSessionCount)

	for i := 0; i < maxSessionCount; i++ {
		_manager._sessionList[i] = new(session)

		index := int32(i)
		_manager._sessionList[i].Init(index)
	}

	return true
}

func AddSession(sessionIndex int32, sessionUniqueID uint64) bool {
	if _validSessionIndex(sessionIndex) == false {
		return false
	}

	if _manager._sessionList[sessionIndex].GetConnectTimeSec() > 0 {
		return false
	}

	_manager._sessionList[sessionIndex].Clear()

	_manager._sessionList[sessionIndex].setConnectTimeSec(time.Now().Unix(), sessionUniqueID)

	return true
}

func RemoveSession(sessionIndex int32, isLoginedUser bool) bool {
	if _validSessionIndex(sessionIndex) == false {
		return false
	}

	if isLoginedUser {
		atomic.AddInt32(&_manager._currentLoginUserCount, -1)

		userID := string(_manager._sessionList[sessionIndex].getUserID())
		_manager._UserIDsessionMap.Delete(userID)
	}

	_manager._sessionList[sessionIndex].Clear()

	return true
}

func _validSessionIndex(index int32) bool {
	if index < 0 || index >= _manager._maxSessionCount {
		return false
	}
	return true
}

func GetNetworkUniqueID(sessionIndex int32) uint64 {
	if _validSessionIndex(sessionIndex) == false {
		return 0
	}

	return _manager._sessionList[sessionIndex].GetNetworkUniqueID()
}

func GetUserID(sessionIndex int32) ([]byte, bool) {
	if _validSessionIndex(sessionIndex) == false {
		return nil, false
	}
	return _manager._sessionList[sessionIndex].getUserID(), true
}

func SetLogin(sessionIndex int32, sessionUniqueID uint64, userId []byte, curTimeSec int64) bool {
	if _validSessionIndex(sessionIndex) == false {
		return false
	}

	newUserID := string(userId)
	if _, ok := _manager._UserIDsessionMap.Load(newUserID); ok {
		return false
	}

	_manager._sessionList[sessionIndex].SetUser(sessionUniqueID, userId, curTimeSec)
	_manager._UserIDsessionMap.Store(newUserID, _manager._sessionList[sessionIndex])

	atomic.AddInt32(&_manager._currentLoginUserCount, 1)
	return true
}

func IsLoginUser(sessionIndex int32) bool {
	if _validSessionIndex(sessionIndex) == false {
		return false
	}

	return _manager._sessionList[sessionIndex].IsAuth()
}

func SetRoomNumber(sessionIndex int32, sessionUniqueId uint64, roomNum int32, curTimeSec int64) bool {
	if _validSessionIndex(sessionIndex) == false {
		return false
	}

	return _manager._sessionList[sessionIndex].setRoomNumber(sessionUniqueId, roomNum, curTimeSec)
}

func GetRoomNumber(sessionIndex int32) (int32, int32) {
	if _validSessionIndex(sessionIndex) == false {
		return -1, -1
	}
	return _manager._sessionList[sessionIndex].getRoomNumber()
}

func GetNetWorkInfoByUserID(userID []byte) (int32, uint64) {
	sessionObj, ok := _manager._UserIDsessionMap.Load(userID);
	userSession := sessionObj.(*session)

	if ok == false {
		return -1, 0
	}

	return userSession.GetNetworkInfo()
}