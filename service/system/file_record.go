package system

import (
	"errors"
	"github.com/tealeg/xlsx"
	"go-apevolo/global"
	"go-apevolo/model"
	"go-apevolo/model/system"
	"go-apevolo/payloads/dto"
	"go-apevolo/payloads/request"
	"go-apevolo/utils"
	"go-apevolo/utils/ext"
	"go-apevolo/utils/upload"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"mime/multipart"
	"strconv"
)

type FileRecordService struct{}

// Create
// @description: 上传文件
// @receiver: fileRecordService
// @param: req
// @param: fileHeader
// @return: error
func (fileRecordService *FileRecordService) Create(req dto.CreateUpdateFileRecordDto, fileHeader *multipart.FileHeader) error {
	fileRecord := &system.FileRecord{}
	var total int64
	err := global.Db.Model(&system.FileRecord{}).Scopes(utils.IsDeleteSoft).Where("description = ?", req.Description).Count(&total).Error
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		return err
	}
	if total > 0 {
		return errors.New("文件描述=>" + req.Description + "=>已存在!")
	}

	req.Generate(fileRecord)
	local := upload.Local{}
	filePath, fileName, fileSize, fileTypeName, fileTypeNameEn, err := local.UploadFile(fileHeader)
	if err != nil {
		return err
	}
	fileRecord.OriginalName = fileHeader.Filename
	fileRecord.NewName = fileName
	fileRecord.FilePath = filePath
	fileRecord.Size = fileSize
	fileRecord.ContentType = fileHeader.Header.Get("Content-Type")
	fileRecord.ContentTypeName = fileTypeName
	fileRecord.ContentTypeNameEn = fileTypeNameEn
	err = global.Db.Create(&fileRecord).Error
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		return err
	}
	return nil
}

// Update
// @description: 修改
// @receiver: fileRecordService
// @param: req
// @return: error
func (fileRecordService *FileRecordService) Update(req *dto.CreateUpdateFileRecordDto) error {
	fileRecord := &system.FileRecord{}
	oldFileRecord := &system.FileRecord{}
	err := global.Db.Scopes(utils.IsDeleteSoft).First(oldFileRecord, req.Id).Error
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("数据不存在或您无权查看！")
		}
		return err
	}
	if oldFileRecord.Description != req.Description {
		var total int64
		err = global.Db.Model(&system.FileRecord{}).Scopes(utils.IsDeleteSoft).Where("description = ?", req.Description).Count(&total).Error
		if err != nil {
			global.Logger.Error("db error: ", zap.Error(err))
			return err
		}
		if total > 0 {
			return errors.New("文件描述=>" + req.Description + "=>已存在!")
		}
	}
	req.Generate(fileRecord)
	err = global.Db.Updates(fileRecord).Error
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		return err
	}
	return nil
}

// Delete
// @description: 删除
// @receiver: fileRecordService
// @param: idArray
// @param: updateBy
// @return: error
func (fileRecordService *FileRecordService) Delete(idArray request.IdCollection, updateBy string) error {
	var fileRecords []system.FileRecord
	err := global.Db.Scopes(utils.IsDeleteSoft).Find(fileRecords, "id in ?", idArray.IdArray).Error
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		return err
	}
	if len(fileRecords) <= 0 {
		return errors.New("数据不存在或您无权查看！")
	}

	localTime := ext.GetCurrentTime()
	err = global.Db.Model(&system.FileRecord{}).Scopes(utils.IsDeleteSoft).Where("id in ?", idArray.IdArray).Updates(
		system.FileRecord{BaseModel: model.BaseModel{
			UpdateBy:   &updateBy,
			UpdateTime: &localTime,
		}, SoftDeleted: model.SoftDeleted{IsDeleted: true}},
	).Error
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		return err
	}
	local := upload.Local{}
	for _, file := range fileRecords {
		_ = local.DeleteFile(file.FilePath)
	}
	return nil
}

// Query
// @description: 查询
// @receiver: fileRecordService
// @param: info
// @param: list
// @param: count
// @return: error
func (fileRecordService *FileRecordService) Query(info *dto.FileRecordQueryCriteria, list *[]system.FileRecord, count *int64) error {
	limit := info.PageSize
	offset := info.PageSize * (info.PageIndex - 1)
	// 创建db
	db := buildFileRecordQuery(global.Db.Model(&system.FileRecord{}), info)
	err := db.Count(count).Limit(limit).Offset(offset).Find(list).Error
	if err != nil {
		global.Logger.Error("db error: ", zap.Error(err))
		return err
	}
	return nil
}

// Download
// @description: 导出
// @receiver: fileRecordService
// @param: info
// @return: filePath
// @return: fileName
// @return: err
func (fileRecordService *FileRecordService) Download(info *dto.FileRecordQueryCriteria) (filePath string, fileName string, err error) {
	var fileRecords []system.FileRecord
	// 创建db并构建查询条件
	err = buildFileRecordQuery(global.Db.Model(&system.AppSecret{}), info).Find(&fileRecords).Error

	if err != nil {
		return
	}
	// 创建一个新的 Excel 文件
	file := xlsx.NewFile()
	sheet, err := file.AddSheet("FileRecord")
	if err != nil {
		return
	}
	style := xlsx.NewStyle()
	style.Fill = *xlsx.NewFill("solid", "1e90ff", "1e90ff")
	style.ApplyFill = true
	// 表头数据
	header := []string{"ID", "文件描述", "文件类型", "文件类型名称", "文件类型名称(EN)", "源名称", "新名称", "存储路径", "文件大小", "创建时间"}
	// 将表头添加到 Excel 工作表
	row := sheet.AddRow()
	for _, cellValue := range header {
		cell := row.AddCell()
		cell.Value = cellValue
		cell.SetStyle(style)
	}

	// 将数据添加到 Excel 工作表
	for _, fileRecord := range fileRecords {
		row := sheet.AddRow()
		row.WriteSlice(&[]interface{}{
			strconv.FormatInt(fileRecord.Id, 10),
			fileRecord.Description,
			fileRecord.ContentType,
			fileRecord.ContentTypeName,
			fileRecord.ContentTypeNameEn,
			fileRecord.OriginalName,
			fileRecord.NewName,
			fileRecord.FilePath,
			fileRecord.Size,
			fileRecord.CreateTime.Format("2006-01-02 15:04:05"),
		}, -1)
	}
	// 保存 Excel 文件到本地
	fileName = "FileRecord_" + utils.GenerateID().String() + ".xlsx"
	filePath = global.Config.Excel.Dir + fileName
	err = file.Save(filePath)
	return filePath, fileName, err
}

// buildFileRecordQuery
// @description: 条件表达式
// @param: db
// @param: info
// @return: *gorm.DB
func buildFileRecordQuery(db *gorm.DB, info *dto.FileRecordQueryCriteria) *gorm.DB {
	if info.KeyWords != "" {
		db = db.Where("description LIKE ? or original_name LIKE ? ", "%"+info.KeyWords+"%", "%"+info.KeyWords+"%")
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
