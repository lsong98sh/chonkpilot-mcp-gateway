// 端口定义（MOD-1 M0.5）：gateway 内核零 MQ 依赖。
//
// 内核只认 Op（自有操作名，原 MQ 相对主题）与 JSON 字节载荷；「Op → 实际主题」
// 的映射与订阅装配都归宿主 adapter（mq 实现见 mqadapter 包）。内核因此不 import
// chonkpilot-lib/mq 与 msgkeys（L3 断言见 deps_test.go）。
//
// 出站**通知**（任务完成回报 / 目录接入变化 / 工具超时待裁决）不经 Transport 发布，
// 改由宿主注入的 Hooks 承接——「发到哪个主题」完全由宿主决定，内核只给语义事件。
package mcpgateway

import (
	"context"
	"encoding/json"
)

// Op 是 gateway 自有的操作名（原 MQ 相对主题），不引用 msgkeys。
type Op string

// Unsubscribe 退订句柄（幂等；重复调用无副作用）。
type Unsubscribe func() error

// Hooks 是 gateway **出站通知**的宿主回调。内核只表达语义事件，主题名与发布方式
// （MQ / HTTP / 落库 / 丢弃）由宿主决定；MQ 实现见 mqadapter 包的 Hooks()。
//
// 约定：
//   - **nil = 该通知丢弃**（内核不 panic）；装配方遗漏时 Start 会打 Warn 提示；
//   - 实现方**必须非阻塞**（fire-and-forget 语义：内核在调用路径上直接调，不 await）；
//   - payload 为结构化 map，序列化由实现方负责（内核不再预先 marshal）。
type Hooks struct {
	// OnTaskReport 异步任务执行完成 / 失败 / 取消后的回报
	//（payload：task_id / tool / state / result_summary / instance_id / session / turn /
	//  tool_call_id / top_session / parent）。
	OnTaskReport func(payload map[string]any)
	// OnMCPChanged 目录 / 接入变化通知
	//（payload：kind / name / status / reason / instance_id）。
	OnMCPChanged func(payload map[string]any)
	// OnToolsTimeout 工具到达超时、待用户裁决
	//（payload：instance_id / tool_call_id / task_id / tool / reason / timeout_s / options）。
	OnToolsTimeout func(payload map[string]any)
}

// Dispatcher 是 gateway 的**出站**端口（宿主实现）：
//   - Call = 同步请求-响应（语义同原 mq Emit + Wait：返回结果 JSON 字节，nil = 无结果）
type Dispatcher interface {
	Call(ctx context.Context, op Op, payload []byte) ([]byte, error)
}

// Subscriber 是 gateway 的**入站**端口：宿主把消息通道上的方法调用接到 handler，
// handler 返回结果 JSON 字节（nil = 无结果）。
type Subscriber interface {
	Subscribe(op Op, h func(ctx context.Context, payload []byte) ([]byte, error)) (Unsubscribe, error)
}

// Transport = 出站 + 入站（宿主装配时注入 Params.Transport）。
type Transport interface {
	Dispatcher
	Subscriber
}

// call 是一次方法调用的内核替身（原 *mq.Value）：req = 原始载荷；订阅者写回 Result。
type call struct {
	req    []byte
	Result any
}

// JSON 把请求载荷反序列化到 out（无载荷 → 按 null 处理，与 mq.Value.JSON 同义）。
func (c *call) JSON(out any) error {
	if len(c.req) == 0 {
		return json.Unmarshal([]byte("null"), out)
	}
	return json.Unmarshal(c.req, out)
}
