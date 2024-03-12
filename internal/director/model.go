package director

import (
	"fmt"
	"strings"
	"time"
)

type Director struct {
	ID        string     `json:"id"`
	FirstName string     `json:"first_name,omitempty"`
	LastName  string     `json:"last_name,omitempty"`
	BirthDate CustomDate `json:"birth_date,omitempty"`
	Country   string     `json:"country,omitempty"`
	HasOscar  bool       `json:"has_oscar,omitempty"`
}

type CustomDate struct {
	time.Time
}

const layout = "2006-01-02"

func (c *CustomDate) UnmarshalJSON(b []byte) (err error) {
	s := strings.Trim(string(b), `"`) // remove quotes
	if s == "null" {
		return
	}
	c.Time, err = time.Parse(layout, s)
	return
}

func (c *CustomDate) MarshalJSON() ([]byte, error) {
	if c.Time.IsZero() {
		return nil, nil
	}
	return []byte(fmt.Sprintf(`"%s"`, c.Time.Format(layout))), nil
}
