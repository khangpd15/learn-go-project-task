package entities
type User struct {
	ID       int    `json:"id"`
	Role	 string `json:"role"`
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}
func NewUser(ID int, Role string, Username string, Password string, Email string) User {
	return User{
		ID:       ID,
		Role:     Role,
		Username: Username,
		Password: Password,
		Email:    Email,
	}
}
const (
	RoleAdmin = "ADMIN"
	RoleCustomer = "CUSTOMER"
	RoleGuest = "GUEST"
)