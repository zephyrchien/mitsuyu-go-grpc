package main

import (
	"flag"
	"fmt"
	"github.com/ZephyrChien/Mitsuyu/client"
	"github.com/ZephyrChien/Mitsuyu/common"
	"github.com/ZephyrChien/Mitsuyu/server"
	"os"
	"strconv"
)

var (
	mode   = flag.String("m", "", "mode, server/client")
	local  = flag.String("l", "", "listen addr, [client] support socks5/http")
	remote = flag.String("r", "", "[client] remote addr")
	sname  = flag.String("sname", "", "service name [path=/service_name/proxy]")

	tls      = flag.Bool("tls", false, "enable tls")
	sni      = flag.String("tls-sni", "", "[client]server name")
	verify   = flag.Bool("tls-verify", false, "[client] enable verify")
	cafile   = flag.String("tls-verify-ca", "", "[client] specify ca-file")
	certfile = flag.String("tls-cert", "cert.pem", "[server] certificate")
	keyfile  = flag.String("tls-key", "key.pem", "[server] private key")

	compress = flag.Bool("compress", false, "[client] enable compress")
	config   = flag.String("c", "", "config file")
)

func init() {
	flag.Parse()
}

func main() {
	if *mode == "server" && *config == "" {
		startSingleServer()
	} else if *mode == "server" && *config != "" {
		startServer()
	} else if *mode == "client" && *config == "" {
		startSingleClient()
	} else if *mode == "client" && *config != "" {
		startClient()
	} else {
		fmt.Println("use cmd flags or specify config file")
		os.Exit(0)
	}
	select {}
}

func startServer() {
	var servers []*common.ServerConfig
	if err := common.LoadServerConfig(*config, &servers); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	for _, s := range servers {
		ss, err := server.New(s)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		go ss.Serve()
	}
}

func startSingleServer() {
	conf := &common.ServerConfig{
		Addr:        *local,
		ServiceName: *sname,
		TLS:         strconv.FormatBool(*tls),
		TLSCert:     *certfile,
		TLSKey:      *keyfile,
	}
	s, err := server.New(conf)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	go s.Serve()
}

func startClient() {
	var clients []*common.ClientConfig
	if err := common.LoadClientConfig(*config, &clients); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	for _, c := range clients {
		cc, err := client.New(c)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		go cc.Run()
	}
}

func startSingleClient() {
	conf := &common.ClientConfig{
		Local:       *local,
		Remote:      *remote,
		ServiceName: *sname,
		TLS:         strconv.FormatBool(*tls),
		TLSCA:       *cafile,
		TLSSNI:      *sni,
		TLSVerify:   strconv.FormatBool(*verify),
		Compress:    strconv.FormatBool(*compress),
	}
	c, err := client.New(conf)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	go c.Run()
}
