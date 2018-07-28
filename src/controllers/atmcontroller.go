package controllers

import (
	"errors"
	"model"

	"github.com/jinzhu/gorm"
)

type atmcontroller interface {
	Dispense(amount int) (notes []model.Note, err error)
}

/*
*
 */
func Dispense(db *gorm.DB, amount int) (retNotes []model.Note, err error) {
	retNotes = make([]model.Note, 0)
	notes := make([]model.Note, 4)

	db.Find(&notes)

	for i := len(notes) - 1; i >= 0; i-- {
		note := model.Notedata{}
		var retNote *model.Notedata
		db.First(&note, i+1)
		for amount >= note.Denomination && note.NumberOfNotes > 0 {
			if retNote == nil {
				retNote = &model.Notedata{}
			}
			amount -= note.Denomination
			note.NumberOfNotes--
			retNote.Denomination = note.Denomination
			retNote.NumberOfNotes++
		}

		if retNote != nil {
			retNotes = append(retNotes, model.Note{Data: retNote})
		}
		db.Save(&note)
		db.Save(&notes[i])
	}

	if amount != 0 {
		return nil, errors.New("InsufficientNotesInAtm")
	}

	return retNotes, nil
}
