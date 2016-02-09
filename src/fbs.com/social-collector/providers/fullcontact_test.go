package providers

import (
	"encoding/json"
	"fbs.com/social-collector/types"
	. "github.com/smartystreets/goconvey/convey"
	_ "log"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

func testBackend(response_code int, payload string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(response_code)
			w.Write([]byte(payload))
		}))
}

func TestRequest(t *testing.T) {

	Convey("Request", t, func() {

		responses := []int{0, 200, 400, 403, 404, 405, 410, 422, 500, 503}

		for _, code := range responses {

			Convey("Test response code:"+strconv.Itoa(code), func() {

				backend := testBackend(code, "{}")
				defer backend.Close()

				provider := Fullcontact{Url: backend.URL, ApiKey: "1"}

				user := types.User{Email: "test@test.com", Id: 1}

				social, err := provider.Request(user)

				if code == 200 {
					So(social, ShouldResemble, types.Social{UserId: 1})
					So(err, ShouldBeNil)
				} else {
					So(social, ShouldResemble, types.Social{UserId: 0})
					So(err, ShouldNotBeNil)
				}

			})

		}

		Convey("Test err Parse url", func() {

			provider := Fullcontact{Url: "///threeslashes", ApiKey: "1"}

			user := types.User{Email: "test@test.com", Id: 1}

			social, err := provider.Request(user)

			So(social, ShouldResemble, types.Social{UserId: 0})
			So(err, ShouldNotBeNil)

		})

		Convey("Test err parse person", func() {

			backend := testBackend(200, "{status:200,socialProfiles:[{type:facebook, url:http://test.com}]}")
			defer backend.Close()

			provider := Fullcontact{Url: backend.URL, ApiKey: "1"}

			user := types.User{Email: "test@test.com", Id: 1}

			social, err := provider.Request(user)

			So(social, ShouldResemble, types.Social{UserId: 0})
			So(err, ShouldNotBeNil)

		})

		Convey("Test invalid json", func() {

			backend := testBackend(200, "not well-formed JSON")
			defer backend.Close()

			provider := Fullcontact{Url: backend.URL, ApiKey: "1"}

			user := types.User{Email: "test@test.com", Id: 1}

			social, err := provider.Request(user)

			So(social, ShouldResemble, types.Social{UserId: 0})
			So(err, ShouldNotBeNil)

		})

		Convey("Test headers", func() {

			backend := httptest.NewServer(http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("X-Rate-Limit-Limit", "60")
					w.Header().Set("X-Rate-Limit-Remaining", "0")
					w.Header().Set("X-Rate-Limit-Reset", "7")
					w.WriteHeader(200)
					w.Write([]byte("{}"))
				}))
			defer backend.Close()

			provider := Fullcontact{Url: backend.URL, ApiKey: "1"}

			user := types.User{Email: "test@test.com", Id: 1}

			social, err := provider.Request(user)

			So(social, ShouldResemble, types.Social{UserId: 1})
			So(err, ShouldBeNil)

		})

		Convey("Test json with social profiles", func() {

			pr := Person{
				Status: 200,
				SocialProfiles: []SocialProfile{
					{
						Type:     "facebook",
						TypeName: "Facebook",
						Url:      "http://facebook.com/test",
						Id:       "123456789",
					},
					{
						Type:     "twitter",
						TypeName: "Twitter",
						Url:      "http://twitter.com/test",
						Id:       "123456789",
					},
				},
			}

			b, _ := json.Marshal(pr)
			backend := testBackend(200, string(b))
			defer backend.Close()

			provider := Fullcontact{Url: backend.URL, ApiKey: "1"}

			user := types.User{Email: "test@test.com", Id: 1}

			social, err := provider.Request(user)

			So(social, ShouldResemble, types.Social{UserId: 1, TwitterUrl: "http://twitter.com/test", FacebookUrl: "http://facebook.com/test"})
			So(social.IsValid(), ShouldBeTrue)
			So(err, ShouldBeNil)
		})

		Convey("Test json with photo", func() {

			pr := Person{
				Status: 200,
				Photos: []Photo{
					{
						Url:      "https://test.jpg",
						TypeId:   "googleplus",
						TypeName: "Google Plus",
					},
					{
						Url:       "https://test2.gif",
						TypeId:    "googleplus",
						TypeName:  "Google Plus",
						IsPrimary: true,
					},
				},
			}

			b, _ := json.Marshal(pr)
			backend := testBackend(200, string(b))
			defer backend.Close()

			provider := Fullcontact{Url: backend.URL, ApiKey: "1"}

			user := types.User{Email: "test@test.com", Id: 1}

			social, err := provider.Request(user)

			So(social, ShouldResemble, types.Social{UserId: 1, PhotoUrl: "https://test2.gif"})
			So(social.IsValid(), ShouldBeTrue)
			So(err, ShouldBeNil)
		})

		Convey("Test empty body", func() {

			backend := testBackend(200, "")
			defer backend.Close()

			provider := Fullcontact{Url: backend.URL, ApiKey: "1"}

			user := types.User{Email: "test@test.com", Id: 1}

			social, err := provider.Request(user)

			So(social, ShouldResemble, types.Social{UserId: 0})
			So(err, ShouldNotBeNil)

		})

		// Create a backend to generate a test URL, then close it to cause a
		// connection error.
		Convey("Expected error when a connection fails", func() {

			backend := testBackend(200, "{}")
			backend.Close()

			provider := Fullcontact{Url: backend.URL, ApiKey: "1"}

			user := types.User{Email: "test@test.com", Id: 1}

			social, err := provider.Request(user)

			So(social, ShouldResemble, types.Social{UserId: 0})
			So(err, ShouldNotBeNil)

		})

	})

}
