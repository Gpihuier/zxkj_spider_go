package model

import (
	"encoding/json"

	"gorm.io/gorm"
)

func init() {
	Register("downyi_com", NewDownyiCom())
}

type DownyiCom struct {
	BashModel
}

func NewDownyiCom() Impl {
	return &DownyiCom{}
}

func (d *DownyiCom) Process(db *gorm.DB, url string, data map[string]any) error {
	dj, err := json.Marshal(data)
	if err != nil {
		return err
	}

	item := DownyiCom{
		BashModel{
			Url:     url,
			Content: string(dj),
		},
	}

	if err = db.Create(&item).Error; err != nil {
		return err
	}
	return nil

}
