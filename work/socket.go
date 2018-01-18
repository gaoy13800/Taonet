package work

import (
	"net"
	"sync"
	"log"
	"time"
	"Taonet/session"
	"Taonet/manage"
	"Taonet/task"
	"Taonet/push"
	"Taonet/common"
	"fmt"
	"Taonet/conf"
)


type ITaoSocket interface {

	Start() ITaoSocket

	Close()

}

type TaoSocket struct {

	listener 	net.Listener

	syncTex	 	sync.RWMutex

	running		bool

	tasks		task.ITask

	//deviceMng 	*manage.DeviceManager

	*manage.SessionManager
}

func (this *TaoSocket) Start() ITaoSocket{


	address := fmt.Sprintf("%s:%s", "0.0.0.0", conf.TaoConf["socket_tcp_port"])

	defer func() { // 必须要先声明defer，否则不能捕获到panic异常
		if err := recover(); err != nil {
			log.Println("recover error", err)
		}
	}()

	tcpAddr, err := net.ResolveTCPAddr("tcp", address)

	if err != nil {
		panic(err)
	}

	listener, err := net.ListenTCP("tcp", tcpAddr)

	if err != nil {
		panic(err)
	}

	this.listener = listener

	go this.acceptor()

	return this
}

func (this *TaoSocket) Close(){
	this.listener.Close()
}

func (this *TaoSocket) acceptor() {

	this.setRunning(true)

	for {

		if !this.isRunning() {
			break
		}

		conn, err := this.listener.Accept()

		if err != nil {
			panic(err)
		}

		go this.handler(conn)
	}

	this.setRunning(false)
}

func (this *TaoSocket) handler(conn net.Conn) {

	sess := session.NewSocketSession(conn)

	this.Add(sess)

	sess.WithClose = func() {
		this.Remove(sess)
	}

	sess.Work(this.tasks)

	this.tasks.PushMessage(push.BuildRuleData(common.Connect, "connect", sess))
}

func (this *TaoSocket) setRunning(isRunning bool) {
	this.syncTex.Lock()
	defer this.syncTex.Unlock()

	this.running = isRunning
}

func (this *TaoSocket) isRunning() bool {
	return this.running
}

func NewTaoSocket(task task.ITask) *TaoSocket {
	this := &TaoSocket{
		tasks:task,
		SessionManager: manage.NewSessionManager(),
	}

	go func() {
		tick := time.Tick(time.Second * 15)

		for {
			select {
			case <-tick:
				log.Println("current session number:", this.SessionManager.Count())
			}
		}
	}()

	return this
}