package models

import (
	"database/sql/driver"
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

// GeoPoint - структура гео-точки.
type GeoPoint struct {
	Lat float64
	Lon float64
}

// Value - значение в которое должен возвращать драйвер БД.
func (gp GeoPoint) Value() (driver.Value, error) {
	return fmt.Sprintf("(%f,%f)", gp.Lat, gp.Lon), nil
}

// Latitude - широта.
func (gp GeoPoint) Latitude() float64 {
	return gp.Lat
}

// Longitude - долгота.
func (gp GeoPoint) Longitude() float64 {
	return gp.Lon
}

// Scan - метод сканирования. Подходит для pgx, pgxpool, pq
// C ORM не проверялся.
func (gp *GeoPoint) Scan(src interface{}) error {
	raw, ok := src.(string)

	if !ok {
		return errors.New("can't switch type")
	}

	start := strings.Index(raw, "(")
	end := strings.Index(raw, ")")

	if start < 0 || end < 0 {
		return errors.New("Index == -1 ")
	}

	coords := raw[start+1 : end]
	values := strings.Split(coords, " ")

	// nolint:gomnd // нужно только 2 значения
	if len(values) < 2 {
		return errors.New("Too few values")
	}

	var err error
	// nolint:gomnd // 64 бита флоат
	if gp.Lat, err = strconv.ParseFloat(values[1], 64); err != nil {
		return errors.Wrap(err, "err when parse latitude")
	}

	// nolint:gomnd // 64 бита флоат
	if gp.Lon, err = strconv.ParseFloat(values[0], 64); err != nil {
		return errors.Wrap(err, "err when parse longitude")
	}

	return nil
}
