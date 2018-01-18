package webHandler

import (
	"net/http"
	"fmt"
	"strings"
	"io"
	"encoding/base64"
	"Taonet/memory"
	"Taonet/common"
	"encoding/json"
	"Taonet/taonetDB"
)


/**
	http 接口
 */

func GetPlatformData(w http.ResponseWriter, r *http.Request){

	validateAuth(w , r)

	var deviceCount int

	out := make(map[string]interface{})

	db := memory.SelectMemory(common.Message_Global)

	lists, ok := db.Get("wt:tao:deviceList")

	if !ok{
		deviceCount = 0

		out["deviceCount"] = deviceCount
	}else {

		deviceList := strings.Split(lists.(string), "|")

		deviceCount = len(deviceList)

		out["deviceList"] = deviceList

		out["deviceCount"] = deviceCount
	}

	re, _ := json.Marshal(&out)

	fmt.Fprint(w, string(re))
	return
}


func DBDispose(w http.ResponseWriter, r *http.Request){

	validateAuth(w , r)

	action_type := r.FormValue("type")

	deviceId := r.FormValue("deviceId")

	redisCommon := taonetDB.NewRedisCommon()

	switch action_type {

	case "client":
		data := redisCommon.GetClientData("client:" + deviceId)

		fmt.Println(data)

		result, _ := json.Marshal(&data)

		fmt.Fprint(w, string(result))

		return
	case "client_set":
		setKey:= r.FormValue("key")
		setValue := r.FormValue("value")

		redisCommon.SetClientMessage("client:" + deviceId, setKey, setValue)

		data := make(map[string]string)

		data["result"] = "success"

		result, _ := json.Marshal(&data)

		fmt.Fprint(w, string(result))

		return

	case "client_publish":

		action := r.FormValue("action")

		redisCommon.PublishMessage(deviceId + "|" + action)

		data := make(map[string]string)

		data["result"] = "success"

		result, _ := json.Marshal(&data)

		fmt.Fprint(w, string(result))

		return
	default:
		data := make(map[string]string)

		data["result"] = "error"
		data["desc"] = "检查请求类型！"

		result, _ := json.Marshal(&data)

		fmt.Fprint(w, string(result))
		return
	}

}

func validateAuth(w http.ResponseWriter, r *http.Request){
	auth := r.Header.Get("Authorization")

	if auth == "" {
		w.Header().Set("WWW-Authenticate", `Basic realm="Dotcoo User Login"`)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	auths := strings.SplitN(auth, " ", 2)

	if len(auths) != 2 {
		fmt.Println("error")
		return
	}
	authMethod := auths[0]
	authB64 := auths[1]

	if authMethod != "Basic"{
		io.WriteString(w, "Auth type invalid")
		return
	}

	authStr, err := base64.StdEncoding.DecodeString(authB64)

	if err != nil {
		io.WriteString(w, "Unauthorized!\n")
		return
	}

	userInfo := strings.SplitN(string(authStr), ":", 2)

	username := userInfo[0]
	password := userInfo[1]

	if password != "wt@123" || username != "admin"{
		io.WriteString(w, "Unauthorized!\n 用户名或密码错误，请重新登陆")

		return
	}
}

