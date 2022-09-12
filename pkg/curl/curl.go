package curl

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"time"
)

const format = `{"http_code":"%{http_code}","time_connect":"%{time_connect}s","time_start_transfer":"%{time_starttransfer}s","time_total":"%{time_total}s","time_name_lookup":"%{time_namelookup}s"}`

type AccessLatencyInfo struct {
	HttpCode          string        `json:"http_code"`           // HTTP状态码
	TimeNameLookup    time.Duration `json:"time_name_lookup"`    // DNS解析域名的时间
	TimeConnect       time.Duration `json:"time_connect"`        // 客户端和服务端建立TCP连接的时间
	TimeStartTransfer time.Duration `json:"time_start_transfer"` // 从客服端发送请求到服务端收到并响应首个字节的时间
	TimeTotal         time.Duration `json:"time_total"`          // 整个请求到响应总耗时
}

func (c *AccessLatencyInfo) unmarshalJSON(data []byte) (err error) {
	m := make(map[string]interface{})
	if err = json.Unmarshal(data, &m); err != nil {
		return
	}

	for key, val := range m {
		switch key {
		case "time_connect":
			if result, ok := val.(string); ok {
				if c.TimeConnect, err = time.ParseDuration(result); err != nil {
					return err
				}
			}
		case "time_start_transfer":
			if result, ok := val.(string); ok {
				if c.TimeStartTransfer, err = time.ParseDuration(result); err != nil {
					return err
				}
			}
		case "time_total":
			if result, ok := val.(string); ok {
				if c.TimeTotal, err = time.ParseDuration(result); err != nil {
					return err
				}
			}
		case "time_name_lookup":
			if result, ok := val.(string); ok {
				if c.TimeNameLookup, err = time.ParseDuration(result); err != nil {
					return err
				}
			}
		case "http_code":
			if result, ok := val.(string); ok {
				c.HttpCode = result
			}
		}
	}
	return
}

func (c *AccessLatencyInfo) Display() {
	fmt.Printf("请求状态: %s\t域名解析: %dms\t建立连接: %dms\t首个字节: %dms\t整体耗时: %dms\n",
		c.HttpCode, c.TimeNameLookup.Milliseconds(), c.TimeConnect.Milliseconds(), c.TimeStartTransfer.Milliseconds(), c.TimeTotal.Milliseconds())
}

func Tracing(ctx context.Context, url string) (ali *AccessLatencyInfo, err error) {
	// 构造curl命令参数列表
	var args []string
	args = append(args, "-o", "/dev/null") // 不保存输出HTTP响应
	args = append(args, "-w", format)      // 对请求测量统计
	args = append(args, "-L")              // HTTP请求跟随服务器的重定向
	args = append(args, url)

	// 执行请求
	cmd := exec.CommandContext(ctx, "curl", args...)
	out, err := cmd.Output()
	if err != nil {
		return
	}

	// 对输出结果进行处理
	ali = &AccessLatencyInfo{}
	if err = ali.unmarshalJSON(out); err != nil {
		return
	}

	return
}
