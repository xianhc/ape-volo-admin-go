package filerecord

import (
	"context"
	"errors"
	"mime/multipart"
	"strconv"

	"github.com/xianhc/ape-volo-admin-go/internal/global"
	"github.com/xianhc/ape-volo-admin-go/internal/utils"
	"github.com/xianhc/ape-volo-admin-go/internal/utils/excel"
	"github.com/xianhc/ape-volo-admin-go/internal/utils/upload"
	"go.uber.org/zap"
)

// Service 实现文件记录的业务规则。
//
// Service 编排文件系统操作，并只通过 repository 访问数据库。
type Service struct {
	repository repository
}

// Create 上传文件并创建文件记录。
//
// 方法先校验描述唯一性，再保存本地文件，最后写入文件元数据。
func (service *Service) Create(ctx context.Context, req CreateAndUpdateFileRecordRequest, header *multipart.FileHeader) error {
	return utils.InTransaction(ctx, func(ctx context.Context) error { return service.create(ctx, req, header) })
}

func (service *Service) create(ctx context.Context, req CreateAndUpdateFileRecordRequest, header *multipart.FileHeader) error {
	exists, err := service.repository.ExistsByDescription(ctx, req.Description)
	if err != nil {
		return logDatabaseError(err)
	}
	if exists {
		return errors.New("文件描述=>" + req.Description + "=>已存在!")
	}
	file := &FileRecord{Description: req.Description}
	file.Id = req.Id
	if file.Id == 0 {
		file.Id = int64(utils.GenerateID())
	}
	local := upload.Local{}
	filePath, fileName, fileSize, typeName, typeNameEn, err := local.UploadFile(header)
	if err != nil {
		return err
	}
	utils.AfterRollback(ctx, func(context.Context) {
		if err := local.DeleteFile(filePath); err != nil {
			global.Logger.Warn("清理未提交的上传文件失败", zap.Error(err))
		}
	})
	file.OriginalName = header.Filename
	file.NewName = fileName
	file.FilePath = filePath
	file.Size = fileSize
	file.ContentType = header.Header.Get("Content-Type")
	file.ContentTypeName = typeName
	file.ContentTypeNameEn = typeNameEn
	if err := service.repository.Create(ctx, file); err != nil {
		return logDatabaseError(err)
	}
	return nil
}

// Update 修改文件记录描述。
//
// 方法先确认目标记录存在，再在描述变化时校验唯一性。
func (service *Service) Update(ctx context.Context, req CreateAndUpdateFileRecordRequest) error {
	oldFile, err := service.repository.FindByID(ctx, req.Id)
	if errors.Is(err, errNotFound) {
		return errors.New("数据不存在或您无权查看！")
	}
	if err != nil {
		return logDatabaseError(err)
	}
	if oldFile.Description != req.Description {
		exists, err := service.repository.ExistsByDescription(ctx, req.Description)
		if err != nil {
			return logDatabaseError(err)
		}
		if exists {
			return errors.New("文件描述=>" + req.Description + "=>已存在!")
		}
	}
	file := &FileRecord{Description: req.Description}
	file.Id = req.Id
	if err := service.repository.Update(ctx, file); err != nil {
		return logDatabaseError(err)
	}
	return nil
}

// Delete 删除文件记录及其本地文件。
//
// 方法保持先逻辑删除数据库记录、再逐个清理本地文件的既有顺序。
func (service *Service) Delete(ctx context.Context, ids []string) error {
	return utils.InTransaction(ctx, func(ctx context.Context) error { return service.delete(ctx, ids) })
}

func (service *Service) delete(ctx context.Context, ids []string) error {
	files, err := service.repository.FindByIDs(ctx, ids)
	if err != nil {
		return logDatabaseError(err)
	}
	if len(files) == 0 {
		return errors.New("数据不存在或您无权查看！")
	}
	if _, err := service.repository.Delete(ctx, ids); err != nil {
		return logDatabaseError(err)
	}
	local := upload.Local{}
	for _, file := range files {
		utils.AfterCommit(ctx, "file-delete:"+file.FilePath, func(context.Context) {
			if err := local.DeleteFile(file.FilePath); err != nil {
				global.Logger.Warn("清理已删除记录的文件失败", zap.Error(err))
			}
		})
	}
	return nil
}

// Query 查询文件记录分页列表。
//
// 返回前统一把 Entity 转换为 FileRecordResponse。
func (service *Service) Query(ctx context.Context, req QueryFileRecordRequest) ([]FileRecordResponse, int64, error) {
	files, total, err := service.repository.List(ctx, newListOptions(req))
	if err != nil {
		return nil, 0, logDatabaseError(err)
	}
	return fromEntities(files), total, nil
}

// Download 导出符合条件的文件记录。
//
// 导出沿用列表筛选和排序条件，并返回生成文件路径与下载文件名。
func (service *Service) Download(ctx context.Context, req QueryFileRecordRequest) (filePath string, fileName string, err error) {
	files, err := service.repository.FindAll(ctx, newListOptions(req))
	if err != nil {
		return "", "", logDatabaseError(err)
	}
	excelFile := excel.NewFile("FileRecord")
	defer func() { _ = excelFile.Close() }()
	excelFile.SetHeader([]string{"ID", "文件描述", "文件类型", "文件类型名称", "文件类型名称(EN)", "源名称", "新名称", "存储路径", "文件大小", "创建时间"})
	for _, file := range files {
		excelFile.AddRow(strconv.FormatInt(file.Id, 10), file.Description, file.ContentType, file.ContentTypeName, file.ContentTypeNameEn, file.OriginalName, file.NewName, file.FilePath, file.Size, file.CreationTime.Format("2006-01-02 15:04:05"))
	}
	fileName = "FileRecord_" + utils.GenerateID().String() + ".xlsx"
	filePath = global.Config.Excel.Dir + fileName
	return filePath, fileName, excelFile.Save(filePath)
}

func newListOptions(req QueryFileRecordRequest) queryOptions {
	return queryOptions{KeyWords: req.KeyWords, CreateTime: req.CreateTime, Sorting: req.Sorting, Limit: req.PageSize, Offset: req.PageSize * (req.PageIndex - 1)}
}

func logDatabaseError(err error) error {
	global.Logger.Error("db error: ", zap.Error(err))
	return err
}
