package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ====================== 视图渲染 ======================

func (m *Model) View() string {
	var builder strings.Builder

	// 标题栏
	builder.WriteString("\n" + m.renderHeader() + "\n\n")

	// 任务列表
	if len(m.items) == 0 {
		builder.WriteString(m.renderEmptyState())
	} else {
		builder.WriteString(m.renderTodoTable())
	}

	// 交互区域
	builder.WriteString(m.renderInteractiveArea())

	// 状态栏
	if m.statusLine != "" {
		builder.WriteString("\n  " + m.styles.Status.Render("● ") + m.statusLine)
	}

	return builder.String()
}

func (m *Model) renderHeader() string {
	doneCount := 0
	for _, item := range m.items {
		if item.Done {
			doneCount++
		}
	}

	title := m.styles.Header.
		Background(lipgloss.Color("62")).
		Foreground(lipgloss.Color("255")).
		Render(" TODO ")

	stats := m.styles.Deadline.Render(
		fmt.Sprintf(" %d/%d 已完成", doneCount, len(m.items)),
	)

	return " " + title + stats
}

func (m *Model) renderEmptyState() string {
	return "  " + m.styles.Help.Render("暂无任务，按 a 开始添加") + "\n"
}

func (m *Model) renderTodoTable() string {
	// 定义列宽
	const (
		statusColWidth   = 6  // [ ] 状态
		priorityColWidth = 8  // "优先级"（3个中文字符，实际宽度为6，加2个空格）
		titleColWidth    = 40 // 任务标题
		deadlineColWidth = 19 // 截止日期 (YYYY-MM-DD HH:MM)
	)

	// 创建表格边框样式
	tableStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("236"))

	// 表格头部
	header := lipgloss.JoinHorizontal(
		lipgloss.Center,
		m.styles.TableHeader.Width(statusColWidth).Render("状态"),
		m.styles.TableHeader.Width(priorityColWidth).Render("优先级"),
		m.styles.TableHeader.Width(titleColWidth).Render("任务"),
		m.styles.TableHeader.Width(deadlineColWidth).Render("截止日期"),
	)

	// 表格行
	var rows []string
	for i, item := range m.items {
		// 判断是否是当前选中的行
		isSelected := item.id == m.selectedID
		row := m.renderTableRow(i, item, statusColWidth, priorityColWidth, titleColWidth, deadlineColWidth, isSelected)
		rows = append(rows, row)
	}

	// 计算总宽度
	totalWidth := statusColWidth + priorityColWidth + titleColWidth + deadlineColWidth

	// 组合表格
	table := lipgloss.JoinVertical(lipgloss.Left,
		header,
		m.styles.TableBorder.Render(strings.Repeat("─", totalWidth)),
	)
	table += "\n" + strings.Join(rows, "\n")

	return tableStyle.Render(table)
}

func (m *Model) renderTableRow(index int, item TodoItem, statusWidth, priorityWidth, titleWidth, deadlineWidth int, isSelected bool) string {
	// 状态列
	var status string
	if item.Done {
		status = m.styles.Checkbox.Render("✓")
	} else {
		status = "○"
	}
	if isSelected {
		status = m.styles.Cursor.Render("▶ ") + status
	}

	statusStyle := m.styles.TableHeader.Width(statusWidth).Align(lipgloss.Left)
	if isSelected {
		statusStyle = statusStyle.Background(lipgloss.Color("235"))
	}
	statusCell := statusStyle.Render(status)

	// 优先级列 - 确保内容居中
	var priority string
	switch item.Priority {
	case PriorityHigh:
		priority = m.styles.PriorityHigh.Render(item.Priority.String())
	case PriorityMedium:
		priority = m.styles.PriorityMid.Render(item.Priority.String())
	case PriorityLow:
		priority = m.styles.PriorityLow.Render(item.Priority.String())
	}

	priorityStyle := m.styles.TableHeader.Width(priorityWidth).Align(lipgloss.Center)
	if isSelected {
		priorityStyle = priorityStyle.Background(lipgloss.Color("235"))
	}
	priorityCell := priorityStyle.Render(priority)

	// 标题列
	title := item.Title
	if len(title) > titleWidth-2 {
		title = title[:titleWidth-5] + "..."
	}
	if item.Done {
		title = m.styles.Done.Render(title)
	}

	titleStyle := m.styles.TableHeader.Width(titleWidth).Align(lipgloss.Left)
	if isSelected {
		titleStyle = titleStyle.Background(lipgloss.Color("235"))
	}
	titleCell := titleStyle.Render(title)

	// 截止日期列
	var deadline string
	if item.HasDeadline {
		deadline = item.DeadlineString()
		if item.IsOverdue() {
			deadline = m.styles.Overdue.Render(deadline)
		} else {
			deadline = m.styles.Deadline.Render(deadline)
		}
	} else {
		deadline = m.styles.Deadline.Render("-")
	}

	deadlineStyle := m.styles.TableHeader.Width(deadlineWidth).Align(lipgloss.Left)
	if isSelected {
		deadlineStyle = deadlineStyle.Background(lipgloss.Color("235"))
	}
	deadlineCell := deadlineStyle.Render(deadline)

	// 组合行 - 不再需要额外的背景色，因为每个单元格已经有了
	return statusCell + priorityCell + titleCell + deadlineCell
}

func (m *Model) renderInteractiveArea() string {
	var content string

	switch m.mode {
	case ModeNormal:
		content = m.renderHelp()
	case ModeInputTitle:
		content = "\n  " + m.input.View() + "\n"
	case ModePickPriority:
		content = "\n" + m.renderPriorityPicker()
	case ModePickDate:
		content = "\n  " + m.renderDatePicker() + "\n"
	}

	return content
}

func (m *Model) renderHelp() string {
	if m.mode != ModeNormal {
		return ""
	}
	return "\n" + m.styles.Help.Render(
		"  ↑/↓ 移动 • a 添加 • e 编辑 • 空格 完成 • x 删除 • q 退出",
	) + "\n"
}

func (m *Model) renderPriorityPicker() string {
	var builder strings.Builder
	builder.WriteString(" 选择优先级：\n\n")

	priorities := []Priority{PriorityHigh, PriorityMedium, PriorityLow}
	for _, p := range priorities {
		pStr := p.String()
		if p == m.priorityPicker.priority {
			builder.WriteString("  " + m.styles.Cursor.Render("┃ ") +
				m.styles.Selected.Render(pStr) + "\n")
		} else {
			builder.WriteString("    " + m.styles.Deadline.Render(pStr) + "\n")
		}
	}

	return builder.String()
}

func (m *Model) renderDatePicker() string {
	date := m.datePicker.date

	// 格式化和高亮选中的字段
	formatField := func(value string, field DateField) string {
		if m.datePicker.field == field {
			return m.styles.Selected.Render(value)
		}
		return value
	}

	year := formatField(fmt.Sprintf("%04d", date.Year()), DateFieldYear)
	month := formatField(fmt.Sprintf("%02d", int(date.Month())), DateFieldMonth)
	day := formatField(fmt.Sprintf("%02d", date.Day()), DateFieldDay)
	hour := formatField(fmt.Sprintf("%02d", date.Hour()), DateFieldHour)
	minute := formatField(fmt.Sprintf("%02d", date.Minute()), DateFieldMinute)
	second := formatField(fmt.Sprintf("%02d", date.Second()), DateFieldSecond)

	return fmt.Sprintf("截止日期：%s-%s-%s %s:%s:%s",
		year, month, day, hour, minute, second)
}
