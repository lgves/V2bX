package node

import (
	"context"
	"time"

	"github.com/InazumaV/V2bX/api/panel"
	"github.com/InazumaV/V2bX/common/monitor"
	log "github.com/sirupsen/logrus"
)

// defaultMonitorInterval 是未在配置里指定 MonitorInterval 时的服务器负载上报间隔。
const defaultMonitorInterval = 60 * time.Second

// reportNodeStatusTask 采集本机负载并上报给面板。采集永不失败（单项降级为零值），
// 上报失败只记日志、不返回 error —— 与 reportUserTrafficTask 一致，避免单个
// 周期的面板抖动把整个任务循环拖垮。
func (c *Controller) reportNodeStatusTask(ctx context.Context) error {
	s := monitor.Collect()
	status := &panel.NodeStatus{
		CPU:  s.CPU,
		Mem:  panel.NodeStatusPair{Total: s.MemTotal, Used: s.MemUsed},
		Swap: panel.NodeStatusPair{Total: s.SwapTotal, Used: s.SwapUsed},
		Disk: panel.NodeStatusPair{Total: s.DiskTotal, Used: s.DiskUsed},
	}
	if err := c.getAPIClient().ReportNodeStatusCtx(ctx, status); err != nil {
		log.WithFields(log.Fields{
			"tag": c.tag,
			"err": err,
		}).Info("Report node status failed")
		return nil
	}
	log.WithField("tag", c.tag).Debugf("Report node status: %+v", status)
	return nil
}
