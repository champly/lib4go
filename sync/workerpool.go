package sync

import (
	"fmt"
	"runtime/debug"

	"github.com/champly/lib4go/tool"
	"k8s.io/klog"
	"mosn.io/pkg/utils"
)

const maxRespawnTimes = 1 << 6

type shard struct {
	index        int
	respawnTimes uint32
	jobChan      chan interface{}
}

type shardWorkerPool struct {
	workerFunc WorkerFunc
	shards     []*shard
	numShards  int
}

func NewShardWokerPool(size int, numShards int, workerFunc WorkerFunc) (ShardWorkerPool, error) {
	if size <= 0 {
		return nil, fmt.Errorf("worker pool size too small: %d", size)
	}

	if size < numShards {
		numShards = size
	}
	shardCap := size / numShards
	shards := make([]*shard, numShards)
	for i := range shards {
		shards[i] = &shard{
			index:   i,
			jobChan: make(chan interface{}, shardCap),
		}
	}

	return &shardWorkerPool{
		workerFunc: workerFunc,
		shards:     shards,
		numShards:  numShards,
	}, nil
}

func (pool *shardWorkerPool) Init() {
	for i := range pool.shards {
		pool.spawnWorker(pool.shards[i])
	}
}

func (pool *shardWorkerPool) spawnWorker(shard *shard) {
	tool.GoWithRecover(
		func() {
			pool.workerFunc(shard.index, shard.jobChan)
		},
		func(r interface{}) {
			if shard.respawnTimes < maxRespawnTimes {
				shard.respawnTimes++
				pool.spawnWorker(shard)
			}
		},
	)
}

func (pool *shardWorkerPool) Shard(source uint32) uint32 {
	return source % uint32(pool.numShards)
}

func (pool *shardWorkerPool) Offer(job ShardJob, block bool) {
	i := pool.Shard(job.Source())
	if block {
		pool.shards[i].jobChan <- job
		return
	}

	select {
	case pool.shards[i].jobChan <- job:
	default:
		klog.Errorf("[syncpool] jobChan over full")
	}
}

type workerPool struct {
	work chan func()
	sem  chan struct{}
}

func NewWorkerPool(size int) WorkerPool {
	return &workerPool{
		work: make(chan func()),
		sem:  make(chan struct{}, size),
	}
}

func (p *workerPool) Schedule(task func()) {
	select {
	case p.work <- task:
	case p.sem <- struct{}{}:
		go p.spawnWorker(task)
	}
}

func (p *workerPool) ScheduleAlways(task func()) {
	select {
	case p.work <- task:
	case p.sem <- struct{}{}:
		go p.spawnWorker(task)
	default:
		klog.Infof("[syncpool] workerpool new goroutine")
		tool.GoWithRecover(func() {
			task()
		}, nil)
	}
}

func (p *workerPool) ScheduleAuto(task func()) {
	select {
	case p.work <- task:
		return
	default:
	}

	select {
	case p.work <- task:
	case p.sem <- struct{}{}:
		go p.spawnWorker(task)
	default:
		klog.Infof("[syncpool] workerpool new goroutine")
		utils.GoWithRecover(func() {
			task()
		}, nil)
	}
}

func (p *workerPool) spawnWorker(task func()) {
	defer func() {
		if r := recover(); r != nil {
			klog.Warningf("syncpool panic %v\n%s", r, string(debug.Stack()))
		}
		<-p.sem
	}()

	for {
		task()
		task = <-p.work
	}
}
