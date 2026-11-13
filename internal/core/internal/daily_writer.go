package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const logDateLayout = "2006-01-02"

// cleanupInterval 控制过期目录清理的最小间隔，避免每次刷盘都遍历目录。
const cleanupInterval = time.Hour

// DailyWriter 按天分目录、按大小切分的日志写入器。
//
// 输出结构：<dir>/<YYYY-MM-DD>/<level>.log，单文件超过上限时追加序号
// 变为 <level>.1.log、<level>.2.log。
type DailyWriter struct {
	mu      sync.Mutex
	dir     string
	level   string
	maxAge  int   // 目录保留天数，<=0 表示不清理
	maxSize int64 // 单文件上限（字节），<=0 表示不限制
	now     func() time.Time

	day         string
	file        *os.File
	size        int64
	seq         int
	lastCleanup time.Time
}

// NewDailyWriter 创建按天轮转的写入器。
// maxAgeDays 为目录保留天数，maxSizeMB 为单文件大小上限（MB）。
func NewDailyWriter(dir, level string, maxAgeDays, maxSizeMB int) *DailyWriter {
	w := &DailyWriter{
		dir:    dir,
		level:  level,
		maxAge: maxAgeDays,
		now:    time.Now,
	}
	if maxSizeMB > 0 {
		w.maxSize = int64(maxSizeMB) << 20
	}
	return w
}

// Write 实现 io.Writer。
func (w *DailyWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	day := w.now().Format(logDateLayout)
	switch {
	case w.file == nil || day != w.day:
		// 首次写入或跨天：从当天第 0 个文件重新开始
		if err := w.open(day, 0); err != nil {
			return 0, err
		}
	case w.maxSize > 0 && w.size+int64(len(p)) > w.maxSize:
		// 超过单文件上限：切到下一个序号
		if err := w.open(day, w.seq+1); err != nil {
			return 0, err
		}
	}

	n, err := w.file.Write(p)
	w.size += int64(n)
	return n, err
}

// Sync 刷盘，并按间隔清理过期目录。
func (w *DailyWriter) Sync() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.file == nil {
		return nil
	}
	if err := w.file.Sync(); err != nil {
		return err
	}

	if w.maxAge > 0 && w.now().Sub(w.lastCleanup) >= cleanupInterval {
		w.lastCleanup = w.now()
		w.cleanup()
	}
	return nil
}

// open 切换到指定日期与序号的日志文件。调用方需持有锁。
func (w *DailyWriter) open(day string, seq int) error {
	if w.file != nil {
		_ = w.file.Sync()
		_ = w.file.Close()
		w.file = nil
	}

	path := w.path(day, seq)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	// 追加模式：同一天内进程重启不会覆盖已有日志
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	info, err := f.Stat()
	if err != nil {
		_ = f.Close()
		return err
	}

	w.file, w.day, w.seq, w.size = f, day, seq, info.Size()
	return nil
}

// path 返回指定日期与序号的日志路径。
func (w *DailyWriter) path(day string, seq int) string {
	name := w.level + ".log"
	if seq > 0 {
		name = fmt.Sprintf("%s.%d.log", w.level, seq)
	}
	return filepath.Join(w.dir, day, name)
}

// cleanup 删除超过保留天数的日期目录。调用方需持有锁。
func (w *DailyWriter) cleanup() {
	entries, err := os.ReadDir(w.dir)
	if err != nil {
		return
	}

	cutoff := w.now().AddDate(0, 0, -w.maxAge)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		day, err := time.ParseInLocation(logDateLayout, entry.Name(), time.Local)
		if err != nil {
			continue // 非日期目录，不处理
		}
		if day.Before(cutoff) {
			_ = os.RemoveAll(filepath.Join(w.dir, entry.Name()))
		}
	}
}

// Close 关闭当前文件，便于测试与进程退出时释放句柄。
func (w *DailyWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.file == nil {
		return nil
	}
	err := w.file.Close()
	w.file = nil
	return err
}
