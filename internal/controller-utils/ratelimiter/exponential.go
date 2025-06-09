// Package ratelimiter contains suggested default ratelimiters for providers.
package ratelimiter

import (
	"time"

	internal_workqueue "github.com/krateoplatformops/azuredevops-provider/internal/controller-utils/workqueue"
	"k8s.io/client-go/util/workqueue"
)

func NewGlobalExponential(baseDelay time.Duration, maxDelay time.Duration) workqueue.RateLimiter {
	return internal_workqueue.NewExponentialTimedFailureRateLimiter(baseDelay, maxDelay)
}
