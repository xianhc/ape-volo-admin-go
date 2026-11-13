package job

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/xianhc/ape-volo-admin-go/internal/model"
	"github.com/xianhc/ape-volo-admin-go/internal/modules/message/emailaccount"
	"github.com/xianhc/ape-volo-admin-go/internal/modules/message/queuedemail"
	emailsend "github.com/xianhc/ape-volo-admin-go/internal/utils/email"
)

var queuedEmailService = queuedemail.Service{}
var emailDeliveryMu sync.Mutex

// SendEmailJob 串行处理待发邮件，同进程内的多个调度配置共用互斥锁。
func SendEmailJob(ctx context.Context) error {
	if !emailDeliveryMu.TryLock() {
		return nil
	}
	defer emailDeliveryMu.Unlock()
	unsent := false
	query := queuedemail.QueryQueuedEmailRequest{IsSend: &unsent, MaxTries: 3}
	query.PageIndex, query.PageSize, query.Sorting = 1, 100, "priority asc, creation_time asc"
	list, _, err := queuedEmailService.Pending(ctx, query)
	if err != nil {
		return err
	}
	var failures []error
	for _, item := range list {
		if err := ctx.Err(); err != nil {
			return errors.Join(append(failures, err)...)
		}
		account, sendErr := (&emailaccount.Service{}).GetCredentials(ctx, item.EmailAccountId)
		if sendErr == nil {
			sendErr = emailsend.SendEmail(ctx, account.Email, account.DisplayName, account.Password, account.Host, account.Port, account.EnableSSL,
				[]string{item.To}, nil, nil, item.Subject, item.Body)
		}
		if sendErr == nil {
			now := model.Now()
			item.IsSend = true
			item.SendTime = &now
		} else {
			item.IsSend = false
		}
		item.SentTries++
		// 发送完成后即使任务已取消，也尝试记录结果，减少下一次重复发送。
		writeCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		updateErr := queuedEmailService.UpdateDelivery(writeCtx, item.Id, item.IsSend, item.SendTime, item.SentTries)
		cancel()
		if err := errors.Join(sendErr, updateErr); err != nil {
			failures = append(failures, fmt.Errorf("邮件 %d 处理失败: %w", item.Id, err))
		}
	}
	return errors.Join(failures...)
}
