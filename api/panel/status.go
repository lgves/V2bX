package panel

import (
	"context"

	"github.com/sirupsen/logrus"
)

// NodeStatusPair 描述一项资源（内存 / Swap / 磁盘）的总量与已用量，单位字节。
type NodeStatusPair struct {
	Total uint64 `json:"total"`
	Used  uint64 `json:"used"`
}

// NodeStatus 是上报给面板的服务器负载快照，CPU 为百分比（0~100）。
type NodeStatus struct {
	CPU  float64        `json:"cpu"`
	Mem  NodeStatusPair `json:"mem"`
	Swap NodeStatusPair `json:"swap"`
	Disk NodeStatusPair `json:"disk"`
}

// ReportNodeStatus reports the server load to the panel.
func (c *Client) ReportNodeStatus(status *NodeStatus) error {
	return c.ReportNodeStatusCtx(context.Background(), status)
}

// ReportNodeStatusCtx is the ctx-aware variant used by the task framework so a
// watchdog timeout can cancel the in-flight HTTP request. W3.2 / W3.4.
//
// 老面板可能没有实现 /status 端点。这里对 HTTP >= 400 的响应只记一条 Debug
// 日志并返回 nil（视为“面板不支持该功能”），避免每个上报周期都刷错误日志；
// 仅网络 / 传输层失败才真正返回 error，交由任务框架处理。
func (c *Client) ReportNodeStatusCtx(ctx context.Context, status *NodeStatus) error {
	var path string
	if c.ApiVersion == 2 {
		path = "/api/v2/server/status"
	} else {
		path = "/api/v1/server/UniProxy/status"
	}
	r, err := c.client.R().
		SetContext(ctx).
		SetBody(status).
		ForceContentType("application/json").
		Post(path)
	if err != nil {
		return c.checkResponse(r, path, err)
	}
	if r != nil && r.StatusCode() >= 400 {
		logrus.WithFields(logrus.Fields{
			"url":    c.assembleURL(path),
			"status": r.StatusCode(),
		}).Debug("report node status: endpoint may not be supported by panel, skipping")
		return nil
	}
	return c.checkResponse(r, path, nil)
}
