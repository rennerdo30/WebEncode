package store

import (
	"github.com/jackc/pgx/v5/pgtype"
)

func ToText(s string) pgtype.Text {
	return pgtype.Text{String: s, Valid: s != ""}
}

func ToInt4(i int32) pgtype.Int4 {
	return pgtype.Int4{Int32: i, Valid: true}
}

func ToBool(b bool) pgtype.Bool {
	return pgtype.Bool{Bool: b, Valid: true}
}

func ToUUID(s string) pgtype.UUID {
	var u pgtype.UUID
	u.Scan(s)
	return u
}
