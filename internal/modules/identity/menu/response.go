package menu

import (
	"github.com/xianhc/ape-volo-admin-go/internal/model"
)

// MenuResponse 表示菜单接口的返回模型。
//
// Children 递归使用相同 Response，数据库实体不会进入 HTTP 输出。
type MenuResponse struct {
	Id                   int64          `json:"id,string"`
	Title                string         `json:"title"`
	Path                 string         `json:"path"`
	AuthCode             string         `json:"authCode"`
	Component            string         `json:"component,omitempty"`
	ComponentName        string         `json:"componentName,omitempty"`
	ParentId             *int64         `json:"parentId,string"`
	Sort                 int            `json:"sort"`
	Icon                 string         `json:"icon"`
	MenuType             int            `json:"menuType"`
	Enabled              bool           `json:"enabled"`
	KeepAlive            bool           `json:"keepAlive"`
	Hidden               bool           `json:"hidden"`
	Children             []MenuResponse `json:"children,omitempty"`
	HasChildren          bool           `json:"hasChildren"`
	Leaf                 bool           `json:"leaf"`
	Label                string         `json:"label"`
	CreatorId            *int64         `json:"creatorId,omitempty,string"`
	CreationTime         model.Time     `json:"creationTime"`
	LastModifierId       *int64         `json:"lastModifierId,omitempty,string"`
	LastModificationTime *model.Time    `json:"lastModificationTime,omitempty"`
	DeleterId            *int64         `json:"deleterId,omitempty,string"`
	DeletionTime         *model.Time    `json:"deletionTime,omitempty"`
}

// fromEntity 将菜单实体及子树转换为接口返回模型。
//
// 转换保持树控件所需的派生字段。
func fromEntity(menu Menu) MenuResponse {
	children := make([]MenuResponse, 0, len(menu.Children))
	for _, child := range menu.Children {
		children = append(children, fromEntity(child))
	}
	return MenuResponse{Id: menu.Id, Title: menu.Title, Path: menu.Path, AuthCode: menu.AuthCode, Component: menu.Component, ComponentName: menu.ComponentName, ParentId: menu.ParentId, Sort: menu.Sort, Icon: menu.Icon, MenuType: menu.MenuType, Enabled: menu.Enabled, KeepAlive: menu.KeepAlive, Hidden: menu.Hidden, Children: children, HasChildren: menu.HasChildren, Leaf: menu.Leaf, Label: menu.Label, CreatorId: menu.CreatorId, CreationTime: menu.CreationTime, LastModifierId: menu.LastModifierId, LastModificationTime: menu.LastModificationTime, DeleterId: menu.DeleterId, DeletionTime: menu.DeletionTime}
}

// FromEntities 将菜单实体列表转换为接口返回模型列表。
//
// 返回值始终是非 nil 切片。
func FromEntities(menus []Menu) []MenuResponse {
	result := make([]MenuResponse, 0, len(menus))
	for _, menu := range menus {
		result = append(result, fromEntity(menu))
	}
	return result
}
