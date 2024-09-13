package model

import (
	"encoding/json"
	"fmt"

	"gorm.io/gorm"
)

func init() {
	Register("www_333ttt_com", NewWww333tttCom())
}

type Www333tttCom struct {
	BashModel
}

func NewWww333tttCom() Impl {
	return &Www333tttCom{}
}

func (d *Www333tttCom) Process(db *gorm.DB, url string, data map[string]any) error {
	if len(data) == 0 {
		return fmt.Errorf("no data, url: %s", url)
	}
	dj, err := json.Marshal(data)
	if err != nil {
		return err
	}

	item := Www333tttCom{
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

func (d *Www333tttCom) TableName() string {
	return "www_333ttt_com"
}
