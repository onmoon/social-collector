package types

import (
	"errors"
)

type Config struct {
	Fullcontact struct {
		Url    string
		ApiKey string `yaml:"key"`
	}
	Database struct {
		Driver   string
		Database string
		Username string
		Password string
		Port     int
		Host     string
	}
}

type Social struct {
	UserId      int    `db:"user_id"`
	FacebookUrl string `db:"facebook_url"`
	TwitterUrl  string `db:"twitter_url"`
	PhotoUrl    string `db:"photo_url"`
}

func (s Social) IsValid() error {
	if s.UserId == 0 {
		return errors.New("UserId not found.")
	}
	if s.TwitterUrl != "" || s.FacebookUrl != "" || s.PhotoUrl != "" {
		return nil
	} else {
		return errors.New("TwitterUrl or FacebookUrl or PhotoUrl not found.")
	}
}

type User struct {
	Id    int
	Email string
}
