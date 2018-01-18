package session

import (
	"net"
	"sync"
	"Taonet/task"
	"Taonet/push"
	"Taonet/common"
	"log"
)

type SocketSession struct {
	sessionId 	int64

	deviceId 	string

	conn 		net.Conn

	syncWait 	sync.WaitGroup

	WithClose 	func()

	tickerContainer int
}

func (this *SocketSession) ID() int64 {
	return this.sessionId
}

func (this *SocketSession) SetSessionId(id int64){
	this.sessionId = id
}

func (this *SocketSession) DeviceId() string {
	return this.deviceId
}

func (this *SocketSession) SetDeviceId(id string){
	this.deviceId = id
}

func (this *SocketSession) Work(task task.ITask) {

	this.syncWait.Add(1)

	go func() {

		this.syncWait.Wait()

		this.Close()
	}()

	go this.pipeReceive(task)
}

func (this *SocketSession) pipeReceive(task task.ITask) {

	//go this.pingSocket(queue)

	for {

		data, err := this.decode()

		if err != nil {
			goto CLOSELABLE
		}

		task.PushMessage(push.BuildRuleData(common.Notice, data, this))

		continue

	CLOSELABLE:
		task.PushMessage(push.BuildRuleData(common.Close, data, this))
		break
	}

	this.syncWait.Done()
}

func (this *SocketSession) decode() (string, error) {
	byt := make([]byte, 1024)
	index, err := this.conn.Read(byt)

	if err != nil {
		return "", err
	}

	return string(byt[0:index]), err
}


func (this *SocketSession) CheckConnect() bool {

	if this.sessionId == 0 || this.deviceId == ""{
		return false
	}

	return true
}


func (this *SocketSession) Encode(data string){

	if data == "" {
		return
	}

	_, err := this.conn.Write([]byte(data))

	if err != nil {
		log.Print("消息发送失败！", err)
	}else {
		log.Print("消息发送成功！")
	}
}

func (this *SocketSession) Close(){
	this.conn.Close()

	this.WithClose()
}

func NewSocketSession(conn net.Conn) *SocketSession{
	this := &SocketSession{
		conn:conn,
	}

	return this
}
