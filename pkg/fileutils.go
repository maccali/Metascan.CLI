package pkg

import (
	"math/big"
	"strings"

	"github.com/rwcarlsen/goexif/exif"
)

func GetStringVal(x *exif.Exif, name exif.FieldName) (string, bool) {
	tag, err := x.Get(name)
	if err != nil {
		return "", false
	}
	valStr := strings.TrimSpace(tag.String())
	valStr = strings.Trim(valStr, "\"")
	return valStr, true
}

func GetIntVal(x *exif.Exif, name exif.FieldName) (int64, bool) {
	tag, err := x.Get(name)
	if err != nil {
		return 0, false
	}
	val, err := tag.Int(0)
	if err != nil {
		ratVal, errRat := tag.Rat(0)
		if errRat == nil {
			if ratVal.IsInt() && ratVal.Denom().Cmp(big.NewInt(1)) == 0 {
				return ratVal.Num().Int64(), true
			}
		}
		return 0, false
	}
	return int64(val), true
}
