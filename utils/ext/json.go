package ext

import (
	"encoding/json"
	"go-apevolo/global"
	"go.uber.org/zap"
)

// JsonMarshal 序列化
func JsonMarshal(data interface{}) ([]byte, error) {
	serializedData, err := json.Marshal(data)
	if err != nil {
		global.Logger.Error("序列化失败", zap.Error(err))
	}
	return serializedData, err
}

// JsonUnmarshal 反序列化
func JsonUnmarshal(byteData []byte, result interface{}) error {
	err := json.Unmarshal(byteData, result)
	if err != nil {
		global.Logger.Error("反序列化失败", zap.Error(err))
	}
	return err
}
