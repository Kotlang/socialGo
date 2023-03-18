package models

type ContentMasterModel struct {
	Language string   `bson:"language"`
	Field    string   `bson:"field"`
	Values   []string `bson:"options"`
}

func (m *ContentMasterModel) Id() string {
	return m.Language + "/" + m.Field
}
