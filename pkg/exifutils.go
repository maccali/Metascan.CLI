package pkg

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/rwcarlsen/goexif/exif"
)

func GetFormattedRational(x *exif.Exif, name exif.FieldName, asDecimal bool, prefix string) (string, bool) {
	tag, err := x.Get(name)
	if err != nil {
		return "", false
	}
	ratVal, errRat := tag.Rat(0)
	if errRat != nil {
		return "", false
	}
	num := ratVal.Num()
	den := ratVal.Denom()
	if den.Sign() == 0 {
		return fmt.Sprintf("%s%s/0 (Invalid)", prefix, num.String()), true
	}
	if asDecimal {
		precision := 2
		floatVal, _ := ratVal.Float64()
		if name == exif.ExposureTime {
			if floatVal < 0.1 {
				precision = 3
			}
			if floatVal < 0.01 {
				precision = 4
			}
			if floatVal < 0.001 {
				precision = 5
			}
		}
		return fmt.Sprintf("%s%.*f", prefix, precision, floatVal), true
	}
	if den.Cmp(big.NewInt(1)) == 0 {
		return fmt.Sprintf("%s%s", prefix, num.String()), true
	}
	return fmt.Sprintf("%s%s/%s", prefix, num.String(), den.String()), true
}

func IsExifExpectedError(err error) bool {
	if err == nil {
		return false
	}
	expectedErrors := []string{
		"exif: missing EXIF mark", "exif: data format not supported",
		"tiff: short tag value", "tiff: invalid ExifIFD pointer", "EOF",
	}
	for _, e := range expectedErrors {
		if strings.Contains(err.Error(), e) {
			return true
		}
	}
	return false
}

type FileInfoData struct {
	FileName         string
	FilePath         string
	FileSize         int64
	LastModified     string
	Permissions      string
	MD5              string
	SHA1             string
	SHA256           string
	ExifMake         string
	ExifModel        string
	ExifDateTime     string
	ExifImageWidth   string
	ExifImageHeight  string
	ExifISO          string
	ExifAperture     string
	ExifExposureTime string
	ExifFocalLength  string
	ExifOrientation  string
	GPSLatitude      string
	GPSLongitude     string
	GPSAltitude      string
	GPSDate          string
	GPSTime          string
	GoogleMapsLink   string
}

