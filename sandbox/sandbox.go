// Package sandbox 提供**进程级写隔离**：以 Low Integrity Level 启动子进程，
// 并用 Job Object 约束其资源与进程树。
//
// 归属（2026-10-21）：本包随 gateway 一起发布（原 src/lib/core/sandbox），因为沙箱是
// gateway spawn 下游 MCP server 时的执行能力。**Rule 配置类型不在此包**——配置形状由
// gateway 包定义（SandboxRule），解析与持有归宿主（llm/server 的 security-*）。
//
// 设计约束：
//   - **只限制写**：Low IL 进程写不了 Medium IL 对象；**读不受限**（Windows IL 语义）。
//   - **不修改任何 DACL**：只改子进程令牌 + Job。允许写目录需提前打 Low IL 标签
//     （[EnsureLowIL]），否则子进程对它也只能读不能写。
//   - **Job Object 不能限制读写**：它只表达「进程数 / 内存 / CPU / 进程树生命周期」。
//   - 非 Windows → [ErrUnsupportedPlatform]，调用方据此 fail-closed 或显式降级。
package sandbox

import "errors"

// ErrUnsupportedPlatform 表示当前平台不支持进程级写隔离。
var ErrUnsupportedPlatform = errors.New("sandbox: 当前平台不支持进程级写隔离")

// Limits 是 Job Object 的资源约束；零值 = 不限制该项。
type Limits struct {
	// MaxProcesses 限制 Job 内同时存活的进程数（0 = 不限）。
	MaxProcesses int
	// MaxMemoryBytes 限制 Job 内进程的总提交内存（0 = 不限）。
	MaxMemoryBytes uint64
}

// Spec 是一次 spawn 的沙箱参数。
type Spec struct {
	// WritableRoots 是允许写的目录（须已通过 [EnsureLowIL] 打标）。
	// 本包不据此施加限制——限制由 Low IL 承担；该字段仅用于调用方预检与审计。
	WritableRoots []string
	// Limits 是 Job 的资源约束。
	Limits Limits
}
