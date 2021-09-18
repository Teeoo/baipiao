package jd

import (
	"bytes"
	"crypto/hmac"
	m "crypto/md5"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	j "encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/robertkrimen/otto"
	. "github.com/teeoo/baipiao/http"
	json "github.com/tidwall/gjson"
	"log"
	"math/big"
	random "math/rand"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

type jx struct {
	c string
}

var (
	randSeek           = int64(1)
	l                  sync.Mutex
	token              string
	rd                 []string
	f                  string
	algo               string
	coins              int       // 金币数量
	foodNum            = 0       // 白菜数量
	activeId           = ""      // 活动ID
	petInfoList        []Petinfo // 小鸡相关信息列表
	cowInfo            Cow       // 牛相关信息
	share_code         string    // 助力码
	phone_id           string    // 设备ID
	egg_num            = 0       // 金蛋数量
	newcomer_task_step = [...]string{"A-1", "A-2", "A-3", "A-4", "A-5", "A-6", "A-7", "A-8", "A-9",
		"A-10", "A-11", "A-12", "B-1", "C-1", "D-1", "E-1", "E-2", "E-3", "E-4", "E-5",
		"F-1", "F-2", "G-1", "G-2", "G-3", "G-4", "G-5", "G-6", "G-7", "G-8", "G-9"}
	curTaskStep string
)

func init() {
	getEncrypt()
}

func request(c *resty.Request, path, body, user string) string {
	phoneId := GetRandomString(16)
	timestamp := time.Now().Unix() * 1000
	jsToken := tom5(fmt.Sprintf("%s%s%stPOamqCuk9NLgVPAljUyIHcPRmKlVxDy", user, strconv.FormatInt(timestamp, 10), phoneId))
	params := url.Values{}
	params.Add("channel", "7")
	params.Add("sceneid", "1001")
	params.Add("activeid", activeId)
	params.Add("activekey", "null")
	params.Add("_ste", "1")
	params.Add("_", strconv.FormatInt(time.Now().Unix()*1000+2, 10))
	params.Add("sceneval", "2")
	params.Add("g_login_type", "1")
	params.Add("callback", "")
	params.Add("g_ty", "ls")
	params.Add("jxmc_jstoken", jsToken)
	//if body != "" {
	//	params.Add("_stk", json.Get(body, "_stk").String())
	//	params.Add("token", json.Get(body, "token").String())
	//}
	var mapBody map[string]string
	_ = j.Unmarshal([]byte(body), &mapBody)
	for s, i := range mapBody {
		if params.Has(s) {
			params.Add(s, i)
		} else {
			params.Set(s, i)
		}
	}
	purl := fmt.Sprintf("https://m.jingxi.com/%s?%s", path, params.Encode())
	h5st := encrypt(purl, "")
	purl = fmt.Sprintf("%s&h5st=%s", purl, h5st)
	log.Println(purl)
	timeResult, _ := rand.Int(rand.Reader, big.NewInt(3))
	time.Sleep(time.Duration(timeResult.Int64()) * time.Second)
	resp, err := c.SetHeaders(map[string]string{"referer": "https://st.jingxi.com/"}).Post(purl)
	if err != nil {
		log.Println(err)
	}
	return string(resp.Body())
}

func tom5(str string) string {
	data := []byte(str)
	has := m.Sum(data)
	return fmt.Sprintf("%x", has)
}

func GetRandomString(num int, str ...string) string {
	s := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	if len(str) > 0 {
		s = str[0]
	}
	l := len(s)
	r := random.New(random.NewSource(getRandSeek()))
	var buf bytes.Buffer
	for i := 0; i < num; i++ {
		x := r.Intn(l)
		buf.WriteString(s[x : x+1])
	}
	return buf.String()
}

func getRandSeek() int64 {
	l.Lock()
	if randSeek >= 100000000 {
		randSeek = 1
	}
	randSeek++
	l.Unlock()
	return time.Now().UnixNano() + randSeek

}

func encrypt(u, stk string) string {
	timestamp := time.Now().Format("20060102150405")
	timestamp = fmt.Sprintf("%s%s", timestamp, strconv.FormatInt(time.Now().UnixNano(), 10)[:3])
	r, _ := url.Parse(u)
	if stk == "" {
		stk = r.Query().Get("_stk")
	}
	s := fmt.Sprintf("%s%s%s%s%s", token, f, timestamp, "10001", rd[1])
	jxx := new(jx)
	method := reflect.ValueOf(jxx).MethodByName(fmt.Sprintf("Call%s", algo))
	var val []reflect.Value
	if strings.Contains(fmt.Sprintf("Call%s", algo), "Hmac") {
		val = method.Call([]reflect.Value{reflect.ValueOf(s), reflect.ValueOf(token)})
	} else {
		val = method.Call([]reflect.Value{reflect.ValueOf(s)})
	}
	var tp []string
	for _, s2 := range strings.Split(stk, ",") {
		tp = append(tp, fmt.Sprintf("%s:%s", s2, r.Query().Get(s2)))
	}
	hash := jxx.CallHmacSHA256(strings.Join(tp, "&"), val[0].String())
	return strings.Join([]string{timestamp, f, "10001", token, hash}, ";")
}

func (t *jx) CallMD5(val string) string {
	data := []byte(val)
	has := m.Sum(data)
	return fmt.Sprintf("%x", has)
}
func (t *jx) CallHmacMD5(key, val string) string {
	h := hmac.New(m.New, []byte(key))
	h.Write([]byte(val))
	return fmt.Sprintf("%x", h.Sum(nil))
}
func (t *jx) CallSHA256(val string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(val)))
}
func (t *jx) CallHmacSHA256(key, val string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(val))
	return fmt.Sprintf("%x", h.Sum(nil))
}
func (t *jx) CallSHA512(val string) string {
	return fmt.Sprintf("%x", sha512.Sum512([]byte(val)))
}
func (t *jx) CallHmacSHA512(key, val string) string {
	h := hmac.New(sha512.New, []byte(key))
	h.Write([]byte(val))
	return fmt.Sprintf("%x", m.Sum(nil))
}

