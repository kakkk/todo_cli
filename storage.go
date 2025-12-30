package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// 数据库模型
type TodoModel struct {
	ID          string    `gorm:"primaryKey;size:50"`
	Title       string    `gorm:"not null"`
	Done        bool      `gorm:"default:false"`
	Priority    int       `gorm:"not null"`
	HasDeadline bool      `gorm:"default:false"`
	Deadline    time.Time `gorm:"default:null"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"`
}

// 转换为内存模型
func (tm *TodoModel) ToTodoItem() TodoItem {
	return TodoItem{
		Title:       tm.Title,
		Done:        tm.Done,
		Priority:    Priority(tm.Priority),
		HasDeadline: tm.HasDeadline,
		Deadline:    tm.Deadline,
		id:          tm.ID,
	}
}

// 从内存模型转换
func TodoItemToModel(item *TodoItem) *TodoModel {
	return &TodoModel{
		ID:          item.id,
		Title:       item.Title,
		Done:        item.Done,
		Priority:    int(item.Priority),
		HasDeadline: item.HasDeadline,
		Deadline:    item.Deadline,
	}
}

type Storage struct {
	db *gorm.DB
}

// 创建新的存储实例
func NewStorage() (*Storage, error) {
	db, err := gorm.Open(sqlite.Open(dataFile()), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("打开数据库失败: %v", err)
	}

	// 自动迁移数据库结构
	err = db.AutoMigrate(&TodoModel{})
	if err != nil {
		return nil, fmt.Errorf("迁移数据库失败: %v", err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) Load() (TodoList, string) {
	var models []TodoModel
	result := s.db.Order("done asc, priority desc, has_deadline desc, deadline asc, title asc").Find(&models)
	if result.Error != nil {
		return nil, fmt.Sprintf("加载失败: %v", result.Error)
	}

	items := make(TodoList, 0, len(models))
	for _, model := range models {
		items = append(items, model.ToTodoItem())
	}

	// 额外排序一次确保内存中的顺序正确
	items.Sort()
	return items, ""
}

func (s *Storage) Save(items TodoList) string {
	// 首先对传入的列表进行排序
	items.Sort() // 这里缺少了这个调用！

	// 开启事务
	tx := s.db.Begin()
	if tx.Error != nil {
		return fmt.Sprintf("开始事务失败: %v", tx.Error)
	}

	// 获取所有现有ID
	var existingIDs []string
	if err := tx.Model(&TodoModel{}).Pluck("id", &existingIDs).Error; err != nil {
		tx.Rollback()
		return fmt.Sprintf("查询现有数据失败: %v", err)
	}

	// 将要保存的ID映射
	savingIDs := make(map[string]bool)
	for _, item := range items {
		savingIDs[item.id] = true
	}

	// 删除不在新列表中的项目
	for _, id := range existingIDs {
		if !savingIDs[id] {
			if err := tx.Where("id = ?", id).Delete(&TodoModel{}).Error; err != nil {
				tx.Rollback()
				return fmt.Sprintf("删除项目失败: %v", err)
			}
		}
	}

	// 更新或创建项目
	for _, item := range items {
		model := TodoItemToModel(&item)
		var count int64
		tx.Model(&TodoModel{}).Where("id = ?", item.id).Count(&count)

		if count > 0 {
			// 更新
			if err := tx.Model(&TodoModel{}).Where("id = ?", item.id).Updates(map[string]interface{}{
				"title":        item.Title,
				"done":         item.Done,
				"priority":     int(item.Priority),
				"has_deadline": item.HasDeadline,
				"deadline":     item.Deadline,
			}).Error; err != nil {
				tx.Rollback()
				return fmt.Sprintf("更新项目失败: %v", err)
			}
		} else {
			// 创建
			if err := tx.Create(model).Error; err != nil {
				tx.Rollback()
				return fmt.Sprintf("创建项目失败: %v", err)
			}
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return fmt.Sprintf("提交事务失败: %v", err)
	}

	return ""
}

// 关闭数据库连接
func (s *Storage) Close() error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func dataFile() string {
	homePath, err := os.UserHomeDir()
	if err != nil {
		// 兜底保存在当前运行目录下
		return "todo_cli.db"
	}
	dir := filepath.Join(homePath, ".todo_cli")
	err = ensureDir(dir)
	if err != nil {
		// 兜底保存在当前运行目录下
		return "todo_cli.db"
	}
	return filepath.Join(dir, "todo_cli.db")
}

func ensureDir(dirPath string) error {
	// 获取文件信息
	info, err := os.Stat(dirPath)
	if err == nil {
		// 路径存在
		if !info.IsDir() {
			return fmt.Errorf("路径 %s 已存在但不是目录", dirPath)
		}
		return nil
	}

	// 检查是否是"不存在"的错误
	if os.IsNotExist(err) {
		// 创建目录
		fmt.Printf("创建目录: %s\n", dirPath)
		return os.MkdirAll(dirPath, 0755)
	}

	return fmt.Errorf("检查目录失败: %v", err)
}
