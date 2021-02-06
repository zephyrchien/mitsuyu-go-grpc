package server

import (
	"context"
	"net"
)

func ipLookup(host, dns string) (string, error) {
	r := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{}
			return d.DialContext(ctx, network, dns)
		},
	}
	addrs, err := r.LookupHost(context.Background(), host)
	if err != nil {
		return "", err
	}
	return addrs[0], nil
}
