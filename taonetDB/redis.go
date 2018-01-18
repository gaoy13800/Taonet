package taonetDB

import (
	"sync"
	"github.com/go-redis/redis"
	"Taonet/common"
	"strconv"
	"Taonet/task"
	"log"
	"time"
	"Taonet/conf"
	"fmt"
)

type ITaoRedis interface {

	InitClient(deviceId string, ipAddress int64) error

	IsExistClient(deviceId string) bool

	IsNormalLockElectric(deviceId string) bool

	SaveLockData(t common.DBLockAction, deviceId string, data interface{})
}


type TaoRedisCenter struct {

	redis    *redis.Client

	wait     sync.WaitGroup

	syncText sync.RWMutex

	closeSub bool
}

const (
	IPSKEY       = "wt:sevice"
	SubKey       = "wtClintChan"
	ClientKey    = "client:"
	ClientEx     = "client:expire:"
	ClientStusEx = "client:expire:stus:"
	Service      = "service:"
	FlagKey      = "wtFlag"

	FlagClientKey = "flag:detail:"


	UdpListKey = "lock:client:list"
	UdpDetailKey = "lock:client:detail:"
)


/**
	ping redis 测试redis是否同行
 */
func (this *TaoRedisCenter) pass() error {

	pong, err := this.redis.Ping().Result()

	if err != nil || pong != "PONG" {
		fmt.Println("redis ping error:", err)

		return err
	}

	return nil
}

// init wt:sevice
func (this *TaoRedisCenter) init(ipAddr int64) error {

	if _, err := this.redis.SAdd(IPSKEY, strconv.FormatInt(ipAddr, 10)).Result(); err != nil {
		return err
	}

	return nil
}

func (this *TaoRedisCenter) IsExistClient(deviceId string) bool{

	if ret, _ := this.redis.Exists(ClientKey + deviceId).Result(); ret > 0 {
		return true
	}

	return false
}

func (this *TaoRedisCenter) IsNormalLockElectric(deviceId string) bool{

	if n, _ := this.redis.TTL(ClientStusEx + deviceId).Result(); !(n > 0){
		return false
	}

	return true
}

func (this *TaoRedisCenter) InitClient(deviceId string, ipAddress int64) error{

	if index, err := this.redis.Exists(ClientKey + deviceId).Result(); err != nil {
		if index > 0 {
			return nil
		}
	}

	if _, err := this.redis.HSet(ClientKey + deviceId, "FmIP", strconv.FormatInt(ipAddress, 10)).Result(); err != nil {
		return err
	}

	//暂时状态为关闭状态

	if _, err := this.redis.HSet(ClientKey + deviceId, "cstu", 4).Result(); err != nil {
		return err
	}

	return nil

}

func (this *TaoRedisCenter) SaveLockData(t common.DBLockAction, deviceId string, data interface{}){

	switch t {
	case common.Lock_Cstu:
		this.saveLockStatus(deviceId, data.(string))
		break
	case common.Lock_Brut:
		this.saveLockBrut(deviceId, data.(string))
		break
	case common.Lock_Stus:
		this.saveLockElectric(deviceId, data.(string))
		break
	case common.Lock_Init:
		this.init(data.(int64))
		break
	case common.Lock_Belong_IP:
		this.belongToIp(deviceId, data.(int64))
	default:
		break
	}

}

func (this *TaoRedisCenter) belongToIp(deviceId string, ipAddress int64) error {

	//service:<ip>   set   ip与deviceId 的关系

	if _, err := this.redis.SAdd(Service + strconv.FormatInt(ipAddress, 10), deviceId).Result(); err != nil {
		return err
	}

	return nil
}

func (this *TaoRedisCenter) saveLockStatus(deviceId, cstu string) bool{

	if _, err := this.redis.HSet(ClientKey + deviceId, "cstu", cstu).Result();  err != nil{
		return false
	}

	return true
}

func (this *TaoRedisCenter) saveLockElectric(deviceId, stus string) bool{

	_, err := this.redis.HSet(ClientKey + deviceId, "stus", stus).Result()

	this.redis.Set(ClientStusEx + deviceId, time.Now().Format("150405"), time.Second * 60 * 60 * 24)

	if err != nil{
		return false
	}

	return true
}

func (this *TaoRedisCenter) saveLockBrut(deviceId, stus string) bool{
	if _, err := this.redis.HSet(ClientKey + deviceId, "bstu", stus).Result();  err != nil{
		return false
	}

	return true
}

func (this *TaoRedisCenter) subThread(task task.ITask) {

	subHandler := this.redis.Subscribe(SubKey)

	for  {
		if !this.closeSub {
			break
		}

		subData, err := subHandler.ReceiveMessage()

		if err != nil {
			subHandler.Close()
			log.Println("subThead receive errors:", err)
			break
		}

		task.PushMessage(subData.Payload)
	}

	subHandler.Close()

	this.wait.Done()
}

func NewRedisInstance(iTask task.ITask)*TaoRedisCenter{

	this := &TaoRedisCenter{
		redis: redis.NewClient(&redis.Options{
			Addr:     conf.TaoConf["redis_addr"],
			Password: conf.TaoConf["redis_passwd"],
			DB:       1,
		}),
		closeSub: true,
	}

	err := this.pass()

	if err != nil {
		log.Println("redis validate connect errors:", err)
		return nil
	}


	localIp := common.GetMyIp()

	ip64, _ := common.TranferIpToint64(localIp)

	log.Println("ip detail", localIp, ip64)

	this.init(ip64)

	this.wait.Add(1)

	go func() {
		this.wait.Wait()

		this.closeSub = false
	}()

	go this.subThread(iTask)

	return this
}







