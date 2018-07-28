package model

import "github.com/jinzhu/gorm"

type Note struct {
	gorm.Model `json:"-"`
	Data       *Notedata
	AtmID      uint `json:"-"`
}

type Notedata struct {
	gorm.Model    `json:"-"`
	Denomination  int  `json:"Denomination,omitempty"`
	NumberOfNotes int  `json:"NumberOfNotes,omitempty"`
	NoteID        uint `json:"-"`
}

type Atm struct {
	gorm.Model `json:"-"`
	Notes      []Note
}
