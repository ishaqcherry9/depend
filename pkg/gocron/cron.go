package gocron

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/robfig/cron/v3"
)

var (
	c      *cron.Cron
	nameID = sync.Map{}
	idName = sync.Map{}
)

type Task struct {
	TimeSpec string

	Name      string
	Fn        func()
	IsRunOnce bool
}

func Init(opts ...Option) error {
	o := defaultOptions()
	o.apply(opts...)

	log := &zapLog{zapLog: o.zapLog, isOnlyPrintError: o.isOnlyPrintError}
	cronOpts := []cron.Option{
		cron.WithLogger(log),
		cron.WithChain(
			cron.Recover(log),
		),
	}
	if o.granularity == SecondType {
		cronOpts = append(cronOpts, cron.WithSeconds())
	}

	c = cron.New(cronOpts...)
	c.Start()

	return nil
}

func Run(tasks ...*Task) error {
	if c == nil {
		return errors.New("cron is not initialized")
	}

	var errs []string
	for _, task := range tasks {
		if IsRunningTask(task.Name) {
			errs = append(errs, fmt.Sprintf("task '%s' is already exists", task.Name))
			continue
		}

		if err := checkRunOnce(task); err != nil {
			errs = append(errs, err.Error())
			continue
		}

		id, err := c.AddFunc(task.TimeSpec, task.Fn)
		if err != nil {
			errs = append(errs, fmt.Sprintf("run task '%s' error: %v", task.Name, err))
			continue
		}
		idName.Store(id, task.Name)
		nameID.Store(task.Name, id)
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, " || "))
	}

	return nil
}

func checkRunOnce(task *Task) error {
	if task.Fn == nil {
		return fmt.Errorf("task '%s' is nil", task.Name)
	}
	if task.IsRunOnce {
		job := task.Fn
		task.Fn = func() {
			job()
			DeleteTask(task.Name)
		}
	}
	return nil
}

func IsRunningTask(name string) bool {
	_, ok := nameID.Load(name)
	return ok
}

func GetRunningTasks() []string {
	var names []string
	nameID.Range(func(key, value interface{}) bool {
		names = append(names, key.(string))
		return true
	})
	return names
}

func DeleteTask(name string) {
	if id, ok := nameID.Load(name); ok {
		entryID, isOk := id.(cron.EntryID)
		if !isOk {
			return
		}
		c.Remove(entryID)
		nameID.Delete(name)
		idName.Delete(entryID)
	}
}

func Stop() {
	if c != nil {
		c.Stop()
	}
}

func EverySecond(size int) string {
	return fmt.Sprintf("@every %ds", size)
}

func EveryMinute(size int) string {
	return fmt.Sprintf("@every %dm", size)
}

func EveryHour(size int) string {
	return fmt.Sprintf("@every %dh", size)
}

func Everyday(size int) string {
	return fmt.Sprintf("@every %dh", size*24)
}
