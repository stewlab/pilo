package api

import (
	"pilo/internal/config"
)

// User represents a user in the system.
type User struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Name     string `json:"name"`
}

// GetUsers reads the users from the users.json file.
func GetUsers() ([]config.User, error) {
	return config.ReadUsersConfig()
}

// AddUser adds a new user to the users.json file.
func AddUser(username, name, email string) error {
	users, err := config.ReadUsersConfig()
	if err != nil {
		return err
	}

	users = append(users, config.User{Username: username, Name: name, Email: email})

	return config.WriteUsersConfig(users)
}

// RemoveUser removes a user from the users.json file.
func RemoveUser(username string) error {
	users, err := config.ReadUsersConfig()
	if err != nil {
		return err
	}

	var updatedUsers []config.User
	for _, user := range users {
		if user.Username != username {
			updatedUsers = append(updatedUsers, user)
		}
	}

	return config.WriteUsersConfig(updatedUsers)
}

// UpdateUser updates an existing user.
func UpdateUser(oldUsername, newUsername, name, email string) error {
	users, err := config.ReadUsersConfig()
	if err != nil {
		return err
	}

	for i, user := range users {
		if user.Username == oldUsername {
			users[i].Username = newUsername
			users[i].Name = name
			users[i].Email = email
			break
		}
	}

	return config.WriteUsersConfig(users)
}
