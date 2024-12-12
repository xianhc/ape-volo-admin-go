package upload

import (
	"errors"
	"fmt"
	"go-apevolo/global"
	"go-apevolo/utils"
	"go.uber.org/zap"
	"io"
	"mime/multipart"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

type Local struct{}

// UploadFile
// @description: 上传文件
// @receiver: *Local
// @param: file
// @return: string
// @return: string
// @return: string
// @return: string
// @return: string
// @return: error
func (*Local) UploadFile(file *multipart.FileHeader) (string, string, string, string, string, error) {
	// 读取文件后缀
	ext := path.Ext(file.Filename)
	name := strconv.FormatInt(int64(utils.GenerateID()), 10)
	//后缀名
	suffixName := strings.ReplaceAll(ext, ".", "")
	// 拼接新文件名
	fileName := time.Now().Format("20060102150405") + "_" + name + ext

	fileSize := getFileSize(file.Size)
	fileTypeName := getFileTypeName(suffixName)
	fileTypeNameEn := getFileTypeNameEn(fileTypeName)
	p := global.Config.Local.StorePath + "/" + fileTypeNameEn + "/" + fileName
	filePath := global.Config.Local.Path + "/" + fileTypeNameEn + "/" + fileName
	// 尝试创建此路径
	mkdirErr := os.MkdirAll(global.Config.Local.StorePath+"/"+fileTypeNameEn, os.ModePerm)
	if mkdirErr != nil {
		global.Logger.Error("function os.MkdirAll() failed", zap.Any("err", mkdirErr.Error()))
		return "", "", "", "", "", errors.New("function os.MkdirAll() failed, err:" + mkdirErr.Error())
	}

	f, openError := file.Open() // 读取文件
	if openError != nil {
		global.Logger.Error("function file.Open() failed", zap.Any("err", openError.Error()))
		return "", "", "", "", "", errors.New("function file.Open() failed, err:" + openError.Error())
	}
	defer f.Close() // 创建文件 defer 关闭

	out, createErr := os.Create(p)
	if createErr != nil {
		global.Logger.Error("function os.Create() failed", zap.Any("err", createErr.Error()))

		return "", "", "", "", "", errors.New("function os.Create() failed, err:" + createErr.Error())
	}
	defer out.Close() // 创建文件 defer 关闭

	_, copyErr := io.Copy(out, f) // 传输（拷贝）文件
	if copyErr != nil {
		global.Logger.Error("function io.Copy() failed", zap.Any("err", copyErr.Error()))
		return "", "", "", "", "", errors.New("function io.Copy() failed, err:" + copyErr.Error())
	}
	return filePath, fileName, fileSize, fileTypeName, fileTypeNameEn, nil
}

func getFileSize(size int64) string {
	//GB
	const GB = 1024 * 1024 * 1024
	//MB
	const MB = 1024 * 1024
	//KB
	const KB = 1024

	var fileSize = ""
	if size/GB >= 1 {
		// 如果当前Byte的值大于等于1GB
		fileSize = fmt.Sprintf("%.2fGB", float64(size)/GB)
	} else if size/MB >= 1 {
		//如果当前Byte的值大于等于1MB
		fileSize = fmt.Sprintf("%.2fMB", float64(size)/MB)
	} else if size/KB >= 1 {
		//如果当前Byte的值大于等于1KB
		fileSize = fmt.Sprintf("%.2fKB", float64(size)/KB)
	} else {
		fileSize = strconv.FormatInt(size, 10) + "B"
	}

	return fileSize
}

func getFileTypeName(fileType string) string {
	documents := []string{"txt", "doc", "pdf", "ppt", "pps", "xlsx", "xls", "docx"}
	musics := []string{"mp3", "wav", "wma", "mpa", "ram", "ra", "aac", "aif", "m4a"}
	videos := []string{
		"mpe", "asf", "mov", "qt", "rm", "mp4", "ogg", "webm", "ogv", "flv", "m4v", "mpg", "wmv", "mpeg",
		"avi",
	}
	images := []string{
		"dib", "tif", "iff", "mpt", "cdr", "bmp", "dif", "jpg", "psd", "pcp", "gif", "jpeg", "png", "pcd",
		"tga", "eps", "wmf",
	}

	if contains(documents, fileType) {
		return "文档"
	}

	if contains(musics, fileType) {
		return "音乐"
	}

	if contains(videos, fileType) {
		return "视频"
	}

	if contains(images, fileType) {
		return "图片"
	}

	return "其他"
}

func contains(slice []string, element string) bool {
	for _, item := range slice {
		if strings.ToLower(item) == strings.ToLower(element) {
			return true
		}
	}
	return false
}

// getFileTypeNameEn 根据文件类型返回对应的英文名称
func getFileTypeNameEn(fileType string) string {
	switch fileType {
	case "文档":
		return "documents"
	case "视频":
		return "videos"
	case "音乐":
		return "musics"
	case "图片":
		return "images"
	default:
		return "other"
	}
}

// DeleteFile
// @description: 删除文件
// @receiver: *Local
// @param: key
// @return: error
func (*Local) DeleteFile(key string) error {
	if err := os.Remove(key); err != nil {
		return errors.New("删除失败:" + err.Error())
	}
	return nil
}
