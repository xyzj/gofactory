package gofactory

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"syscall"
	"time"

	"github.com/tidwall/sjson"
	"github.com/xyzj/toolbox/cache"
	"github.com/xyzj/toolbox/crypto"
	"github.com/xyzj/toolbox/json"
	"github.com/xyzj/toolbox/logger"
	"github.com/xyzj/toolbox/loopfunc"
)

type discoverType byte

const (
	byUDP discoverType = iota
	byRedis
)

type ProtocolType byte

const (
	ProtocolTCP ProtocolType = iota
	ProtocolHTTP
	ProtocolHTTPS
	ProtocolMQTT
	ProtocolMQTTTLS
	ProtocolMQTTWS
)

type svrinfo struct {
	RegisterAddress map[ProtocolType]string `json:"register_address,omitempty"` // 注册地址
	SvrName         string                  `json:"svr_name,omitempty"`         // 注册服务名称
	SvrAlias        string                  `json:"svr_alias,omitempty"`        // 注册别名
	SvrSource       string                  `json:"svr_source,omitempty"`       // 服务本地原始地址
	RootPath        string                  `json:"root_path,omitempty"`        // 业务标识
	UpdateTime      string                  `json:"update_time,omitempty"`      // 更新时间,可读
	UpdateStamp     int64                   `json:"update_stamp,omitempty"`     // 更新时间戳
}

func (s *svrinfo) Json() string {
	s.UpdateStamp = time.Now().Unix()
	b, _ := json.Marshal(s)
	return json.String(b)
}

type svrinfoUDP struct {
	Info string `json:"info"`
	Key  string `json:"key"`
	Sign string `json:"sign"`
}

func (s *svrinfoUDP) json() []byte {
	b, _ := json.Marshal(s)
	return b
}

type discover struct {
	rediscli        cliRedis
	svrInfo         *svrinfo
	infos           *cache.AnyCache[svrinfo]
	cryp            *crypto.SM2
	infoTimeout     time.Duration // 服务消息超时
	publishInterval time.Duration // 服务消息更新间隔
	info            string
	udpPort         int // udp端口
	discoverType    discoverType
	enable          bool
}

func (opt *discover) packInfo(encode bool) string {
	opt.info, _ = sjson.Set(opt.info, "update_stamp", time.Now().Unix())
	opt.info, _ = sjson.Set(opt.info, "update_time", time.Now().Format("2006-01-02 15:04:05"))
	if !encode {
		return opt.info
	}
	bb, err := opt.cryp.Encode([]byte(opt.info))
	if err != nil {
		return ""
	}
	bbb, err := opt.cryp.Sign([]byte(bb.Base64String()))
	if err != nil {
		return ""
	}
	a := svrinfoUDP{
		Key:  fmt.Sprintf("%s/discover/%s/%s", opt.svrInfo.RootPath, opt.svrInfo.SvrName, time.Now().Format("Jan-01-02 15:04:05.000000")),
		Info: bb.Base64String(),
		Sign: bbb.Base64String(),
	}
	return json.String(a.json())
}

