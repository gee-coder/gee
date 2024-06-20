package pool

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gee-coder/gee/config"
)

const DefaultExpire = 5

type Pool struct {
	// cap 容量 pool max cap
	cap int32
	// running 正在运行的worker的数量
	running int32
	// 空闲worker
	workers []*Worker
	// expire 过期时间 空闲的worker超过这个时间 回收掉
	expire time.Duration
	// release 释放资源  pool就不能使用了
	release chan struct{}
	// lock 去保护pool里面的相关资源的安全
	lock sync.Mutex
	// once 释放只能调用一次 不能多次调用
	once sync.Once
	// workerCache 缓存
	workerCache sync.Pool
	// cond
	cond *sync.Cond
	// PanicHandler
	PanicHandler func()
}

func (p *Pool) expireWorker() {
	// 定时清理过期的空闲worker
	ticker := time.NewTicker(p.expire)
	for range ticker.C {
		if p.IsClosed() {
			break
		}
		// 循环空闲的workers 如果当前时间和worker的最后运行任务的时间 差值大于expire 进行清理
		p.lock.Lock()
		idleWorkers := p.workers
		n := len(idleWorkers) - 1
		if n >= 0 {
			var clearNum = -1
			for i, w := range idleWorkers {
				if time.Now().Sub(w.lastTime) <= p.expire {
					break
				}
				clearNum = i
				w.task <- nil
				idleWorkers[i] = nil
			}
			// 有机会重构一下，这块不应该跟切片元素顺序有关联
			if clearNum != -1 {
				if clearNum == len(idleWorkers)-1 {
					p.workers = idleWorkers[:0]
				} else {
					p.workers = idleWorkers[clearNum+1:]
				}
				fmt.Printf("清除完成,running:%d, workers:%v \n", p.running, p.workers)
			}
		}
		p.lock.Unlock()
	}
}

func NewTimePool(cap int, expire int) (*Pool, error) {
	if cap <= 0 {
		return nil, errors.New("pool cap can not <= 0")
	}
	if expire <= 0 {
		return nil, errors.New("pool expire can not <= 0")
	}
	p := &Pool{
		cap:     int32(cap),
		expire:  time.Duration(expire) * time.Second,
		release: make(chan struct{}, 1),
	}
	p.workerCache.New = func() any {
		return &Worker{
			pool: p,
			task: make(chan func(), 1),
		}
	}
	p.cond = sync.NewCond(&p.lock)

	go p.expireWorker()
	return p, nil
}

func NewPool(cap int) (*Pool, error) {
	return NewTimePool(cap, DefaultExpire)
}

func NewPoolConf() (*Pool, error) {
	cap, ok := config.Conf.Pool["cap"]
	if !ok {
		return nil, errors.New("cap config not exist")
	}
	return NewTimePool(int(cap.(int64)), DefaultExpire)
}

// 提交任务
func (p *Pool) Submit(task func()) error {
	if len(p.release) > 0 {
		return errors.New("pool has bean released")
	}
	// 获取池里面的一个worker，然后执行任务就可以了
	w := p.GetWorker()
	w.task <- task
	return nil
}

func (p *Pool) waitIdleWorker() *Worker {
	p.lock.Lock()
	p.cond.Wait()
	idleWorkers := p.workers
	n := len(idleWorkers) - 1
	if n < 0 {
		p.lock.Unlock()
		if p.running < p.cap {
			// 还不够pool的容量，直接新建一个
			c := p.workerCache.Get()
			var w *Worker
			if c == nil {
				w = &Worker{
					pool: p,
					task: make(chan func(), 1),
				}
			} else {
				w = c.(*Worker)
			}
			w.run()
			return w
		}
		return p.waitIdleWorker()
	}
	w := idleWorkers[n]
	idleWorkers[n] = nil
	p.workers = idleWorkers[:n]
	p.lock.Unlock()
	return w
}

// 获取pool里面的worker
func (p *Pool) GetWorker() (w *Worker) {
	p.lock.Lock()
	defer p.lock.Unlock()
	// 如果有空闲的 worker 直接获取
	idleWorkers := p.workers
	n := len(idleWorkers) - 1
	if n >= 0 {
		w = idleWorkers[n]
		idleWorkers[n] = nil
		p.workers = idleWorkers[:n]
		return
	}
	// 如果没有空闲的worker，要新建一个worker
	if p.running < p.cap {
		w = p.workerCache.Get().(*Worker)
		w.run()
		return
	} else {
		// 如果正在运行的workers 如果大于pool容量，阻塞等待，worker释放
		return p.waitIdleWorker()
	}
}

func (p *Pool) incRunning() {
	atomic.AddInt32(&p.running, 1)
}

func (p *Pool) decRunning() {
	atomic.AddInt32(&p.running, -1)
}

func (p *Pool) PutWorker(w *Worker) {
	w.lastTime = time.Now()
	p.lock.Lock()
	defer p.lock.Unlock()
	p.workers = append(p.workers, w)
	p.cond.Signal()
}

func (p *Pool) Release() {
	p.once.Do(func() {
		// 只执行一次
		p.lock.Lock()
		defer p.lock.Unlock()
		workers := p.workers
		for i, w := range workers {
			if w == nil {
				continue
			}
			w.task = nil
			w.pool = nil
			workers[i] = nil
		}
		p.workers = nil
		p.release <- struct{}{}
	})
}

func (p *Pool) IsClosed() bool {
	return len(p.release) > 0
}

func (p *Pool) Restart() bool {
	if len(p.release) <= 0 {
		return true
	}
	_ = <-p.release
	go p.expireWorker()
	return true
}

func (p *Pool) Running() int {
	return int(atomic.LoadInt32(&p.running))
}

func (p *Pool) Free() int {
	return int(p.cap - p.running)
}
