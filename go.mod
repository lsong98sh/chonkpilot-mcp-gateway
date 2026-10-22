// gateway 内核 module（2026-10-21）：**零本仓依赖**（生产包 = gateway/ + sandbox/ + winproc/；
// 同包测试经 internal/testbus 代替 mq，不再引用 core）。
// 适配外壳（mqadapter / facadeapi / internal/facade）见 github.com/chonkpilot/chonkpilot-gwadapter。
module github.com/chonkpilot/chonkpilot-mcp-gateway

go 1.26.0

require github.com/modelcontextprotocol/go-sdk v1.7.0

require (
	github.com/google/jsonschema-go v0.4.3 // indirect
	github.com/segmentio/asm v1.1.3 // indirect
	github.com/segmentio/encoding v0.5.4 // indirect
	github.com/yosida95/uritemplate/v3 v3.0.2 // indirect
	golang.org/x/oauth2 v0.35.0 // indirect
	golang.org/x/sync v0.22.0 // indirect
	golang.org/x/sys v0.47.0 // indirect
	golang.org/x/time v0.15.0 // indirect
	golang.org/x/tools v0.48.0 // indirect
)
