package service

import "sync"

type reportExportCache struct {
	mu    sync.RWMutex
	items map[string]ReportExport
}

func newReportExportCache() *reportExportCache {
	return &reportExportCache{items: map[string]ReportExport{}}
}

func (c *reportExportCache) Get(key string) (ReportExport, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, ok := c.items[key]
	if !ok {
		return ReportExport{}, false
	}
	return cloneReportExport(item), true
}

func (c *reportExportCache) Set(key string, item ReportExport) ReportExport {
	c.mu.Lock()
	defer c.mu.Unlock()

	item = cloneReportExport(item)
	c.items[key] = item
	return cloneReportExport(item)
}

func cloneReportExport(item ReportExport) ReportExport {
	if item.Body == nil {
		return item
	}
	body := make([]byte, len(item.Body))
	copy(body, item.Body)
	item.Body = body
	return item
}
