package main

import (
	"sort"
	"time"

	"github.com/google/uuid"
)

// ====================== 常量定义 ======================

type Priority int

const (
	PriorityLow    Priority = iota // P2
	PriorityMedium                 // P1
	PriorityHigh                   // P0
)

func (p Priority) String() string {
	switch p {
	case PriorityHigh:
		return "P0"
	case PriorityMedium:
		return "P1"
	case PriorityLow:
		return "P2"
	default:
		return "?"
	}
}

type Mode int

const (
	ModeNormal Mode = iota
	ModeInputTitle
	ModePickDate
	ModePickPriority
)

type InputContext int

const (
	InputContextNone InputContext = iota
	InputContextAddTitle
	InputContextEditTitle
	InputContextAddPriority
	InputContextEditPriority
)

type DateField int

const (
	DateFieldYear DateField = iota
	DateFieldMonth
	DateFieldDay
	DateFieldHour
	DateFieldMinute
	DateFieldSecond
)

// ====================== 数据结构 ======================

type TodoItem struct {
	Title       string
	Done        bool
	Priority    Priority
	HasDeadline bool
	Deadline    time.Time
	id          string
}

// 用于生成唯一ID
func generateID() string {
	return uuid.New().String()
}

func (ti *TodoItem) IsOverdue() bool {
	return !ti.Done && ti.HasDeadline && time.Now().After(ti.Deadline)
}

func (ti *TodoItem) DeadlineString() string {
	if !ti.HasDeadline {
		return "-"
	}
	return ti.Deadline.Format("2006-01-02 15:04")
}

type TodoList []TodoItem

func (tl TodoList) Sort() {
	sort.SliceStable(tl, func(i, j int) bool {
		a, b := tl[i], tl[j]

		// 1. 未完成的在前
		if a.Done != b.Done {
			return !a.Done && b.Done
		}
		// 2. 优先级高的在前
		if a.Priority != b.Priority {
			return a.Priority > b.Priority
		}
		// 3. 有截止日期的在前
		if a.HasDeadline != b.HasDeadline {
			return a.HasDeadline && !b.HasDeadline
		}
		// 4. 截止日期早的在前
		if a.HasDeadline && b.HasDeadline {
			return a.Deadline.Before(b.Deadline)
		}
		// 5. 按标题字母排序
		return a.Title < b.Title
	})
}
