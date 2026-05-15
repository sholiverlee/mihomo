//go:build wahoo_slim

package outbound

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"
	"time"

	N "github.com/metacubex/mihomo/common/net"
	tlsC "github.com/metacubex/mihomo/component/tls"
	C "github.com/metacubex/mihomo/constant"
	"github.com/metacubex/mihomo/transport/vless"
	"github.com/metacubex/mihomo/transport/vless/encryption"
	"github.com/metacubex/mihomo/transport/vmess"
	"github.com/metacubex/mihomo/transport/xhttp"

	"github.com/metacubex/http"
)

type Vless struct {
	*Base
	client *vless.Client
	option *VlessOption

	encryption *encryption.ClientInstance

	xhttpClient *xhttp.Client

	realityConfig *tlsC.RealityConfig
}

type ECHOptions struct {
	Enable bool   `proxy:"enable,omitempty" obfs:"enable,omitempty"`
	Config string `proxy:"config,omitempty" obfs:"config,omitempty"`

	QueryServerName string `proxy:"query-server-name,omitempty" obfs:"query-server-name,omitempty"`
}

type HTTPOptions struct {
	Method  string              `proxy:"method,omitempty"`
	Path    []string            `proxy:"path,omitempty"`
	Headers map[string][]string `proxy:"headers,omitempty"`
}

type HTTP2Options struct {
	Host []string `proxy:"host,omitempty"`
	Path string   `proxy:"path,omitempty"`
}

type GrpcOptions struct {
	GrpcServiceName string `proxy:"grpc-service-name,omitempty"`
	GrpcUserAgent   string `proxy:"grpc-user-agent,omitempty"`
	PingInterval    int    `proxy:"ping-interval,omitempty"`
	MaxConnections  int    `proxy:"max-connections,omitempty"`
	MinStreams      int    `proxy:"min-streams,omitempty"`
	MaxStreams      int    `proxy:"max-streams,omitempty"`
}

type WSOptions struct {
	Path                     string            `proxy:"path,omitempty"`
	Headers                  map[string]string `proxy:"headers,omitempty"`
	MaxEarlyData             int               `proxy:"max-early-data,omitempty"`
	EarlyDataHeaderName      string            `proxy:"early-data-header-name,omitempty"`
	V2rayHttpUpgrade         bool              `proxy:"v2ray-http-upgrade,omitempty"`
	V2rayHttpUpgradeFastOpen bool              `proxy:"v2ray-http-upgrade-fast-open,omitempty"`
}

type VlessOption struct {
	BasicOption
	Name              string            `proxy:"name"`
	Server            string            `proxy:"server"`
	Port              int               `proxy:"port"`
	UUID              string            `proxy:"uuid"`
	Flow              string            `proxy:"flow,omitempty"`
	TLS               bool              `proxy:"tls,omitempty"`
	ALPN              []string          `proxy:"alpn,omitempty"`
	UDP               bool              `proxy:"udp,omitempty"`
	PacketAddr        bool              `proxy:"packet-addr,omitempty"`
	XUDP              bool              `proxy:"xudp,omitempty"`
	PacketEncoding    string            `proxy:"packet-encoding,omitempty"`
	Encryption        string            `proxy:"encryption,omitempty"`
	Network           string            `proxy:"network,omitempty"`
	ECHOpts           ECHOptions        `proxy:"ech-opts,omitempty"`
	RealityOpts       RealityOptions    `proxy:"reality-opts,omitempty"`
	HTTPOpts          HTTPOptions       `proxy:"http-opts,omitempty"`
	HTTP2Opts         HTTP2Options      `proxy:"h2-opts,omitempty"`
	GrpcOpts          GrpcOptions       `proxy:"grpc-opts,omitempty"`
	WSOpts            WSOptions         `proxy:"ws-opts,omitempty"`
	XHTTPOpts         XHTTPOptions      `proxy:"xhttp-opts,omitempty"`
	WSHeaders         map[string]string `proxy:"ws-headers,omitempty"`
	SkipCertVerify    bool              `proxy:"skip-cert-verify,omitempty"`
	Fingerprint       string            `proxy:"fingerprint,omitempty"`
	Certificate       string            `proxy:"certificate,omitempty"`
	PrivateKey        string            `proxy:"private-key,omitempty"`
	ServerName        string            `proxy:"servername,omitempty"`
	ClientFingerprint string            `proxy:"client-fingerprint,omitempty"`
}

