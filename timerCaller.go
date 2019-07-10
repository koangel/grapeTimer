/// Author:koangel
/// jackliu100@gmail.com
/// 负责分析并解析字符串格式并转换成一个time
package grapeTimer

import "strconv"

// 以下函数均为自动ID
// 间隔为毫秒 运行一个tick并返回一个Id
func NewTickerOnce(tick int, fn GrapeExecFn, args ...interface{}) int {
	return NewTickerLoop(tick, LoopOnce, fn, args...)
}

// 通过json构造timer
func NewFromJson(json string, fn GrapeExecFn, args ...interface{}) int {
	nowId := GScheduler.autoId
	GScheduler.autoId++

	newTimer := newTimerFromJson(json, fn, args...)

	// 覆盖Id
	newTimer.Id = nowId // 防止ID出错

	GScheduler.listLocker.Lock()
	GScheduler.timerContiner.PushBack(newTimer)
	GScheduler.listLocker.Unlock()

	return nowId
}

// 循环可控版本
func NewTickerLoop(tick, count int, fn GrapeExecFn, args ...interface{}) int {
	nowId := GScheduler.autoId
	GScheduler.autoId++
	newTimer := newTimer(nowId,
		timerTickMode,
		count,
		strconv.FormatInt(int64(tick), 10),
		fn, args...)

	//GScheduler.appendTimer <- newTimer

	GScheduler.listLocker.Lock()
	GScheduler.timerContiner.PushBack(newTimer)
	GScheduler.listLocker.Unlock()

	return nowId
}

// 格式分析时钟
func NewTimeDataOnce(data string, fn GrapeExecFn, args ...interface{}) int {
	return NewTimeDataLoop(data, LoopOnce, fn, args...)
}

func NewTimeDataLoop(data string, count int, fn GrapeExecFn, args ...interface{}) int {
	nowId := GScheduler.autoId
	GScheduler.autoId++
	newTimer := newTimer(nowId,
		timerDateMode,
		count,
		data,
		fn, args...)

	GScheduler.listLocker.Lock()
	GScheduler.timerContiner.PushBack(newTimer)
	GScheduler.listLocker.Unlock()

	return nowId
}

func StopTimer(Id int) {
	GScheduler.StopTimer(Id)
}

func String(Id int) string {
	return GScheduler.String(Id)
}

func Format(Id int, layout string) string {
	return GScheduler.Format(Id, layout)
}

func List() map[int]string {
	return GScheduler.List()
}

// 所有的时间参数保存到JSON
func ToJson(Id int) string {
	return GScheduler.ToJson(Id)
}

func SaveAll() []string {
	return GScheduler.SaveAll()
}
