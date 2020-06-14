package grapeTimer

import (
	"fmt"
	"sync/atomic"
	"testing"
	"time"
)

func CallBackOnce(arg1 string, arg2 int, arg3 float32) {
	fmt.Println(arg1, arg2, arg3)
}

func CallBackResult(arg1 string, arg2 int, arg3 float32, resultcall func(val float32)) {
	fmt.Println(arg1, arg2, arg3)

	resultcall(arg3) // 调用CALLBACK去返回数值
}

func Test_CallResult(t *testing.T) {
	cb, err := reflectFunc(CallBackResult, "this arg1", 2000, float32(300.5), func(v float32) {
		fmt.Println("i'm call back:", v)
	})
	if err != nil {
		t.Error(err)
		return
	}

	callFunc(cb)
}

func Test_CallReflect(t *testing.T) {
	cb, err := reflectFunc(CallBackOnce, "this arg1", 2000, float32(300.5))
	if err != nil {
		t.Error(err)
		return
	}

	callFunc(cb)
}

func Test_JsonSave(t *testing.T) {
	InitGrapeScheduler(2*time.Second, false)
	Id := NewTickerOnce(1000, CallBackOnce, "this arg1", 2000, float32(300.5))
	Json := ToJson(Id)

	if len(Json) == 0 {
		t.Error("save error")
		return
	}

	fmt.Print(Json)

	Id = NewFromJson(Json, CallBackOnce, "this arg1", 2000, float32(300.5))

}

func Test_JsonSaveAll(t *testing.T) {
	InitGrapeScheduler(2*time.Second, false)
	NewTickerOnce(1000, CallBackOnce, "this arg1", 2000, float32(300.5))
	Json := SaveAll()

	if len(Json) == 0 {
		t.Error("save error")
		return
	}

	fmt.Print(Json)
}

func Test_CreateGUIDFnc(t *testing.T) {
	InitGrapeScheduler(time.Second, false)
	autoId := int64(10222)
	SetCreateGUID(func() int64 {
		return atomic.AddInt64(&autoId, 1)
	})

	nextId := autoId + 1
	eqNextId := NewTickerLoop(1000, -1, func() {
		fmt.Println("ticker 1 sec")
	})

	if nextId != eqNextId {
		t.Fail()
	}

	fmt.Println(nextId, eqNextId)

	SetGuidSeed(nextId)
	nextId = PeekNextId()
	eqNextId = NewTickerLoop(1000, -1, func() {
		fmt.Println("ticker 1 sec")
	})
	if nextId != eqNextId {
		t.Fail()
	}

}

func Benchmark_Parallel(b *testing.B) {
	cb, err := reflectFunc(CallBackOnce, "this arg1", 2000, float32(300.5))
	if err != nil {
		b.Error(err)
		return
	}

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			callFunc(cb)
		}
	})
}
