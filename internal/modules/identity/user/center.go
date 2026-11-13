package user

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"strconv"

	"github.com/xianhc/ape-volo-admin-go/internal/global"
	"github.com/xianhc/ape-volo-admin-go/internal/global/constants/cachePrefix"
	"github.com/xianhc/ape-volo-admin-go/internal/utils"
	"github.com/xianhc/ape-volo-admin-go/internal/utils/cache"
	"github.com/xianhc/ape-volo-admin-go/internal/utils/ext"
	"github.com/xianhc/ape-volo-admin-go/internal/utils/upload"
	"go.uber.org/zap"
)

// UpdateCenter 修改当前用户昵称、性别和电话。
//
// 用户主键来自认证上下文，唯一性校验排除当前用户。
func (service *Service) UpdateCenter(ctx context.Context, userId int64, req UpdateCenterRequest) error {
	return utils.InTransaction(ctx, func(ctx context.Context) error { return service.updateCenter(ctx, userId, req) })
}

func (service *Service) updateCenter(ctx context.Context, userId int64, req UpdateCenterRequest) error {
	user, err := service.repository.FindByID(ctx, userId, false, true)
	if err != nil {
		return userNotFound(err)
	}
	if req.NickName != user.NickName {
		if total, err := service.repository.CountByField(ctx, "nick_name", req.NickName, userId); err != nil {
			return err
		} else if total > 0 {
			return errors.New("昵称=>" + req.NickName + "=>已存在!")
		}
	}
	if req.Phone != "" && req.Phone != user.Phone {
		if total, err := service.repository.CountByField(ctx, "phone", req.Phone, userId); err != nil {
			return err
		} else if total > 0 {
			return errors.New("电话=>" + req.Phone + "=>已存在!")
		}
	}
	return service.repository.UpdateFields(ctx, userId, map[string]any{"nick_name": req.NickName, "gender_code": req.GenderCode, "phone": req.Phone})
}

// UpdatePassword 校验旧密码并修改当前用户密码。
//
// 修改成功后撤销该用户全部会话。
func (service *Service) UpdatePassword(ctx context.Context, userId int64, req UpdatePasswordRequest) error {
	return utils.InTransaction(ctx, func(ctx context.Context) error { return service.updatePassword(ctx, userId, req) })
}

func (service *Service) updatePassword(ctx context.Context, userId int64, req UpdatePasswordRequest) error {
	oldPlain, err := utils.Decrypt(req.OldPassword)
	if err != nil {
		return errors.New("旧密码解密失败:" + err.Error())
	}
	newPlain, err := utils.Decrypt(req.NewPassword)
	if err != nil {
		return errors.New("新密码解密失败:" + err.Error())
	}
	confirm, err := utils.Decrypt(req.ConfirmPassword)
	if err != nil {
		return errors.New("确认密码解密失败:" + err.Error())
	}
	if newPlain != confirm {
		return errors.New("两次输入的新密码不一致")
	}
	if oldPlain == newPlain {
		return errors.New("新密码不能与旧密码相同")
	}
	if len(newPlain) < 6 {
		return errors.New("新密码长度不能少于 6 位")
	}
	user, err := service.repository.FindByID(ctx, userId, false, true)
	if err != nil {
		return userNotFound(err)
	}
	if !utils.BcryptCheck(oldPlain, user.Password) {
		return errors.New("旧密码错误")
	}
	now := ext.GetCurrentTime()
	if err = service.repository.UpdateFields(ctx, userId, map[string]any{"password": utils.BcryptHash(newPlain), "password_re_set_time": now}); err != nil {
		return logDatabaseError(err)
	}
	service.revoke(ctx, userId, "改密后撤销会话失败")
	return nil
}

// UpdateEmail 校验密码与验证码后修改当前用户邮箱。
//
// 验证码成功使用后立即从缓存删除。
func (service *Service) UpdateEmail(ctx context.Context, userId int64, req UpdateEmailRequest) error {
	return utils.InTransaction(ctx, func(ctx context.Context) error { return service.updateEmail(ctx, userId, req) })
}

