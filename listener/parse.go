//go:build wahoo_slim

package listener

import (
	"fmt"

	"github.com/metacubex/mihomo/common/structure"
	C "github.com/metacubex/mihomo/constant"
	IN "github.com/metacubex/mihomo/listener/inbound"
)

func ParseListener(mapping map[string]any) (C.InboundListener, error) {
	decoder := structure.NewDecoder(structure.Option{TagName: "inbound", WeaklyTypedInput: true, KeyReplacer: structure.DefaultKeyReplacer})
	proxyType, existType := mapping["type"].(string)
	if !existType {
		return nil, fmt.Errorf("missing type")
	}

	var (
		listener C.InboundListener
		err      error
	)
	switch proxyType {
	case "tun":
		tunOption := &IN.TunOption{
			Stack:     C.TunGvisor,
			DNSHijack: []string{"0.0.0.0:53"},
		}
		err = decoder.Decode(mapping, tunOption)
		if err != nil {
			return nil, err
		}
		listener, err = IN.NewTun(tunOption)
	default:
		return nil, fmt.Errorf("wahoo_slim unsupported listener type: %s", proxyType)
	}
	return listener, err
}
