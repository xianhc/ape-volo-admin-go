package initialize

import (
	"context"
	"time"

	"github.com/xianhc/ape-volo-admin-go/internal/global"
	"github.com/xianhc/ape-volo-admin-go/internal/logstore"
	"github.com/xianhc/ape-volo-admin-go/internal/modules/log/exceptionlog"
	"github.com/xianhc/ape-volo-admin-go/internal/modules/log/operatelog"
	"github.com/xianhc/ape-volo-admin-go/internal/modules/log/scheduledjoblog"
)

// InitLogTables 装配三类日志模型并维护日志连接上的月表。
//
// 成功后才发布存储，避免生产者读取未完成初始化的元数据。
func InitLogTables(ctx context.Context) error {
	loc, err := time.LoadLocation(global.Config.LogStorage.TimeZone)
	if err != nil {
		return err
	}
	s, err := logstore.New(global.LogDb, loc, global.Config.System.IsInitTable,
		logstore.Schema{Prefix: logstore.Operate, Model: &operatelog.OperateLog{}, Indexes: [][]string{{"creation_time", "id"}, {"creator_id", "creation_time", "id"}}},
		logstore.Schema{Prefix: logstore.Exception, Model: &exceptionlog.ExceptionLog{}, Indexes: [][]string{{"creation_time", "id"}}},
		logstore.Schema{Prefix: logstore.Job, Model: &scheduledjoblog.ScheduledJobLog{}, Indexes: [][]string{{"creation_time", "id"}, {"job_group", "creation_time", "id"}, {"is_success", "creation_time", "id"}}},
	)
	if err != nil {
		return err
	}
	if err := s.Initialize(ctx); err != nil {
		return err
	}
	logstore.SetDefault(s)
	return nil
}
