package sync

type WorkerFunc func(shard int, jobCh <-chan interface{})

type ShardJob interface {
	Source() uint32
}

type ShardWorkerPool interface {
	Init()

	Shard(source uint32) uint32

	Offer(job ShardJob, block bool)
}

type WorkerPool interface {
	Schedule(task func())

	ScheduleAlways(task func())

	ScheduleAuto(task func())
}
