package manage

import (
	"sync"
	"sync/atomic"
	"Taonet/taonetBase"
)

const totalTryCount = 1000


type SessionManager struct {
	sessionList  map[int64]taonetBase.ITaoSession
	sessionIDAcc int64
	syncTex sync.RWMutex
}

func (this *SessionManager) Count() int {
	this.syncTex.Lock()
	defer this.syncTex.Unlock()

	return len(this.sessionList)
}

func (this *SessionManager) Add(session taonetBase.ITaoSession) {

	this.syncTex.Lock()
	defer this.syncTex.Unlock()

	var tryCount int = totalTryCount

	var id int64

	for tryCount > 0 {
		id = atomic.AddInt64(&this.sessionIDAcc, 1)
		if _, ok := this.sessionList[id]; !ok {
			break
		}
		tryCount--
	}

	session.SetSessionId(id)

	this.sessionList[id] = session
}

func (this *SessionManager) Remove(session taonetBase.ITaoSession) {
	this.syncTex.Lock()
	defer this.syncTex.Unlock()

	delete(this.sessionList, session.ID())
}

func (this *SessionManager) GetSessionById(id int64) taonetBase.ITaoSession {

	this.syncTex.Lock()
	defer this.syncTex.Unlock()

	v, ok := this.sessionList[id]

	if ok {
		return v
	}

	return nil
}

func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessionList: make(map[int64]taonetBase.ITaoSession),
	}
}