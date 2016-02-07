package providers

import (
	"encoding/json"
	"errors"
	"fbs.com/social-collector/types"
	"io"
	_ "log"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type Fullcontact struct {
	Url    string
	ApiKey string
}

func (f Fullcontact) Request(user types.User) (social types.Social, err error) {

	var apiUrl *url.URL

	var person *Person

	apiUrl, err = url.Parse(f.Url)

	if err != nil {
		return
	}

	parameters := url.Values{}
	parameters.Add("email", user.Email)
	parameters.Add("apiKey", f.ApiKey)

	apiUrl.RawQuery = parameters.Encode()

	client := &http.Client{}

	req, err := http.NewRequest("GET", apiUrl.String(), nil)
	if err != nil {
		return
	}

	res, err := client.Do(req)
	if err != nil {
		return
	}

	defer res.Body.Close()

	responseLimit := res.Header.Get("X-Rate-Limit-Limit")
	responseRemaining := res.Header.Get("X-Rate-Limit-Remaining")
	responseReset := res.Header.Get("X-Rate-Limit-Reset")

	limit, err := strconv.ParseInt(responseLimit, 10, 64)
	if err != nil {
		limit = 60
	}
	remaining, err := strconv.ParseInt(responseRemaining, 10, 64)
	if err != nil {
		remaining = 60
	}
	reset, err := strconv.ParseInt(responseReset, 10, 64)
	if err != nil {
		reset = 0
	}

	time.Sleep(time.Second * time.Duration(60/limit))

	if remaining == 0 {
		time.Sleep(time.Second * time.Duration(reset+1))
	}

	if res.StatusCode != 200 {
		err = errors.New("Request:response status:" + strconv.Itoa(res.StatusCode))
		return
	}

	decoder := json.NewDecoder(res.Body)
	if err = decoder.Decode(&person); err == io.EOF {
		return
	}

	if person == nil {
		err = errors.New("Request:not parse")
		return
	}

	for _, SocialProfile := range person.SocialProfiles {
		if SocialProfile.Type == "twitter" {
			social.TwitterUrl = SocialProfile.Url
		}
		if SocialProfile.Type == "facebook" {
			social.FacebookUrl = SocialProfile.Url
		}
	}

	for _, Photo := range person.Photos {
		if Photo.IsPrimary {
			social.PhotoUrl = Photo.Url
		}
	}
	social.UserId = user.Id

	return social, nil
}

type Person struct {
	Status           int              `json:"status"`
	RequestId        string           `json:"requestId"`
	Likelihood       float64          `json:"likelihood"`
	ContactInfo      ContactInfo      `json:"contactInfo,omitempty"`
	Demographics     Demographics     `json:"demographics,omitempty"`
	Photos           []Photo          `json:"photos,omitempty"`
	SocialProfiles   []SocialProfile  `json:"socialProfiles,omitempty"`
	DigitalFootprint DigitalFootprint `json:"digitalFootprint,omitempty"`
	Organizations    []Organization   `json:"organizations,omitempty"`
}

type ContactInfo struct {
	FamilyName  string    `json:"familyName,omitempty"`
	GivenName   string    `json:"givenName,omitempty"`
	FullName    string    `json:"fullName,omitempty"`
	MiddleNames []string  `json:"middleNames,omitempty"`
	Websites    []Website `json:"websites,omitempty"`
	Chats       []Chat    `json:"chats,omitempty"`
}

type Demographics struct {
	LocationGeneral string          `json:"locationGeneral,omitempty"`
	LocationDeduced LocationDeduced `json:"locationDeduced,omitempty"`
	Age             string          `json:"age,omitempty"`
	Gender          string          `json:"gender,omitempty"`
	AgeRange        string          `json:"ageRange,omitempty"`
}

type Photo struct {
	Type      string `json:"type"`
	TypeId    string `json:"typeId"`
	TypeName  string `json:"typeName"`
	Url       string `json:"url"`
	IsPrimary bool   `json:"isPrimary"`
}

type SocialProfile struct {
	Type      string `json:"type"`
	TypeId    string `json:"typeId"`
	TypeName  string `json:"typeName"`
	Id        string `json:"id"`
	Username  string `json:"username"`
	Url       string `json:"url"`
	Bio       string `json:"bio,omitempty"`
	Rss       string `json:"rss,omitempty"`
	Following int    `json:"following,omitempty"`
	Followers int    `json:"followers,omitempty"`
}

type DigitalFootprint struct {
	Topics []Topic `json:"topics,omitempty"`
	Scores []Score `json:"scores,omitempty"`
}

type Organization struct {
	Title     string `json:"title"`
	Name      string `json:"name"`
	StartDate string `json:"startDate"`
	EndDate   string `json:"endDate"`
	IsPrimary bool   `json:"isPrimary"`
	Current   bool   `json:"current"`
}

type Website struct {
	Url string `json:"url"`
}

type Chat struct {
	Handle string `json:"handle"`
	Client string `json:"client"`
}

type LocationDeduced struct {
	NormalizedLocation string    `json:"normalizedLocation"`
	DeducedLocation    string    `json:"deducedLocation"`
	City               City      `json:"city"`
	State              State     `json:"state"`
	Country            Country   `json:"country"`
	Continent          Continent `json:"continent"`
	County             County    `json:"county"`
	Likelihood         float64   `json:"likelihood"`
}

type City struct {
	Location
}

type State struct {
	Location
	Code string `json:"code"`
}

type Country struct {
	Location
	Code string `json:"code"`
}

type Continent struct {
	Location
}

type County struct {
	Location
	Code string `json:"code"`
}

type Location struct {
	Deduced bool   `json:"deduced"`
	Name    string `json:"name"`
}

type Topic struct {
	Value    string `json:"value"`
	Provider string `json:"provider"`
}

type Score struct {
	Type     string `json:"type"`
	Value    int    `json:"value"`
	Provider string `json:"provider"`
}