type XHTTPOptions struct {
	Path                 string                 `proxy:"path,omitempty"`
	Host                 string                 `proxy:"host,omitempty"`
	Mode                 string                 `proxy:"mode,omitempty"`
	Headers              map[string]string      `proxy:"headers,omitempty"`
	NoGRPCHeader         bool                   `proxy:"no-grpc-header,omitempty"`
	XPaddingBytes        string                 `proxy:"x-padding-bytes,omitempty"`
	XPaddingObfsMode     bool                   `proxy:"x-padding-obfs-mode,omitempty"`
	XPaddingKey          string                 `proxy:"x-padding-key,omitempty"`
	XPaddingHeader       string                 `proxy:"x-padding-header,omitempty"`
	XPaddingPlacement    string                 `proxy:"x-padding-placement,omitempty"`
	XPaddingMethod       string                 `proxy:"x-padding-method,omitempty"`
	UplinkHTTPMethod     string                 `proxy:"uplink-http-method,omitempty"`
	SessionPlacement     string                 `proxy:"session-placement,omitempty"`
	SessionKey           string                 `proxy:"session-key,omitempty"`
	SeqPlacement         string                 `proxy:"seq-placement,omitempty"`
	SeqKey               string                 `proxy:"seq-key,omitempty"`
	UplinkDataPlacement  string                 `proxy:"uplink-data-placement,omitempty"`
	UplinkDataKey        string                 `proxy:"uplink-data-key,omitempty"`
	UplinkChunkSize      string                 `proxy:"uplink-chunk-size,omitempty"`
	ScMaxEachPostBytes   string                 `proxy:"sc-max-each-post-bytes,omitempty"`
	ScMinPostsIntervalMs string                 `proxy:"sc-min-posts-interval-ms,omitempty"`
	ReuseSettings        *XHTTPReuseSettings    `proxy:"reuse-settings,omitempty"`
	DownloadSettings     *XHTTPDownloadSettings `proxy:"download-settings,omitempty"`
}

type XHTTPReuseSettings struct {
	MaxConcurrency   string `proxy:"max-concurrency,omitempty"`
	MaxConnections   string `proxy:"max-connections,omitempty"`
	CMaxReuseTimes   string `proxy:"c-max-reuse-times,omitempty"`
	HMaxRequestTimes string `proxy:"h-max-request-times,omitempty"`
	HMaxReusableSecs string `proxy:"h-max-reusable-secs,omitempty"`
	HKeepAlivePeriod int    `proxy:"h-keep-alive-period,omitempty"`
}