func getEncrypt() {
	f = fp()
	response, err := HttpClient.SetDebug(false).
		R().
		SetHeaders(map[string]string{
			"Authority":       "cactus.jd.com",
			"Pragma":          "no-cache",
			"Cache-Control":   "no-cache",
			"Accept":          "application/json",
			"Content-Type":    "application/json",
			"Origin":          "https://st.jingxi.com",
			"Sec-Fetch-Site":  "cross-site",
			"Sec-Fetch-Mode":  "cors",
			"Sec-Fetch-Dest":  "empty",
			"Referer":         "https://st.jingxi.com/",
			"Accept-Language": "zh-CN,zh;q=0.9,zh-TW;q=0.8,en;q=0.7",
		}).
		SetBody(map[string]string{
			"version":      "1.0",
			"fp":           f,
			"appId":        "10001",
			"timestamp":    strconv.FormatInt(time.Now().Unix()*1000, 10),
			"platform":     "web",
			"expandParams": "",
		}).
		Post("https://cactus.jd.com/request_algo?g_ty=ajax")
	body := string(response.Body())
	if err != nil {
		log.Println("签名算法获取失败:", err)
	}
	if json.Get(body, "status").Int() == 200 {
		token = json.Get(body, "data.result.tk").String()
		rd = regexp.MustCompile("rd='(.*)';").FindStringSubmatch(json.Get(body, "data.result.algo").String())
		algo = regexp.MustCompile(`algo\.(.*)\(`).FindStringSubmatch(json.Get(body, "data.result.algo").String())[1]
		log.Printf("获取到签名算法为: %s tk为: %s", algo, token)
	}

}