func (opt *discover) unpackInfo(s string, encode bool) (*svrinfo, error) {
	var err error
	if encode {
		aa := svrinfoUDP{}
		err = json.UnmarshalFromString(s, &aa)
		if err != nil {
			return nil, err
		}
		ok, _ := opt.cryp.VerifySignFromBase64(aa.Sign, json.Bytes(aa.Info))
		if !ok {
			return nil, errors.New("verify sign failed")
		}
		s, err = opt.cryp.DecodeBase64(aa.Info)
		if err != nil {
			return nil, err
		}
	}
	a := svrinfo{}
	err = json.Unmarshal(json.Bytes(s), &a)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (opt *discover) build(l logger.Logger) error {
	opt.infos = cache.NewAnyCache[svrinfo](opt.infoTimeout)
	opt.cryp = crypto.NewSM2()
	opt.cryp.SetPrivateKey(getSM2())
	key := fmt.Sprintf("%s/discover/%s/%s", opt.svrInfo.RootPath, opt.svrInfo.SvrName, time.Now().Format("Jan-01-02 15:04:05.000000"))
	switch opt.discoverType {
	case byRedis:
		go loopfunc.LoopFunc(func(params ...any) {
			err := opt.rediscli.build(l)
			if err != nil {
				return
			}
			t1 := time.NewTicker(opt.publishInterval)
			fWrite := func() {
				if err := opt.rediscli.write(key, opt.packInfo(false), opt.infoTimeout); err != nil {
					l.Error("[discover] write info error:" + err.Error())
				}
			}
			fRead := func() {
				val, err := opt.rediscli.keys(fmt.Sprintf("%s/discover/*", opt.svrInfo.RootPath))
				if err != nil {
					l.Error("[discover] read keys error:" + err.Error())
					return
				}
				for _, k := range val {
					v, err := opt.rediscli.read(k)
					if err != nil {
						l.Error("[discover] read info error:" + err.Error())
						continue
					}
					x, err := opt.unpackInfo(v, false)
					if err != nil {
						l.Error("[discover] unpack info error:" + err.Error())
						continue
					}
					if time.Since(time.Unix(x.UpdateStamp, 0)) > opt.infoTimeout {
						continue
					}
					opt.infos.Store(k, *x)
				}
			}
			for range t1.C {
				fWrite()
				fRead()
			}
		}, "discover", l.DefaultWriter())
	case byUDP:
		go loopfunc.LoopFunc(func(params ...any) {
			p, err := net.ResolveUDPAddr("udp", "255.255.255.255:"+strconv.Itoa(opt.udpPort))
			if err != nil {
				l.Error("[discover] resolve udp addr error:" + err.Error())
				return
			}
			fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_DGRAM, syscall.IPPROTO_UDP)
			if err != nil {
				l.Error("[discover] create udp socket error:" + err.Error())
				return
			}
			defer syscall.Close(fd)
			syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1) // 允许多个程序绑定同一端口
			syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_BROADCAST, 1) // 允许广播

			addr := syscall.SockaddrInet4{Port: opt.udpPort}
			// copy(addr.Addr[:], net.IPv4zero)
			if err := syscall.Bind(fd, &addr); err != nil {
				l.Error("[discover] bind udp port error:" + err.Error())
				return
			}
			u, err := net.FilePacketConn(os.NewFile(uintptr(fd), ""))
			if err != nil {
				l.Error("[discover] create udp conn error:" + err.Error())
				return
			}
			l.System("[discover] start udp server on port:" + strconv.Itoa(opt.udpPort))
			go func() {
				t1 := time.NewTicker(opt.publishInterval)
				var err error
				for range t1.C {
					_, err = u.WriteTo(json.Bytes(opt.packInfo(true)), p)
					if err != nil {
						l.Error("[discover] publish info error:" + err.Error())
					}
					l.Debug("[discover] publish info")
				}
			}()
			b := make([]byte, 4096)
			a := &svrinfo{}
			aa := svrinfoUDP{}
			var n int
			var ra net.Addr
			for {
				n, ra, err = u.ReadFrom(b)
				if err != nil || n == 0 {
					continue
				}
				a, err = opt.unpackInfo(string(b[:n]), true)
				if err != nil {
					l.Error("[discover] unpack error from " + ra.String() + ":" + err.Error())
					continue
				}
				if time.Since(time.Unix(a.UpdateStamp, 0)) > opt.infoTimeout {
					continue
				}
				opt.infos.Store(aa.Key, *a)
				l.Debug("[discover] unpack info from " + ra.String() + ":" + a.Json())
			}
		}, "discover", l.DefaultWriter())
	default:
		return errors.New("[discover] type error")
	}
	return nil
}

type discoverOpts func(o *discover)

func OptDiscoverInfo(name, alias, source, rootpath string, address map[ProtocolType]string) discoverOpts {
	return func(o *discover) {
		if name == "" {
			name = time.Now().Format("x_150405.000")
		}
		o.svrInfo.SvrName = name
		o.svrInfo.SvrAlias = alias
		o.svrInfo.SvrSource = source
		o.svrInfo.RootPath = rootpath
		o.svrInfo.RegisterAddress = address
		o.info = o.svrInfo.Json()
	}
}

func OptDiscoverByRedis(opts ...redisOpts) discoverOpts {
	return func(o *discover) {
		o.discoverType = byRedis
		o.rediscli = cliRedis{
			user:         "",
			pwd:          "",
			addr:         "127.0.0.1:6379",
			database:     0,
			readTimeout:  time.Second * 5,
			writeTimeout: time.Second * 10,
		}
		for _, v := range opts {
			v(&o.rediscli)
		}
	}
}

func OptDiscoverByUDP(port int) discoverOpts {
	return func(o *discover) {
		o.discoverType = byUDP
		o.udpPort = min(max(port, 1024), 65535)
	}
}

// OptDiscoverInfoTimeout sets the timeout of service info.
//
// The timeout should between 1 minute and 3 seconds.
func OptDiscoverInfoTimeout(timeo, interval time.Duration) discoverOpts {
	return func(o *discover) {
		o.publishInterval = min(max(interval, time.Second*30), time.Second)
		if int64(timeo.Seconds()) <= int64(o.publishInterval.Seconds()) {
			timeo = o.publishInterval + time.Second
		}
		o.infoTimeout = min(max(timeo, time.Minute), time.Second*3)
	}
}
