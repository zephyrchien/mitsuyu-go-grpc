package main

import (
	"flag"
	"fmt"
	"mitsuyu/client"
	"mitsuyu/common"
	"mitsuyu/manager"
	"mitsuyu/server"
	"mitsuyu/terminal"
	"os"
	"strconv"
)

var (
	mode   = flag.String("m", "", "mode, server/client/client_terminal")
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
	color    = flag.String("color", "black", "terminal background color")
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
	} else if (*mode == "client" || *mode == "client_terminal") && *config == "" {
		loadSingleClient(m)
	} else if (*mode == "client" || *mode == "client_terminal") && *config != "" {
		loadClient(m)
	} else {
		fmt.Println("use cmd flags or specify config file")
		os.Exit(0)
	}
	m.Start()
	if *mode == "client_terminal" {
		t, err := terminal.NewTerminal(m, *color, 0.2, 0.7)
		if err != nil {
			fmt.Println(err)
			os.Exit(0)
		}
		r := manager.NewLogRecorder()
		m.SetRecorder(r)
		m.StartLog(r)
		m.StartConnector()
		m.StartStatistician()
		go t.Run()
	} else {
		m.StartLog(os.Stdout)
	}
	select {}
}

func loadServer(m *manager.Manager) {
	var s common.ServerConfig
	if err := common.LoadServerConfig(*config, &s); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	ss, err := server.New(&s)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	m.Add(ss, false)
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
	m.Add(s, false)
}

func loadClient(m *manager.Manager) {
	var c common.ClientConfig
	if err := common.LoadClientConfig(*config, &c); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	cc, err := client.New(&c)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	m.Add(cc, *mode == "client_terminal")
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
	m.Add(c, *mode == "client_terminal")
}
