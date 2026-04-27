package geotypes

import (
	"database/sql/driver"
	"encoding/binary"
	"fmt"

	"github.com/twpayne/go-geom"
	"github.com/twpayne/go-geom/encoding/ewkb"
)

type NullPoint struct {
	Point *geom.Point
	Valid bool
}

func (n *NullPoint) Scan(value any) error {
	if value == nil {
		n.Point, n.Valid = nil, false
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("NullPoint: expected []byte, got %T", value)
	}
	g, err := ewkb.Unmarshal(b)
	if err != nil {
		return err
	}
	pt, ok := g.(*geom.Point)
	if !ok {
		return fmt.Errorf("NullPoint: expected *geom.Point, got %T", g)
	}
	n.Point, n.Valid = pt, true
	return nil
}

func (n NullPoint) Value() (driver.Value, error) {
	if !n.Valid || n.Point == nil {
		return nil, nil
	}
	return ewkb.Marshal(n.Point, binary.LittleEndian)
}
