package main

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// ====================== 更新逻辑 ======================

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyPress(msg)
	case tea.WindowSizeMsg:
		m.terminalWidth = msg.Width
		// 更新输入框宽度，确保能够显示完整内容
		width := msg.Width - 4 // 减去边距
		if width > 60 {
			width = 60 // 限制最大宽度
		}
		if width < 40 {
			width = 40 // 保证最小宽度
		}
		m.input.Width = width
		return m, nil
	}
	return m, nil
}

func (m *Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.mode {
	case ModeNormal:
		return m.handleNormalMode(msg)
	case ModeInputTitle:
		return m.handleInputMode(msg)
	case ModePickPriority:
		return m.handlePriorityPicker(msg)
	case ModePickDate:
		return m.handleDatePicker(msg)
	}
	return m, nil
}

func (m *Model) handleNormalMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		// 关闭存储连接
		if m.storage != nil {
			m.storage.Close()
		}
		return m, tea.Quit
	case "up", "k":
		m.moveCursor(-1)
		// 更新选中的ID
		if m.cursor >= 0 && m.cursor < len(m.items) {
			m.selectedID = m.items[m.cursor].id
		}
	case "down", "j":
		m.moveCursor(1)
		// 更新选中的ID
		if m.cursor >= 0 && m.cursor < len(m.items) {
			m.selectedID = m.items[m.cursor].id
		}
	case "a":
		m.startAddingItem()
		return m, m.input.Focus()
	case "e":
		m.startEditingItem()
		return m, m.input.Focus()
	case " ":
		m.toggleCompletion()
		return m, nil
	case "x":
		m.deleteCurrentItem()
		return m, nil
	}
	return m, nil
}

func (m *Model) handleInputMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		return m.cancelInput()
	case "enter":
		return m.confirmInput()
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m *Model) handlePriorityPicker(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		return m.exitToNormalMode(), nil
	case "enter":
		return m.confirmPrioritySelection(), nil
	case "up", "k", "left", "h":
		m.rotatePriority(-1)
	case "down", "j", "right", "l":
		m.rotatePriority(1)
	}
	return m, nil
}

func (m *Model) handleDatePicker(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		return m.exitToNormalMode(), nil
	case "enter":
		return m.confirmDateSelection(), nil
	case "left", "h":
		m.rotateDateField(-1)
	case "right", "l":
		m.rotateDateField(1)
	case "up", "k":
		m.adjustDate(1)
	case "down", "j":
		m.adjustDate(-1)
	}
	return m, nil
}

// ====================== 操作辅助方法 ======================

func (m *Model) moveCursor(delta int) {
	newPos := m.cursor + delta
	if newPos >= 0 && newPos < len(m.items) {
		m.cursor = newPos
	}
}

func (m *Model) startAddingItem() {
	m.mode = ModeInputTitle
	m.inputContext = InputContextAddTitle
	m.draftItem = TodoItem{Priority: PriorityMedium, id: generateID()}
	m.input.SetValue("")
	m.input.Placeholder = "新任务内容"
	m.input.Focus()
	m.input.CursorEnd()
	m.statusLine = "输入内容后回车，Esc 取消"
}

func (m *Model) startEditingItem() {
	if len(m.items) == 0 {
		return
	}
	m.mode = ModeInputTitle
	m.inputContext = InputContextEditTitle
	m.input.SetValue(m.items[m.cursor].Title)
	m.input.Placeholder = "编辑内容"
	m.input.Focus()
	m.input.CursorEnd()
	m.statusLine = "修改后回车保存，Esc 取消"
}

// 修改完成任务状态的方法
func (m *Model) toggleCompletion() {
	if len(m.items) == 0 {
		return
	}

	// 记录当前选中任务的ID
	currentID := m.items[m.cursor].id

	// 切换完成状态
	m.items[m.cursor].Done = !m.items[m.cursor].Done

	// 保存更改（这会触发排序）
	m.saveChanges()

	// 根据ID重新定位光标
	m.findItemByID(currentID)
}

// 删除当前项目的方法
func (m *Model) deleteCurrentItem() {
	if len(m.items) == 0 {
		return
	}

	// 删除当前项目
	m.items = append(m.items[:m.cursor], m.items[m.cursor+1:]...)

	// 如果删除后还有项目，调整光标位置
	if len(m.items) > 0 {
		// 如果删除的是最后一个项目，将光标移到新的最后一个项目
		if m.cursor >= len(m.items) {
			m.cursor = len(m.items) - 1
		}
		// 更新选中的ID为当前光标位置的项目ID
		if m.cursor >= 0 && m.cursor < len(m.items) {
			m.selectedID = m.items[m.cursor].id
		}
	} else {
		// 没有项目了，重置光标和选中ID
		m.cursor = 0
		m.selectedID = ""
	}

	m.saveChanges()
}

// 根据ID查找项目并设置光标位置
func (m *Model) findItemByID(id string) {
	for i, item := range m.items {
		if item.id == id {
			m.cursor = i
			m.selectedID = id
			return
		}
	}
	// 如果没找到，尝试选中第一个未完成的项目
	for i, item := range m.items {
		if !item.Done {
			m.cursor = i
			m.selectedID = item.id
			return
		}
	}
	// 如果所有项目都已完成，选中第一个
	if len(m.items) > 0 {
		m.cursor = 0
		m.selectedID = m.items[0].id
	}
}

