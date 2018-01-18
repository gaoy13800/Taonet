package manage

import (
	"sync"
	"errors"
	"Taonet/taonetBase"
)

type DeviceManager struct {
	deviceList map[string]taonetBase.ITaoSession

	syncTex sync.RWMutex
}

func (this *DeviceManager) Add(deviceId string, sess taonetBase.ITaoSession) {

	this.syncTex.Lock()
	defer this.syncTex.Unlock()

	this.deviceList[deviceId] = sess
}

func (this *DeviceManager) Remove(deviceId string) {
	this.syncTex.Lock()
	defer this.syncTex.Unlock()

	delete(this.deviceList, deviceId)
}

func (this *DeviceManager) GetSessionByDeviceId(deviceId string) (taonetBase.ITaoSession, error) {
	if v, ok := this.deviceList[deviceId]; ok {
		return v, nil
	}

	return nil, errors.New("not exits")
}

func (this *DeviceManager) IsExists(deviceId string) bool {

	if _, ok := this.deviceList[deviceId]; !ok {
		return false
	}

	return true
}

func (this *DeviceManager) GetDeviceIds()[]string{
	list := make([]string, 0, 1000)

	for key, _ := range this.deviceList{
		list = append(list, key)
	}

	return list
}

func NewDeviceManage() *DeviceManager {

	return &DeviceManager{
		deviceList: make(map[string]taonetBase.ITaoSession),
	}

}