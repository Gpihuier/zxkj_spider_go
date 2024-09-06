package model

import (
	"encoding/json"

	"gorm.io/gorm"
)

func init() {
	Register("www_7k7k7_com", NewWww7K7k7Com())
}

type Www7K7k7Com struct {
	BashModel
}

func NewWww7K7k7Com() Impl {
	return &Www7K7k7Com{}
}

func (d *Www7K7k7Com) Process(db *gorm.DB, url string, data map[string]any) error {
	dj, err := json.Marshal(data)
	if err != nil {
		return err
	}

	item := Www7K7k7Com{
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

func (d *Www7K7k7Com) TableName() string {
	return "www_7k7k7_com"
}
