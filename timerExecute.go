/// Author:koangel
/// jackliu100@gmail.com
/// 每个单独的执行模块，结构中越简单越好
package grapeTimer

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"reflect"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

const (
	// 通过字符串转换下一次的日期
	timerDateMode = iota
	// 通过转成的Tick日期进行下一次时间计算
	timerTickMode
)

type GrapeExecFn interface{}

// 每个执行周期的执行者
// 每个时间执行器
// 可Json化并反Json化

type grapeCallFunc struct {
	exeCall GrapeExecFn
	args    []reflect.Value
	status  int32
	mux     sync.Mutex
}

type GrapeTimer struct {
	Id        int64  `json:"TimerId"`
	NextTime  int64  `json:"nextUnix"`
	RunMode   int    `json:"Mode"`
	TimeData  string `json:"timeData"`
	LoopCount int32  `json:"loopCount"` // -1为无限执行

	cbFunc     *grapeCallFunc
	tickSecond int // 临时使用 不可保存
	nextVTime  time.Time
}

func reflectFunc(fn GrapeExecFn, args ...interface{}) (cb *grapeCallFunc, err error) {
	cb = nil
	err = nil

	t := reflect.TypeOf(fn) // 获得对象类型,从而知道有多少个参数

	if t.Kind() != reflect.Func {
		err = errors.New("callback must be a function")
		return
	}

	argArr := []interface{}(args) // 先把参数都转成ARRAY

	if len(argArr) < t.NumIn() {
		err = errors.New("Not enough arguments")
		return
	}

	// 解析全部参数
	var in = make([]reflect.Value, t.NumIn()) //MAKE要保存的参数
	for i := 0; i < t.NumIn(); i++ {
		argType := t.In(i)
		if argType != reflect.TypeOf(argArr[i]) {
			err = errors.New(fmt.Sprintf("Value not found for type %v", argType))
			return
		}
		in[i] = reflect.ValueOf(argArr[i]) // 参数保存下来
	}

	cb = &grapeCallFunc{
		exeCall: fn,
		args:    in,
		status:  0,
	}
	return
}

func PanicTrace(kb int) []byte {
	s := []byte("/src/runtime/panic.go")
	e := []byte("\ngoroutine ")

	line := []byte("\n")
	stack := make([]byte, kb<<10) //4KB
	length := runtime.Stack(stack, true)
	start := bytes.Index(stack, s)
	stack = stack[start:length]
	start = bytes.Index(stack, line) + 1
	stack = stack[start:]
	end := bytes.LastIndex(stack, line)
	if end != -1 {
		stack = stack[:end]
	}
	end = bytes.Index(stack, e)
	if end != -1 {
		stack = stack[:end]
	}
	stack = bytes.TrimRight(stack, "\n")
	return stack
}

func callFunc(timer *grapeCallFunc) {
	// 修正BUG，防止各种异常
	defer func() {
		if p := recover(); p != nil {
			stacks := PanicTrace(4)
			panic := fmt.Sprintf("recover panics: %v call:%v", p, string(stacks))
			if RecoverPanic != nil {
				RecoverPanic(fmt.Errorf(panic))
			} else {
				fmt.Println(string(panic))
			}
		}
	}()

	if timer.status >= 1 && SkipWaitTask {
		return
	}

	// 否则只计数不处理
	atomic.AddInt32(&timer.status, 1)
	reflect.ValueOf(timer.exeCall).Call(timer.args) // 正确调用
	atomic.AddInt32(&timer.status, -1)
}

/// 直接创建一个timer 内部函数
func newTimer(Id int64, Mode, Count int, timeData string, fn GrapeExecFn, args ...interface{}) *GrapeTimer {

	cbo, err := reflectFunc(fn, args...)
	if err != nil {
		fmt.Printf("error:%v", err)
		return nil
	}

	newValue := &GrapeTimer{
		Id:         Id,
		RunMode:    Mode,
		TimeData:   timeData,
		LoopCount:  int32(Count),
		NextTime:   0,
		tickSecond: 0,

		cbFunc: cbo,
	}

	newValue.makeNextTime()

	return newValue
}

