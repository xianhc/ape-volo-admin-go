package utils

import (
	"github.com/bwmarrin/snowflake"
	"github.com/xianhc/ape-volo-admin-go/internal/global"
	"go.uber.org/zap"
)

var node *snowflake.Node

func init() {
	var err error
	node, err = snowflake.NewNode(1) // 传入一个唯一的节点 ID
	if err != nil {
		global.Logger.Error(err.Error(), zap.Error(err))
	}
}

func GenerateID() snowflake.ID {
	return node.Generate()
}
