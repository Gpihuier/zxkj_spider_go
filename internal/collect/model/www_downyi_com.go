package model

import (
	"encoding/json"

	"gorm.io/gorm"
)

func init() {
	Register("www_downyi_com", NewDownyiCom())
}

type WwwDownyiCom struct {
	BashModel
}

func NewDownyiCom() Impl {
	return &WwwDownyiCom{}
}

func (d *WwwDownyiCom) Process(db *gorm.DB, url string, data map[string]any) error {
	dj, err := json.Marshal(data)
	if err != nil {
		return err
	}

	item := WwwDownyiCom{
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

func (d *WwwDownyiCom) TableName() string {
	return "www_downyi_com"
}
