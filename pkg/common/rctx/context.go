package rctx

import "context"

type MD map[string]interface{}

type mdKey struct{}

func New(m map[string]interface{}) MD {
	md := MD{}
	for k, val := range m {
		md[k] = val
	}

	return md
}

func Join(mds ...MD) MD {
	out := MD{}
	for _, md := range mds {
		for k, v := range md {
			out[k] = v
		}
	}
	return out
}

// set metadata in context
func NewContext(ctx context.Context, md MD) context.Context {
	return context.WithValue(ctx, mdKey{}, md)
}

// get metadata from context
func FromContext(ctx context.Context) (md MD, ok bool) {
	md, ok = ctx.Value(mdKey{}).(MD)
	return
}

func String(ctx context.Context, key string) string {
	md, ok := ctx.Value(mdKey{}).(MD)
	if !ok {
		return ""
	}

	str, _ := md[key].(string)
	return str
}

func Int64(ctx context.Context, key string) int64 {
	md, ok := ctx.Value(mdKey{}).(MD)
	if !ok {
		return 0
	}

	i64, _ := md[key].(int64)
	return i64
}

func Value(ctx context.Context, key string) interface{} {
	md, ok := ctx.Value(mdKey{}).(MD)
	if !ok {
		return nil
	}
	return md[key]
}
