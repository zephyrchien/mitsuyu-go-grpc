package api

import (
	"fmt"
	"mitsuyu/common"
	"mitsuyu/manager"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type Api struct {
	manager *manager.Manager
	handler *http.ServeMux
	addr    string
	base    string
	token   string
	// busy ctrls
	stats *common.Statistician
	conns *common.Connector
}

func NewApi(manager *manager.Manager, addr, token string) *Api {
	handler := http.NewServeMux()
	api := &Api{
		manager: manager,
		handler: handler,
		addr:    addr,
		base:    "/" + manager.GetClient().GetServiceName(),
		token:   token,
		//
		stats: manager.GetStatistician(),
		conns: manager.GetConnector(),
	}
	handler.Handle(api.base+"/traffic", api.handleAuth(api.handleGetTraffic))
	handler.Handle(api.base+"/connection", api.handleAuth(api.handleGetConnection))
	return api
}

func (api *Api) Serve() {
	srv := &http.Server{
		Handler:        api.handler,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   5 * time.Second,
		MaxHeaderBytes: 2048,
	}
	lis, err := net.Listen("tcp", api.addr)
	if err != nil {
		fmt.Printf("Api: Unable to listen %s, %v\n", api.addr, err)
		os.Exit(0)
	}
	go srv.Serve(lis)
}

func (api *Api) handleAuth(next func(http.ResponseWriter, *http.Request)) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("token") != api.token {
				w.WriteHeader(http.StatusForbidden)
				return
			}
			next(w, r)
		},
	)
}

func (api *Api) handleGetTraffic(w http.ResponseWriter, r *http.Request) {
	up, down := api.stats.GetTraffic()
	upstr := strconv.FormatUint(up, 10)
	downstr := strconv.FormatUint(down, 10)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte(upstr + "," + downstr))
}

func (api *Api) handleGetConnection(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte(strings.Join(api.conns.GetReport(), "\n")))
}