func (service *Service) updateEmail(ctx context.Context, userId int64, req UpdateEmailRequest) error {
	user, err := service.repository.FindByID(ctx, userId, false, true)
	if err != nil {
		return userNotFound(err)
	}
	plain, err := utils.Decrypt(req.Password)
	if err != nil {
		return errors.New("密码解密失败:" + err.Error())
	}
	if !utils.BcryptCheck(plain, user.Password) {
		return errors.New("密码错误")
	}
	var code int
	key := cachePrefix.EmailCaptcha + req.Email
	if err = global.Cache.Get(ctx, key, &code); err != nil {
		if errors.Is(err, cache.ErrNotFound) {
			return errors.New("验证码已失效，请重新获取")
		}
		return err
	}
	if strconv.Itoa(code) != req.Code {
		return errors.New("验证码错误")
	}
	if total, err := service.repository.CountByField(ctx, "email", req.Email, userId); err != nil {
		return err
	} else if total > 0 {
		return errors.New("邮箱=>" + req.Email + "=>已存在!")
	}
	if err = service.repository.UpdateFields(ctx, userId, map[string]any{"email": req.Email}); err != nil {
		return logDatabaseError(err)
	}
	utils.AfterCommit(ctx, "email-code:"+key, func(ctx context.Context) {
		if err := global.Cache.Del(ctx, key); err != nil {
			global.Logger.Warn("清理已使用的邮箱验证码失败", zap.Error(err))
		}
	})
	return nil
}

// UpdatePositions 替换请求指定用户的岗位关联。
//
// 写操作复用请求事务。
func (service *Service) UpdatePositions(ctx context.Context, req UpdatePositionsRequest) error {
	return utils.InTransaction(ctx, func(ctx context.Context) error { return service.updatePositions(ctx, req) })
}

func (service *Service) updatePositions(ctx context.Context, req UpdatePositionsRequest) error {
	if _, err := service.repository.FindByID(ctx, req.Id, false, false); err != nil {
		return userNotFound(err)
	}
	return service.repository.ReplacePositions(ctx, req.Id, []int64(req.PositionIdArray))
}

// UpdateRoles 替换请求指定用户的角色关联。
//
// 写操作复用请求事务并撤销旧会话权限快照。
func (service *Service) UpdateRoles(ctx context.Context, req UpdateRolesRequest) error {
	return utils.InTransaction(ctx, func(ctx context.Context) error { return service.updateRoles(ctx, req) })
}

func (service *Service) updateRoles(ctx context.Context, req UpdateRolesRequest) error {
	if err := requireAdmin(ctx); err != nil {
		return err
	}
	if _, err := service.repository.FindByID(ctx, req.Id, false, false); err != nil {
		return userNotFound(err)
	}
	if err := service.repository.ReplaceRoles(ctx, req.Id, []int64(req.RoleIdArray)); err != nil {
		return err
	}
	service.revoke(ctx, req.Id, "修改角色后撤销会话失败")
	return nil
}

// UpdatePreferences 修改当前用户偏好配置。
//
// 配置按既有 JSON 字符串原样保存。
func (service *Service) UpdatePreferences(ctx context.Context, userId int64, req UpdatePreferencesRequest) error {
	return utils.InTransaction(ctx, func(ctx context.Context) error { return service.updatePreferences(ctx, userId, req) })
}

func (service *Service) updatePreferences(ctx context.Context, userId int64, req UpdatePreferencesRequest) error {
	if _, err := service.repository.FindByID(ctx, userId, false, true); err != nil {
		return userNotFound(err)
	}
	return service.repository.UpdateFields(ctx, userId, map[string]any{"preferences_config": req.PreferencesConfig})
}

// UpdateAvatar 保存头像并更新当前用户头像路径。
//
// 文件大小受 system.file-limit-size 配置限制。
func (service *Service) UpdateAvatar(ctx context.Context, userId int64, file *multipart.FileHeader) (string, error) {
	var path string
	err := utils.InTransaction(ctx, func(ctx context.Context) error {
		var err error
		path, err = service.updateAvatar(ctx, userId, file)
		return err
	})
	return path, err
}

func (service *Service) updateAvatar(ctx context.Context, userId int64, file *multipart.FileHeader) (string, error) {
	if _, err := service.repository.FindByID(ctx, userId, false, true); err != nil {
		return "", userNotFound(err)
	}
	if limit := global.Config.System.FileLimitSize; limit > 0 && file.Size > int64(limit)*1024*1024 {
		return "", fmt.Errorf("头像过大，请选择小于等于 %dMB 的图片", limit)
	}
	path, err := upload.UploadAvatar(file)
	if err != nil {
		return "", err
	}
	utils.AfterRollback(ctx, func(context.Context) {
		if err := (&upload.Local{}).DeleteFile(path); err != nil {
			global.Logger.Warn("清理未提交的头像失败", zap.Error(err))
		}
	})
	if err = service.repository.UpdateFields(ctx, userId, map[string]any{"avatar_path": path}); err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		return "", err
	}
	return path, nil
}

func userNotFound(err error) error {
	if errors.Is(err, errNotFound) {
		return errors.New("数据不存在或您无权查看！")
	}
	return err
}
