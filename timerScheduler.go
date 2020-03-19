/// Author:koangel
/// jackliu100@gmail.com
/// 调度器负责单独的timer处理以及调度器行为
package grapeTimer

import (
	"container/list"
	"fmt"
	"log"
	"runtime"
	"sync"
	"time"
)

const (
	LoopOnce    = 1
	UnlimitLoop = -1
)

// 开启日志调试模式 默认开启
// 错误类信息不受该控制
var CDebugMode bool = true

// 创建GO去执行到期的任务 默认开启
var UseAsyncExec bool = true

// 如果上次任务未执行完成，默认方式还是跳过
// 默认跳过,单协程模式下，不处理
var SkipWaitTask bool = true

// 对于默认Panic Recover函数处理
var RecoverPanic = func(err error) {
	fmt.Println(err.Error())
}

// 默认的通用时区字符串，通过修改他会更改分析后的日期结果
// 默认为上海时区
var LocationFormat = "Asia/Shanghai"

type GrapeScheduler struct {
	done chan bool // 是否关闭

	schedulerTimer *time.Ticker

	timerContiner *list.List
	autoId        int // 自动计数的Id

	listLocker sync.Mutex
}

var GScheduler *GrapeScheduler = nil

// 初始化全局整个调度器
// 调度器的粒度不建议小于1秒，会导致Cpu爆炸
// 不建议低于1秒钟,如果低于100毫秒 则自动设置为100毫秒有效防止CPU爆炸
// ars = auto set runtime,自动设置Cpu数量
func InitGrapeScheduler(t time.Duration, ars bool) {

	if ars {
		runtime.GOMAXPROCS(runtime.NumCPU()) // 启动时钟时 自动设置Go的最大执行数，以便提高性能
	}

	chkTick := t
	if chkTick <= (time.Millisecond * 100) {
		chkTick = time.Duration(time.Millisecond * 100)
	}

	GScheduler = &GrapeScheduler{
		done:           make(chan bool),
		timerContiner:  list.New(),
		autoId:         1000,
		schedulerTimer: time.NewTicker(chkTick),
	}

	go GScheduler.procScheduler() // 启动执行线程
	//go GScheduler.procAddTimer()
}

// 停止这个timer
func (c *GrapeScheduler) StopTimer(Id int) {
	c.listLocker.Lock()
	defer c.listLocker.Unlock()

	for e := c.timerContiner.Front(); e != nil; e = e.Next() {
		vnTimer := e.Value.(*GrapeTimer)
		if vnTimer.IsDestroy() {
			continue
		}

		if vnTimer.Id == Id {
			vnTimer.Stop()
			return
		}
	}
}

// 列出全部timer的下一次执行周期
func (c *GrapeScheduler) List() map[int]string {
	var temp map[int]string = map[int]string{}

	c.listLocker.Lock()
	defer c.listLocker.Unlock()

	for e := c.timerContiner.Front(); e != nil; e = e.Next() {
		vnTimer := e.Value.(*GrapeTimer)
		if vnTimer.IsDestroy() {
			continue
		}

		temp[vnTimer.Id] = vnTimer.String()
	}

	return temp
}

func (c *GrapeScheduler) String(id int) string {
	c.listLocker.Lock()
	defer c.listLocker.Unlock()

	for e := c.timerContiner.Front(); e != nil; e = e.Next() {
		vnTimer := e.Value.(*GrapeTimer)
		if vnTimer.IsDestroy() {
			continue
		}

		if vnTimer.Id == id {
			return vnTimer.String()
		}
	}

	return ""
}

func (c *GrapeScheduler) Format(id int, layout string) string {
	c.listLocker.Lock()
	defer c.listLocker.Unlock()

	for e := c.timerContiner.Front(); e != nil; e = e.Next() {
		vnTimer := e.Value.(*GrapeTimer)
		if vnTimer.IsDestroy() {
			continue
		}

		if vnTimer.Id == id {
			return vnTimer.Format(layout)
		}
	}

	return ""
}

func (c *GrapeScheduler) ToJson(id int) string {
	c.listLocker.Lock()
	defer c.listLocker.Unlock()

	for e := c.timerContiner.Front(); e != nil; e = e.Next() {
		vnTimer := e.Value.(*GrapeTimer)
		if vnTimer.IsDestroy() {
			continue
		}

		if vnTimer.Id == id {
			return vnTimer.toJson() // 参数转为JSON
		}
	}

	return ""
}

func (c *GrapeScheduler) SaveAll() []string {
	c.listLocker.Lock()
	defer c.listLocker.Unlock()

	var jsonArr []string = []string{}

	for e := c.timerContiner.Front(); e != nil; e = e.Next() {
		vnTimer := e.Value.(*GrapeTimer)
		if vnTimer.IsDestroy() {
			continue
		}

		jsonArr = append(jsonArr, vnTimer.toJson()) // 参数转为JSON
	}

	return jsonArr
}

func (c *GrapeScheduler) procScheduler() {
	defer func() {
		close(c.done)

		c.schedulerTimer.Stop()
	}()

	for {
		select {
		case <-c.schedulerTimer.C:
			c.listLocker.Lock()
			if CDebugMode {
				log.Printf("[grapeTimer] Timer TickOnce |time:%v|", time.Now())
			}

			var nextE *list.Element = nil
			for e := c.timerContiner.Front(); e != nil; e = nextE {
				nextE = e.Next()
				vnTimer := e.Value.(*GrapeTimer)
				vnTimer.Execute()
				if vnTimer.IsDestroy() {
					if CDebugMode {
						log.Printf("[grapeTimer] Timer RemoveId:%v |time:%v|", vnTimer.Id, time.Now())
					}
					c.timerContiner.Remove(e) // 直接删除
				}
			}

			if CDebugMode {
				log.Printf("[grapeTimer] Timer TickOnce End |time:%v|", time.Now())
			}
			c.listLocker.Unlock()
			break
		case <-c.done:
			return
		}
	}
}