type XHTTPDownloadSettings struct {
	Path                 *string             `proxy:"path,omitempty"`
	Host                 *string             `proxy:"host,omitempty"`
	Headers              *map[string]string  `proxy:"headers,omitempty"`
	NoGRPCHeader         *bool               `proxy:"no-grpc-header,omitempty"`
	XPaddingBytes        *string             `proxy:"x-padding-bytes,omitempty"`
	XPaddingObfsMode     *bool               `proxy:"x-padding-obfs-mode,omitempty"`
	XPaddingKey          *string             `proxy:"x-padding-key,omitempty"`
	XPaddingHeader       *string             `proxy:"x-padding-header,omitempty"`
	XPaddingPlacement    *string             `proxy:"x-padding-placement,omitempty"`
	XPaddingMethod       *string             `proxy:"x-padding-method,omitempty"`
	UplinkHTTPMethod     *string             `proxy:"uplink-http-method,omitempty"`
	SessionPlacement     *string             `proxy:"session-placement,omitempty"`
	SessionKey           *string             `proxy:"session-key,omitempty"`
	SeqPlacement         *string             `proxy:"seq-placement,omitempty"`
	SeqKey               *string             `proxy:"seq-key,omitempty"`
	UplinkDataPlacement  *string             `proxy:"uplink-data-placement,omitempty"`
	UplinkDataKey        *string             `proxy:"uplink-data-key,omitempty"`
	UplinkChunkSize      *string             `proxy:"uplink-chunk-size,omitempty"`
	ScMaxEachPostBytes   *string             `proxy:"sc-max-each-post-bytes,omitempty"`
	ScMinPostsIntervalMs *string             `proxy:"sc-min-posts-interval-ms,omitempty"`
	ReuseSettings        *XHTTPReuseSettings `proxy:"reuse-settings,omitempty"`

	Server            *string         `proxy:"server,omitempty"`
	Port              *int            `proxy:"port,omitempty"`
	TLS               *bool           `proxy:"tls,omitempty"`
	ALPN              *[]string       `proxy:"alpn,omitempty"`
	ECHOpts           *ECHOptions     `proxy:"ech-opts,omitempty"`
	RealityOpts       *RealityOptions `proxy:"reality-opts,omitempty"`
	SkipCertVerify    *bool           `proxy:"skip-cert-verify,omitempty"`
	Fingerprint       *string         `proxy:"fingerprint,omitempty"`
	Certificate       *string         `proxy:"certificate,omitempty"`
	PrivateKey        *string         `proxy:"private-key,omitempty"`
	ServerName        *string         `proxy:"servername,omitempty"`
	ClientFingerprint *string         `proxy:"client-fingerprint,omitempty"`
}

func (v *Vless) StreamConnContext(ctx context.Context, c net.Conn, metadata *C.Metadata) (_ net.Conn, err error) {
	switch v.option.Network {
	case "xhttp":
		break // already handled by xhttpClient.Dial
	case "tcp":
		c, err = v.streamTLSConn(ctx, c, false)
	default:
		return nil, fmt.Errorf("wahoo_slim VLESS supports xhttp and tcp reality only")
	}
	if err != nil {
		return nil, err
	}
	return v.streamConnContext(ctx, c, metadata)
}

func (v *Vless) streamConnContext(ctx context.Context, c net.Conn, metadata *C.Metadata) (conn net.Conn, err error) {
	if ctx.Done() != nil {
		done := N.SetupContextForConn(ctx, c)
		defer done(&err)
	}
	if metadata.NetWork == C.UDP {
		return nil, C.ErrNotSupport
	}
	if v.encryption != nil {
		c, err = v.encryption.Handshake(c)
		if err != nil {
			return nil, err
		}
	}
	conn, err = v.client.StreamConn(c, parseVlessAddr(metadata, false))
	if err != nil {
		conn = nil
	}
	return conn, err
}

func (v *Vless) streamTLSConn(ctx context.Context, conn net.Conn, isH2 bool) (net.Conn, error) {
	if !v.option.TLS {
		return conn, nil
	}

	host, _, _ := net.SplitHostPort(v.addr)
	tlsOpts := vmess.TLSConfig{
		Host:              host,
		SkipCertVerify:    v.option.SkipCertVerify,
		FingerPrint:       v.option.Fingerprint,
		Certificate:       v.option.Certificate,
		PrivateKey:        v.option.PrivateKey,
		ClientFingerprint: v.option.ClientFingerprint,
		Reality:           v.realityConfig,
		NextProtos:        v.option.ALPN,
	}

	if isH2 {
		tlsOpts.NextProtos = []string{"h2"}
	}

	if v.option.ServerName != "" {
		tlsOpts.Host = v.option.ServerName
	}

	return vmess.StreamTLSConn(ctx, conn, &tlsOpts)
}

func (v *Vless) dialContext(ctx context.Context) (net.Conn, error) {
	switch v.option.Network {
	case "xhttp":
		return v.xhttpClient.Dial()
	case "tcp":
		return v.dialer.DialContext(ctx, "tcp", v.addr)
	default:
		return nil, fmt.Errorf("wahoo_slim VLESS supports xhttp and tcp reality only")
	}
}

