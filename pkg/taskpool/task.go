package taskpool

import (
	"context"
	"sync"

	config2 "github.com/jukylin/esim/core/config"

	"github.com/jukylin/esim/config"

	"github.com/jukylin/esim/log"
)

type TaskPool struct {
	pool       chan chan IJob
	jobList    chan IJob
	works      []*worker
	maxWorkers int
	wg         *sync.WaitGroup
	cancel     context.CancelFunc
	logger     log.Logger
	conf       config2.Config
}

type Option func(*TaskPool)

var (
	once       sync.Once
	TaskClient *TaskPool
)

const defaultPoolSize = 5

func NewTaskPool(opts ...Option) *TaskPool {
	once.Do(func() {
		TaskClient = &TaskPool{
			pool:       make(chan chan IJob, defaultPoolSize),
			works:      make([]*worker, 0),
			jobList:    make(chan IJob),
			maxWorkers: defaultPoolSize,
			wg:         &sync.WaitGroup{},
		}

		for _, opt := range opts {
			opt(TaskClient)
		}

		if TaskClient.logger == nil {
			TaskClient.logger = log.NewLogger()
		}

		if TaskClient.conf == nil {
			TaskClient.conf = config.NewMemConfig()
		}

		poolSize := TaskClient.conf.GetInt("taskpool_max_count")
		if poolSize > 0 && poolSize < 500 {
			TaskClient.pool = make(chan chan IJob, poolSize)
			TaskClient.maxWorkers = poolSize
		}

		TaskClient.logger.Infof("开启了【%v】个taskPool", poolSize)

	})
	return TaskClient
}

func (t *TaskPool) WithTaskSize(taskSize int) {
	TaskClient.maxWorkers = taskSize
	TaskClient.pool = make(chan chan IJob, taskSize)
}

func WithTaskLogger(log log.Logger) Option {
	return func(task *TaskPool) {
		task.logger = log
	}
}

func WithTaskConf(conf config2.Config) Option {
	return func(task *TaskPool) {
		task.conf = conf
	}
}

func (t *TaskPool) Concurrency() int {
	return t.maxWorkers - len(t.pool)
}

func (t *TaskPool) AddJobs(jobs ...IJob) {
	for _, job := range jobs {
		t.jobList <- job
	}
}

func GetTaskClient() *TaskPool {
	if TaskClient == nil {
		TaskClient = NewTaskPool()
	}

	return TaskClient
}

func (t *TaskPool) Run() {
	ctx, cancel := context.WithCancel(context.Background())
	t.cancel = cancel
	// starting n number of workers
	for i := 0; i < t.maxWorkers; i++ {
		worker := newWorker(i, t.pool)
		go func(wg *sync.WaitGroup, idx int) {
			defer wg.Done()

			t.logger.Infof("worker[%v], 开始执行....", idx+1)
			wg.Add(1)
			worker.start()
			t.logger.Infof("worker[%v], 结束执行....", idx+1)
		}(t.wg, i)

		t.works = append(t.works, worker)
	}

	go t.process(ctx)
}

func (t *TaskPool) process(ctx context.Context) {
	for {
		select {
		case job := <-t.jobList:
			// a job request has been received
			// 直接在当前routine完成到worker的分配*/
			jobChannel, ok := <-t.pool
			if !ok {
				t.logger.Info("Task failure")
				return
			}

			// dispatch the job to the worker job channel
			jobChannel <- job

		case <-ctx.Done():
			for i := 0; i < t.maxWorkers; i++ {
				jobChannel := <-t.pool
				close(jobChannel)
			}
			close(t.pool)

			t.logger.Info("congratulations, Task over")
			return
		}
	}
}

func (t *TaskPool) Stop() {
	t.cancel()
	t.wg.Wait()

	for _, w := range t.works {
		w.stop()
	}
}
