package work

import (
	"Taonet/webHandler"
	"log"
	"net/http"
	"fmt"
	"Taonet/conf"
)

type WebServer struct {

}


func (web *WebServer) RunWork(){

	router :=  webHandler.NewRouter()

	addr := fmt.Sprintf("%s:%s", "0.0.0.0", conf.TaoConf["web_listen_port"])

	log.Fatal(http.ListenAndServe(addr, router))
}


func NewWebServer() * WebServer{

	return &WebServer{}
}