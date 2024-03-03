package components

import (
	"github.com/qor5/ui/vuetify"
	"github.com/theplant/htmlgo"
	"gorm.io/gorm"
)

func Dropdown[T any](db *gorm.DB, value uint, label, item, field string) htmlgo.HTMLComponent {
	var comps []T
	db.Find(&comps)
	return htmlgo.Div(
		vuetify.VSelect().
			Label(label).
			Items(comps).
			ItemText(item).
			ItemValue("ID").
			Value(value).
			FieldName(field),
	)
}

type DropdownListItem struct {
	Name  string
	Value string
}

func DropdownList(items []DropdownListItem, value string, label, field string) htmlgo.HTMLComponent {
	return htmlgo.Div(
		vuetify.VSelect().
			Label(label).
			Items(items).
			ItemText("Name").
			ItemValue("Value").
			Value(value).
			FieldName(field),
	)
}
