package pressure_metric

import (
	"time"
)

type PressureMetric struct {
	epollStartTime    time.Time
	epollEndTime      time.Time
	estimateD         time.Duration
	estimnateWorkingD time.Duration
	estimateP         time.Duration
	onEstimateResult  func(e float32)
}

func NewPressureMetric(period time.Duration, onEstimateResult func(e float32)) *PressureMetric {
	return &PressureMetric{
		estimateP:        period,
		onEstimateResult: onEstimateResult,
		epollStartTime:   time.Now(),
		epollEndTime:     time.Now(),
	}
}

func (p *PressureMetric) IdleStart() {
	epollStartTime := time.Now()
	p.epollStartTime = epollStartTime
	workingD := epollStartTime.Sub(p.epollEndTime)
	p.estimateD += workingD
	p.estimnateWorkingD += workingD
	// fmt.Print(float32(p.estimnateWorkingD)/float32(p.estimateD), p.estimateD, "\n")
	if p.estimateD > p.estimateP {
		load := float32(p.estimnateWorkingD) / float32(p.estimateD)
		p.onEstimateResult(load)
		p.estimateD = 0
		p.estimnateWorkingD = 0
	}
}

func (p *PressureMetric) IdleEnd() {
	epollEndTime := time.Now()
	p.epollEndTime = epollEndTime
	waitingD := epollEndTime.Sub(p.epollStartTime)
	p.estimateD += waitingD
}
