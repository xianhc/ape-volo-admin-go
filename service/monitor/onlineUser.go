package monitor

import (
	"context"
	"go-apevolo/global"
	"go-apevolo/global/constants/cachePrefix"
	"go-apevolo/model/system"
	"go-apevolo/payloads/auth"
	"go-apevolo/payloads/request"
	"go-apevolo/utils"
	"go-apevolo/utils/redis"
)

type OnlineUserService struct{}

// Query
// @description: 查询在线用户
// @receiver: onlineUserService
// @param: pagination
// @return: list
// @return: total
// @return: err
func (onlineUserService *OnlineUserService) Query(pagination request.Pagination) (list interface{}, total int64, err error) {
	// 使用Keys方法进行模糊匹配
	keys, err := global.Redis.Keys(context.Background(), cachePrefix.OnlineKey+"*").Result()
	if err != nil {
		return
	}
	var loginUserInfo auth.LoginUserInfo
	var loginUserInfos []auth.LoginUserInfo
	// 处理匹配到的keys
	for _, key := range keys {
		err = redis.Get(key, &loginUserInfo)
		if err != nil {
			return
		}
		loginUserInfo.AccessToken = utils.MD5(loginUserInfo.AccessToken)
		loginUserInfos = append(loginUserInfos, loginUserInfo)
	}
	if len(loginUserInfos) > 0 {
		// 计算要跳过的元素数量
		elementsToSkip := pagination.PageSize * (pagination.PageIndex - 1)

		// 使用切片实现 Skip 分页
		slicedData := loginUserInfos[elementsToSkip:]
		return slicedData, int64(len(keys)), err
	}
	return loginUserInfos, int64(len(keys)), err
}

// DropOut
// @description: 用户登出
// @receiver: onlineUserService
// @param: idArray
// @return: err
func (onlineUserService *OnlineUserService) DropOut(idArray request.IdCollection) (err error) {

	var tokenBlacklists []system.TokenBlacklist
	for _, i := range idArray.IdArray {
		black := system.TokenBlacklist{AccessToken: i}
		tokenBlacklists = append(tokenBlacklists, black)
	}

	err = global.Db.Create(&tokenBlacklists).Error
	if err == nil {
		for _, key := range idArray.IdArray {
			_, _ = global.Redis.Del(context.Background(), cachePrefix.OnlineKey+key).Result()
		}
	}
	return err
}
