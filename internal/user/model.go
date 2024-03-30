package user

type User struct {
	ID       string `json:"id" example:"a9aec972-2c52-441a-8f17-79506cd34366"`
	Name     string `json:"name,omitempty" example:"example_name"`
	Password string `json:"password,omitempty" example:"example_pass"`
	Email    string `json:"email,omitempty" example:"example@mail.com"`
}
