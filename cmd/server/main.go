/*
 * 项目名称：APE-VOLO-ADMIN-GO
 * 开源地址：https://github.com/xianhc/ApeVolo.AdminGO
 * 系统官网：https://www.apevolo.com
 * 联系邮箱：apevolo@gmail.com
 *
 * Copyright © 2022 xianhc（冼浩春）。保留所有权利。
 * SPDX-License-Identifier: Apache-2.0
 *
 * 本项目采用 Apache License 2.0，完整许可条款详见根目录 LICENSE 文件。
 *
 * 再分发本项目或其衍生作品时，须向接收者提供许可证副本。以源码形式
 * 分发衍生作品时，须保留相关版权及署名通知；修改过的文件须显著注明。
 * 完整义务以 LICENSE 条款为准。
 */

package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"

	"github.com/xianhc/ape-volo-admin-go/internal/core"
	"github.com/xianhc/ape-volo-admin-go/internal/global"
	"github.com/xianhc/ape-volo-admin-go/internal/initialize"
	"github.com/xianhc/ape-volo-admin-go/internal/utils"
	"go.uber.org/zap"
)

//go:generate go env -w GO111MODULE=on
//go:generate go env -w GOPROXY=https://goproxy.cn,direct
//go:generate go mod tidy
//go:generate go mod download

func main() {
	if err := run(); err != nil {
		os.Exit(1)
	}
}

func run() (err error) {
	global.Viper = core.Viper()
	global.Logger = core.Zap()
	zap.ReplaceGlobals(global.Logger)
	defer func() { err = errors.Join(err, core.CloseLogger()) }()
	defer func() {
		if err != nil {
			global.Logger.Error("服务退出", zap.Error(err))
		}
	}()
	if err := utils.ValidateRsaKey(); err != nil {
		return err
	}
	if err := initialize.Gorm(); err != nil {
		return err
	}
	defer func() { err = errors.Join(err, initialize.CloseDatabases()) }()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	return core.Run(ctx)
}
