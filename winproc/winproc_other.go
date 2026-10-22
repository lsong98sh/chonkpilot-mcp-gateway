//go:build !windows

// Package winproc 提供 spawn 外部进程时的进程创建属性（gateway 专用副本，2026-10-21）。
// 非 Windows 平台仅提供退化实现，保证交叉编译与测试不受影响。
package winproc

import (
	"os"
	"syscall"
)

// SysProcAttr 非 Windows 平台无「控制台窗口」概念，返回 nil（不设置进程属性）。
func SysProcAttr() *syscall.SysProcAttr {
	return nil
}

// KillTree 非 Windows 平台仅终止指定进程（无 taskkill /T 等价树杀；调用方主要用于 Windows）。
func KillTree(pid int) error {
	if pid <= 0 {
		return nil
	}
	p, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	return p.Kill()
}
