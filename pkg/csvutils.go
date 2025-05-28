package pkg

import (
	"encoding/csv"
	"strconv"
)

// Exportar o cabeçalho padrão
func WriteCSVHeader(csvWriter *csv.Writer) error {
	header := []string{
		"FileName", "FilePath", "FileSize", "LastModified", "Permissions",
		"MD5", "SHA1", "SHA256",
		"Make", "Model", "DateTime", "ImageWidth", "ImageHeight", "ISO",
		"Aperture", "ExposureTime", "FocalLength", "Orientation",
		"GPSLatitude", "GPSLongitude", "GPSAltitude", "GPSDate", "GPSTime",
		"GoogleMapsLink",
	}
	return csvWriter.Write(header)
}

// Exportar um registro de FileInfoData
func WriteCSVRecord(csvWriter *csv.Writer, data *FileInfoData) error {
	record := []string{
		data.FileName, data.FilePath, strconv.FormatInt(data.FileSize, 10), data.LastModified, data.Permissions,
		data.MD5, data.SHA1, data.SHA256,
		data.ExifMake, data.ExifModel, data.ExifDateTime, data.ExifImageWidth, data.ExifImageHeight, data.ExifISO,
		data.ExifAperture, data.ExifExposureTime, data.ExifFocalLength, data.ExifOrientation,
		data.GPSLatitude, data.GPSLongitude, data.GPSAltitude, data.GPSDate, data.GPSTime,
		data.GoogleMapsLink,
	}
	return csvWriter.Write(record)
}