func (m *Model) cancelInput() (tea.Model, tea.Cmd) {
	m.mode = ModeNormal
	m.statusLine = ""
	m.input.Blur()
	return m, nil
}

func (m *Model) confirmInput() (tea.Model, tea.Cmd) {
	value := strings.TrimSpace(m.input.Value())
	if value == "" {
		m.statusLine = "内容不能为空"
		return m, nil
	}

	switch m.inputContext {
	case InputContextAddTitle:
		m.draftItem.Title = value
		m.draftItem.id = generateID() // 确保有ID
		m.startPriorityPicker(InputContextAddPriority, PriorityMedium)
	case InputContextEditTitle:
		m.items[m.cursor].Title = value
		m.startPriorityPicker(InputContextEditPriority, m.items[m.cursor].Priority)
	}

	m.input.Blur()
	return m, nil
}

func (m *Model) startPriorityPicker(context InputContext, initial Priority) {
	m.mode = ModePickPriority
	m.inputContext = context
	m.priorityPicker.priority = initial
	m.statusLine = "↑/↓ 选择优先级 • Enter 确认 • Esc 取消"
}

func (m *Model) rotatePriority(delta int) {
	current := int(m.priorityPicker.priority)
	newPriority := (current - delta + 3) % 3
	m.priorityPicker.priority = Priority(newPriority)
}

func (m *Model) confirmPrioritySelection() tea.Model {
	if m.inputContext == InputContextAddPriority {
		m.draftItem.Priority = m.priorityPicker.priority
		m.startDatePicker(-1)
	} else {
		// 记录当前选中任务的ID
		currentID := m.items[m.cursor].id

		m.items[m.cursor].Priority = m.priorityPicker.priority

		// 保存更改并排序
		m.saveChanges()

		// 根据ID重新定位光标
		m.findItemByID(currentID)

		m.startDatePicker(m.cursor)
	}
	return m
}

func (m *Model) startDatePicker(index int) {
	m.mode = ModePickDate
	m.datePicker.index = index
	if index >= 0 && m.items[index].HasDeadline {
		m.datePicker.date = m.items[index].Deadline
	} else {
		now := time.Now()
		m.datePicker.date = time.Date(now.Year(), now.Month(), now.Day(),
			17, 0, 0, 0, now.Location())
	}
	m.datePicker.field = DateFieldHour
	m.statusLine = "←/→ 切换字段 • ↑/↓ 调整时间 • Enter 确认 • Esc 取消"
}

func (m *Model) rotateDateField(delta int) {
	current := int(m.datePicker.field)
	newField := (current + delta + 6) % 6
	m.datePicker.field = DateField(newField)
}

func (m *Model) adjustDate(delta int) {
	switch m.datePicker.field {
	case DateFieldYear:
		m.datePicker.date = m.datePicker.date.AddDate(delta, 0, 0)
	case DateFieldMonth:
		m.datePicker.date = m.datePicker.date.AddDate(0, delta, 0)
	case DateFieldDay:
		m.datePicker.date = m.datePicker.date.AddDate(0, 0, delta)
	case DateFieldHour:
		m.datePicker.date = m.datePicker.date.Add(time.Duration(delta) * time.Hour)
	case DateFieldMinute:
		m.datePicker.date = m.datePicker.date.Add(time.Duration(delta) * 10 * time.Minute)
	case DateFieldSecond:
		m.datePicker.date = m.datePicker.date.Add(time.Duration(delta) * 10 * time.Second)
	}
}

func (m *Model) confirmDateSelection() tea.Model {
	if m.inputContext == InputContextAddPriority {
		m.draftItem.HasDeadline = true
		m.draftItem.Deadline = m.datePicker.date
		m.items = append(m.items, m.draftItem)
		m.draftItem = TodoItem{}
		// 保存更改并排序
		m.saveChanges()
		// 选中新添加的项目（应该在最后）
		if len(m.items) > 0 {
			m.cursor = len(m.items) - 1
			m.selectedID = m.items[m.cursor].id
		}
	} else {
		// 记录当前选中任务的ID
		currentID := m.items[m.datePicker.index].id

		m.items[m.datePicker.index].HasDeadline = true
		m.items[m.datePicker.index].Deadline = m.datePicker.date

		// 保存更改并排序
		m.saveChanges()

		// 根据ID重新定位光标
		m.findItemByID(currentID)
	}

	m.mode = ModeNormal
	m.statusLine = ""
	m.input.Blur()
	return m
}

func (m *Model) exitToNormalMode() *Model {
	m.mode = ModeNormal
	m.statusLine = ""
	m.input.Blur()
	return m
}

func (m *Model) saveChanges() {
	if m.storage != nil {
		if msg := m.storage.Save(m.items); msg != "" {
			m.statusLine = msg
		} else {
			m.statusLine = ""
		}
	}
}
