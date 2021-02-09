package main

import (
	"flag"
	"fmt"
	"github.com/ZephyrChien/Mitsuyu/client"
	"github.com/ZephyrChien/Mitsuyu/common"
	"github.com/ZephyrChien/Mitsuyu/manager"
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
	m := manager.NewManager()
	if *mode == "server" && *config == "" {
		loadSingleServer(m)
	} else if *mode == "server" && *config != "" {
		loadServer(m)
	} else if *mode == "client" && *config == "" {
		loadSingleClient(m)
	} else if *mode == "client" && *config != "" {
		loadClient(m)
	} else {
		fmt.Println("use cmd flags or specify config file")
		os.Exit(0)
	}
	m.StartAll()
	m.StartLogAll(os.Stdout)
	select {}
}

func loadServer(m *manager.Manager) {
	var servers []*common.ServerConfig
	if err := common.LoadServerConfig(*config, &servers); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	for i, s := range servers {
		tag := s.Tag
		if tag == "" {
			tag = strconv.Itoa(i)
		}
		ss, err := server.New(s)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		m.Add(tag, ss)
	}
}

func loadSingleServer(m *manager.Manager) {
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
	m.Add("single", s)
}

func loadClient(m *manager.Manager) {
	var clients []*common.ClientConfig
	if err := common.LoadClientConfig(*config, &clients); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	for i, c := range clients {
		tag := c.Tag
		if tag == "" {
			tag = strconv.Itoa(i)
		}
		cc, err := client.New(c)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		m.Add(tag, cc)
	}
}

func loadSingleClient(m *manager.Manager) {
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
	m.Add("single", c)
}
