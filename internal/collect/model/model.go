package model

import (
	"fmt"
	"sync"
	"time"

	"gorm.io/gorm"
)

type BashModel struct {
	ID         int64     `gorm:"primarykey;comment:主键ID"`
	CreateTime time.Time `gorm:"autoCreateTime;comment:创建时间"`
	UpdateTime time.Time `gorm:"autoUpdateTime;comment:更新时间"`
	Url        string    `gorm:"index:uk_url,unique,comment:唯一地址索引;type:varchar(200);not null;comment:地址"`
	Rewrite    uint8     `gorm:"type:tinyint(1);not null;default:2;comment:重新写入 1是 2否"`
	Content    string    `gorm:"type:text;not null;comment:内容数据"`
}

type PushSite struct {
	ID         uint      `gorm:"primarykey;comment:主键ID"`
	CreateTime time.Time `gorm:"autoCreateTime;comment:创建时间"`
	UpdateTime time.Time `gorm:"autoUpdateTime;comment:更新时间"`
	State      uint8     `gorm:"type:tinyint(1);not null;default:2;comment:状态 1成功 2失败"`
	Retry      uint      `gorm:"type:int;unsigned;not null;default:0;comment:重试次数"`
	DataId     int64     `gorm:"index:uk_push,unique,comment:唯一索引;unsigned;not null;default:0;comment:内容序号"`
	Domain     string    `gorm:"index:uk_push,unique,comment:唯一索引;type:varchar(50);not null;comment:站内域名"`
	DataTable  string    `gorm:"index:uk_push,unique,comment:唯一索引;type:varchar(50);not null;comment:内容表名"`
}

func Migrator(db *gorm.DB) error {
	if !db.Migrator().HasTable(&PushSite{}) {
		if err := db.
			Set("gorm:table_options", "ENGINE=InnoDB CHARSET=utf8mb4 COMMENT='发布表'").
			AutoMigrate(&PushSite{}); err != nil {
			return err
		}
	}

	if err := createTableIfNotExists(db, &WwwDownyiCom{}); err != nil {
		return err
	}

	if err := createTableIfNotExists(db, &Www7K7k7Com{}); err != nil {
		return err
	}

	if err := createTableIfNotExists(db, &Www333tttCom{}); err != nil {
		return err
	}

	return nil
}

// 创建表的辅助函数
func createTableIfNotExists(db *gorm.DB, model interface{}) error {
	if !db.Migrator().HasTable(model) {
		if err := db.Set("gorm:table_options", "ENGINE=InnoDB CHARSET=utf8mb4 COMMENT='数据表'").
			AutoMigrate(model); err != nil {
			return fmt.Errorf("failed to create table for %v: %w", model, err)
		}
	}
	return nil
}

type Impl interface {
	Process(db *gorm.DB, url string, data map[string]any) error
}

var (
	Services = make(map[string]Impl)
	mutex    sync.Mutex
)

// Register a service
func Register(key string, val Impl) {
	mutex.Lock()
	defer mutex.Unlock()
	Services[key] = val
}
