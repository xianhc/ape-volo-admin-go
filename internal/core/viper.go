package core

import (
	"flag"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	serverconfig "github.com/xianhc/ape-volo-admin-go/internal/config"
	"github.com/xianhc/ape-volo-admin-go/internal/core/internal"

	"github.com/xianhc/ape-volo-admin-go/internal/global"
)

// setDefaults 注册关键配置项的默认值。
//
// 缺省值保证配置缺项时行为完整，并为环境变量覆盖登记对应键。
func setDefaults(v *viper.Viper) {
	v.SetDefault("system.port", 9002)
	v.SetDefault("system.main-database", "sqlite")
	v.SetDefault("system.is-init-table", true)
	v.SetDefault("system.is-init-table-data", true)
	for _, id := range []string{"sqlite", "mysql", "mssql", "pgsql", "oracle"} {
		base := "data-connection.connections." + id + "."
		v.SetDefault(base+"enabled", id == "sqlite")
		v.SetDefault(base+"db-type", id)
		v.SetDefault(base+"max-idle-conns", 1)
		v.SetDefault(base+"max-open-conns", 1)
	}
	v.SetDefault("data-connection.connections.sqlite.host", "./data")
	v.SetDefault("data-connection.connections.sqlite.db-name", "ape-volo-admin-main")
	v.SetDefault("log-data-connection.enabled", true)
	v.SetDefault("log-data-connection.db-type", "sqlite")
	v.SetDefault("log-data-connection.host", "./data")
	v.SetDefault("log-data-connection.db-name", "ape-volo-admin-log")
	v.SetDefault("log-data-connection.max-idle-conns", 1)
	v.SetDefault("log-data-connection.max-open-conns", 1)
	v.SetDefault("log-storage.time-zone", "Asia/Shanghai")
	v.SetDefault("log-queue.capacity", 10000)
	v.SetDefault("log-queue.batch-size", 200)
	v.SetDefault("log-queue.write-timeout-seconds", 5)
	v.SetDefault("log-queue.shutdown-timeout-seconds", 30)
	v.SetDefault("swagger.enabled", true)
	v.SetDefault("swagger.name", "v1")
	v.SetDefault("swagger.version", "3.6.2")
	v.SetDefault("swagger.title", "Ape-Volo-Admin-Go 接口文档")
	v.SetDefault("swagger.contact-name", "xianhc")
	v.SetDefault("swagger.contact-email", "apevolo@gamil.com")
	v.SetDefault("swagger.contact-url", "https://www.apevolo.com")
	// 查询缓存：默认开启，保持升级前行为
	v.SetDefault("aop.cache", true)
	v.SetDefault("aop.transactions", true)
	// 单机模式默认：不依赖 Redis
	v.SetDefault("system.use-redis", false)
	// 密码传输私钥的默认位置，缺失时启动会直接报错而不是运行期才暴露
	v.SetDefault("rsa.private-key-path", "resource/keys/password-private.pem")
	// 日志单文件上限 100MB，0 表示不限制
	v.SetDefault("zap.max-size", 100)
	// 上传文件大小上限（MB）
	v.SetDefault("system.file-limit-size", 10)
	// 默认用户密码，未配置时使用
	v.SetDefault("system.user-default-password", "123456")
	// 验证码尺寸：配置缺项时若为 0，画出来的图片是空白的
	v.SetDefault("captcha.code-length", 6)
	v.SetDefault("captcha.image-width", 240)
	v.SetDefault("captcha.image-height", 80)
	v.SetDefault("captcha.font-size", 48)
	v.SetDefault("captcha.counter-ttl-seconds", 3600)
}

