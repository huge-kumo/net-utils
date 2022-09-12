package curl

import (
	"context"
	"github.com/huge-kumo/net-utils/pkg/log"
	"testing"
)

func TestTracing(t *testing.T) {
	ali, err := Tracing(context.Background(), "https://www.taobao.com")
	if err != nil {
		log.GetInstance().Error("页面请求失败 ", err.Error())
		return
	}
	ali.Display()
}
