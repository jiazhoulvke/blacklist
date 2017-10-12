package blacklist

import (
	"sort"
	"sync"
	"time"

	"github.com/labstack/echo"
)

//BlackListFunc 当ip属于黑名单中时调用的函数
var BlackListFunc = func(c echo.Context) error {
	return c.HTML(200, "You're Not Allowed to Visit!")
}

var ipblacklist *IPBlackList

//Item 黑名单项目
type Item struct {
	//IP IP地址
	IP string
	//EndTime unix时间戳格式的封锁结束时间,0或负数表示永久
	EndTime int64
}

type iplist []Item

func (l iplist) Len() int {
	return len(l)
}
func (l iplist) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}
func (l iplist) Less(i, j int) bool {
	if l[i].EndTime < 1 {
		return false
	}
	return l[i].EndTime < l[j].EndTime
}

//IPBlackList ip data
type IPBlackList struct {
	mux  sync.RWMutex
	data map[string]Item
	list iplist
}

func init() {
	ipblacklist = New()
}

//New 新的黑名单
func New() *IPBlackList {
	l := IPBlackList{
		data: make(map[string]Item),
		list: make([]Item, 0, 64),
	}
	//开启一个协程定时删除过期的ip
	go func(*IPBlackList) {
		for {
			now := time.Now().Unix()
			l.mux.Lock()
			for _, item := range l.list {
				if item.EndTime < 1 {
					break
				}
				if item.EndTime < now {
					delete(l.data, item.IP)
					l.list = remove(l.list, item.IP)
				} else {
					sort.Sort(l.list)
					break
				}
				sort.Sort(l.list)
			}
			l.mux.Unlock()
			time.Sleep(time.Second)
		}
	}(&l)
	return &l
}

//remove 删除ip列表中指定的项目
func remove(slice []Item, ip string) []Item {
	for i := 0; i < len(slice); i++ {
		if slice[i].IP == ip {
			slice = append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}

//Add 添加IP到黑名单
//t 表示需要封锁的秒数，0表示永久
func (i *IPBlackList) Add(ip string, t int64) error {
	i.mux.Lock()
	defer i.mux.Unlock()
	if t > 0 {
		t += time.Now().Unix()
	}
	item := Item{
		IP:      ip,
		EndTime: t,
	}
	i.data[ip] = item
	i.list = remove(i.list, ip)
	i.list = append(i.list, item)
	sort.Sort(i.list)
	return nil
}

//Del 删除IP
func (i *IPBlackList) Del(ip string) error {
	i.mux.Lock()
	delete(i.data, ip)
	i.list = remove(i.list, ip)
	if len(i.list) > 1 {
		sort.Sort(i.list)
	}
	i.mux.Unlock()
	return nil
}

//Exist 判断IP是否在黑名单中
func (i *IPBlackList) Exist(ip string) bool {
	_, exist := i.data[ip]
	return exist
}

//List 黑名单列表
func (i *IPBlackList) List() []Item {
	return i.list
}

//Add 添加ip到黑名单
func Add(ip string, t int64) error {
	return ipblacklist.Add(ip, t)
}

//Del 从黑名单中删除IP
func Del(ip string) error {
	return ipblacklist.Del(ip)
}

//List 获取黑名单列表
func List() []Item {
	return ipblacklist.List()
}

//BlackList 黑名单
func BlackList(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if ipblacklist.Exist(c.RealIP()) {
			return BlackListFunc(c)
		}
		return next(c)
	}
}
