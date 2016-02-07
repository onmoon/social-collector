package types

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

func (s Social) Valid() bool {
	if s.UserId == 0 {
		return false
	}
	if s.TwitterUrl != "" || s.FacebookUrl != "" || s.PhotoUrl != "" {
		return true
	} else {
		return false
	}
}

type User struct {
	Id    int
	Email string
}