// Viper 在启动时选择并加载配置文件，再应用环境变量和缺省值。
// 配置值优先级为环境变量、配置文件、缺省值；修改文件后需重启服务。
func Viper(path ...string) *viper.Viper {
	var config string

	if len(path) == 0 {
		flag.StringVar(&config, "c", "", "choose config file.")
		flag.Parse()
		if config == "" { // 判断命令行参数是否为空
			if configEnv := os.Getenv(internal.ConfigEnv); configEnv == "" { // 判断 internal.ConfigEnv 常量存储的环境变量是否为空
				switch gin.Mode() {
				case gin.DebugMode:
					config = internal.ConfigDefaultFile
					fmt.Printf("您正在使用gin模式的%s环境,config的路径为%s\n", gin.DebugMode, internal.ConfigDefaultFile)
				case gin.ReleaseMode:
					config = internal.ConfigReleaseFile
					fmt.Printf("您正在使用gin模式的%s环境,config的路径为%s\n", gin.ReleaseMode, internal.ConfigReleaseFile)
				case gin.TestMode:
					config = internal.ConfigTestFile
					fmt.Printf("您正在使用gin模式的%s环境,config的路径为%s\n", gin.TestMode, internal.ConfigTestFile)
				}
			} else { // internal.ConfigEnv 常量存储的环境变量不为空 将值赋值于config
				config = configEnv
				fmt.Printf("您正在使用%s环境变量,config的路径为%s\n", internal.ConfigEnv, config)
			}
		} else { // 命令行参数不为空 将值赋值于config
			fmt.Printf("您正在使用命令行的-c参数传递的值,config的路径为%s\n", config)
		}
	} else { // 函数传递的可变参数的第一个值赋值于config
		config = path[0]
		fmt.Printf("您正在使用func Viper()传递的值,config的路径为%s\n", config)
	}

	v := viper.New()
	v.SetConfigFile(config)
	v.SetConfigType("yaml")

	// 环境变量覆盖：APEVOLO_JWT_AUTH_OPTIONS_SECURITY_KEY 之类的键可覆盖配置文件，
	// 用于把密钥、数据库密码从仓库中剥离（部署时由容器注入）。
	v.SetEnvPrefix("APEVOLO")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	v.AutomaticEnv()

	// 为缺失的配置项设置默认值。
	setDefaults(v)

	err := v.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("配置文件致命错误: %s \n", err))
	}
	if v.IsSet("system.db-type") {
		panic("不支持 system.db-type，请使用 system.main-database 和 data-connection.connections")
	}
	for _, key := range []string{"mysql", "mssql", "sqlite", "oracle", "pgsql"} {
		if v.InConfig(key) {
			panic("数据库连接必须配置在 data-connection.connections 中")
		}
	}
	bindConnectionEnv(v)
	var cfg serverconfig.Server
	if err = v.Unmarshal(&cfg); err != nil {
		panic(fmt.Errorf("解析配置失败: %w", err))
	}
	if err = validateConfig(cfg); err != nil {
		panic(err)
	}
	global.Config = cfg

	return v
}

func validateConfig(cfg serverconfig.Server) error {
	if cfg.System.Port < 1 || cfg.System.Port > 65535 {
		return fmt.Errorf("system.port 必须在 1 到 65535 之间")
	}
	if err := cfg.ValidateConnections(); err != nil {
		return err
	}
	if _, err := time.LoadLocation(cfg.LogStorage.TimeZone); err != nil {
		return fmt.Errorf("log-storage.time-zone 无效: %w", err)
	}
	if cfg.LogQueue.Capacity < 1 || cfg.LogQueue.BatchSize < 1 || cfg.LogQueue.BatchSize > 200 || cfg.LogQueue.WriteTimeoutSeconds < 1 || cfg.LogQueue.ShutdownTimeoutSeconds < 1 {
		return fmt.Errorf("log-queue 容量及超时须为正数，batch-size 必须在 1 到 200 之间")
	}
	if cfg.System.UseRedis && strings.TrimSpace(cfg.Redis.Addr) == "" {
		return fmt.Errorf("开启 Redis 时 redis.addr 不能为空")
	}
	if cfg.Redis.DB < 0 {
		return fmt.Errorf("redis.db 不能小于零")
	}
	if cfg.System.FileLimitSize < 0 {
		return fmt.Errorf("system.file-limit-size 不能小于零")
	}
	if cfg.Swagger.Enabled && !regexp.MustCompile(`^[A-Za-z0-9_-]+$`).MatchString(cfg.Swagger.Name) {
		return fmt.Errorf("swagger.name 只能包含字母、数字、下划线和短横线，且不能为空")
	}
	return nil
}

func bindConnectionEnv(v *viper.Viper) {
	prefixes := []string{"log-data-connection"}
	ids := map[string]bool{"sqlite": true, "mysql": true, "mssql": true, "pgsql": true, "oracle": true}
	for id := range v.GetStringMap("data-connection.connections") {
		ids[id] = true
	}
	for id := range ids {
		prefixes = append(prefixes, "data-connection.connections."+id)
	}
	var bind func(string, reflect.Type)
	bind = func(prefix string, t reflect.Type) {
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			if f.Anonymous {
				bind(prefix, f.Type)
				continue
			}
			key := strings.Split(f.Tag.Get("mapstructure"), ",")[0]
			if key != "" {
				_ = v.BindEnv(prefix + "." + key)
			}
		}
	}
	for _, prefix := range prefixes {
		bind(prefix, reflect.TypeOf(serverconfig.Connection{}))
	}
}
