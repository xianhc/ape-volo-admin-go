package onlineuser

import (
	"context"
	"strconv"

	"github.com/xianhc/ape-volo-admin-go/internal/global"
	"github.com/xianhc/ape-volo-admin-go/internal/model"
	sessionmodule "github.com/xianhc/ape-volo-admin-go/internal/modules/authorization/session"
	"github.com/xianhc/ape-volo-admin-go/internal/utils"
	"github.com/xianhc/ape-volo-admin-go/internal/utils/excel"
)

// Service 实现在线会话列表、强制下线和导出业务。
//
// 在线状态完全来自 AuthSession 缓存，不访问数据库。
type Service struct{ sessions sessionmodule.Service }

// Query 查询在线用户分页列表。
//
// 损坏的会话登录信息会被跳过，不影响其余会话。
func (service *Service) Query(ctx context.Context, req QueryRequest) ([]OnlineUserResponse, int64, error) {
	sessions, err := service.sessions.ListActive(context.Background())
	if err != nil {
		return nil, 0, err
	}
	items := make([]OnlineUserResponse, 0, len(sessions))
	for _, session := range sessions {
		info, infoErr := session.LoginUserInfo()
		if infoErr != nil {
			continue
		}
		items = append(items, fromLoginInfo(*info, session.Id, session.CreatedAt))
	}
	total := int64(len(items))
	page, size := req.PageIndex, req.PageSize
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 10
	}
	offset := size * (page - 1)
	if offset >= len(items) {
		return []OnlineUserResponse{}, total, nil
	}
	end := offset + size
	if end > len(items) {
		end = len(items)
	}
	return items[offset:end], total, nil
}

// DropOut 撤销指定会话集合。
//
// 任一撤销失败时立即返回错误并记录会话标识。
func (service *Service) DropOut(ctx context.Context, ids []string) error {
	for _, id := range ids {
		if err := service.sessions.RevokeSession(ctx, id); err != nil {
			global.Logger.Error("强制下线时撤销会话失败: " + id)
			return err
		}
	}
	return nil
}

// Download 导出全部在线用户。
//
// 导出最多读取十万条会话并保持既有列顺序。
func (service *Service) Download(ctx context.Context) (filePath, fileName string, err error) {
	req := QueryRequest{}
	req.PageIndex, req.PageSize = 1, 100000
	items, _, err := service.Query(ctx, req)
	if err != nil {
		return "", "", err
	}
	file := excel.NewFile("OnlineUsers")
	defer func() { _ = file.Close() }()
	file.SetHeader([]string{"用户ID", "账号", "昵称", "部门", "登录IP", "登录地点", "操作系统", "浏览器", "版本", "设备类型", "登录时间"})
	for _, item := range items {
		file.AddRow(strconv.FormatInt(item.UserId, 10), item.Account, item.NickName, item.DeptName, item.Ip, item.Address, item.OperatingSystem, item.BrowserName, item.Version, item.DeviceType, item.LoginTime.Format(model.TimeLayout))
	}
	fileName = "OnlineUsers_" + utils.GenerateID().String() + ".xlsx"
	filePath = global.Config.Excel.Dir + fileName
	return filePath, fileName, file.Save(filePath)
}