func ProcessFile(filePath string) (*FileInfoData, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("error getting info for '%s': %w", filePath, err)
	}
	if fileInfo.IsDir() {
		return nil, nil
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("error opening '%s': %w", filePath, err)
	}
	defer file.Close()

	exifDataMap := make(map[string]string)
	var gpsLat, gpsLong string

	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		log.Printf("Warning (EXIF Seek): Could not rewind '%s': %v", filePath, err)
	} else {
		x, errDecode := exif.Decode(file)
		if errDecode == nil {
			if val, ok := GetStringVal(x, exif.Make); ok {
				exifDataMap["Make"] = val
			}
			if val, ok := GetStringVal(x, exif.Model); ok {
				exifDataMap["Model"] = val
			}
			if val, ok := GetStringVal(x, exif.Software); ok {
				exifDataMap["Software"] = val
			}
			if val, ok := GetStringVal(x, exif.DateTimeOriginal); ok {
				exifDataMap["DateTimeOriginal"] = val
			} else if val, ok := GetStringVal(x, exif.DateTime); ok {
				exifDataMap["DateTime"] = val
			}
			if val, ok := GetIntVal(x, exif.PixelXDimension); ok {
				exifDataMap["ImageWidth"] = fmt.Sprintf("%d", val)
			}
			if val, ok := GetIntVal(x, exif.PixelYDimension); ok {
				exifDataMap["ImageHeight"] = fmt.Sprintf("%d", val)
			}
			if val, ok := GetIntVal(x, exif.ISOSpeedRatings); ok {
				exifDataMap["ISO"] = fmt.Sprintf("%d", val)
			}
			if orientVal, ok := GetIntVal(x, exif.Orientation); ok {
				switch orientVal {
				case 1:
					exifDataMap["Orientation"] = "Normal"
				default:
					exifDataMap["Orientation"] = fmt.Sprintf("%d", orientVal)
				}
			}
			if val, ok := GetFormattedRational(x, exif.FNumber, true, ""); ok {
				exifDataMap["Aperture"] = "f/" + val
			}
			if val, ok := GetFormattedRational(x, exif.ExposureTime, false, ""); ok {
				exifDataMap["ExposureTime"] = val
			}
			if val, ok := GetFormattedRational(x, exif.FocalLength, true, ""); ok {
				exifDataMap["FocalLength"] = val
			}

			latGPS, longGPS, errLatLong := x.LatLong()
			if errLatLong == nil {
				gpsLat = fmt.Sprintf("%.6f", latGPS)
				gpsLong = fmt.Sprintf("%.6f", longGPS)

				if altTag, errAlt := x.Get(exif.GPSAltitude); errAlt == nil {
					ratVal, errRat := altTag.Rat(0)
					if errRat == nil && ratVal.Denom().Sign() != 0 {
						altitude, _ := ratVal.Float64()
						altRefStr := ""
						if altRefVal, okRef := GetIntVal(x, exif.GPSAltitudeRef); okRef && altRefVal == 1 {
							altRefStr = " (Below Sea Level)"
						}
						exifDataMap["GPSAltitude"] = fmt.Sprintf("%.2f%s", altitude, altRefStr)
					}
				}
				if val, ok := GetStringVal(x, exif.GPSDateStamp); ok {
					exifDataMap["GPSDate"] = val
				}
				if tsTag, errTs := x.Get(exif.GPSTimeStamp); errTs == nil {
					var timePartsStr [3]string
					validTimeParts := 0
					for i := 0; i < 3; i++ {
						ratVal, errRat := tsTag.Rat(i)
						if errRat == nil && ratVal.Denom().Sign() != 0 {
							if ratVal.IsInt() {
								timePartsStr[i] = ratVal.Num().String()
								validTimeParts++
							} else {
								floatVal, _ := ratVal.Float64()
								timePartsStr[i] = fmt.Sprintf("%d", int(floatVal))
								validTimeParts++
							}
						} else {
							break
						}
					}
					if validTimeParts == 3 {
						h, _ := strconv.Atoi(timePartsStr[0])
						m, _ := strconv.Atoi(timePartsStr[1])
						s, _ := strconv.Atoi(timePartsStr[2])
						exifDataMap["GPSTime"] = fmt.Sprintf("%02d:%02d:%02d", h, m, s)
					}
				}
			}
		} else if !IsExifExpectedError(errDecode) {
			log.Printf("Warning (EXIF Decode): Could not decode '%s': %v", filePath, errDecode)
		}
	}

	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		return nil, fmt.Errorf("critical error rewinding '%s' for hash: %w", filePath, err)
	}
	hMd5 := md5.New()
	hSha1 := sha1.New()
	hSha256 := sha256.New()
	multiWriter := io.MultiWriter(hMd5, hSha1, hSha256)
	if _, err := io.Copy(multiWriter, file); err != nil {
		return nil, fmt.Errorf("error calculating hashes for '%s': %w", filePath, err)
	}
	md5Sum := hex.EncodeToString(hMd5.Sum(nil))
	sha1Sum := hex.EncodeToString(hSha1.Sum(nil))
	sha256Sum := hex.EncodeToString(hSha256.Sum(nil))

	absPath, _ := filepath.Abs(filePath)

	googleMapsLink := ""
	if gpsLat != "" && gpsLong != "" {
		googleMapsLink = fmt.Sprintf("https://www.google.com/maps?q=%s,%s", url.QueryEscape(gpsLat), url.QueryEscape(gpsLong))
	}

	return &FileInfoData{
		FileName:         fileInfo.Name(),
		FilePath:         absPath,
		FileSize:         fileInfo.Size(),
		LastModified:     fileInfo.ModTime().Format(time.RFC3339),
		Permissions:      fileInfo.Mode().String(),
		MD5:              md5Sum,
		SHA1:             sha1Sum,
		SHA256:           sha256Sum,
		ExifMake:         exifDataMap["Make"],
		ExifModel:        exifDataMap["Model"],
		ExifDateTime:     exifDataMap["DateTimeOriginal"],
		ExifImageWidth:   exifDataMap["ImageWidth"],
		ExifImageHeight:  exifDataMap["ImageHeight"],
		ExifISO:          exifDataMap["ISO"],
		ExifAperture:     exifDataMap["Aperture"],
		ExifExposureTime: exifDataMap["ExposureTime"],
		ExifFocalLength:  exifDataMap["FocalLength"],
		ExifOrientation:  exifDataMap["Orientation"],
		GPSLatitude:      gpsLat,
		GPSLongitude:     gpsLong,
		GPSAltitude:      exifDataMap["GPSAltitude"],
		GPSDate:          exifDataMap["GPSDate"],
		GPSTime:          exifDataMap["GPSTime"],
		GoogleMapsLink:   googleMapsLink,
	}, nil
}
