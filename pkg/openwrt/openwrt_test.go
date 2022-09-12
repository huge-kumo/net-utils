package openwrt

import "testing"

func TestRouter_Login(t *testing.T) {
	r := NewRouterInstance(&RouterConfig{
		Addr:     "192.168.2.1:80",
		Username: "root",
		Password: "password",
	})

	err := r.Login()
	if err != nil {
		panic(err)
	}
}

func TestRouter_ListAllProxyNodeInfo(t *testing.T) {
	r := NewRouterInstance(&RouterConfig{
		Addr:     "192.168.2.1:80",
		Username: "root",
		Password: "password",
	})

	_, err := r.ListAllProxyNodeInfo()
	if err != nil {
		panic(err)
	}
}

func TestRouter_TestProxyNodeLatency(t *testing.T) {
	r := NewRouterInstance(&RouterConfig{
		Addr:     "192.168.2.1:80",
		Username: "root",
		Password: "password",
	})

	list, err := r.ListAllProxyNodeInfo()
	if err != nil {
		panic(err)
	}

	err = r.TestProxyNodeLatency(list[0])
	if err != nil {
		panic(err)
	}
}

func TestRouter_ApplyProxyNodeToGlobal(t *testing.T) {
	r := NewRouterInstance(&RouterConfig{
		Addr:     "192.168.2.1:80",
		Username: "root",
		Password: "password",
	})

	list, err := r.ListAllProxyNodeInfo()
	if err != nil {
		panic(err)
	}

	err = r.ApplyProxyNodeToGlobal(list[0])
	if err != nil {
		panic(err)
	}
}
