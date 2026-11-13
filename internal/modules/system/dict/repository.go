package dict

import (
	"context"
	"errors"

	"github.com/xianhc/ape-volo-admin-go/internal/utils"
	"gorm.io/gorm"
)

// errNotFound 表示未找到可见的字典或字典明细。
//
// repository 用该错误隔离 GORM 的记录不存在错误。
var errNotFound = errors.New("dict data not found")

// queryOptions 描述字典列表查询的数据库参数。
//
// Service 负责从 HTTP 请求转换该结构。
type queryOptions struct {
	KeyWords string
	Sorting  string
	Limit    int
	Offset   int
}

// repository 封装字典主表和明细的数据库访问。
//
// 主从实体位于同一模块，共享查询和事务上下文。
type repository struct{}

// DictExistsByName 判断字典名称是否已存在。
//
// 查询继承软删除过滤，但字典按项目约定不接入数据权限。
func (repository) DictExistsByName(ctx context.Context, name string) (bool, error) {
	var total int64
	err := utils.FromContext(ctx).Model(&Dict{}).Where("name = ?", name).Count(&total).Error
	return total > 0, err
}

// FindDictByID 按主键读取字典主表。
//
// 未找到记录时返回 errNotFound。
func (repository) FindDictByID(ctx context.Context, id int64) (Dict, error) {
	var dict Dict
	err := utils.FromContext(ctx).Where("id = ?", id).First(&dict).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Dict{}, errNotFound
	}
	return dict, err
}

// FindDictByName 按名称读取字典及其明细。
//
// 明细通过 GORM 预加载，并保留数据库默认顺序。
func (repository) FindDictByName(ctx context.Context, name string) (Dict, error) {
	var dict Dict
	err := utils.FromContext(ctx).Preload("DictDetails").Where("name = ?", name).First(&dict).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Dict{}, errNotFound
	}
	return dict, err
}

// CreateDict 持久化新的字典主表。
//
// 创建审计字段由 ORM 回调自动填充。
func (repository) CreateDict(ctx context.Context, dict *Dict) error {
	return utils.FromContext(ctx).Create(dict).Error
}

// UpdateDict 按主键更新字典主表。
//
// 修改审计字段由 ORM 回调自动填充。
func (repository) UpdateDict(ctx context.Context, dict *Dict) error {
	return utils.FromContext(ctx).Select("Name", "Description", "LastModifierId", "LastModificationTime").Updates(dict).Error
}

// DeleteDicts 逻辑删除指定字典主键集合。
//
// 返回受影响行数，Service 据此判断数据是否存在。
func (repository) DeleteDicts(ctx context.Context, ids []string) (int64, error) {
	result := utils.FromContext(ctx).Delete(&Dict{}, ids)
	return result.RowsAffected, result.Error
}

// ListDicts 查询一页字典及其明细。
//
// 返回分页前总数，明细通过 GORM 预加载。
func (repository) ListDicts(ctx context.Context, options queryOptions) ([]Dict, int64, error) {
	dicts := make([]Dict, 0)
	var total int64
	db := buildDictQuery(utils.FromContext(ctx).Model(&Dict{}), options)
	if err := db.Count(&total).Limit(options.Limit).Offset(options.Offset).Preload("DictDetails").Find(&dicts).Error; err != nil {
		return nil, 0, err
	}
	return dicts, total, nil
}

// FindAllDicts 查询符合条件的全部字典及其明细。
//
// 该方法用于导出，不应用分页参数。
func (repository) FindAllDicts(ctx context.Context, options queryOptions) ([]Dict, error) {
	dicts := make([]Dict, 0)
	err := buildDictQuery(utils.FromContext(ctx).Model(&Dict{}), options).Preload("DictDetails").Find(&dicts).Error
	return dicts, err
}

// DetailExists 判断同一字典下的标签和值组合是否已存在。
//
// 该检查用于创建字典明细前的唯一性校验。
func (repository) DetailExists(ctx context.Context, dictId int64, label, value string) (bool, error) {
	var total int64
	err := utils.FromContext(ctx).Model(&DictDetail{}).Where("dict_id = ? and label = ? and value = ?", dictId, label, value).Count(&total).Error
	return total > 0, err
}

// FindDetailByID 按主键读取字典明细。
//
// 未找到记录时返回 errNotFound。
func (repository) FindDetailByID(ctx context.Context, id int64) (DictDetail, error) {
	var detail DictDetail
	err := utils.FromContext(ctx).First(&detail, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return DictDetail{}, errNotFound
	}
	return detail, err
}

// CreateDetail 持久化新的字典明细。
//
// 创建审计字段由 ORM 回调自动填充。
func (repository) CreateDetail(ctx context.Context, detail *DictDetail) error {
	return utils.FromContext(ctx).Create(detail).Error
}

// UpdateDetail 按主键更新字典明细。
//
// 修改审计字段由 ORM 回调自动填充。
func (repository) UpdateDetail(ctx context.Context, detail *DictDetail) error {
	return utils.FromContext(ctx).Select("DictId", "Label", "Value", "DictSort", "LastModifierId", "LastModificationTime").Updates(detail).Error
}

// DeleteDetail 逻辑删除指定字典明细。
//
// 返回受影响行数，Service 据此判断数据是否存在。
func (repository) DeleteDetail(ctx context.Context, id int64) (int64, error) {
	result := utils.FromContext(ctx).Delete(&DictDetail{}, id)
	return result.RowsAffected, result.Error
}

// FindDetailsByDictName 按字典名称查询排序后的明细。
//
// 返回所属字典主键，用于重建明细响应中的 `dict.id`。
func (repository) FindDetailsByDictName(ctx context.Context, name string) (Dict, error) {
	var dict Dict
	err := utils.FromContext(ctx).Preload("DictDetails", func(db *gorm.DB) *gorm.DB { return db.Order("dict_sort") }).Where("name = ?", name).First(&dict).Error
	return dict, err
}

func buildDictQuery(db *gorm.DB, options queryOptions) *gorm.DB {
	if options.KeyWords != "" {
		db = db.Where("name LIKE ?", "%"+options.KeyWords+"%").Or("description LIKE ?", "%"+options.KeyWords+"%")
	}
	if options.Sorting != "" {
		db = utils.ApplySorting(db, options.Sorting)
	}
	return db
}
