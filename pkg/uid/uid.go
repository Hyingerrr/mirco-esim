package uid

import (
	"bytes"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/segmentio/ksuid"
	"github.com/sony/sonyflake"
)

type UIDRepo interface {
	GenID() (string, error)
	TraceID() string
	TradeID() string
}

type UID struct {
	sf       *sonyflake.Sonyflake
	currDate atomic.Value
}

var timeFunc = func() string {
	return time.Now().Format("20060102")
}

func NewUIDRepo() UIDRepo {
	u := new(UID)
	u.currDate.Store(timeFunc())
	//初始化sn
	t, _ := time.Parse("20060102", "20200101")
	settings := sonyflake.Settings{
		StartTime: t,
	}
	u.sf = sonyflake.NewSonyflake(settings)
	go func() {
		t := time.NewTicker(3 * time.Second)
		for {
			select {
			case <-t.C:
				u.currDate.Store(timeFunc())
			}
		}
	}()
	return u
}

func (u *UID) GenID() (string, error) {
	id, err := u.sf.NextID()
	if err != nil {
		return "", err
	}
	return strconv.FormatUint(id, 10), nil
}

func (u *UID) date() string {
	return u.currDate.Load().(string)
}

func (u *UID) TradeID() string {
	var buf bytes.Buffer
	buf.WriteString(u.date())
	id, _ := u.GenID()
	buf.WriteString(id)
	return buf.String()
}

func (u *UID) TraceID() string {
	var buf bytes.Buffer
	buf.WriteString(u.date())
	buf.WriteString(ksuid.New().String())
	return buf.String()

}