func (v *Vless) DialContext(ctx context.Context, metadata *C.Metadata) (_ C.Conn, err error) {
	c, err := v.dialContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s connect error: %s", v.addr, err.Error())
	}
	defer func(c net.Conn) {
		safeConnClose(c, err)
	}(c)

	c, err = v.StreamConnContext(ctx, c, metadata)
	if err != nil {
		return nil, fmt.Errorf("%s connect error: %s", v.addr, err.Error())
	}
	return NewConn(c, v), err
}

func (v *Vless) ListenPacketContext(ctx context.Context, metadata *C.Metadata) (_ C.PacketConn, err error) {
	return nil, C.ErrNotSupport
}

func (v *Vless) SupportUOT() bool {
	return false
}

func (v *Vless) ProxyInfo() C.ProxyInfo {
	info := v.Base.ProxyInfo()
	info.DialerProxy = v.option.DialerProxy
	return info
}

func (v *Vless) Close() error {
	var errs []error
	if v.xhttpClient != nil {
		if err := v.xhttpClient.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func parseVlessAddr(metadata *C.Metadata, xudp bool) *vless.DstAddr {
	var addrType byte
	var addr []byte
	switch metadata.AddrType() {
	case C.AtypIPv4:
		addrType = vless.AtypIPv4
		addr = make([]byte, net.IPv4len)
		copy(addr[:], metadata.DstIP.AsSlice())
	case C.AtypIPv6:
		addrType = vless.AtypIPv6
		addr = make([]byte, net.IPv6len)
		copy(addr[:], metadata.DstIP.AsSlice())
	case C.AtypDomainName:
		addrType = vless.AtypDomainName
		addr = make([]byte, len(metadata.Host)+1)
		addr[0] = byte(len(metadata.Host))
		copy(addr[1:], metadata.Host)
	}

	return &vless.DstAddr{
		UDP:      metadata.NetWork == C.UDP,
		AddrType: addrType,
		Addr:     addr,
		Port:     metadata.DstPort,
		Mux:      metadata.NetWork == C.UDP && xudp,
	}
}

func NewVless(option VlessOption) (*Vless, error) {
	if option.Network != "xhttp" && option.Network != "tcp" {
		return nil, fmt.Errorf("wahoo_slim VLESS supports xhttp and tcp reality only")
	}
	if option.ECHOpts.Enable {
		return nil, fmt.Errorf("wahoo_slim VLESS does not support ECH")
	}
	if option.Network == "xhttp" {
		if option.XHTTPOpts.DownloadSettings != nil {
			return nil, fmt.Errorf("wahoo_slim VLESS xhttp does not support download-settings")
		}
		if len(option.ALPN) == 1 && option.ALPN[0] == "h3" {
			return nil, fmt.Errorf("wahoo_slim VLESS xhttp does not support HTTP/3")
		}
	}

	var addons *vless.Addons
	if len(option.Flow) >= 16 {
		option.Flow = option.Flow[:16]
		if option.Flow != vless.XRV {
			return nil, fmt.Errorf("unsupported xtls flow type: %s", option.Flow)
		}
		addons = &vless.Addons{
			Flow: option.Flow,
		}
	}

	option.UDP = false
	option.PacketAddr = false
	option.XUDP = false

	client, err := vless.NewClient(option.UUID, addons)
	if err != nil {
		return nil, err
	}

	v := &Vless{
		Base: NewBase(BaseOption{
			Name:         option.Name,
			Addr:         net.JoinHostPort(option.Server, strconv.Itoa(option.Port)),
			Type:         C.Vless,
			ProviderName: option.ProviderName,
			UDP:          false,
			XUDP:         false,
			TFO:          option.TFO,
			MPTCP:        option.MPTCP,
			Interface:    option.Interface,
			RoutingMark:  option.RoutingMark,
			Prefer:       option.IPVersion,
		}),
		client: client,
		option: &option,
	}
	v.dialer = option.NewDialer(v.DialOptions())

	v.encryption, err = encryption.NewClient(option.Encryption)
	if err != nil {
		return nil, err
	}

	v.realityConfig, err = v.option.RealityOpts.Parse()
	if err != nil {
		return nil, err
	}
	if option.Network == "tcp" {
		if !option.TLS || v.realityConfig == nil {
			return nil, fmt.Errorf("wahoo_slim VLESS tcp requires reality")
		}
		return v, nil
	}

	requestHost := v.option.XHTTPOpts.Host
	if requestHost == "" {
		if v.option.ServerName != "" {
			requestHost = v.option.ServerName
		} else {
			requestHost = v.option.Server
		}
	}

	var hKeepAlivePeriod time.Duration

	var reuseCfg *xhttp.ReuseConfig
	if option.XHTTPOpts.ReuseSettings != nil {
		reuseCfg = &xhttp.ReuseConfig{
			MaxConcurrency:   option.XHTTPOpts.ReuseSettings.MaxConcurrency,
			MaxConnections:   option.XHTTPOpts.ReuseSettings.MaxConnections,
			CMaxReuseTimes:   option.XHTTPOpts.ReuseSettings.CMaxReuseTimes,
			HMaxRequestTimes: option.XHTTPOpts.ReuseSettings.HMaxRequestTimes,
			HMaxReusableSecs: option.XHTTPOpts.ReuseSettings.HMaxReusableSecs,
		}
		hKeepAlivePeriod = time.Duration(option.XHTTPOpts.ReuseSettings.HKeepAlivePeriod) * time.Second
	}

	cfg := &xhttp.Config{
		Host:                 requestHost,
		Path:                 v.option.XHTTPOpts.Path,
		Mode:                 v.option.XHTTPOpts.Mode,
		Headers:              v.option.XHTTPOpts.Headers,
		NoGRPCHeader:         v.option.XHTTPOpts.NoGRPCHeader,
		XPaddingBytes:        v.option.XHTTPOpts.XPaddingBytes,
		XPaddingObfsMode:     v.option.XHTTPOpts.XPaddingObfsMode,
		XPaddingKey:          v.option.XHTTPOpts.XPaddingKey,
		XPaddingHeader:       v.option.XHTTPOpts.XPaddingHeader,
		XPaddingPlacement:    v.option.XHTTPOpts.XPaddingPlacement,
		XPaddingMethod:       v.option.XHTTPOpts.XPaddingMethod,
		UplinkHTTPMethod:     v.option.XHTTPOpts.UplinkHTTPMethod,
		SessionPlacement:     v.option.XHTTPOpts.SessionPlacement,
		SessionKey:           v.option.XHTTPOpts.SessionKey,
		SeqPlacement:         v.option.XHTTPOpts.SeqPlacement,
		SeqKey:               v.option.XHTTPOpts.SeqKey,
		UplinkDataPlacement:  v.option.XHTTPOpts.UplinkDataPlacement,
		UplinkDataKey:        v.option.XHTTPOpts.UplinkDataKey,
		UplinkChunkSize:      v.option.XHTTPOpts.UplinkChunkSize,
		ScMaxEachPostBytes:   v.option.XHTTPOpts.ScMaxEachPostBytes,
		ScMinPostsIntervalMs: v.option.XHTTPOpts.ScMinPostsIntervalMs,
		ReuseConfig:          reuseCfg,
	}

	makeTransport := func() http.RoundTripper {
		return xhttp.NewTransport(
			func(ctx context.Context) (net.Conn, error) {
				return v.dialer.DialContext(ctx, "tcp", v.addr)
			},
			func(ctx context.Context, raw net.Conn, isH2 bool) (net.Conn, error) {
				return v.streamTLSConn(ctx, raw, isH2)
			},
			nil,
			v.option.ALPN,
			hKeepAlivePeriod,
		)
	}

	v.xhttpClient, err = xhttp.NewClient(cfg, makeTransport, nil, v.realityConfig != nil)
	if err != nil {
		return nil, err
	}

	return v, nil
}
