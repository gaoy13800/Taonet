package taonetBase

import (
	"Taonet/task"
)

type ITaoSession interface {

	ID() int64

	DeviceId() string

	Work(task task.ITask)

	Encode(data string)

	Close()

	SetSessionId(id int64)

	SetDeviceId(id string)

	CheckConnect() bool
}


type IDevices interface {
	Add(string, ITaoSession)

	Remove(string)

	GetSessionByDeviceId(string) (ITaoSession, error)

	IsExists(string) bool

	GetDeviceIds() []string
}

type ITaoSessionManage interface {
	Count() int
	Add(session ITaoSession)
	Remove(session ITaoSession)
	GetSessionById(id int64) ITaoSession
}