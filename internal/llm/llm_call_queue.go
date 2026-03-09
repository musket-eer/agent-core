package llm

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"sync"
)

const (
	defaultLLMQueueWorkers = 12
	defaultLLMQueueSize    = 512
)

type llmQueueResult struct {
	output string
	err    error
}

type llmQueueJob struct {
	ctx    context.Context
	call   func() (string, error)
	result chan llmQueueResult
}

type llmCallQueue struct {
	jobs chan llmQueueJob
}

var (
	llmQueueOnce sync.Once
	llmQueueInst *llmCallQueue
)

func getLLMCallQueue() *llmCallQueue {
	llmQueueOnce.Do(func() {
		workers := getenvInt("LLM_QUEUE_WORKERS", defaultLLMQueueWorkers)
		queueSize := getenvInt("LLM_QUEUE_SIZE", defaultLLMQueueSize)
		if workers < 1 {
			workers = defaultLLMQueueWorkers
		}
		if queueSize < 1 {
			queueSize = defaultLLMQueueSize
		}
		q := &llmCallQueue{
			jobs: make(chan llmQueueJob, queueSize),
		}
		for i := 0; i < workers; i++ {
			go q.worker()
		}
		llmQueueInst = q
	})
	return llmQueueInst
}

func (q *llmCallQueue) worker() {
	for job := range q.jobs {
		if err := job.ctx.Err(); err != nil {
			select {
			case job.result <- llmQueueResult{err: err}:
			default:
			}
			continue
		}

		out, err := job.call()
		select {
		case job.result <- llmQueueResult{output: out, err: err}:
		default:
		}
	}
}

func (q *llmCallQueue) submit(ctx context.Context, call func() (string, error)) (string, error) {
	job := llmQueueJob{
		ctx:    ctx,
		call:   call,
		result: make(chan llmQueueResult, 1),
	}

	select {
	case q.jobs <- job:
	case <-ctx.Done():
		return "", ctx.Err()
	default:
		return "", fmt.Errorf("LLM_QUEUE_OVERLOADED")
	}

	select {
	case res := <-job.result:
		return res.output, res.err
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

func getenvInt(key string, fallback int) int {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return v
}
