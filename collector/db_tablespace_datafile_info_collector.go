package collector

import (
	"context"
	"dameng_exporter/config"
	"dameng_exporter/logger"
	"database/sql"
	"encoding/json"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
	"time"
)

type TableSpaceDateFileInfoCollector struct {
	db        *sql.DB
	totalDesc *prometheus.Desc
	freeDesc  *prometheus.Desc
}

type TableSpaceDateFileInfo struct {
	Path       string
	TotalSize  float64
	FreeSize   float64
	AutoExtend string
	NextSize   string
	MaxSize    string
}

func NewTableSpaceDateFileInfoCollector(db *sql.DB) MetricCollector {
	return &TableSpaceDateFileInfoCollector{
		db: db,
		totalDesc: prometheus.NewDesc(
			dmdbms_tablespace_file_total_info,
			"Tablespace file information",
			[]string{"host_name", "tablespace_name", "auto_extend", "next_size", "max_size"}, // 添加标签
			nil,
		),
		freeDesc: prometheus.NewDesc(
			dmdbms_tablespace_file_free_info,
			"Tablespace file information",
			[]string{"host_name", "tablespace_name", "auto_extend", "next_size", "max_size"}, // 添加标签
			nil,
		),
	}
}

func (c *TableSpaceDateFileInfoCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.totalDesc
	ch <- c.freeDesc
}

func (c *TableSpaceDateFileInfoCollector) Collect(ch chan<- prometheus.Metric) {
	funcStart := time.Now()
	// 时间间隔的计算发生在 defer 语句执行时，确保能够获取到正确的函数执行时间。
	defer func() {
		duration := time.Since(funcStart)
		logger.Logger.Debugf("func exec time：%vms", duration.Milliseconds())
	}()

	//保存全局结果对象，可以用来做缓存以及序列化
	var tablespaceInfos []TableSpaceDateFileInfo

	// 从缓存中获取数据
	if cachedJSON, found := config.GetFromCache(dmdbms_tablespace_file_total_info); found {
		// 将缓存中的 JSON 字符串转换为 TablespaceInfo 切片
		if err := json.Unmarshal([]byte(cachedJSON), &tablespaceInfos); err != nil {
			// 处理反序列化错误
			logger.Logger.Error("Error unmarshaling cached data", zap.Error(err))
			// 反序列化失败，忽略缓存中的数据，继续查询数据库
			cachedJSON = "" // 清空缓存数据，确保后续不使用过期或损坏的数据
		} else {
			logger.Logger.Infof("Use cache TablespaceDateFile data")
			// 使用缓存的数据
			for _, info := range tablespaceInfos {
				ch <- prometheus.MustNewConstMetric(c.totalDesc, prometheus.GaugeValue, info.TotalSize, config.GetHostName(), info.Path, info.AutoExtend, info.NextSize, info.MaxSize)
				ch <- prometheus.MustNewConstMetric(c.freeDesc, prometheus.GaugeValue, info.FreeSize, config.GetHostName(), info.Path, info.AutoExtend, info.NextSize, info.MaxSize)
			}
			return
		}
	}

	if err := c.db.Ping(); err != nil {
		logger.Logger.Error("Database connection is not available: %v", zap.Error(err))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(config.GlobalConfig.QueryTimeout)*time.Second)
	defer cancel()

	rows, err := c.db.QueryContext(ctx, config.QueryTablespaceFileSqlStr)
	if err != nil {
		handleDbQueryError(err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var info TableSpaceDateFileInfo
		if err := rows.Scan(&info.Path, &info.TotalSize, &info.FreeSize, &info.AutoExtend, &info.NextSize, &info.MaxSize); err != nil {
			logger.Logger.Error("Error scanning row", zap.Error(err))
			continue
		}
		tablespaceInfos = append(tablespaceInfos, info)
	}
	if err := rows.Err(); err != nil {
		logger.Logger.Error("Error with rows", zap.Error(err))
	}
	// 发送数据到 Prometheus
	for _, info := range tablespaceInfos {
		ch <- prometheus.MustNewConstMetric(c.totalDesc, prometheus.GaugeValue, info.TotalSize, config.GetHostName(), info.Path, info.AutoExtend, info.NextSize, info.MaxSize)
		ch <- prometheus.MustNewConstMetric(c.freeDesc, prometheus.GaugeValue, info.FreeSize, config.GetHostName(), info.Path, info.AutoExtend, info.NextSize, info.MaxSize)
	}

	// 将 TablespaceInfo 切片序列化为 JSON 字符串
	valueJSON, err := json.Marshal(tablespaceInfos)
	if err != nil {
		// 处理序列化错误
		logger.Logger.Error("TablespaceInfo ", zap.Error(err))
		return
	}
	// 将查询结果存入缓存
	config.SetCache(dmdbms_tablespace_file_total_info, string(valueJSON), time.Minute*time.Duration(config.GlobalConfig.AlarmKeyCacheTime)) // 设置缓存有效时间为5分钟
	logger.Logger.Infof("TablespaceFileInfoCollector exec finish")

}
