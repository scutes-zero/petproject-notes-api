package models

type Validator interface {
	Validate() (problems map[string]string)
}

type User struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type Note struct {
	ID        int    `json:"id"`
	UserID    int    `json:"user_id"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func (u *User) Validate() map[string]string {
	problems := make(map[string]string)

	if u.Username == "" {
		problems["username"] = "Username cannot be empty"
	}

	if len(u.Password) < 4 {
		problems["password"] = "Password cannot be less than 4 characters"
	}

	return problems
}

func (n *Note) Validate() map[string]string {
	problems := make(map[string]string)

	if n.Title == "" {
		problems["title"] = "Title cannot be empty"
	}

	return problems
}
