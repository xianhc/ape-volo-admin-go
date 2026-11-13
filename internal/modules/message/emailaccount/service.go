package emailaccount

import (
	"context"
	"errors"

	"github.com/xianhc/ape-volo-admin-go/internal/global"
	"github.com/xianhc/ape-volo-admin-go/internal/modules/system/setting"
	"github.com/xianhc/ape-volo-admin-go/internal/utils"
	"go.uber.org/zap"
)

// Service 实现邮件账户的业务规则。
//
// Service 对 HTTP 返回进行密码脱敏，并向内部发信流程提供 Credentials DTO。
type Service struct{ repository repository }

// Create 创建邮件账户。
//
// 方法校验邮箱唯一性，并在缺少主键时生成主键。
func (service *Service) Create(ctx context.Context, req CreateAndUpdateAccountRequest) error {
	exists, err := service.repository.ExistsByEmail(ctx, req.Email)
	if err != nil {
		return logDatabaseError(err)
	}
	if exists {
		return errors.New("邮箱账户=>" + req.Email + "=>已存在!")
	}
	account := newEntity(req.Id, req.Email, req.DisplayName, req.Host, req.Port, req.UserName, req.Password, req.EnableSsl, req.UseDefaultCredentials)
	if err := service.repository.Create(ctx, account); err != nil {
		return logDatabaseError(err)
	}
	return nil
}

// Update 修改邮件账户。
//
// Password 为空时沿用原密码，邮箱变化时重新校验唯一性。
func (service *Service) Update(ctx context.Context, req CreateAndUpdateAccountRequest) error {
	old, err := service.repository.FindByID(ctx, req.Id)
	if errors.Is(err, errNotFound) {
		return errors.New("数据不存在或您无权查看！")
	}
	if err != nil {
		return logDatabaseError(err)
	}
	if old.Email != req.Email {
		exists, err := service.repository.ExistsByEmail(ctx, req.Email)
		if err != nil {
			return logDatabaseError(err)
		}
		if exists {
			return errors.New("邮箱账户=>" + req.Email + "=>已存在!")
		}
	}
	password := req.Password
	if password == "" {
		password = old.Password
	}
	account := newEntity(req.Id, req.Email, req.DisplayName, req.Host, req.Port, req.UserName, password, req.EnableSsl, req.UseDefaultCredentials)
	if err := service.repository.Update(ctx, account); err != nil {
		return logDatabaseError(err)
	}
	return nil
}

// Delete 逻辑删除一组邮件账户。
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

// Query 查询邮件账户分页列表。
//
// 返回前统一转换为密码脱敏的 AccountResponse。
func (service *Service) Query(ctx context.Context, req QueryAccountRequest) ([]AccountResponse, int64, error) {
	accounts, total, err := service.repository.List(ctx, listOptions(req))
	if err != nil {
		return nil, 0, logDatabaseError(err)
	}
	return fromEntities(accounts), total, nil
}

// All 查询全部邮件账户。
//
// 返回结果按创建时间升序排列并完成密码脱敏。
func (service *Service) All(ctx context.Context) ([]AccountResponse, error) {
	accounts, err := service.repository.All(ctx)
	if err != nil {
		return nil, logDatabaseError(err)
	}
	return fromEntities(accounts), nil
}

// GetCredentials 按账户主键获取内部发信凭据。
//
// 主键为零时读取默认邮件账户设置；凭据只供服务端发信，不进入 HTTP 响应或日志。
func (service *Service) GetCredentials(ctx context.Context, id int64) (Credentials, error) {
	if id == 0 {
		var err error
		id, err = (&setting.Service{}).GetInt64(ctx, setting.DefaultEmailAccountId)
		if err != nil {
			return Credentials{}, err
		}
	}
	account, err := service.repository.FindByID(ctx, id)
	if err != nil {
		return Credentials{}, logDatabaseError(err)
	}
	return Credentials{Id: account.Id, Email: account.Email, DisplayName: account.DisplayName, Password: account.Password, Host: account.Host, Port: account.Port, EnableSSL: account.EnableSsl}, nil
}

func newEntity(id int64, email, displayName, host string, port int32, username, password string, enableSSL, useDefault bool) *Account {
	account := &Account{Email: email, DisplayName: displayName, Host: host, Port: port, UserName: username, Password: password, EnableSsl: enableSSL, UseDefaultCredentials: useDefault}
	if id == 0 {
		id = int64(utils.GenerateID())
	}
	account.Id = id
	return account
}
func listOptions(req QueryAccountRequest) queryOptions {
	return queryOptions{Email: req.Email, DisplayName: req.DisplayName, UserName: req.UserName, CreateTime: req.CreateTime, Sorting: req.Sorting, Limit: req.PageSize, Offset: req.PageSize * (req.PageIndex - 1)}
}
func logDatabaseError(err error) error { global.Logger.Error("db error: ", zap.Error(err)); return err }
