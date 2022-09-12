package openwrt

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type Router struct {
	addr          string
	username      string
	password      string
	timeout       time.Duration
	cookieSysAuth string
}

type RouterConfig struct {
	Addr     string
	Username string
	Password string
}

type ProxyNodeInfo struct {
	Name    string // 代理名称
	Id      string // 代理编号
	Host    string // 代理主机地址
	Port    string // 代理服务端口
	Latency int    // 延迟时间
	Offline bool   // 下线状态
}

const (
	LoginPath                  = "/cgi-bin/luci/"                              // 登录路径
	ListAllProxyNodeInfoPath   = "/cgi-bin/luci/admin/services/vssr/servers"   // 列出所有代理服务器节点信息
	UpdateSubscribePath        = "/cgi-bin/luci/admin/services/vssr/subscribe" // 更新订阅
	TestProxyNodeLatencyPath   = "/cgi-bin/luci/admin/services/vssr/checkport" // 节点测速延迟
	ApplyProxyNodeToGlobalPath = "/cgi-bin/luci/admin/services/vssr/change"    // 应用节点配置为全局代理
)

// NewRouterInstance 获取路由实例
func NewRouterInstance(conf *RouterConfig) (r *Router) {
	r = &Router{
		addr:     conf.Addr,
		username: conf.Username,
		password: conf.Password,
	}
	return
}

// Login 登录接口
func (r *Router) Login() (err error) {
	// 构建请求
	path := fmt.Sprintf("http://%s%s", r.addr, LoginPath)
	req, err := http.NewRequest(http.MethodPost, path, strings.NewReader(fmt.Sprintf(`luci_username=%s&luci_password=%s`, r.username, r.password)))
	if err != nil {
		return
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	// 服务请求
	client := http.Client{
		Timeout: r.timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return errors.New("break")
		},
	}
	var rsp *http.Response
	rsp, err = client.Do(req)
	if err != nil {
		if !strings.Contains(err.Error(), "break") {
			return
		}
		err = nil

		if rsp == nil || rsp.StatusCode != 302 {
			err = errors.New("get cookie err, rsp is nil or status code not equal 302")
			return
		}
	}
	defer func() {
		_ = rsp.Close
	}()

	// 序列化响应结果
	cookie := rsp.Header.Get("Set-Cookie")
	items := strings.Split(cookie, ";")
	for _, item := range items {
		if strings.Contains(item, "sysauth") {
			r.cookieSysAuth = item
			return
		}
	}

	err = errors.New("cookie 'sysauth' not found")
	return
}

// ListAllProxyNodeInfo 列出所有代理节点信息
func (r *Router) ListAllProxyNodeInfo() (pns []*ProxyNodeInfo, err error) {
	if r.cookieSysAuth == "" {
		if err = r.Login(); err != nil {
			return
		}
	}

	// 构建请求
	path := fmt.Sprintf("http://%s%s", r.addr, ListAllProxyNodeInfoPath)
	req, err := http.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return
	}
	req.Header.Add("Cookie", r.cookieSysAuth)

	// 服务请求
	client := http.Client{
		Timeout: r.timeout,
	}
	var rsp *http.Response
	rsp, err = client.Do(req)
	if err != nil {
		return
	}
	defer func() {
		_ = rsp.Close
	}()

	// 解析html
	doc, err := goquery.NewDocumentFromReader(rsp.Body)
	if err != nil {
		return
	}
	doc.Find(".cbi-section-table-row").Each(func(i int, selection *goquery.Selection) {
		ins := &ProxyNodeInfo{}

		// 获取主机地址
		var exist bool
		ins.Host, exist = selection.Attr("server")
		if !exist {
			return
		}

		// 获取服务端口
		ins.Port, exist = selection.Attr("server_port")
		if !exist {
			return
		}

		// 获取编号
		selection.Find(".incon").Each(func(i int, idDoc *goquery.Selection) {
			ins.Id, _ = idDoc.Attr("data-setction")
		})
		if ins.Id == "" {
			return
		}

		// 获取名称
		selection.Find(".alias").Each(func(i int, nameDoc *goquery.Selection) {
			ins.Name = nameDoc.Text()
			ins.Name = strings.ReplaceAll(ins.Name, "\n", "")
			ins.Name = strings.ReplaceAll(ins.Name, " ", "")
		})
		pns = append(pns, ins)
	})
	return
}

