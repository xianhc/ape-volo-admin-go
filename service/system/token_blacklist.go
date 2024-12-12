package system

import (
	"go-apevolo/global"
	"go-apevolo/model/system"
	"go-apevolo/utils"
	"go.uber.org/zap"
)

type TokenBlacklistService struct{}

// DoesItExist
// @description: 查询
// @receiver: tokenBlacklistService
// @param: tokenMd5
// @return: isExist
// @return: err
func (tokenBlacklistService *TokenBlacklistService) DoesItExist(tokenMd5 string) (isExist bool, err error) {
	var total int64
	err = global.Db.Model(&system.TokenBlacklist{}).Scopes(utils.IsDeleteSoft).Where("access_token = ?", tokenMd5).Count(&total).Error
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		return false, err
	}
	if total > 0 {
		return true, nil
	}
	return false, err
}
