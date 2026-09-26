// Package monitor 采集节点所在服务器的负载指标（CPU / 内存 / Swap / 磁盘），
// 供 node 控制器周期性上报给面板。采集全部走 gopsutil，跨平台；任一指标
// 采集失败都只降级为零值并记 Debug 日志，绝不让整个上报流程失败。
package monitor

import (
	"runtime"

	log "github.com/sirupsen/logrus"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/mem"
)

// Status 是一次采集到的服务器负载快照。字段单位：
// CPU 为百分比（0~100），其余 Total/Used 均为字节。
type Status struct {
	CPU       float64
	MemTotal  uint64
	MemUsed   uint64
	SwapTotal uint64
	SwapUsed  uint64
	DiskTotal uint64
	DiskUsed  uint64
}

// diskMountPoint 返回统计磁盘用量的挂载点。Linux/类 Unix 用根分区 "/"，
// Windows 用系统盘 "C:"（gopsutil 在 Windows 上对 "/" 会直接报错）。
func diskMountPoint() string {
	if runtime.GOOS == "windows" {
		return "C:"
	}
	return "/"
}

// Collect 采集当前服务器负载。任何单项失败都只记 Debug 日志并保留零值，
// 保证调用方永远拿到一个可用的快照。
//
// CPU 用 cpu.Percent(0, ...)：它返回「距上一次 Percent(0) 调用之间」的平均占用率。
// gopsutil 的 cpu 包在自身 init 里已种下首个基线，因此无需额外预热；
// 由于 Collect 按上报周期被调用，每次读数恰好是该周期内的平均 CPU。
func Collect() Status {
	var s Status

	if percents, err := cpu.Percent(0, false); err != nil {
		log.WithField("err", err).Debug("monitor: get cpu usage failed")
	} else if len(percents) > 0 {
		s.CPU = percents[0]
	}

	if vm, err := mem.VirtualMemory(); err != nil {
		log.WithField("err", err).Debug("monitor: get virtual memory failed")
	} else {
		s.MemTotal = vm.Total
		s.MemUsed = vm.Used
	}

	if swap, err := mem.SwapMemory(); err != nil {
		log.WithField("err", err).Debug("monitor: get swap memory failed")
	} else {
		s.SwapTotal = swap.Total
		s.SwapUsed = swap.Used
	}

	if usage, err := disk.Usage(diskMountPoint()); err != nil {
		log.WithField("err", err).Debug("monitor: get disk usage failed")
	} else {
		s.DiskTotal = usage.Total
		s.DiskUsed = usage.Used
	}

	return s
}
