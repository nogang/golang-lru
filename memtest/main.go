package main

import (
			"fmt"
	"github.com/hashicorp/golang-lru"
				"time"
	"github.com/ground-x/go-gxplatform/common"
	"sync"
)

var sw sync.WaitGroup

const goroutineNum = 1000
const shardsNum = 512
const allwork = 10000000
const workpergoroutine = allwork/goroutineNum
const cacheSize = allwork * 4

func main(){
	lruShardTest()
	lruTest()
}

func lruTest() {
	lru, _ := lru.NewWithEvict(cacheSize, nil)
	lruAddTest(lru)
	lruGetTest(lru)
}

func lruAddTest(lru *lru.Cache) {
	start := time.Now()
	sw.Add(goroutineNum)
	for i := 0 ; i < goroutineNum; i++ {
		go lruAddWork(lru, i)
	}
	sw.Wait()
	duration := time.Since(start)
	fmt.Println("LRU Add Time : ",duration/allwork)
}

func lruGetTest(lru *lru.Cache) {
	sw.Add(goroutineNum)
	start := time.Now()
	for i := 0 ; i < goroutineNum; i++ {
		go lruGetWork(lru, i)
	}
	sw.Wait()
	duration := time.Since(start)
	fmt.Println("LRU Get Time : ",duration/allwork)
}

func lruAddWork(lru *lru.Cache,goNum int) {
	defer sw.Done()
	for i := 0 ; i < workpergoroutine ; i++ {
		ev := lru.Add(getKey(goNum, i), getValue())
		if ev {
			println("add Error")
		}
	}
}

func lruGetWork(lru *lru.Cache, goNum int) {
	defer sw.Done()
	for i := 0 ; i < workpergoroutine ; i++ {
		_, ok := lru.Get(getKey(goNum, i))
		if !ok {
			println("get Error")
		}
	}
}

func lruShardTest() {
	lruShard := InitLruShard(shardsNum, cacheSize/shardsNum)
	lruShardAddTest(lruShard)
	lruShardGetTest(lruShard)
}

func lruShardAddTest(lru *LRUShard) {
	start := time.Now()
	sw.Add(goroutineNum)
	for i := 0 ; i < goroutineNum; i++ {
		go lruShardAddWork(lru, i)
	}
	sw.Wait()
	duration := time.Since(start)
	fmt.Println("shard LRU Add Time : ",duration/allwork)
}

func lruShardGetTest(lru *LRUShard) {
	sw.Add(goroutineNum)
	start := time.Now()
	for i := 0 ; i < goroutineNum; i++ {
		go lruShardGetWork(lru, i)
	}
	sw.Wait()
	duration := time.Since(start)
	fmt.Println("shard LRU Get Time : ",duration/allwork)
}

func lruShardAddWork(shardLRU *LRUShard, goNum int) {
	defer sw.Done()
	for i := 0 ; i < workpergoroutine ; i++ {
		ev := shardLRU.Add(getKey(goNum, i), getValue())
		if ev {
			println("add Error")
		}
	}
}

func lruShardGetWork(shardLRU *LRUShard, goNum int) {
	defer sw.Done()
	for i := 0 ; i < workpergoroutine ; i++ {
		_, ok := shardLRU.Get(getKey(goNum, i))
		if !ok {
			println("get Error")
		}
	}
}

type LRUShard struct {
	shards     []*lru.Cache
	shardCount int
}

func (s *LRUShard) Add(key, val interface{}) (bool) {
	shardIndex := s.getShardIndex(key)
	return s.shards[shardIndex].Add(key, val)
}

func (s *LRUShard) Get(key interface{}) (interface{},bool) {
	shardIndex := s.getShardIndex(key)
	return s.shards[shardIndex].Get(key)
}

func (s *LRUShard) getShardIndex(key interface{}) int {
	switch k := key.(type) {
	case common.Hash:
		return int((k[3] << 24) + (k[2] << 16) + (k[1] << 8) + k[0]) % s.shardCount
	case common.Address:
		return int((k[3] << 24) + (k[2] << 16) + (k[1] << 8) + k[0]) % s.shardCount
	default:
		return 0
	}
}

//if key is not common.Hash nor common.Address then you should set numShard 1
func InitLruShard(shardCount int, shardSize int) *LRUShard {
	lruShard := &LRUShard{shards : make([]*lru.Cache,shardCount), shardCount:shardCount}
	for i := 0 ; i < shardCount ; i++ {
		lruShard.shards[i], _ = lru.NewWithEvict(shardSize, nil)
	}
	return lruShard
}

func getKey(goNum, i int) common.Hash {
	var key common.Hash
	mixNum := goNum * 7717 + i * 5009
	for i := 0; i < common.HashLength; i++ {
		key[i] = byte(mixNum & 0xff)
		mixNum >>= 8
		if mixNum == 0 {
			break
		}
	}
	return key
}

func getValue() []byte {
	data := make([]byte, 256)
	data[0] = 10
	return data
}


/*
runtime.MemProfileRate=1

var sumDurationg time.Duration
for i := 0 ; i < testTime ; i++ {
	testName := fmt.Sprintf("Single_Add_Test_dataSize_%d",i)

	cache, e := lru.New(10000000)
	runtime.GC()
	if e != nil {
		fmt.Printf("cache generate error : %s\n",e)
	}
	//b.ReportAllocs()

	f, err := os.Create("./" + testName + "_cpu")

	if err != nil {
		log.Fatal("could not create CPU profile: ", err)
	}
	if err := pprof.StartCPUProfile(f); err != nil {
		log.Fatal("could not start CPU profile: ", err)
	}

	start := time.Now()

	for i:= 0 ; i < operationCount ; i++{
		cache.Add(key(i), getData(byte(i)))
	}

	duration := time.Since(start)
	sumDurationg += duration


	mf, err := os.Create("./" + testName +"_mem")

	if err := pprof.WriteHeapProfile(mf) ; err != nil {
		log.Fatal(err)
	}
	pprof.StopCPUProfile()
	mf.Close()
}

fmt.Println((sumDurationg - makeDataDuration)/operationCount/testTime)
*/