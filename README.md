# chonkpilot-mcp-gateway

ChonkPilot 的 **MCP 工具面聚合内核**：把 registered provider 与本地沙箱进程统一到一个方法面上，承载同步调用、异步任务、超时裁决与出站通知。

## 定位

- **零本仓依赖**：只用 `github.com/modelcontextprotocol/go-sdk` + `golang.org/x/sys`，不引用 ChonkPilot 主仓任何代码
- **不认识 MQ**：出站通知以语义回调 `Hooks` 暴露，主题名的决定权在宿主（ChonkPilot 主仓的 `src/lib/gwadapter`）
- **零 MQ 概念**：连测试也不模拟消息总线——同包测试用内存直连 `testTransport`（`gateway/transport_test.go`）

## 目录

| 路径 | 说明 |
|---|---|
| `gateway/` | 内核：`New(Params)`、方法面分发、异步任务、并发池、熔断、工具覆盖、资产元数据 |
| `gateway/port.go` | `Transport`（出站调用 / 方法面订阅）与 `Hooks`（出站通知）契约 |
| `sandbox/` | 进程级写隔离（Windows Low IL + Job Object；非 Windows 退化实现） |
| `winproc/` | 隐藏子进程控制台窗口、进程树终止（Windows / 退化实现） |

## 契约

```go
// 宿主实现：把 Op 映射到自己的传输（MQ / 内存 / HTTP）
type Transport interface {
    Call(ctx context.Context, op Op, payload []byte) ([]byte, error)
    Subscribe(op Op, h func(ctx context.Context, payload []byte) ([]byte, error)) (Unsubscribe, error)
}

// 宿主实现：把语义事件映射到自己的主题名
type Hooks struct {
    OnTaskReport   func(payload map[string]any) // 异步任务完成 / 失败 / 取消
    OnMCPChanged   func(payload map[string]any) // 工具 / 接入 / 状态变化
    OnToolsTimeout func(payload map[string]any) // 工具到达超时，待用户裁决
}
```

`Hooks` 为 `nil` 时静默丢弃，并在 `Start` 打一条 Warn——不会 panic，但也不会自动重试。

## 用法

```go
g, err := gateway.New(gateway.Params{
    Transport:   myTransport,        // 必须
    Hooks:       myHooks,            // 可选；nil = 不发出站通知
    CallTimeout: 60 * time.Second,
    MaxTasks:    8,
    Logf:        log.Printf,
})
if err != nil { return err }
if err := g.Start(ctx); err != nil { return err }
defer g.Stop(ctx)
```

工具调用的超时裁决、异步任务状态机、沙箱可写目录（`SandboxRule`）都在 `Params` / `ServerEntry` 上配置。

## 测试

```bash
go test ./...
```

- `gateway/deps_test.go` 与 `purity_test.go` 断言内核不 import MQ / 主仓包
- `go list -test -deps ./gateway` 应只输出本模块的包

## 与主仓的关系

ChonkPilot 主仓通过 `go.mod` 的 `replace` 引用本模块。MQ 适配（`mqadapter` / `facadeapi`）与 `exe` 外壳**不在本仓**，留在主仓的 `src/lib/gwadapter` 与 `src/others/mcp-gateway`。
