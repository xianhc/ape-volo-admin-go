package ipregion

import (
	"errors"
	"fmt"
	"net/netip"
	"strings"
	"sync"

	"github.com/lionsoul2014/ip2region/binding/golang/xdb"
)

// Region 保存统一的 IP 归属地信息。
//
// 数据库中的未知字段（0）统一转为空字符串，内网地址的 Country 为“内网IP”。
type Region struct {
	Country     string `json:"country"`
	Province    string `json:"province"`
	City        string `json:"city"`
	ISP         string `json:"isp"`
	CountryCode string `json:"countryCode"`
}

// String 返回竖线分隔的归属地文本。
//
// 无查询结果时返回空字符串，避免把分隔符写入日志。
func (r Region) String() string {
	if r == (Region{}) {
		return ""
	}
	return strings.Join([]string{r.Country, r.Province, r.City, r.ISP, r.CountryCode}, "|")
}

type database struct {
	mu       sync.Mutex
	searcher *xdb.Searcher
}

// Service 提供 IPv4 和 IPv6 归属地查询。
//
// 两份数据库常驻内存，每个搜索器串行访问，实例可安全地供多个请求共用。
type Service struct {
	v4 *database
	v6 *database
}

// New 从指定的两份 xdb 文件创建归属地服务。
//
// 每份文件只读取一次，构造成功后查询不再访问文件系统；调用方应复用该实例。
func New(v4Path, v6Path string) (*Service, error) {
	v4, err := load(v4Path, xdb.IPv4)
	if err != nil {
		return nil, err
	}
	v6, err := load(v6Path, xdb.IPv6)
	if err != nil {
		return nil, err
	}
	return &Service{v4: v4, v6: v6}, nil
}

func load(path string, version *xdb.Version) (*database, error) {
	content, err := xdb.LoadContentFromFile(path)
	if err != nil {
		return nil, fmt.Errorf("加载归属地数据库 %s: %w", path, err)
	}
	header, err := xdb.LoadHeaderFromBuff(content)
	if err != nil {
		return nil, fmt.Errorf("读取归属地数据库头 %s: %w", path, err)
	}
	actual, err := xdb.VersionFromHeader(header)
	if err != nil || actual != version || header.IndexPolicy != xdb.VectorIndexPolicy ||
		(header.Version == xdb.Structure30 && header.RuntimePtrBytes != 4) ||
		len(content) < xdb.HeaderInfoLength+xdb.VectorIndexRows*xdb.VectorIndexCols*xdb.VectorIndexSize {
		return nil, fmt.Errorf("归属地数据库格式或 IP 版本不匹配: %s", path)
	}
	searcher, err := xdb.NewWithBuffer(version, content)
	if err != nil {
		return nil, fmt.Errorf("创建归属地搜索器 %s: %w", path, err)
	}
	return &database{searcher: searcher}, nil
}

// Lookup 根据 IP 地址类型查询对应的归属地数据库。
//
// 映射 IPv4 地址按 IPv4 查询；私有、回环、链路本地地址返回内网标记，
// 未指定或组播地址返回空结果，无效地址及数据库错误返回 error。
func (s *Service) Lookup(ip string) (Region, error) {
	addr, err := netip.ParseAddr(strings.TrimSpace(ip))
	if err != nil {
		return Region{}, errors.New("无效的 IP 地址")
	}
	addr = addr.Unmap()
	if addr.IsPrivate() || addr.IsLoopback() || addr.IsLinkLocalUnicast() {
		return Region{Country: "内网IP"}, nil
	}
	if addr.IsUnspecified() || addr.IsMulticast() {
		return Region{}, nil
	}
	if s == nil {
		return Region{}, errors.New("归属地服务未初始化")
	}
	db := s.v6
	if addr.Is4() {
		db = s.v4
	}
	if db == nil || db.searcher == nil {
		return Region{}, errors.New("归属地数据库未初始化")
	}
	raw, err := db.lookup(addr.WithZone("").AsSlice())
	if err != nil {
		return Region{}, err
	}
	return parseRegion(raw)
}

func (db *database) lookup(ip []byte) (region string, err error) {
	db.mu.Lock()
	defer db.mu.Unlock()
	// 损坏的数据库指针可能导致越界，将其转为查询错误，避免中断请求。
	defer func() {
		if recover() != nil {
			region = ""
			err = errors.New("归属地数据库损坏")
		}
	}()
	return db.searcher.Search(ip)
}

func parseRegion(raw string) (Region, error) {
	if raw == "" {
		return Region{}, nil
	}
	fields := strings.Split(raw, "|")
	if len(fields) != 5 {
		return Region{}, errors.New("归属地结果格式无效")
	}
	for i := range fields {
		fields[i] = strings.TrimSpace(fields[i])
		if fields[i] == "0" {
			fields[i] = ""
		}
	}
	return Region{Country: fields[0], Province: fields[1], City: fields[2], ISP: fields[3], CountryCode: fields[4]}, nil
}
