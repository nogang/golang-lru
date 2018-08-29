package main

import (
			"fmt"
	"github.com/hashicorp/golang-lru"
				"time"
	"github.com/ground-x/go-gxplatform/common"
	"sync"
		)

var sw sync.WaitGroup

const shardsNum = 65536
const allwork = 10000000
const cacheSize = allwork * 3

func main(){
	//lruShardTest()
	for i := 1 ; i <= 20 ; i += 1 {
		shardCountTest(2)
	}

	for i := 1 ; i <= 4000 ; i += 100 {
		shardCountTest(2)
	}

	/*
	lruTest(1)
	lruTest(1)

	lruTest(1)
	lruTest(1)
	lruTest(1)
	lruTest(1)

*/
	//lruTest(1)
	//lruTest(2000)
	//lruTest(4000)



	//shardCountTest(1)
	//shardCountTest(2000)
	//shardCountTest(4000)
}

func parallelTest() {

}

func cachSizeTest() {

}

func dataSizeTest() {

}

func shardCountTest(gonum int) {
	testName := "shardCountTest"
	//for s := 1 ; s <= 65536 ; s *= 4 {
	s:= 4096
	lruShard := InitLruShard(s, cacheSize/s)
	lruShardAddTest(testName, lruShard, gonum)
	lruShardGetTest(testName, lruShard, gonum)
	//}
}

func lruTest(gonum int) {
	lru, _ := lru.NewWithEvict(cacheSize, nil)
	lruAddTest("lruBaseTest",lru, gonum)
	lruGetTest("lruBaseTest",lru, gonum)
}


func lruAddTest(testName string,lru *lru.Cache, gonum int) {
	start := time.Now()
	sw.Add(gonum)
	for i := 0 ; i < gonum; i++ {
		go lruAddWork(lru, gonum)
	}
	sw.Wait()
	duration := time.Since(start)
	printResult(testName+"AddNoShard",0,gonum,int(duration/allwork))
}

func lruGetTest(testName string, lru *lru.Cache, gonum int) {
	sw.Add(gonum)
	start := time.Now()
	for i := 0 ; i < gonum; i++ {
		go lruGetWork(lru, gonum)
	}
	sw.Wait()
	duration := time.Since(start)
	printResult(testName+"GetNoShard",0,gonum,int(duration/allwork))
}

func lruAddWork(lru *lru.Cache,goNum int) {
	defer sw.Done()
	for i := 0 ; i < allwork/goNum ; i++ {
		ev := lru.Add(getKey(goNum, i), getValue())
		if ev {
			println("add Error")
		}
	}
}

func lruGetWork(lru *lru.Cache, goNum int) {
	defer sw.Done()
	for i := 0 ; i < allwork/goNum ; i++ {
		_, ok := lru.Get(getKey(goNum, i))
		if !ok {
			println("get Error")
		}
	}
}
/*
func lruShardTest() {
	lruShard := InitLruShard(shardsNum, cacheSize/shardsNum)
	lruShardAddTest(lruShard)
	lruShardGetTest(lruShard)
}
*/
func lruShardAddTest(testName string, lru *LRUShard, grNum int) {
	start := time.Now()
	sw.Add(grNum)
	for i := 0 ; i < grNum; i++ {
		go lruShardAddWork(lru, grNum)
	}
	sw.Wait()
	duration := time.Since(start)
	printResult(testName+"AddShard",lru.maxShard+1,grNum,int(duration/allwork))
}

func lruShardGetTest(testName string, lru *LRUShard, grNum int) {
	sw.Add(grNum)
	start := time.Now()
	for i := 0 ; i < grNum; i++ {
		go lruShardGetWork(lru, grNum)
	}
	sw.Wait()
	duration := time.Since(start)
	printResult(testName+"GetShard",lru.maxShard + 1,grNum,int(duration/allwork))
}

func lruShardAddWork(shardLRU *LRUShard, goNum int) {
	defer sw.Done()
	for i := 0 ; i < allwork/goNum ; i++ {
		backupKey := getKey(goNum, i)
		ev := shardLRU.Add(backupKey, getValue())
		if ev {
			println("add Error")
		}
	}
}

func lruShardGetWork(shardLRU *LRUShard, goNum int) {
	defer sw.Done()
	for i := 0 ; i < allwork/goNum ; i++ {
		_, ok := shardLRU.Get(getKey(goNum, i))
		if !ok {
			println("get Error")
		}
	}
}

type LRUShard struct {
	shards     []*lru.Cache
	maxShard int
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
		return ((int(k[2]) << 16) + (int(k[1]) << 8) + int(k[0])) & s.maxShard
	case common.Address:
		return ((int(k[2]) << 16) + (int(k[1]) << 8) + int(k[0])) & s.maxShard
	default:
		return 0
	}
}

//if key is not common.Hash nor common.Address then you should set numShard 1
func InitLruShard(shardCount int, shardSize int) *LRUShard {
	preShardCount := shardCount
	for shardCount > 0 {
		preShardCount = shardCount
		shardCount = shardCount & (shardCount - 1)
	}
	shardCount = preShardCount
	lruShard := &LRUShard{shards : make([]*lru.Cache,shardCount), maxShard:shardCount - 1}
	for i := 0 ; i < shardCount ; i++ {
		lruShard.shards[i], _ = lru.NewWithEvict(shardSize, nil)
	}
	return lruShard
}

func getKey(goNum, i int) common.Hash {
	var key common.Hash
	mixNum := goNum * 5009 + i * 7717
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

func printResult(testName string, shardNum, goroutine, time int) {
	fmt.Println(testName, "Shards", shardNum, "Goroutine", goroutine, "Time", time)
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