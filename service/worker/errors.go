package worker

import "errors"

// ErrQueueFull is returned when trying to enqueue a session but the queue is full
var ErrQueueFull = errors.New("session queue is full")
