package routes

// Pipeline chains a Deduplicator, RateLimiter, and Throttler together to
// provide a single ShouldForward gate before a diff is acted upon.
type Pipeline struct {
	dedup    *Deduplicator
	limiter  *RateLimiter
	throttle *Throttler
}

// PipelineConfig bundles the configuration for each stage.
type PipelineConfig struct {
	Dedupe    DedupeConfig
	RateLimit RateLimitConfig
	Throttle  ThrottleConfig
}

// DefaultPipelineConfig returns a PipelineConfig with all defaults.
func DefaultPipelineConfig() PipelineConfig {
	return PipelineConfig{
		Dedupe:    DefaultDedupeConfig(),
		RateLimit: DefaultRateLimitConfig(),
		Throttle:  DefaultThrottleConfig(),
	}
}

// NewPipeline constructs a Pipeline from the given config.
func NewPipeline(cfg PipelineConfig) *Pipeline {
	return &Pipeline{
		dedup:    NewDeduplicator(cfg.Dedupe),
		limiter:  NewRateLimiter(cfg.RateLimit),
		throttle: NewThrottler(cfg.Throttle),
	}
}

// ShouldForward returns true only when the diff passes all three stages:
// it is not a duplicate, the rate limit has not been exceeded, and the
// throttle interval has elapsed.
func (p *Pipeline) ShouldForward(d Diff) bool {
	if p.dedup.IsDuplicate(d) {
		return false
	}
	if !p.limiter.Allow() {
		return false
	}
	if !p.throttle.Allow() {
		return false
	}
	return true
}

// Reset clears state across all pipeline stages.
func (p *Pipeline) Reset() {
	p.dedup.Reset()
	p.limiter.Reset()
	p.throttle.Reset()
}
