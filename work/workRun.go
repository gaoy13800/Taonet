package work

import (
	"Taonet/task"
	"Taonet/taonetDB"
	"Taonet/message"
	"Taonet/push"
	"Taonet/memory"
	"Taonet/common"
	"Taonet/taonetBase"
	"time"
	"strings"
)

var (

	subTask     task.ITask = task.TaskCreate()
	socketTask  task.ITask = task.TaskCreate()
)

func StartTaoNet(){

	go NewTaoSocket(socketTask).Start()

	go NewWebServer().RunWork()

	MessageDispose()
}

func MessageDispose(){

	redis1 := taonetDB.NewRedisInstance(subTask)

	msgCenter :=  message.NewMessageCenter(redis1)

	go timerDispose(msgCenter.IDevices)

	go func() {

		for msg := range socketTask.GetTask()  {

			msgCenter.Notice(msg.(*push.RuleData))
		}

	}()

	go func() {

		for  {
			subMessage :=  <- subTask.GetTask()

			msgCenter.FocusSub(subMessage.(string))
		}

	}()

}

func timerDispose(devices taonetBase.IDevices){

	time.Sleep(time.Second * 5)


	ticker := time.NewTicker(time.Second * 3)

	globalDB := memory.SelectMemory(common.Message_Global)


	for  {
		select {
		case <- ticker.C:

			sessManage :=devices.GetDeviceIds()

			if len(sessManage) > 0 {
				deviceIds := strings.Join(sessManage, "|")
				globalDB.Set("wt:tao:deviceList", deviceIds, common.Long_Time_Expires)
			}
		}
	}


}
