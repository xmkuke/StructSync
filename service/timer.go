package service

import (
	"fmt"
	"time"
)

type MyTimer struct {
	start time.Time
	end   time.Time
}

/*
* Create Timer
 */
func NewMyTimer() *MyTimer {
	return &MyTimer{
		start: time.Now(),
	}
}

/*
* Close Timer
 */
func (mt *MyTimer) Stop() {
	mt.end = time.Now()
}

/*
* Duration
 */
func (mt *MyTimer) UsedSecond() string {
	return fmt.Sprintf("%f s", mt.end.Sub(mt.start).Seconds())
}
