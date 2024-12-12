package job

import (
	"fmt"
	"go-apevolo/global"
	"go-apevolo/model"
	"go-apevolo/model/message/email"
	"go-apevolo/model/queued"
	"go-apevolo/payloads/dto"
	"go-apevolo/payloads/request"
	"go-apevolo/service"
	"go-apevolo/utils"
	emailSend "go-apevolo/utils/email"
	"go-apevolo/utils/ext"
	"go.uber.org/zap"
)

var queuedEmailService = service.ServiceGroupApp.QueuedServiceGroup.EmailQueuedService

func SendEmailJob() {

	b := false
	emailQueuedQueryCriteria := dto.EmailQueuedQueryCriteria{
		IsSend:   &b,
		MaxTries: 3,
		Pagination: request.Pagination{
			PageIndex:  1,
			PageSize:   100,
			SortFields: []string{"priority asc", "create_time asc"},
		},
	}
	list := make([]queued.Email, 0)
	var count int64
	err := queuedEmailService.Query(&emailQueuedQueryCriteria, &list, &count)
	if err != nil {
		return
	}

	if count > 0 {
		var emailAccount email.Account

		for _, item := range list {
			err = global.Db.Scopes(utils.IsDeleteSoft).First(&emailAccount, item.EmailAccountId).Error
			if err != nil {
				global.Logger.Error(fmt.Sprintf("邮箱账户 %d 出错\n", item.EmailAccountId))
				continue
			}
			localTime := ext.GetCurrentTime()
			err = emailSend.SendEmail(emailAccount.Email, emailAccount.DisplayName, emailAccount.Password, emailAccount.Host, emailAccount.Port, emailAccount.EnableSsl,
				[]string{item.To}, []string{}, []string{}, item.Subject, item.Body)
			if err == nil {
				item.SendTime = &localTime
			}
			item.SentTries += 1
			uy := "SendEmailJob"
			item.UpdateBy = &uy

			emailQueuedDto := &dto.CreateUpdateEmailQueuedDto{
				RootKey:   model.RootKey{Id: item.Id},
				SendTime:  item.SendTime,
				SentTries: item.SentTries,
				BaseModel: model.BaseModel{UpdateBy: item.UpdateBy, UpdateTime: &localTime},
			}
			err := queuedEmailService.Update(emailQueuedDto)
			if err != nil {
				global.Logger.Error(err.Error(), zap.Error(err))
			}
		}
	}
}
