```
                                   ,--------.,--.                         ,----.           
 ,---. ,--.--. ,--,--. ,---.  ,---.'--.  .--'`--',--,--,--. ,---. ,--.--.'  .-./    ,---.  
| .-. ||  .--'' ,-.  || .-. || .-. :  |  |   ,--.|        || .-. :|  .--'|  | .---.| .-. | 
' '-' '|  |   \ '-'  || '-' '\   --.  |  |   |  ||  |  |  |\   --.|  |   '  '--'  |' '-' ' 
.`-  / `--'    `--`--'|  |-'  `----'  `--'   `--'`--`--`--' `----'`--'    `------'  `---'  
`---'                 `--'                                              grapeTimer for time scheduling 
```

[![Go Report Card](https://goreportcard.com/badge/github.com/koangel/grapeTimer)](https://goreportcard.com/report/github.com/koangel/grapeTimer)  [![Build Status](https://secure.travis-ci.org/koangel/grapeTimer.png)](http://travis-ci.org/koangel/grapeTimer)

---
# **grapeTimer 时间调度器**

一款粗粒度的时间调度器，可以帮你通过一些字符串快速并简单的创建时间任务。

用于游戏服务端的优化设计，大量并行的时间调度方式。

目前可支持任意类型的函数(无返回值)以及任意参数数量和参数类型。

grapeTimer填坑完毕的库，已做测试。

- Author: Koangel
- Weibo: [@koangel](http://weibo.com/koangel)
- Homepage: [个人博客](http://grapec.me)

#### 简单介绍：
- 通过命令格式创建`time.Time`
- 简洁的Api格式，轻度且可拆分的函数库
- 快速创建调度器
- 可控的调度器时间粒度
- 高性能的并发调度
- 支持任意类型函数的任意参数`[自动推导参数以及类型]`
- 时间周期，次数多模式可控`[支持每天、每周、每月]`
- 可以获取下一次执行时间的字符串`[支持自定义格式]`
- 可选择对调度器保存或内存执行
- 生成可保存的调度器字符串并反向分析他生成调度器`[保存到Json再通过Json创建Timer]`
- 不依赖第三方库
- 处理Panic打印相关的数据信息，用于记录崩溃原因
- 通过atomic预测下一个ID`[该方法不适合并发场景，可能存在争抢ID错误的风险]`
- 自定义起始TimerId的种子
- 自定义TimerId的生成函数`[自生成ID请注意并发场景下的线程争抢]`
- TimerId扩展为int64，支持大ID和CRC64ID生成器对应

## **简单测试**

100W个TIMER压入4秒钟，每个执行1S的话，完成执行大概不到4S，基本上达到性能需求。
如果超过100W个并发TIMER，建议切分服务，所有执行采用异步行为。

## **安装方法**

```
go get -u github.com/koangel/grapeTimer
```

## **基本用法**

``` Go
// 初始化一个1秒钟粒度的调度器，ars代表是否自动设置运行为并行模式
grapeTimer.InitGrapeScheduler(1*time.Second, true)
// 设置GUID种子
grapeTimer.SetGuidSeed(1000)
// 设置GUID生成函数
autoId := int64(10222)
grapeTimer.SetCreateGUID(func() int64 {
	return atomic.AddInt64(&autoId,1)
})
// 获取但不增加下一个TimerId（预测函数，可能存在偏差，请自行管理）
nextId := grapeTimer.PeekNextId()
// 启动一个单次执行的调度器，1秒时间，基本tick单位为毫秒
Id := grapeTimer.NewTickerOnce(1000, exec100,"exec100 this arg1",2000,float32(200.5))
// 启动一个单次执行的调度器，1秒时间，基本tick单位为毫秒 (需要返回参数的代码)
Id := grapeTimer.NewTickerOnce(1000, exec100Result,"exec100 this arg1",2000,float32(200.5),func(v float32){
	fmt.Println("i'm call back:", v)
})
// 启动一个1秒为周期的 循环timer 循环100次 -1为永久循环
Id = grapeTimer.NewTickerLoop(1000,100, exec100Loop,"exec100Loop this arg1",2000,float32(200.5))
// 启动一个每日规则的定时器，参数为args data
Id = grapeTimer.NewTimeDataOnce("Day 13:59:59", exeDayTime,"exeDayTime this arg1",2000,float32(200.5))
// 启动一个每日循环规则的定时器，参数为args data 循环100次 -1为永久循环
Id = grapeTimer.NewTimeDataLoop("Day 13:59:59",100, exeDayTime,"exeDayTime this arg1",2000,float32(200.5))
// 通过json启动一个定时器
Id = grapeTimer.NewFromJson(jsonBody,"exec100Loop this arg1",2000,float32(200.5))
```

函数可以为任意类型的任意参数数量的函数，会自动保存参数以及数值，CALLBACK线程安全：
```
func exec100(arg1 string, arg2 int, arg3 float32) {
	fmt.Println(arg1, arg2, arg3)
}

