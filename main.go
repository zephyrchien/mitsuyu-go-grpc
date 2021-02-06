package main

import (
	"flag"
	"fmt"
	"github.com/ZephyrChien/Mitsuyu/client"
	"github.com/ZephyrChien/Mitsuyu/common"
	"github.com/ZephyrChien/Mitsuyu/server"
)

var (
	mode = flag.String("m", "server", "mode")
	local = flag.String("l","127.0.0.1","8080")
	sname = flag.String("name","Mitsuyu","service name")
	tls = flag.Bool("tls",false,"enable tls")
	certfile = flag.String("cert","cert.pem","certificate directory")
	keyfile = flag.String("key","key.pem","private key directory")
	//
	config = flag.String("c","","config file directory")
)

func init() {
	flag.Parse()
}


func main() {
	if *mode == "server" && *config ==""{
		startSingleServer()
	}else if *mode == "server" && *config !=""{
		startServer()
	}else if *mode == "client" && *config !=""{
		startClient()
	}else{
		fmt.Println("use cmd flags or specify config file")
		return
	}
	select{}
}

func startServer(){
	var servers []*common.ServerConfig
	if err:=common.LoadServerConfig(*config,&servers);err!=nil{
		fmt.Println(err)
		return
	}
	for _,s:=range servers{
		ss,err:=server.New(s)
		if err!=nil{
			fmt.Println(err)
			continue
		}
		go ss.Serve()
	}
}

func startSingleServer() {
	var enableTLS = "false"
	if *tls{
		enableTLS="true"
	}
	conf := &common.ServerConfig{
		Addr:        *local,
		ServiceName: *sname,
		TLS:         enableTLS,
		TLSCert:     *certfile,
		TLSKey:      *keyfile,
	}
	s, err := server.New(conf)
	if err != nil {
		fmt.Println(err)
		return
	}
	go s.Serve()
}

func startClient(){
	var clients []*common.ClientConfig
	if err:=common.LoadClientConfig(*config,&clients);err!=nil{
		fmt.Println(err)
		return
	}
	for _,c:=range clients{
		cc,err:=client.New(c)
		if err!=nil{
			fmt.Println(err)
			continue
		}
		go cc.Run()
	}
}
