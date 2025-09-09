package api

import (
	"fmt"
	"pilo/internal/config"
)

// User represents a user in the system.
type User struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Name     string `json:"name"`
}

// GetUsers reads the users from the base-config.json file.
func GetUsers() ([]config.User, error) {
	cfg, err := config.ReadConfig()
	if err != nil {
		return nil, fmt.Errorf("error loading base config: %w", err)
	}
	return cfg.Users, nil
}

// AddUser adds a new user to the base-config.json file.
func AddUser(username, name, email string) error {
	cfg, err := config.ReadConfig()
	if err != nil {
		return err
	}

	cfg.Users = append(cfg.Users, config.User{Username: username, Name: name, Email: email})

	return config.WriteConfig(cfg)
}

// RemoveUser removes a user from the base-config.json file.
func RemoveUser(username string) error {
	cfg, err := config.ReadConfig()
	if err != nil {
		return err
	}

	var updatedUsers []config.User
	for _, user := range cfg.Users {
		if user.Username != username {
			updatedUsers = append(updatedUsers, user)
		}
	}
	cfg.Users = updatedUsers

	return config.WriteConfig(cfg)
}

// UpdateUser updates an existing user.
func UpdateUser(oldUsername, newUsername, name, email string) error {
	cfg, err := config.ReadConfig()
	if err != nil {
		return err
	}

	for i, user := range cfg.Users {
		if user.Username == oldUsername {
			cfg.Users[i].Username = newUsername
			cfg.Users[i].Name = name
			cfg.Users[i].Email = email
			break
		}
	}

	return config.WriteConfig(cfg)
}