// TestProxyNodeLatency 测试代理节点延迟
func (r *Router) TestProxyNodeLatency(pn *ProxyNodeInfo) (err error) {
	if r.cookieSysAuth == "" {
		if err = r.Login(); err != nil {
			return
		}
	}

	// 构建请求
	path := fmt.Sprintf("http://%s%s", r.addr, TestProxyNodeLatencyPath+fmt.Sprintf("?host=%s&port=%s", pn.Host, pn.Port))
	req, err := http.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return
	}
	req.Header.Add("Cookie", r.cookieSysAuth)

	// 服务请求
	client := http.Client{
		Timeout: r.timeout,
	}
	var rsp *http.Response
	rsp, err = client.Do(req)
	if err != nil {
		return
	}
	defer func() {
		_ = rsp.Close
	}()

	// 序列化数据
	body, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return
	}
	data := &struct {
		Ret  string `json:"ret"`
		Used int    `json:"used"`
	}{}
	err = json.Unmarshal(body, data)
	if err != nil {
		return
	}
	pn.Offline = data.Ret == "0"
	pn.Latency = data.Used
	return
}

// ApplyProxyNodeToGlobal 应用代理节点到全局
func (r *Router) ApplyProxyNodeToGlobal(p *ProxyNodeInfo) (err error) {
	if r.cookieSysAuth == "" {
		if err = r.Login(); err != nil {
			return
		}
	}

	// 获取节点编号
	var id string
	if p == nil {
		id = "nil"
	} else {
		id = p.Id
	}

	// 准备请求
	path := fmt.Sprintf("http://%s%s", r.addr, ApplyProxyNodeToGlobalPath+fmt.Sprintf("?server=global&set=%s", id))
	req, err := http.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return
	}
	req.Header.Add("Cookie", r.cookieSysAuth)

	// 接口请求
	client := http.Client{
		Timeout: r.timeout,
	}
	var rsp *http.Response
	rsp, err = client.Do(req)
	if err != nil {
		return
	}
	defer func() {
		_ = rsp.Close
	}()

	// 序列化数据
	body, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return
	}
	data := &struct {
		Status bool   `json:"status,omitempty"`
		Sid    string `json:"sid,omitempty"`
	}{}
	err = json.Unmarshal(body, data)
	if err != nil {
		return
	}
	if data.Status != true {
		err = errors.New(data.Sid)
	}
	return
}

// UpdateSubscribeInfo 更新订阅信息
func (r *Router) UpdateSubscribeInfo(urls ...string) (err error) {
	if r.cookieSysAuth == "" {
		if err = r.Login(); err != nil {
			return
		}
	}

	// 构建订阅列表
	if len(urls) == 0 {
		return err
	}
	var s string
	for idx, url := range urls {
		s = fmt.Sprintf(`"%s"`, url)
		if idx != len(urls)-1 {
			s = s + `,`
		}
	}
	s = fmt.Sprintf(`[%s]`, s)

	// 构建请求
	path := fmt.Sprintf("http://%s%s", r.addr, UpdateSubscribePath)
	req, err := http.NewRequest(http.MethodPost, path, strings.NewReader(fmt.Sprintf(`auto_update=1&auto_update_time=2&subscribe_url=%s&proxy=0&filter_words=过期时间/剩余流量`, s)))
	if err != nil {
		return
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Cookie", r.cookieSysAuth)

	// 服务请求
	client := http.Client{
		Timeout: r.timeout,
	}
	var rsp *http.Response
	rsp, err = client.Do(req)
	if err != nil {
		return
	}
	defer func() {
		_ = rsp.Close
	}()

	// 序列化数据
	body, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return
	}
	data := &struct {
		Error int `json:"error"`
	}{}
	err = json.Unmarshal(body, data)
	if err != nil {
		return
	}
	if data.Error != 0 {
		err = fmt.Errorf("error not equal zero %d", data.Error)
		return
	}

	return
}
