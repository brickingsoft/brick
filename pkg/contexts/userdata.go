package contexts

import "context"

type userdata struct {
	key   any
	value any
}

type UserdataContext interface {
	context.Context
	Userdata(key any) any
	WithUserdata(key any, value any)
}

type userdataContext struct {
	context.Context
	entries []userdata
}

func (ctx *userdataContext) Value(key any) any {
	if v := ctx.Userdata(key); v != nil {
		return v
	}
	return ctx.Context.Value(key)
}

func (ctx *userdataContext) Userdata(key any) any {
	if key == nil {
		return nil
	}
	if ctx.entries == nil {
		return nil
	}
	for _, entry := range ctx.entries {
		if entry.key == key {
			return entry.value
		}
	}
	return nil
}

func (ctx *userdataContext) WithUserdata(key any, value any) {
	if key == nil {
		return
	}
	for i := range ctx.entries {
		if ctx.entries[i].key == key {
			ctx.entries[i].value = value
			return
		}
	}
	ctx.entries = append(ctx.entries, userdata{key, value})
	return
}

func WrapUserdataContext(ctx context.Context) UserdataContext {
	return &userdataContext{
		Context: ctx,
		entries: nil,
	}
}
