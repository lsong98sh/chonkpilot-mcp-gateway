//go:build windows

// Package winproc 提供 spawn 外部进程时的进程创建属性（gateway 专用副本，2026-10-21）。
//
// 来源：src/lib/core/winproc。**只保留 gateway 实际用到的两个 API**（SysProcAttr / KillTree），
// 使 gateway 不依赖 core（可独立成仓）。core/winproc 仍是宿主侧（mcp-server / plugins / gui）
// 的公共实现——两份代码修改时须同步（语义相同，仅本副本裁剪了 Job 相关部分）。
//
// 目的：console 子系统 exe（下游 MCP server 等）被 GUI 或服务进程 spawn 时，
// 抑制瞬时/可见的控制台窗口（黑窗），且不影响 stdin/stdout/stderr 管道通信。
package winproc

import (
	"os/exec"
	"strconv"
	"syscall"
)

// createNoWindow 为 Windows CREATE_NO_WINDOW（0x08000000）：
// 控制台进程以「无控制台窗口」方式创建，从根源避免控制台被创建/闪现。
const createNoWindow = 0x08000000

// SysProcAttr 返回隐藏子进程控制台窗口的进程创建属性：
// HideWindow（STARTF_USESHOWWINDOW + SW_HIDE）抑制窗口显示，
// CREATE_NO_WINDOW 避免控制台本身被创建——两者并用，覆盖瞬时闪窗。
// 非控制台（GUI 子系统）子进程不受影响；stdin/stdout/stderr 管道通信保持不变。
func SysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: createNoWindow,
	}
}

// KillTree 终止指定进程及其**全部子孙进程**（Windows：`taskkill /T /F`）。
//
// 用途：取消/超时强杀执行器（直接子进程）时，执行器自行拉起的孙进程
// （如 script 执行器 → cmd.exe → ping.exe）会孤儿化；仅 TerminateProcess 直接子进程不够，
// 必须按进程树终止（无孤儿判据）。
func KillTree(pid int) error {
	if pid <= 0 {
		return nil
	}
	c := exec.Command("taskkill", "/T", "/F", "/PID", strconv.Itoa(pid))
	c.SysProcAttr = SysProcAttr()
	return c.Run()
}
