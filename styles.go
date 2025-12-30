package main

import "github.com/charmbracelet/lipgloss"

// ====================== 样式管理 ======================

type Styles struct {
	Header       lipgloss.Style
	Cursor       lipgloss.Style
	Checkbox     lipgloss.Style
	PriorityHigh lipgloss.Style
	PriorityMid  lipgloss.Style
	PriorityLow  lipgloss.Style
	Done         lipgloss.Style
	Overdue      lipgloss.Style
	Deadline     lipgloss.Style
	Selected     lipgloss.Style
	Help         lipgloss.Style
	Status       lipgloss.Style
	TableHeader  lipgloss.Style
	TableBorder  lipgloss.Style
	TableCell    lipgloss.Style
	SelectedRow  lipgloss.Style // 添加选中行样式
}

func NewStyles() *Styles {
	return &Styles{
		Header:       lipgloss.NewStyle().Bold(true),
		Cursor:       lipgloss.NewStyle().Foreground(lipgloss.Color("62")).Bold(true),
		Checkbox:     lipgloss.NewStyle().Foreground(lipgloss.Color("62")),
		PriorityHigh: lipgloss.NewStyle().Foreground(lipgloss.Color("160")).Bold(true),
		PriorityMid:  lipgloss.NewStyle().Foreground(lipgloss.Color("214")),
		PriorityLow:  lipgloss.NewStyle().Foreground(lipgloss.Color("34")),
		Done:         lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Strikethrough(true),
		Overdue:      lipgloss.NewStyle().Foreground(lipgloss.Color("160")).Bold(true),
		Deadline:     lipgloss.NewStyle().Foreground(lipgloss.Color("242")),
		Selected:     lipgloss.NewStyle().Foreground(lipgloss.Color("14")).Bold(true),
		Help:         lipgloss.NewStyle().Foreground(lipgloss.Color("242")),
		Status:       lipgloss.NewStyle().Foreground(lipgloss.Color("220")),
		TableHeader:  lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("240")),
		TableBorder:  lipgloss.NewStyle().Foreground(lipgloss.Color("236")),
		TableCell:    lipgloss.NewStyle().Padding(0, 1),
		SelectedRow:  lipgloss.NewStyle().Background(lipgloss.Color("235")), // 选中行背景色
	}
}
