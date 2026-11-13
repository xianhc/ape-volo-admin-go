// Package routeauth 把私有路由注册与授权策略声明合并为一次操作。
//
// 私有路由必须通过 Group 注册；直接使用 Gin 注册的私有路由不会进入策略表，
// 运行期权限中间件将按 fail closed 拒绝访问。
package routeauth

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
)

// Policy 描述访问一条私有路由需要的权限码和附加角色码。
//
// Policy 只能通过 LoginOnly 或 Require 创建，零值不能用于注册路由。
type Policy struct {
	code     string
	roles    []string
	declared bool
}

// LoginOnly 声明路由只要求有效登录会话。
//
// 条件型角色检查仍可在 Service 中读取中间件写入 context 的角色快照。
func LoginOnly() Policy {
	return Policy{declared: true}
}

// Require 声明路由要求权限码，并可追加一个或多个角色码。
//
// 权限码和角色码会去除首尾空白；空权限码必须改用 LoginOnly。
func Require(code string, roles ...string) Policy {
	code = strings.TrimSpace(code)
	if code == "" {
		panic("routeauth: empty permission code; use LoginOnly")
	}
	cleanRoles := make([]string, 0, len(roles))
	seen := make(map[string]struct{}, len(roles))
	for _, role := range roles {
		role = strings.TrimSpace(role)
		if role == "" {
			panic("routeauth: empty role code")
		}
		if _, exists := seen[role]; exists {
			continue
		}
		seen[role] = struct{}{}
		cleanRoles = append(cleanRoles, role)
	}
	return Policy{code: code, roles: cleanRoles, declared: true}
}

// Code 返回策略要求的权限码。
//
// LoginOnly 策略返回空字符串。
func (policy Policy) Code() string { return policy.code }

// Roles 返回策略要求的角色码只读副本。
//
// 调用方修改返回切片不会改变注册表中的策略。
func (policy Policy) Roles() []string { return append([]string(nil), policy.roles...) }

// Rule 描述注册表中的一条方法、路径与授权策略。
//
// Path 不包含 system.router-prefix，供中间件按 Gin 路由模板匹配。
type Rule struct {
	Method string
	Path   string
	Policy Policy
}

// Group 包装 Gin 路由组，在注册路由的同时登记授权策略。
//
// Group 保留 Gin 的分组与中间件行为，但不提供无策略的 Handle 方法。
type Group struct {
	routes   *gin.RouterGroup
	basePath string
}

var policies = struct {
	sync.RWMutex
	rules map[string]Rule
}{rules: make(map[string]Rule)}

// NewGroup 创建私有路由授权组。
//
// root 应已经挂载 JWT、权限和事务中间件；逻辑路径从 root 下一级开始计算。
func NewGroup(root *gin.RouterGroup) *Group {
	if root == nil {
		panic("routeauth: nil gin router group")
	}
	return &Group{routes: root}
}

// Group 创建子路由组并继承授权注册表。
//
// handlers 按 Gin 组级中间件语义执行，不参与策略匹配。
func (group *Group) Group(relativePath string, handlers ...gin.HandlerFunc) *Group {
	return &Group{
		routes:   group.routes.Group(relativePath, handlers...),
		basePath: joinPath(group.basePath, relativePath),
	}
}

// GET 注册带授权策略的 GET 路由。
//
// 策略先写入注册表，再由 Gin 完成路由注册。
func (group *Group) GET(path string, policy Policy, handlers ...gin.HandlerFunc) gin.IRoutes {
	return group.handle("GET", path, policy, handlers...)
}

// POST 注册带授权策略的 POST 路由。
//
// 策略先写入注册表，再由 Gin 完成路由注册。
func (group *Group) POST(path string, policy Policy, handlers ...gin.HandlerFunc) gin.IRoutes {
	return group.handle("POST", path, policy, handlers...)
}

// PUT 注册带授权策略的 PUT 路由。
//
// 策略先写入注册表，再由 Gin 完成路由注册。
func (group *Group) PUT(path string, policy Policy, handlers ...gin.HandlerFunc) gin.IRoutes {
	return group.handle("PUT", path, policy, handlers...)
}

// DELETE 注册带授权策略的 DELETE 路由。
//
// 策略先写入注册表，再由 Gin 完成路由注册。
func (group *Group) DELETE(path string, policy Policy, handlers ...gin.HandlerFunc) gin.IRoutes {
	return group.handle("DELETE", path, policy, handlers...)
}

func (group *Group) handle(method, relativePath string, policy Policy, handlers ...gin.HandlerFunc) gin.IRoutes {
	if !policy.declared {
		panic("routeauth: undeclared route policy")
	}
	register(method, joinPath(group.basePath, relativePath), policy)
	return group.routes.Handle(method, relativePath, handlers...)
}

func register(method, path string, policy Policy) {
	method = strings.ToUpper(method)
	key := method + " " + path
	rule := Rule{Method: method, Path: path, Policy: clonePolicy(policy)}
	policies.Lock()
	defer policies.Unlock()
	if existing, exists := policies.rules[key]; exists {
		if samePolicy(existing.Policy, policy) {
			return
		}
		panic(fmt.Sprintf("routeauth: conflicting policy for %s", key))
	}
	policies.rules[key] = rule
}

// Lookup 按 HTTP 方法与无前缀路由模板读取授权策略。
//
// 未登记返回 false，权限中间件必须据此 fail closed。
func Lookup(method, path string) (Policy, bool) {
	policies.RLock()
	defer policies.RUnlock()
	rule, exists := policies.rules[strings.ToUpper(method)+" "+path]
	return clonePolicy(rule.Policy), exists
}

// Rules 返回全部已注册授权规则的稳定排序副本。
//
// 该结果仅用于启动自检、种子一致性检查和测试。
func Rules() []Rule {
	policies.RLock()
	rules := make([]Rule, 0, len(policies.rules))
	for _, rule := range policies.rules {
		rule.Policy = clonePolicy(rule.Policy)
		rules = append(rules, rule)
	}
	policies.RUnlock()
	sort.Slice(rules, func(i, j int) bool {
		if rules[i].Method == rules[j].Method {
			return rules[i].Path < rules[j].Path
		}
		return rules[i].Method < rules[j].Method
	})
	return rules
}

func joinPath(base, relative string) string {
	joined := strings.Trim(strings.TrimSpace(base)+"/"+strings.TrimSpace(relative), "/")
	if joined == "" {
		return "/"
	}
	return "/" + joined
}

func clonePolicy(policy Policy) Policy {
	policy.roles = append([]string(nil), policy.roles...)
	return policy
}

func samePolicy(left, right Policy) bool {
	if left.declared != right.declared || left.code != right.code || len(left.roles) != len(right.roles) {
		return false
	}
	for index := range left.roles {
		if left.roles[index] != right.roles[index] {
			return false
		}
	}
	return true
}
