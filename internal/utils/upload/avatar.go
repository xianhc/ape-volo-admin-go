package upload

import (
	"errors"
	"io"
	"mime/multipart"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/xianhc/ape-volo-admin-go/internal/global"
	"github.com/xianhc/ape-volo-admin-go/internal/utils"
)

// avatarExts 是允许的头像扩展名（小写、带点）。
var avatarExts = map[string]struct{}{
	".jpg": {}, ".jpeg": {}, ".png": {}, ".gif": {}, ".bmp": {}, ".webp": {},
}

// AvatarDir 是头像在存储目录下的子目录，同时也是对外路径的一段。
const AvatarDir = "avatar"

// UploadAvatar 保存头像文件并返回可访问的相对路径。
//
// 单独实现而不复用 UploadFile：头像固定落在 avatar 目录，不按文件类型分目录，
// 否则历史头像路径会随分类规则变化而失效。返回值形如
// `/uploads/file/avatar/20260101120000_123.png`，直接写进 `sys_user.avatar_path`。
func UploadAvatar(file *multipart.FileHeader) (string, error) {
	if file == nil || file.Size <= 0 {
		return "", errors.New("不能上传空文件")
	}
	ext := strings.ToLower(path.Ext(file.Filename))
	if _, ok := avatarExts[ext]; !ok {
		return "", errors.New("头像仅支持 jpg、jpeg、png、gif、bmp、webp 格式")
	}

	fileName := time.Now().Format("20060102150405") + "_" + strconv.FormatInt(int64(utils.GenerateID()), 10) + ext
	storeDir := global.Config.Local.StorePath + "/" + AvatarDir
	if err := os.MkdirAll(storeDir, os.ModePerm); err != nil {
		return "", err
	}

	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer func() { _ = src.Close() }()

	dst, err := os.Create(storeDir + "/" + fileName)
	if err != nil {
		return "", err
	}
	saved := false
	defer func() {
		_ = dst.Close()
		if !saved {
			_ = os.Remove(storeDir + "/" + fileName)
		}
	}()

	if _, err = io.Copy(dst, src); err != nil {
		return "", err
	}
	if err := dst.Close(); err != nil {
		return "", err
	}
	saved = true
	return global.Config.Local.Path + "/" + AvatarDir + "/" + fileName, nil
}
