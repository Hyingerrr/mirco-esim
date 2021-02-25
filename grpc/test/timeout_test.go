package test

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func Test_Timeout(t *testing.T) {
	var cancel context.CancelFunc
	ctx := context.Background()
	timeout := 5000 * time.Millisecond

	tt := time.Now()
	fmt.Println(tt)

	ctx, cancel = context.WithTimeout(ctx, 6000*time.Millisecond)
	if dl, ok := ctx.Deadline(); ok {
		out := time.Until(dl)
		if out-time.Millisecond*20 > 0 {
			out = out - time.Millisecond*20 // 减去本进程中的消耗 略估计
		}

		if timeout > out {
			timeout = out
		}
	}

	// debug
	tt2 := time.Now()
	fmt.Println(tt2)
	ctx, cancel = context.WithTimeout(ctx, timeout)
	defer cancel()
	dls, _ := ctx.Deadline()
	fmt.Println(dls)
}
