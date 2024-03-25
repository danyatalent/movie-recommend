package director

type Director struct {
	ID        string `json:"id" example:"0ac7ee25-2ebf-4edb-91eb-3d160a0428a8"`
	FirstName string `json:"first_name,omitempty" example:"Alexandr"`
	LastName  string `json:"last_name,omitempty" example:"Levin"`
	BirthDate string `json:"birth_date,omitempty" example:"2004-03-17"`
	Country   string `json:"country,omitempty" example:"Russia"`
	HasOscar  bool   `json:"has_oscar,omitempty" example:"true"`
}

//type CustomDate struct {
//	time.Time `json:"-"`
//}
//
//const layout = "2006-01-02"
//
//func (c *CustomDate) UnmarshalJSON(b []byte) (err error) {
//	s := strings.Trim(string(b), `"`) // remove quotes
//	if s == "null" {
//		return
//	}
//	c.Time, err = time.Parse(layout, s)
//	return
//}
//
//func (c *CustomDate) MarshalJSON() ([]byte, error) {
//	if c.Time.IsZero() {
//		return nil, nil
//	}
//	return []byte(fmt.Sprintf(`"%s"`, c.Time.Format(layout))), nil
//}
