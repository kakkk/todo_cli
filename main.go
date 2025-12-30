package main

import (
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// ====================== 模型定义 ======================

type Model struct {
	items        TodoList
	cursor       int
	mode         Mode
	input        textinput.Model
	inputContext InputContext
	statusLine   string
	styles       *Styles
	storage      *Storage

	// 用于添加新项目
	draftItem TodoItem

	// 日期选择器状态
	datePicker struct {
		index int
		date  time.Time
		field DateField
	}

	// 优先级选择器状态
	priorityPicker struct {
		priority Priority
	}

	terminalWidth int
	selectedID    string // 跟踪当前选中的任务ID
}

func NewModel() *Model {
	ti := textinput.New()
	ti.Placeholder = "输入内容后回车确认，Esc 取消"
	ti.Prompt = "» "
	ti.CharLimit = 200
	ti.Width = 40 // 设置默认宽度

	// 初始化存储
	storage, err := NewStorage()
	var status string
	var items TodoList

	if err != nil {
		status = fmt.Sprintf("存储初始化失败: %v", err)
		items = TodoList{}
	} else {
		items, status = storage.Load()
	}

	model := &Model{
		items:        items,
		cursor:       0,
		mode:         ModeNormal,
		input:        ti,
		inputContext: InputContextNone,
		statusLine:   status,
		styles:       NewStyles(),
		draftItem:    TodoItem{Priority: PriorityMedium},
		storage:      storage,
	}

	// 初始化选中的ID
	if len(items) > 0 {
		model.selectedID = items[0].id
	}

	return model
}

func (m *Model) Init() tea.Cmd {
	return nil
}

// ====================== 主函数 ======================

func main() {
	program := tea.NewProgram(NewModel())
	if _, err := program.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "运行出错: %v\n", err)
		os.Exit(1)
	}
}