// 需要返回参数
func exec100Result(arg1 string, arg2 int, arg3 float32,rscall func(v float32)) {
	fmt.Println(arg1, arg2, arg3)

	rscall(arg3)
}


func exec100Loop(arg1 string, arg2 int, arg3 float32) {
	fmt.Println(arg1, arg2, arg3)
}

func exeDayTime(arg1 string, arg2 int, arg3 float32) {
	fmt.Println(arg1, arg2, arg3)
}
```
## **停止计时器**

```Go
// 将自动返回的ID作为参数传入可停止持续循环的TIMER
grapeTimer.StopTimer(Id)
```

## **参数设置**

```
// 设置启用日志调试模式，建议正式版本关闭他
grapeTimer.CDebugMode = true
// 调用分析器使用的时区，可以个根据不同国家地区设置 
grapeTimer.LocationFormat = "Asia/Shanghai"
// 开启异步调度模式，在此模式下 timer执行时会建立一个go，不会阻塞其他timer执行调度，建议开启
grapeTimer.UseAsyncExec = true
```

## **获取下一次执行的时间**

```Go
// 将自动返回的ID作为参数传入可停止持续循环的TIMER
 nextTimeStr := grapeTimer.String(Id) // 标准格式获取
 nextTimeStr = grapeTimer.Format("2006-01-02 15:04:05") // 通过自定义格式化获取
 allTimer := grapeTimer.List() // 全部Id对应日期的下一次执行时间
```

## **保存调度器**

```Go
nextJson := grapeTimer.ToJson(Id) // 获取单个调度器保存到JSON格式下
allJson := grapeTimer.SaveAll()
```

## **生成调度器字符串**

```Go
Id = grapeTimer.NewFromJson(jsonBody,"exec100Loop this arg1",2000,float32(200.5))
```

## **可用格式说明**

调度器有轻度的日期模式分析体系，可提供每日，每周，每月的时间日期生成方式，具体格式如下：

|关键字|格式|说明|
|:----------:|:-------:|:----------:|
|Day|Day 00:00:00|生成每日的日期时间|
|Week|Week 1 00:00:00|生成每周的日期时间， 0~6 分别代表周日到周六|
|Month|Month 1 00:00:00|生成每月该日期的时间，建议不要使用20日之后的日期|

以上日期如果超过时间自动生成下一个时间，每月时间使用时，如本月不存在该日期，则直接返回错误。

代码示例：

```go
	vtime, err := grapeTimer.Parser("Day 13:59:59") // 返回值为标准的*time.Time
	if err != nil {
		// 处理错误...
	}

        vtime, err = grapeTimer.Parser("Week 6 23:59:59")
	if err != nil {
		// 处理错误...
	}

	vtime, err = grapeTimer.Parser("Month 26 13:59:59")
	if err != nil {
		// 处理错误...
	}
```
代码输出：

```
2017-05-24 13:59:59 +0800 CST
```