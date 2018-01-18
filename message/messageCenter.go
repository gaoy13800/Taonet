package message

import (
	"Taonet/manage"
	"Taonet/push"
	"Taonet/common"
	"Taonet/taonetDB"
	"Taonet/memory"
	"sync"
	"log"
	"time"
	"strings"
	"fmt"
	"Taonet/taonetBase"
	"github.com/fatih/color"
)

const (
	Lock_Open_Cstu = "3"
	Lock_Close_Cstu = "4"
	Lock_Contact_Cstu = "2"
)

type MessageCenter struct {

	IDevices taonetBase.IDevices

	iRedis   taonetDB.ITaoRedis

	syncTex  sync.RWMutex

	taoMemory *memory.Cache
}


func (msg *MessageCenter) Notice(pushData *push.RuleData){

	switch pushData.DataType {
	case common.Connect:
		log.Printf("新会话建立  session id is %d", pushData.DataSource.ID())
		return
	case common.Notice:
		msg.disposeMessage(pushData)
		return
	case common.Close:
		log.Printf(" 会话结束  session id is %d", pushData.DataSource.ID())

		msg.iRedis.SaveLockData(common.Lock_Cstu, pushData.DataSource.DeviceId(), "2")
		return
	default:
		log.Println("未知的命令！ 不会处理")
		return
	}
}

func (this *MessageCenter) disposeMessage(pushData *push.RuleData){

	cmsg := pushData.DataDetail

	color.Green("收到信息 %s", cmsg)

	color.Green("deviceId:%s", pushData.DataSource.DeviceId())

	if !strings.HasPrefix(cmsg, "wt") {
		log.Println("未知消息信息, 会话将在5s后关闭")

		time.Sleep(time.Second * 5)

		pushData.DataSource.Close()

		return
	}

	if len(cmsg) != 21 && len(cmsg) != 6{

		log.Println("未知消息长度, 会话将在5s后关闭")

		time.Sleep(time.Second * 5)

		pushData.DataSource.Close()

		return
	}

	var action, deviceId string

	if len(cmsg) == 21 {
		action, deviceId = resolve(cmsg)
	}else if len(cmsg) == 6 {
		action = cmsg

		deviceId = pushData.DataSource.DeviceId()
	}

	//todo 接受电量和状态的指令

	switch action {
	case "wtoveo":
		this.over_deal(deviceId, pushData.DataSource, Lock_Open_Cstu)
		break
	case "wtgoid":
		this.build_connect(deviceId, pushData.DataSource)
		break
	case "wtovec":
		this.over_deal(deviceId, pushData.DataSource, Lock_Close_Cstu)
		break
	case "wtoveb":
		this.over_brut(deviceId, "1")

		time.Sleep(time.Second * 20)

		this.over_brut(deviceId, "0")
		break
	default:
		break
	}
}

// 地锁终端回应over所做处理
func (this *MessageCenter) over_deal(deviceId string, sess taonetBase.ITaoSession, status string){

	if ok := sess.CheckConnect(); !ok{
		log.Println("请检查是否建立真实连接")
		return
	}

	if len(deviceId) != 15 {
		log.Println("deviceId 长度无效！ ", deviceId)
		return
	}

	if status == Lock_Open_Cstu {
		this.taoMemory.Delete(deviceId + "_action_" + "open")
	}else {
		this.taoMemory.Delete(deviceId + "_action_" + "clse")
	}

	this.iRedis.SaveLockData(common.Lock_Cstu, deviceId, status)
}


func (this *MessageCenter) over_brut(deviceId string, status string){

	this.taoMemory.Delete(deviceId + "_action_" + "brut")

	this.iRedis.SaveLockData(common.Lock_Brut, deviceId, status)
}

//终端建立连接 存储设备id、新增sessionId
func (this *MessageCenter) build_connect(deviceId string, sess taonetBase.ITaoSession){

	this.syncTex.Lock()

	defer this.syncTex.Unlock()

	//干掉 之前的session 清除会话相关信息

	oldSession, err := this.IDevices.GetSessionByDeviceId(deviceId)

	if err == nil {
		this.closeAll(oldSession, deviceId)
	}

	sess.SetDeviceId(deviceId)

	if !this.IDevices.IsExists(deviceId){

		this.IDevices.Add(deviceId, sess)
	}

	if !this.iRedis.IsExistClient(deviceId){

		localIp := common.GetMyIp()

		ip64, _ := common.TranferIpToint64(localIp)

		log.Println("ip detail", localIp, ip64)

		this.iRedis.InitClient(deviceId, ip64)

		this.iRedis.SaveLockData(common.Lock_Belong_IP, deviceId, ip64)
	}

	this.iRedis.SaveLockData(common.Lock_Cstu, deviceId, "4")

	//if !this.iRedis.IsNormalLockElectric(deviceId){
	//	sess.Encode("stus")
	//}

}

func (this *MessageCenter) FocusSub(data string) {
	log.Print("接收到订阅信息:")

	color.Green(data)

	if _, ok := this.taoMemory.Get(data + "_sub"); ok{
		return
	}

	this.taoMemory.Set(data + "_sub", data, time.Second * 40)

	list := strings.Split(data, "|")

	if len(list) == 2 {
		deviceId := list[0]
		action := list[1]

		sess, err := this.IDevices.GetSessionByDeviceId(deviceId)

		if err == nil {
			//sendAction := rebuildAction(action)

			log.Print("向终端发送信息:")
			color.Green( deviceId, " wt" + action)

			this.taoMemory.Set(deviceId + "_action_" + action, 1, time.Second * 10)

			sess.Encode("wt" + action)
		} else {
			fmt.Println("不存在此会话！", err.Error())
		}
	}
}

/**
	缓存过期回调
 */
func (this *MessageCenter) cacheCallBack(key string, value interface{}){


	if strings.Contains(key, "_sub") {
		return
	} else if strings.Contains(key,"_action_") {

		callValue := value.(int) + 1

		log.Println("动作回调发送！ 次数:", callValue)

		list := strings.Split(key, "_")

		if len(list) != 3 {
			return
		}

		deviceId, action:= list[0], list[2]

		sess, err := this.IDevices.GetSessionByDeviceId(deviceId)

		if err != nil {
			log.Println("cache callback getSessionByDeviceId errors:", err)
			return
		}

		if callValue == 5 {
			this.iRedis.SaveLockData(common.Lock_Cstu, deviceId, Lock_Contact_Cstu)

			//todo  如果终端不回应是否干掉会话？  yes

			log.Println("发送次数超限 会话会在5s后断开")

			time.Sleep(time.Second * 5)

			this.closeAll(sess, deviceId)

			this.taoMemory.Delete(key)

			return
		}

		log.Println("发送消息 to:", deviceId)

		sess.Encode("wt" + action)

		this.taoMemory.Set(key, callValue, time.Second * 10)
	}
}

func (this *MessageCenter) closeAll(sess taonetBase.ITaoSession, deviceId string){

	sess.Close()

	this.IDevices.Remove(deviceId)

}

func resolve(data string)(string, string){

	action := data[:6]

	deviceId := data[6:]

	return action, deviceId
}


func NewMessageCenter(redis taonetDB.ITaoRedis)*MessageCenter{

	msgInstance := &MessageCenter{
		IDevices:manage.NewDeviceManage(),
		iRedis:redis,
		taoMemory:memory.SelectMemory(common.Message_Global),
	}

	msgInstance.taoMemory.OnEvicted(msgInstance.cacheCallBack)

	return msgInstance
}