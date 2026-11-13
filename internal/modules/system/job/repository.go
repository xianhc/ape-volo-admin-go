package job

import (
	"context"
	"errors"

	"github.com/xianhc/ape-volo-admin-go/internal/utils"
	"gorm.io/gorm"
)

// errNotFound 表示未找到可见的调度任务。
//
// repository 用该错误隔离 GORM 的记录不存在错误。
var errNotFound = errors.New("scheduled job not found")

// queryOptions 描述调度任务列表查询的数据库参数。
//
// Service 负责从 HTTP 请求转换该结构。
type queryOptions struct {
	JobName       string
	CreateTime    []string
	Sorting       string
	Limit, Offset int
}

// repository 封装调度任务的数据库访问。
//
// 零值即可使用，数据库连接和事务从 context 中获取。
type repository struct{}

// ExistsByName 判断是否存在同名作业。
//
// exceptID 非零时排除当前修改目标。
func (repository) ExistsByName(ctx context.Context, name string, exceptID int64) (bool, error) {
	var total int64
	db := utils.FromContext(ctx).Model(&Job{}).Where("job_name = ?", name)
	if exceptID != 0 {
		db = db.Where("id <> ?", exceptID)
	}
	err := db.Count(&total).Error
	return total > 0, err
}

// ExistsByExecution 判断程序集与执行类组合是否已被其他作业使用。
//
// exceptID 非零时排除当前修改目标。
func (repository) ExistsByExecution(ctx context.Context, assemblyName, className string, exceptID int64) (bool, error) {
	var total int64
	db := utils.FromContext(ctx).Model(&Job{}).Where("assembly_name = ? AND class_name = ?", assemblyName, className)
	if exceptID != 0 {
		db = db.Where("id <> ?", exceptID)
	}
	err := db.Count(&total).Error
	return total > 0, err
}

// FindByID 按主键读取调度任务。
//
// 未找到记录时返回 errNotFound。
func (repository) FindByID(ctx context.Context, id int64) (Job, error) {
	var task Job
	err := utils.FromContext(ctx).First(&task, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Job{}, errNotFound
	}
	return task, err
}

// FindByIDs 查询一组调度任务。
//
// 返回结果供删除流程同步清理运行时调度器。
func (repository) FindByIDs(ctx context.Context, ids []string) ([]Job, error) {
	tasks := make([]Job, 0)
	err := utils.FromContext(ctx).Find(&tasks, "id in (?)", ids).Error
	return tasks, err
}

// Create 持久化新的调度任务。
//
// 创建审计字段由 ORM 回调自动填充。
func (repository) Create(ctx context.Context, task *Job) error {
	return utils.FromContext(ctx).Create(task).Error
}

// Update 按主键更新调度任务。
//
// 修改审计字段由 ORM 回调自动填充。
func (repository) Update(ctx context.Context, task *Job) error {
	return utils.FromContext(ctx).Select("JobName", "JobGroup", "Cron", "ClassName", "AssemblyName", "Description", "Principal", "AlertEmail", "PauseAfterFailure", "StartTime", "EndTime", "TriggerType", "IntervalSecond", "CycleRunTimes", "Enabled", "RunParams", "LastModifierId", "LastModificationTime").Updates(task).Error
}

// SetEnabled 更新任务启用状态。
//
// 该方法用于手动执行未启用任务时持久化启用标记。
func (repository) SetEnabled(ctx context.Context, id int64, enabled bool) error {
	return utils.FromContext(ctx).Model(&Job{}).Where("id = ?", id).Update("enabled", enabled).Error
}

// Delete 逻辑删除指定调度任务。
//
// 返回受影响行数和数据库错误。
func (repository) Delete(ctx context.Context, ids []string) (int64, error) {
	result := utils.FromContext(ctx).Delete(&Job{}, ids)
	return result.RowsAffected, result.Error
}

// List 查询一页调度任务及分页前总数。
//
// 查询支持任务名称、创建时间和排序条件。
func (repository) List(ctx context.Context, options queryOptions) ([]Job, int64, error) {
	tasks := make([]Job, 0)
	var total int64
	db := buildQuery(utils.FromContext(ctx).Model(&Job{}), options)
	if err := db.Count(&total).Limit(options.Limit).Offset(options.Offset).Find(&tasks).Error; err != nil {
		return nil, 0, err
	}
	return tasks, total, nil
}

// FindAll 查询全部调度任务。
//
// 该方法供启动装配、下拉列表和导出使用。
func (repository) FindAll(ctx context.Context, options queryOptions) ([]Job, error) {
	tasks := make([]Job, 0)
	err := buildQuery(utils.FromContext(ctx).Model(&Job{}), options).Find(&tasks).Error
	return tasks, err
}

func buildQuery(db *gorm.DB, options queryOptions) *gorm.DB {
	if options.JobName != "" {
		db = db.Where("job_name LIKE ? ", "%"+options.JobName+"%")
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