func fp() string {
	e := "0123456789"
	a := 13
	i := ""
	for a > 0 {
		result, _ := rand.Int(rand.Reader, big.NewInt(int64(len(e))))
		i += fmt.Sprintf("%s", result)
		a -= 1
	}
	i += fmt.Sprintf("%s", strconv.FormatInt(time.Now().Unix()*100, 10))
	return i[0:16]
}

func getToken(args string) string {
	vm := otto.New()
	js := `
	function t(n, t) {
      var r = (65535 & n) + (65535 & t);
      return (n >> 16) + (t >> 16) + (r >> 16) << 16 | 65535 & r
    }

    function r(n, t) {
      return n << t | n >>> 32 - t
    }

    function e(n, e, o, u, c, f) {
      return t(r(t(t(e, n), t(u, f)), c), o)
    }

    function o(n, t, r, o, u, c, f) {
      return e(t & r | ~t & o, n, t, u, c, f)
    }

    function u(n, t, r, o, u, c, f) {
      return e(t & o | r & ~o, n, t, u, c, f)
    }

    function c(n, t, r, o, u, c, f) {
      return e(t ^ r ^ o, n, t, u, c, f)
    }

    function f(n, t, r, o, u, c, f) {
      return e(r ^ (t | ~o), n, t, u, c, f)
    }

    function i(n, r) {
      n[r >> 5] |= 128 << r % 32, n[14 + (r + 64 >>> 9 << 4)] = r;
      var e, i, a, d, h, l = 1732584193, g = -271733879, v = -1732584194, m = 271733878;
      for (e = 0; e < n.length; e += 16) i = l, a = g, d = v, h = m, g = f(g = f(g = f(g = f(g = c(g = c(g = c(g = c(g = u(g = u(g = u(g = u(g = o(g = o(g = o(g = o(g, v = o(v, m = o(m, l = o(l, g, v, m, n[e], 7, -680876936), g, v, n[e + 1], 12, -389564586), l, g, n[e + 2], 17, 606105819), m, l, n[e + 3], 22, -1044525330), v = o(v, m = o(m, l = o(l, g, v, m, n[e + 4], 7, -176418897), g, v, n[e + 5], 12, 1200080426), l, g, n[e + 6], 17, -1473231341), m, l, n[e + 7], 22, -45705983), v = o(v, m = o(m, l = o(l, g, v, m, n[e + 8], 7, 1770035416), g, v, n[e + 9], 12, -1958414417), l, g, n[e + 10], 17, -42063), m, l, n[e + 11], 22, -1990404162), v = o(v, m = o(m, l = o(l, g, v, m, n[e + 12], 7, 1804603682), g, v, n[e + 13], 12, -40341101), l, g, n[e + 14], 17, -1502002290), m, l, n[e + 15], 22, 1236535329), v = u(v, m = u(m, l = u(l, g, v, m, n[e + 1], 5, -165796510), g, v, n[e + 6], 9, -1069501632), l, g, n[e + 11], 14, 643717713), m, l, n[e], 20, -373897302), v = u(v, m = u(m, l = u(l, g, v, m, n[e + 5], 5, -701558691), g, v, n[e + 10], 9, 38016083), l, g, n[e + 15], 14, -660478335), m, l, n[e + 4], 20, -405537848), v = u(v, m = u(m, l = u(l, g, v, m, n[e + 9], 5, 568446438), g, v, n[e + 14], 9, -1019803690), l, g, n[e + 3], 14, -187363961), m, l, n[e + 8], 20, 1163531501), v = u(v, m = u(m, l = u(l, g, v, m, n[e + 13], 5, -1444681467), g, v, n[e + 2], 9, -51403784), l, g, n[e + 7], 14, 1735328473), m, l, n[e + 12], 20, -1926607734), v = c(v, m = c(m, l = c(l, g, v, m, n[e + 5], 4, -378558), g, v, n[e + 8], 11, -2022574463), l, g, n[e + 11], 16, 1839030562), m, l, n[e + 14], 23, -35309556), v = c(v, m = c(m, l = c(l, g, v, m, n[e + 1], 4, -1530992060), g, v, n[e + 4], 11, 1272893353), l, g, n[e + 7], 16, -155497632), m, l, n[e + 10], 23, -1094730640), v = c(v, m = c(m, l = c(l, g, v, m, n[e + 13], 4, 681279174), g, v, n[e], 11, -358537222), l, g, n[e + 3], 16, -722521979), m, l, n[e + 6], 23, 76029189), v = c(v, m = c(m, l = c(l, g, v, m, n[e + 9], 4, -640364487), g, v, n[e + 12], 11, -421815835), l, g, n[e + 15], 16, 530742520), m, l, n[e + 2], 23, -995338651), v = f(v, m = f(m, l = f(l, g, v, m, n[e], 6, -198630844), g, v, n[e + 7], 10, 1126891415), l, g, n[e + 14], 15, -1416354905), m, l, n[e + 5], 21, -57434055), v = f(v, m = f(m, l = f(l, g, v, m, n[e + 12], 6, 1700485571), g, v, n[e + 3], 10, -1894986606), l, g, n[e + 10], 15, -1051523), m, l, n[e + 1], 21, -2054922799), v = f(v, m = f(m, l = f(l, g, v, m, n[e + 8], 6, 1873313359), g, v, n[e + 15], 10, -30611744), l, g, n[e + 6], 15, -1560198380), m, l, n[e + 13], 21, 1309151649), v = f(v, m = f(m, l = f(l, g, v, m, n[e + 4], 6, -145523070), g, v, n[e + 11], 10, -1120210379), l, g, n[e + 2], 15, 718787259), m, l, n[e + 9], 21, -343485551), l = t(l, i), g = t(g, a), v = t(v, d), m = t(m, h);
      return [l, g, v, m]
    }

    function a(n) {
      var t, r = "", e = 32 * n.length;
      for (t = 0; t < e; t += 8) r += String.fromCharCode(n[t >> 5] >>> t % 32 & 255);
      return r
    }

    function d(n) {
      var t, r = [];
      for (r[(n.length >> 2) - 1] = void 0, t = 0; t < r.length; t += 1) r[t] = 0;
      var e = 8 * n.length;
      for (t = 0; t < e; t += 8) r[t >> 5] |= (255 & n.charCodeAt(t / 8)) << t % 32;
      return r
    }

    function h(n) {
      return a(i(d(n), 8 * n.length))
    }

    function l(n, t) {
      var r, e, o = d(n), u = [], c = [];
      for (u[15] = c[15] = void 0, o.length > 16 && (o = i(o, 8 * n.length)), r = 0; r < 16; r += 1) u[r] = 909522486 ^ o[r], c[r] = 1549556828 ^ o[r];
      return e = i(u.concat(d(t)), 512 + 8 * t.length), a(i(c.concat(e), 640))
    }

    function g(n) {
      var t, r, e = "";
      for (r = 0; r < n.length; r += 1) t = n.charCodeAt(r), e += "0123456789abcdef".charAt(t >>> 4 & 15) + "0123456789abcdef".charAt(15 & t);
      return e
    }

    function v(n) {
      return unescape(encodeURIComponent(n))
    }

    function m(n) {
      return h(v(n))
    }

    function p(n) {
      return g(m(n))
    }

    function s(n, t) {
      return l(v(n), v(t))
    }

    function C(n, t) {
      return g(s(n, t))
    }

    function A(n, t, r) {
      return t ? r ? s(t, n) : C(t, n) : r ? m(n) : p(n)
    }
`
	_, err := vm.Run(js)
	if err != nil {
		log.Println(err)
	}
	value, err := vm.Call("A", nil, args)
	if err != nil {
		log.Println(err)
	}
	return value.String()
}
