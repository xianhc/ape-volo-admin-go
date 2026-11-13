package emailtemplate

import (
	"context"
	"errors"

	"github.com/xianhc/ape-volo-admin-go/internal/global"
	"github.com/xianhc/ape-volo-admin-go/internal/model"
	"github.com/xianhc/ape-volo-admin-go/internal/utils"
	"go.uber.org/zap"
)

// Service 实现邮件模板的业务规则。
//
// Service 向 HTTP 层返回独立 Response，并向队列业务提供 TemplateContent DTO。
type Service struct{ repository repository }

// Create 创建邮件模板。
//
// 方法校验模板名称唯一性，并在缺少主键时生成主键。
func (service *Service) Create(ctx context.Context, req CreateAndUpdateTemplateRequest) error {
	exists, err := service.repository.ExistsByName(ctx, req.Name)
	if err != nil {
		return logDatabaseError(err)
	}
	if exists {
		return errors.New("模板名称=>" + req.Name + "=>已存在!")
	}
	template := newEntity(req.Id, req.Name, req.BccEmailAddresses, req.Subject, req.Body, req.Enabled, req.EmailAccountId)
	if err := service.repository.Create(ctx, template); err != nil {
		return logDatabaseError(err)
	}
	return nil
}

// Update 修改邮件模板。
//
// 方法先确认记录存在，再在名称变化时校验唯一性。
func (service *Service) Update(ctx context.Context, req CreateAndUpdateTemplateRequest) error {
	old, err := service.repository.FindByID(ctx, req.Id)
	if errors.Is(err, errNotFound) {
		return errors.New("数据不存在或您无权查看！")
	}
	if err != nil {
		return logDatabaseError(err)
	}
	if old.Name != req.Name {
		exists, err := service.repository.ExistsByName(ctx, req.Name)
		if err != nil {
			return logDatabaseError(err)
		}
		if exists {
			return errors.New("模板名称=>" + req.Name + "=>已存在!")
		}
	}
	template := newEntity(req.Id, req.Name, req.BccEmailAddresses, req.Subject, req.Body, req.Enabled, req.EmailAccountId)
	if err := service.repository.Update(ctx, template); err != nil {
		return logDatabaseError(err)
	}
	return nil
}

// Delete 逻辑删除一组邮件模板。
//
// 没有记录受影响时返回统一不可见错误。
func (service *Service) Delete(ctx context.Context, ids []string) error {
	rows, err := service.repository.Delete(ctx, ids)
	if err != nil {
		return logDatabaseError(err)
	}
	if rows == 0 {
		return errors.New("数据不存在或您无权查看！")
	}
	return nil
}

// Query 查询邮件模板分页列表。
//
// 返回前统一把 Entity 转换为 TemplateResponse。
func (service *Service) Query(ctx context.Context, req QueryTemplateRequest) ([]TemplateResponse, int64, error) {
	templates, total, err := service.repository.List(ctx, queryOptions{Name: req.Name, Enabled: req.Enabled, CreateTime: req.CreateTime, Sorting: req.Sorting, Limit: req.PageSize, Offset: req.PageSize * (req.PageIndex - 1)})
	if err != nil {
		return nil, 0, logDatabaseError(err)
	}
	return fromEntities(templates), total, nil
}

// GetByName 按名称获取内部模板内容。
//
// 返回 DTO 供邮件队列组装消息，不暴露数据库 Entity。
func (service *Service) GetByName(ctx context.Context, name string) (TemplateContent, error) {
	template, err := service.repository.FindByName(ctx, name)
	if errors.Is(err, errNotFound) {
		return TemplateContent{}, errors.New("数据不存在或您无权查看！")
	}
	if err != nil {
		return TemplateContent{}, logDatabaseError(err)
	}
	return TemplateContent{BccEmailAddresses: template.BccEmailAddresses, Subject: template.Subject, Body: string(template.Body), EmailAccountId: template.EmailAccountId}, nil
}

func newEntity(id int64, name, bcc, subject, body string, active bool, accountId int64) *MessageTemplate {
	template := &MessageTemplate{Name: name, BccEmailAddresses: bcc, Subject: subject, Body: model.CustomFieldText(body), Enabled: active, EmailAccountId: accountId}
	if id == 0 {
		id = int64(utils.GenerateID())
	}
	template.Id = id
	return template
}
func logDatabaseError(err error) error { global.Logger.Error("db error: ", zap.Error(err)); return err }
