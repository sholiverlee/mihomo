//go:build wahoo_slim

package adapter

import (
	"fmt"

	"github.com/metacubex/mihomo/adapter/outbound"
	"github.com/metacubex/mihomo/common/structure"
	C "github.com/metacubex/mihomo/constant"
)

func ParseProxy(mapping map[string]any, options ...ProxyOption) (C.Proxy, error) {
	decoder := structure.NewDecoder(structure.Option{TagName: "proxy", WeaklyTypedInput: true, KeyReplacer: structure.DefaultKeyReplacer})
	proxyType, existType := mapping["type"].(string)
	if !existType {
		return nil, fmt.Errorf("missing type")
	}

	opt := applyProxyOptions(options...)
	basicOption := outbound.BasicOption{
		DialerForAPI: opt.DialerForAPI,
		ProviderName: opt.ProviderName,
	}

	var (
		proxy outbound.ProxyAdapter
		err   error
	)
	switch proxyType {
	case "vless":
		vlessOption := &outbound.VlessOption{BasicOption: basicOption}
		err = decoder.Decode(mapping, vlessOption)
		if err != nil {
			break
		}
		proxy, err = outbound.NewVless(*vlessOption)
	case "direct":
		directOption := &outbound.DirectOption{BasicOption: basicOption}
		err = decoder.Decode(mapping, directOption)
		if err != nil {
			break
		}
		proxy = outbound.NewDirectWithOption(*directOption)
	case "dns":
		dnsOptions := &outbound.DnsOption{BasicOption: basicOption}
		err = decoder.Decode(mapping, dnsOptions)
		if err != nil {
			break
		}
		proxy = outbound.NewDnsWithOption(*dnsOptions)
	case "reject":
		rejectOption := &outbound.RejectOption{BasicOption: basicOption}
		err = decoder.Decode(mapping, rejectOption)
		if err != nil {
			break
		}
		proxy = outbound.NewRejectWithOption(*rejectOption)
	default:
		return nil, fmt.Errorf("wahoo_slim unsupported proxy type: %s", proxyType)
	}

	if err != nil {
		return nil, err
	}

	if _, muxExist := mapping["smux"]; muxExist {
		return nil, fmt.Errorf("wahoo_slim does not support smux")
	}

	proxy = outbound.NewAutoCloseProxyAdapter(proxy)
	return NewProxy(proxy), nil
}

type proxyOption struct {
	DialerForAPI C.Dialer
	ProviderName string
}

func applyProxyOptions(options ...ProxyOption) proxyOption {
	opt := proxyOption{}
	for _, o := range options {
		o(&opt)
	}
	return opt
}

type ProxyOption func(opt *proxyOption)

func WithDialerForAPI(dialer C.Dialer) ProxyOption {
	return func(opt *proxyOption) {
		opt.DialerForAPI = dialer
	}
}

func WithProviderName(name string) ProxyOption {
	return func(opt *proxyOption) {
		opt.ProviderName = name
	}
}
