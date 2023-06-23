package wxalarm

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Hyingerrr/mirco-esim/config"

	"github.com/Hyingerrr/mirco-esim/pkg/hepler"

	"github.com/stretchr/testify/assert"
)

func TestWXAlarm_SendMessage(t *testing.T) {
	var (
		it  = assert.New(t)
		ctx = context.Background()
		c   *WXAlarm
	)

	options := config.ViperConfOptions{}
	conf := config.NewViperConfig(options.WithConfigType("yaml"),
		options.WithConfFile([]string{"../../config/a.yaml"}))

	it.NotPanics(func() {
		c = New(WithConf(conf))
	})

	// text
	ip, _ := hepler.GetLocalIp()
	message := fmt.Sprintf(AlarmTemplate, "dev", "esim", ip,
		time.Now().Format("2006-01-02 15:04:05"), "测试微信推送消息", "暂无")
	output, err := c.SendTextMessage(ctx, message, nil, []string{"13816076064"})
	it.Nil(err)
	it.Equal(0, output.ErrCode)

	// markdown
	msg := `实时新增用户反馈<font color="warning">132例</font>，请相关同事注意。
> 类型:<font color="comment">用户反馈</font>
> 普通用户反馈:<font color="comment">117例</font>
> VIP用户反馈:<font color="comment">15例</font>`
	output, err = c.SendMarkdownMessageWithRetry(ctx, msg)
	it.Nil(err)
	it.Equal(0, output.ErrCode)
}
