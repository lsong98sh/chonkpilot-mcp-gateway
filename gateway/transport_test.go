package mcpgateway

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
)

// testTransport 是 `Transport` 的**内存直连实现**（仅测试用）。
//
// 与生产实现的区别：生产 = gwadapter/mqadapter（把 Op / Hooks 映射到 61 消息主题）；
// 测试 = 本文件（Call 同步调用已注册 handler，Hooks 把语义事件直接广播给测试订阅者）。
// **不涉及任何消息总线**——gateway 内核不知道主题名，测试也不需要模拟 MQ。
type testTransport struct {
	mu        sync.RWMutex
	handlers  map[Op][]func(ctx context.Context, payload []byte) ([]byte, error)
	listeners map[eventKind][]func(map[string]any)
}

// eventKind 是 gateway 三类出站通知（与 Hooks 的三个回调一一对应）。
type eventKind int

const (
	eventTaskReport eventKind = iota
	eventMCPChanged
	eventToolsTimeout
)

func newTestTransport() *testTransport {
	return &testTransport{
		handlers:  map[Op][]func(context.Context, []byte) ([]byte, error){},
		listeners: map[eventKind][]func(map[string]any){},
	}
}

// Call 同步调用该 op 的 handler（无订阅者 = 明确失败，便于发现漏装配）。
func (t *testTransport) Call(ctx context.Context, op Op, payload []byte) ([]byte, error) {
	t.mu.RLock()
	var h func(context.Context, []byte) ([]byte, error)
	if list := t.handlers[op]; len(list) > 0 {
		h = list[0]
	}
	t.mu.RUnlock()
	if h == nil {
		return nil, fmt.Errorf("testTransport: op %q 无订阅者", op)
	}
	return h(ctx, payload)
}

// Subscribe 注册方法面 handler（同 op 多订阅者取首个）。
func (t *testTransport) Subscribe(op Op, h func(ctx context.Context, payload []byte) ([]byte, error)) (Unsubscribe, error) {
	if h == nil {
		return nil, fmt.Errorf("testTransport: op %q handler 为 nil", op)
	}
	t.mu.Lock()
	t.handlers[op] = append(t.handlers[op], h)
	t.mu.Unlock()
	return func() error { return nil }, nil
}

// Hooks 返回把语义事件广播给测试订阅者的回调（等价生产 mqadapter.Hooks() 的装配位置）。
func (t *testTransport) Hooks() Hooks {
	return Hooks{
		OnTaskReport:   func(payload map[string]any) { t.broadcast(eventTaskReport, payload) },
		OnMCPChanged:   func(payload map[string]any) { t.broadcast(eventMCPChanged, payload) },
		OnToolsTimeout: func(payload map[string]any) { t.broadcast(eventToolsTimeout, payload) },
	}
}

// on 订阅某类出站通知（测试用）。返回退订句柄。
func (t *testTransport) on(kind eventKind, fn func(map[string]any)) Unsubscribe {
	t.mu.Lock()
	t.listeners[kind] = append(t.listeners[kind], fn)
	t.mu.Unlock()
	return func() error { return nil }
}

// broadcast 同步投递给该类通知的全部测试订阅者（payload 传副本，避免串扰）。
func (t *testTransport) broadcast(kind eventKind, payload map[string]any) {
	t.mu.RLock()
	list := append([]func(map[string]any){}, t.listeners[kind]...)
	t.mu.RUnlock()
	if len(list) == 0 {
		return
	}
	// 深拷贝：内核与测试共享同一 map 时，任一侧修改不应影响另一侧。
	raw, err := json.Marshal(payload)
	if err != nil {
		return
	}
	for _, fn := range list {
		var cp map[string]any
		if json.Unmarshal(raw, &cp) != nil {
			continue
		}
		fn(cp)
	}
}

// Close 清理订阅（测试清理入口；与生产 Transport 无对应语义）。
func (t *testTransport) Close() error {
	t.mu.Lock()
	t.handlers = map[Op][]func(context.Context, []byte) ([]byte, error){}
	t.listeners = map[eventKind][]func(map[string]any){}
	t.mu.Unlock()
	return nil
}
