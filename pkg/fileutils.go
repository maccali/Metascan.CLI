package pkg

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"math/big"
	"os"
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

type FileHashes struct {
	MD5    string `json:"md5"`
	SHA1   string `json:"sha1"`
	SHA256 string `json:"sha256"`
}

func CalcFileHashes(filePath string) (FileHashes, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return FileHashes{}, err
	}
	defer file.Close()

	hMd5 := md5.New()
	hSha1 := sha1.New()
	hSha256 := sha256.New()

	multiWriter := io.MultiWriter(hMd5, hSha1, hSha256)
	if _, err := io.Copy(multiWriter, file); err != nil {
		return FileHashes{}, err
	}

	return FileHashes{
		MD5:    hex.EncodeToString(hMd5.Sum(nil)),
		SHA1:   hex.EncodeToString(hSha1.Sum(nil)),
		SHA256: hex.EncodeToString(hSha256.Sum(nil)),
	}, nil
}
