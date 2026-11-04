// Copyright IBM Corp. 2015, 2026
// SPDX-License-Identifier: BUSL-1.1

package allocrunner

import (
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/nomad/client/lib/clientsockets"
	"github.com/hashicorp/nomad/nomad/structs"
	"github.com/hashicorp/nomad/plugins/drivers"
)

// taskDriverCapabilities looks up the driver capabilities of an alloc's tasks.
// It is implemented by allocRunner and faked in tests.
type taskDriverCapabilities interface {
	GetTaskDriverCapabilities(taskName string) (*drivers.Capabilities, error)
}

var _ taskDriverCapabilities = (*allocRunner)(nil)

// taskSocketLabeler decides whether the alloc-level sockets that tasks connect
// to are for confined tasks, which may need them created with the container
// SELinux type. A nil labeler reports no confined tasks.
type taskSocketLabeler struct {
	logger  hclog.Logger
	drivers taskDriverCapabilities
}

func newTaskSocketLabeler(logger hclog.Logger, taskDrivers taskDriverCapabilities) *taskSocketLabeler {
	return &taskSocketLabeler{logger: logger, drivers: taskDrivers}
}

// labelForAlloc reports whether the alloc's shared sockets are for confined
// tasks. They are shared by every task in the group, so they are when any task
// uses a driver that runs it as a confined container. It must only be called
// after the task runners exist, such as from a Prerun hook: the hooks are built
// before the task runners, so from a constructor every task is "not found" and
// no task would be confined.
func (l *taskSocketLabeler) labelForAlloc(alloc *structs.Allocation) bool {
	if l == nil {
		return false
	}
	tg := alloc.Job.LookupTaskGroup(alloc.TaskGroup)
	if tg == nil {
		return false
	}
	for _, task := range tg.Tasks {
		caps, err := l.drivers.GetTaskDriverCapabilities(task.Name)
		if err != nil {
			l.logger.Warn("unable to get task driver capabilities", "task", task.Name, "error", err)
			continue
		}
		if caps != nil && clientsockets.Confined(caps.FSIsolation) {
			return true
		}
	}
	return false
}
