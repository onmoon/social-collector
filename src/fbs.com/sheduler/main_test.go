package main

import (
	"database/sql/driver"
	_ "errors"
	"fbs.com/sheduler/types"
	"github.com/erikstmartin/go-testdb"
	. "github.com/smartystreets/goconvey/convey"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

type testResult struct {
	lastId       int64
	affectedRows int64
}

func (r testResult) LastInsertId() (int64, error) {
	return r.lastId, nil
}

func (r testResult) RowsAffected() (int64, error) {
	return r.affectedRows, nil
}

func testBackend(response_code int, payload string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(response_code)
			w.Write([]byte(payload))
		}))
}

func TestWorker(t *testing.T) {

	Convey("Main", t, func() {

		Convey("Check db", func() {

			Cfg.Database.Driver = `testdb`
			err := initDb()

			So(err, ShouldBeNil)

			responses := []int{0, 200, 400, 403, 404, 405, 410, 422, 500, 503}
			for _, code := range responses {

				Convey("Check search func with code:"+strconv.Itoa(code), func() {

					backend := testBackend(code, `{"status":200, "socialProfiles":[{"type":"facebook", "url":"http://test.com"}]}`)
					defer backend.Close()

					Cfg.Fullcontact.Url = backend.URL

					testdb.SetExecWithArgsFunc(func(query string, args []driver.Value) (result driver.Result, err error) {
						return testResult{1, 1}, nil
					})

					user := types.User{Email: "test@test.com", Id: 1}

					err := search(user)

					if code == 200 {
						So(err, ShouldBeNil)
					} else {
						So(err, ShouldNotBeNil)
					}

				})

			}

			Convey("Check worker func", func() {

				maxId := 0
				messages := make(chan types.User, 1)

				Convey("Create empty channel", func() {

					test := types.User{Id: 1, Email: "test@test.com"}

					messages <- test

					user := <-messages

					So(user, ShouldResemble, types.User{Id: 1, Email: "test@test.com"})

					Convey("Run worker with user", func() {

						testdb.SetQueryWithArgsFunc(func(query string, args []driver.Value) (result driver.Rows, err error) {

							columns := []string{"id", "email"}
							rows := "2,test@test.ru"

							return testdb.RowsFromCSVString(columns, rows), nil
						})

						go worker(messages, maxId)

						user := <-messages

						So(user, ShouldResemble, types.User{Id: 2, Email: "test@test.ru"})

						testdb.Reset()

					})
					Convey("Run worker without user", func() {

						testdb.SetQueryWithArgsFunc(func(query string, args []driver.Value) (result driver.Rows, err error) {

							columns := []string{"id", "email"}
							rows := ""

							return testdb.RowsFromCSVString(columns, rows), nil
						})

						go worker(messages, maxId)

						So(len(messages), ShouldEqual, 0)
						So(maxId, ShouldEqual, 0)

						testdb.Reset()

					})

				})

			})

			DbMap.Db.Close()

		})

		Convey("Check config", func() {

			Convey("Empty Config Url", func() {
				ConfigUrl = ""

				err := initCfg()

				So(err, ShouldNotBeNil)
			})

			Convey("Check database config", func() {
				Convey("Driver:", func() {
					So(Cfg.Database.Driver, ShouldHaveSameTypeAs, "")
					So(Cfg.Database.Driver, ShouldNotBeNil)
				})
				Convey("Database:", func() {
					So(Cfg.Database.Database, ShouldHaveSameTypeAs, "")
					So(Cfg.Database.Database, ShouldNotBeNil)
				})
				Convey("Username:", func() {
					So(Cfg.Database.Username, ShouldHaveSameTypeAs, "")
					So(Cfg.Database.Username, ShouldNotBeNil)
				})
				Convey("Password:", func() {
					So(Cfg.Database.Password, ShouldHaveSameTypeAs, "")
					So(Cfg.Database.Password, ShouldNotBeNil)
				})
				Convey("Port:", func() {
					So(Cfg.Database.Port, ShouldHaveSameTypeAs, 1)
					So(Cfg.Database.Port, ShouldNotBeNil)
				})
				Convey("Host:", func() {
					So(Cfg.Database.Host, ShouldHaveSameTypeAs, "")
					So(Cfg.Database.Host, ShouldNotBeNil)
				})
			})

			Convey("Check Fullcontact config", func() {
				Convey("Url:", func() {
					So(Cfg.Fullcontact.Url, ShouldHaveSameTypeAs, "")
					So(Cfg.Fullcontact.Url, ShouldNotBeNil)
				})
				Convey("ApiKey:", func() {
					So(Cfg.Fullcontact.ApiKey, ShouldHaveSameTypeAs, "")
					So(Cfg.Fullcontact.ApiKey, ShouldNotBeNil)
				})
			})

			Convey("Test generateDataSourceName()", func() {
				Cfg.Database.Driver = "test"
				Cfg.Database.Database = "test"
				Cfg.Database.Username = "test"
				Cfg.Database.Password = "test"
				Cfg.Database.Port = 1234
				Cfg.Database.Host = "localhost"
				So(generateDataSourceName(), ShouldEqual, "host=localhost port=1234 user=test password=test dbname=test sslmode=disable")
			})

		})

	})
}
