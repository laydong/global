package gtools

import (
	"container/list"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/oschwald/geoip2-golang"
	uuid "github.com/satori/go.uuid"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
	"unsafe"
)

// Md5 md5
func Md5(s string) string {
	m := md5.Sum([]byte(s))
	return hex.EncodeToString(m[:])
}

// GetRandomString 获取随机字符串
func GetRandomString(l int) string {
	str := "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := []byte(str)
	var result []byte
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < l; i++ {
		result = append(result, b[r.Intn(len(b))])
	}
	return string(result)
}

// GetRandomString6 获取6位随机字符串
func GetRandomString6(n uint64) []byte {
	baseStr := "0123456789ABCDEFGHJKLMNPQRSTUVWXYZ"
	base := []byte(baseStr)
	quotient := n
	mod := uint64(0)
	l := list.New()
	for quotient != 0 {
		mod = quotient % 34
		quotient = quotient / 34
		l.PushFront(base[int(mod)])
	}
	listLen := l.Len()
	if listLen >= 6 {
		res := make([]byte, 0, listLen)
		for i := l.Front(); i != nil; i = i.Next() {
			res = append(res, i.Value.(byte))
		}
		return res
	} else {
		res := make([]byte, 0, 6)
		for i := 0; i < 6; i++ {
			if i < 6-listLen {
				res = append(res, base[0])
			} else {
				res = append(res, l.Front().Value.(byte))
				l.Remove(l.Front())
			}

		}
		return res
	}
}

// GenValidateCode 生成6位随机验证码
func GenValidateCode(width int) string {
	numeric := [10]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	r := len(numeric)
	rand.Seed(time.Now().UnixNano())

	var sb strings.Builder
	for i := 0; i < width; i++ {
		_, _ = fmt.Fprintf(&sb, "%d", numeric[rand.Intn(r)])
	}
	return sb.String()
}

// RemoteIp 返回远程客户端的 IP，如 192.168.1.1
func RemoteIp(r *http.Request) string {
	remoteAddr := r.RemoteAddr
	if ip := r.Header.Get(XRealIP); ip != "" {
		remoteAddr = ip
	} else if ip = r.Header.Get(XForwardedFor); ip != "" {
		remoteAddr = ip
	} else {
		remoteAddr, _, _ = net.SplitHostPort(remoteAddr)
	}

	if remoteAddr == "::1" {
		remoteAddr = "127.0.0.1"
	}

	return remoteAddr
}

// CreateOrder 生成订单号
func CreateOrder() int64 {
	return int64(rand.New(rand.NewSource(time.Now().UnixNano())).Int31n(1000000))
}

// GetAddressByIP 获取省市区通过ip
func GetAddressByIP(ipA string) string {
	db, err := geoip2.Open("GeoLite2-City.mmdb")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	// If you are using strings that may be invalid, check that ip is not nil
	ip := net.ParseIP(ipA)
	record, err := db.City(ip)
	if err != nil {
		log.Fatal(err)
	}
	var province string
	if len(record.Subdivisions) > 0 {
		province = record.Subdivisions[0].Names["zh-CN"]
	}

	return record.Country.Names["zh-CN"] + "-" + province + "-" + record.City.Names["zh-CN"]
}

// InSliceString string是否在[]string里面
func InSliceString(k string, s []string) bool {
	for _, v := range s {
		if k == v {
			return true
		}
	}
	return false
}

// Exists 判断文件或目录是否存在
func Exists(filePath string) bool {
	_, err := os.Stat(filePath)
	if err == nil {
		return true
	}
	return os.IsExist(err)
}

func IsNil(obj interface{}) bool {
	type eFace struct {
		data unsafe.Pointer
	}
	if obj == nil {
		return true
	}
	return (*eFace)(unsafe.Pointer(&obj)).data == nil
}

// Base64URLDecode 因为Base64转码后可能包含有+,/,=这些不安全的URL字符串，所以要进行换字符
//'+' -> '-'
//'/' -> '_'
//'=' -> ''
//字符串长度不足4倍的位补"="
func Base64URLDecode(data string) string {
	var missing = (4 - len(data)%4) % 4
	data += strings.Repeat("=", missing) //字符串长度不足4倍的位补"="
	data = strings.Replace(data, "_", "/", -1)
	data = strings.Replace(data, "-", "+", -1)
	return data
}

func Base64UrlSafeEncode(data string) string {
	safeUrl := strings.Replace(data, "/", "_", -1)
	safeUrl = strings.Replace(safeUrl, "+", "-", -1)
	safeUrl = strings.Replace(safeUrl, "=", "", -1)
	return safeUrl
}

func Base64Encode(s string) string {
	encodeString := base64.StdEncoding.EncodeToString([]byte(s))
	return encodeString
}

func Base64Decode(code string) string {
	decodeBytes, err := base64.StdEncoding.DecodeString(code)
	if err != nil {
		log.Fatalln(err)
	}
	return string(decodeBytes)
}

// GenerateLogId 获取logId
func GenerateLogId() string {
	s := uuid.NewV4().String()
	return Md5(s)
}

// GetString 只能是map和slice
func GetString(d interface{}) string {
	bytesD, err := CJson.Marshal(d)
	if err != nil {
		return fmt.Sprintf("%v", d)
	} else {
		return string(bytesD)
	}
}

func GetLocalIPs() (ips []string) {
	interfaceAddr, err := net.InterfaceAddrs()
	if err != nil {
		fmt.Printf("fail to get net interface addrs: %v", err)
		return ips
	}

	for _, address := range interfaceAddr {
		ipNet, isValidIpNet := address.(*net.IPNet)
		if isValidIpNet && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				ips = append(ips, ipNet.IP.String())
			}
		}
	}
	return ips
}
