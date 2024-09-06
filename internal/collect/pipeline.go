package collect

import (
	"gorm.io/gorm"
	"zxkj.com/zxkj_spider_go/internal/collect/model"
)

type Pipeline interface {
	Process(item *Item) error
}

type MySQLPipeline struct {
	db *gorm.DB
}

func NewMySQLPipeline(db *gorm.DB) Pipeline {
	return &MySQLPipeline{
		db: db,
	}
}

// Process
// data 数据
// name 表名
// 通过name实例化表，写入相关数据表
func (m *MySQLPipeline) Process(item *Item) error {
	// 遍历服务
	for key, serv := range model.Services {
		if key == item.Name {
			err := serv.Process(m.db, item.Url, item.Data)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
