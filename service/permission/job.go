package permission

import (
	"errors"
	"github.com/tealeg/xlsx"
	"go-apevolo/global"
	"go-apevolo/model"
	"go-apevolo/model/permission"
	"go-apevolo/payloads/dto"
	"go-apevolo/payloads/request"
	"go-apevolo/utils"
	"go-apevolo/utils/ext"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"strconv"
)

type JobService struct{}

// Create
// @description: 创建
// @receiver: jobService
// @param: req
// @return: error
func (jobService *JobService) Create(req *dto.CreateUpdateJobDto) error {
	job := &permission.Job{}
	var total int64
	err := global.Db.Model(&permission.Job{}).Scopes(utils.IsDeleteSoft).Where("name = ?", req.Name).Count(&total).Error
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		return err
	}
	if total > 0 {
		return errors.New("岗位名称=>" + req.Name + "=>已存在!")
	}
	req.Generate(job)
	err = global.Db.Create(job).Error
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		return err
	}
	return nil
}

// Update
// @description: 编辑
// @receiver: jobService
// @param: req
// @return: error
func (jobService *JobService) Update(req *dto.CreateUpdateJobDto) error {
	var oldJob permission.Job
	job := &permission.Job{}
	err := global.Db.Scopes(utils.IsDeleteSoft).First(&oldJob, req.Id).Error
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("数据不存在或您无权查看！")
		}
		return err
	}
	if oldJob.Name != req.Name {
		var total int64
		err = global.Db.Model(&permission.Job{}).Scopes(utils.IsDeleteSoft).Where("name = ?", req.Name).Count(&total).Error
		if err != nil {
			global.Logger.Error("db error: ", zap.Error(err))
			return err
		}
		if total > 0 {
			return errors.New("岗位名称=>" + req.Name + "=>已存在!")
		}
	}
	req.Generate(job)
	err = global.Db.Updates(job).Error
	return err
}

// Delete
// @description: 删除
// @receiver: jobService
// @param: idArray
// @param: updateBy
// @return: error
func (jobService *JobService) Delete(idArray request.IdCollection, updateBy string) error {
	var jobs []permission.Job
	err := global.Db.Scopes(utils.IsDeleteSoft).Find(&jobs, "id in ?", idArray.IdArray).Error
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		return err
	}
	if len(jobs) == 0 {
		return errors.New("数据不存在")
	}
	var JobIds []int64
	for _, job := range jobs {
		JobIds = utils.AppendInt64(JobIds, job.Id)
	}
	var userJobs []permission.UserJob
	err = global.Db.Find(&userJobs, "job_id in ?", JobIds).Error
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		return err
	}
	if len(userJobs) > 0 {
		return errors.New("数据被使用,无法删除")
	}
	localTime := ext.GetCurrentTime()
	err = global.Db.Model(&permission.Job{}).Scopes(utils.IsDeleteSoft).Where("id in (?)", idArray.IdArray).Updates(
		permission.Job{BaseModel: model.BaseModel{
			UpdateBy:   &updateBy,
			UpdateTime: &localTime,
		}, SoftDeleted: model.SoftDeleted{IsDeleted: true}},
	).Error
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		return err
	}
	return nil
}

// Query
// @description: 查询
// @receiver: jobService
// @param: info
// @param: list
// @param: total
// @return: error
func (jobService *JobService) Query(info *dto.JobQueryCriteria, list *[]permission.Job, total *int64) error {
	limit := info.PageSize
	offset := info.PageSize * (info.PageIndex - 1)
	// 创建db并构建查询条件
	db := buildJobQuery(global.Db.Model(&permission.Job{}), info)
	err := db.Count(total).Limit(limit).Offset(offset).Find(list).Error
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		return err
	}
	return nil
}

// Download
// @description: 导出
// @receiver: jobService
// @param: info
// @return: filePath
// @return: fileName
// @return: err
func (jobService *JobService) Download(info *dto.JobQueryCriteria) (filePath string, fileName string, err error) {
	var jobs []permission.Job
	// 创建db并构建查询条件
	err = buildJobQuery(global.Db.Model(&permission.Job{}), info).Find(&jobs).Error

	if err != nil {
		return
	}
	// 创建一个新的 Excel 文件
	file := xlsx.NewFile()
	sheet, err := file.AddSheet("Jobs")
	if err != nil {
		return
	}
	style := xlsx.NewStyle()
	style.Fill = *xlsx.NewFill("solid", "1e90ff", "1e90ff")
	style.ApplyFill = true
	// 表头数据
	header := []string{"ID", "岗位名称", "排序", "状态", "创建时间"}
	// 将表头添加到 Excel 工作表
	row := sheet.AddRow()
	for _, cellValue := range header {
		cell := row.AddCell()
		cell.Value = cellValue
		cell.SetStyle(style)
	}

	// 将数据添加到 Excel 工作表
	for _, job := range jobs {
		row := sheet.AddRow()
		row.WriteSlice(&[]interface{}{
			strconv.FormatInt(job.Id, 10),
			job.Name,
			job.Sort,
			func() string {
				if job.Enabled {
					return "启用"
				}
				return "禁用"
			}(),
			job.CreateTime.Format("2006-01-02 15:04:05"),
		}, -1)
	}
	// 保存 Excel 文件到本地
	fileName = "Jobs_" + utils.GenerateID().String() + ".xlsx"
	filePath = global.Config.Excel.Dir + fileName
	err = file.Save(filePath)
	return filePath, fileName, err
}

// GetAllJob
// @description: 获取全部岗位
// @receiver: jobService
// @param: jobs
// @return: error
func (jobService *JobService) GetAllJob(jobs *[]permission.Job) error {
	err := global.Db.Where("enabled = 1").Order("sort asc").Find(jobs).Error
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		return err
	}
	return nil
}

// buildJobQuery
// @description: 条件表达式
// @param: db
// @param: info
// @return: *gorm.DB
func buildJobQuery(db *gorm.DB, info *dto.JobQueryCriteria) *gorm.DB {
	if info.JobName != "" {
		db = db.Where("name LIKE ?", "%"+info.JobName+"%")
	}
	if info.Enabled != nil {
		db = db.Where("enabled = ?", info.Enabled)
	}
	if len(info.CreateTime) > 0 {
		db = db.Where("create_time >= ? and create_time <= ? ", info.CreateTime[0], info.CreateTime[1])
	}
	if len(info.SortFields) > 0 {
		for _, sort := range info.SortFields {
			db = db.Order(sort)
		}
	}
	db = db.Scopes(utils.IsDeleteSoft)

	return db
}