/// 通过Json创建一个Timer 内部函数
func newTimerFromJson(s string, fn GrapeExecFn, args ...interface{}) *GrapeTimer {
	cbo, err := reflectFunc(fn, args...)
	if err != nil {
		fmt.Printf("error:%v", err)
		return nil
	}

	newValue := &GrapeTimer{}
	newValue = newValue.ParserJson(s)
	if newValue != nil {
		newValue.cbFunc = cbo
		newValue.makeNextTime()
	}

	return newValue
}

func (c *GrapeTimer) String() string {
	return c.nextVTime.Format(TimeFormat)
}

func (c *GrapeTimer) Format(layout string) string {
	return c.nextVTime.Format(layout)
}

/// 执行Timer
func (c *GrapeTimer) Execute() {
	if c.IsDestroy() {
		return // 销毁中的不可执行
	}

	if c.IsExpired() {
		if CDebugMode {
			log.Printf("[grapeTimer] Timer Execute:%v |time:%v| Begin", c.Id, time.Now())
		}

		// 执行一下
		if UseAsyncExec {
			go callFunc(c.cbFunc)
		} else {
			callFunc(c.cbFunc)
		}

		if CDebugMode {
			log.Printf("[grapeTimer] Timer Execute:%v |time:%v| End", c.Id, time.Now())
		}

		c.nextTime() // 计数
	}
}

/// 是否到时间
func (c *GrapeTimer) IsExpired() bool {
	if time.Now().Unix() >= c.NextTime {
		if CDebugMode {
			log.Printf("[grapeTimer] Timer Expired:%v |time:%v|", c.Id, time.Now())
		}
		return true
	}

	return false
}

/// 是否可销毁
func (c *GrapeTimer) IsDestroy() bool {
	val := atomic.LoadInt32(&c.LoopCount)
	if val == -1 {
		return false
	}

	if val == 0 {
		return true
	}

	return false
}

func (c *GrapeTimer) makeNextTime() {
	if CDebugMode {
		log.Printf("[grapeTimer] Timer Id:%v |time:%v|LoopCount:%v|", c.Id, time.Now(), c.LoopCount)
	}

	switch c.RunMode {
	case timerDateMode:
		vtime, err := Parser(c.TimeData)
		if err != nil {
			atomic.StoreInt32(&c.LoopCount, 0) // 出错销毁
			log.Printf("[grapeTimer] Timer Id:%v |time:%v|Error:%v|", c.Id, time.Now(), err)
			return
		}

		c.nextVTime = *vtime
		c.NextTime = vtime.Unix() // 生成下一次时间
		break
	case timerTickMode:
		if c.tickSecond == 0 {
			c.tickSecond, _ = strconv.Atoi(c.TimeData)
		}

		vtime := time.Now().Add(time.Duration(c.tickSecond) * time.Millisecond)
		c.nextVTime = vtime
		c.NextTime = vtime.Unix()
		break
	}
}

/// 计算下一次时间
func (c *GrapeTimer) nextTime() {
	if c.IsDestroy() {
		return
	}

	val := atomic.LoadInt32(&c.LoopCount)
	// 先进行计数
	if val != -1 {
		atomic.AddInt32(&c.LoopCount, -1)
		val = atomic.LoadInt32(&c.LoopCount)
		if val <= 0 {
			val = 0
		}
	}

	c.makeNextTime()
}

func (c *GrapeTimer) Stop() {
	atomic.StoreInt32(&c.LoopCount, 0) // 停止这个timer继续执行
}

func (c *GrapeTimer) toJson() string {
	bJson, err := json.Marshal(c)
	if err != nil {
		log.Printf("[grapeTimer] toJson:%v|Error:%v|", c.Id, err)
		return ""
	}

	return string(bJson)
}

func (c *GrapeTimer) ParserJson(s string) *GrapeTimer {
	err := json.Unmarshal([]byte(s), c)
	if err != nil {
		log.Printf("[grapeTimer] ParserJson:%v|Error:%v|", c.Id, err)
		return nil
	}

	return c
}
