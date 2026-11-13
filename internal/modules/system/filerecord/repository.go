package filerecord

import (
	"context"
	"errors"

	"github.com/xianhc/ape-volo-admin-go/internal/utils"
	"gorm.io/gorm"
)

// errNotFound 表示按主键未找到可见的文件记录。
//
// repository 用该错误隔离 GORM 的记录不存在错误。
var errNotFound = errors.New("file record not found")

// queryOptions 描述文件记录列表查询的数据库参数。
//
// Service 负责从 HTTP 请求转换该结构，repository 不依赖请求模型。
type queryOptions struct {
	KeyWords   string
	CreateTime []string
	Sorting    string
	Limit      int
	Offset     int
}

// repository 封装文件记录的数据库访问。
//
// 零值即可使用，数据库连接和事务从 context 中获取。
type repository struct{}

// ExistsByDescription 判断指定文件描述是否已存在。
//
// 查询继承当前请求的数据权限与软删除过滤。
func (repository) ExistsByDescription(ctx context.Context, description string) (bool, error) {
	var total int64
	err := utils.FromContext(ctx).Model(&FileRecord{}).Where("description = ?", description).Count(&total).Error
	return total > 0, err
}

// FindByID 按主键读取文件记录。
//
// 未找到记录时返回 errNotFound，其余数据库错误原样返回。
func (repository) FindByID(ctx context.Context, id int64) (FileRecord, error) {
	var file FileRecord
	err := utils.FromContext(ctx).First(&file, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return FileRecord{}, errNotFound
	}
	return file, err
}

// FindByIDs 读取指定主键集合对应的文件记录。
//
// 返回结果用于删除数据库记录后清理本地文件。
func (repository) FindByIDs(ctx context.Context, ids []string) ([]FileRecord, error) {
	files := make([]FileRecord, 0)
	err := utils.FromContext(ctx).Find(&files, "id in ?", ids).Error
	return files, err
}

// Create 持久化新的文件记录。
//
// 创建审计字段由 ORM 回调根据 context 自动填充。
func (repository) Create(ctx context.Context, file *FileRecord) error {
	return utils.FromContext(ctx).Create(file).Error
}

// Update 按实体主键更新文件记录。
//
// 修改审计字段由 ORM 回调根据 context 自动填充。
func (repository) Update(ctx context.Context, file *FileRecord) error {
	return utils.FromContext(ctx).Select("Description", "LastModifierId", "LastModificationTime").Updates(file).Error
}

// Delete 逻辑删除指定主键集合对应的文件记录。
//
// 返回受影响行数和数据库错误。
func (repository) Delete(ctx context.Context, ids []string) (int64, error) {
	result := utils.FromContext(ctx).Delete(&FileRecord{}, ids)
	return result.RowsAffected, result.Error
}

// List 查询一页文件记录及分页前总数。
//
// 筛选、排序和分页在同一个查询构造器上执行。
func (repository) List(ctx context.Context, options queryOptions) ([]FileRecord, int64, error) {
	files := make([]FileRecord, 0)
	var total int64
	db := buildQuery(utils.FromContext(ctx).Model(&FileRecord{}), options)
	if err := db.Count(&total).Limit(options.Limit).Offset(options.Offset).Find(&files).Error; err != nil {
		return nil, 0, err
	}
	return files, total, nil
}

// FindAll 查询符合条件的全部文件记录。
//
// 该方法用于导出，不应用分页参数。
func (repository) FindAll(ctx context.Context, options queryOptions) ([]FileRecord, error) {
	files := make([]FileRecord, 0)
	err := buildQuery(utils.FromContext(ctx).Model(&FileRecord{}), options).Find(&files).Error
	return files, err
}

func buildQuery(db *gorm.DB, options queryOptions) *gorm.DB {
	if options.KeyWords != "" {
		db = db.Where("description LIKE ? or original_name LIKE ? ", "%"+options.KeyWords+"%", "%"+options.KeyWords+"%")
	}
	if len(options.CreateTime) != 0 && len(options.CreateTime) != 2 {
		_ = db.AddError(errors.New("创建时间必须包含起止两个值"))
		return db
	}
	if len(options.CreateTime) == 2 {
		db = db.Where("creation_time >= ? and creation_time <= ? ", options.CreateTime[0], options.CreateTime[1])
	}
	if options.Sorting != "" {
		db = utils.ApplySorting(db, options.Sorting)
	}
	return db
}
