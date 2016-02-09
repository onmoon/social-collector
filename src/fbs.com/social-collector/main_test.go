package main

import (
	"database/sql/driver"
	_ "errors"
	"fbs.com/social-collector/providers"
	"fbs.com/social-collector/types"
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

			cfg.Database.Driver = `testdb`
			err := initDb()

			So(err, ShouldBeNil)

			responses := []int{0, 200, 400, 403, 404, 405, 410, 422, 500, 503}
			for _, code := range responses {

				Convey("Check search func with code:"+strconv.Itoa(code), func() {

					backend := testBackend(code, `{"status":200, "socialProfiles":[{"type":"facebook", "url":"http://test.com"}]}`)
					defer backend.Close()

					testdb.SetExecWithArgsFunc(func(query string, args []driver.Value) (result driver.Result, err error) {
						return testResult{1, 1}, nil
					})

					user := types.User{Email: "test@test.com", Id: 1}

					provider := providers.Fullcontact{Url: backend.URL, ApiKey: cfg.Fullcontact.ApiKey}

					err := search(user, &provider)

					if code == 200 {
						So(err, ShouldBeNil)
					} else {
						So(err, ShouldNotBeNil)
					}

				})

			}

			Convey("Check worker func", func() {

				messages := make(chan types.User, 1)
				maxId := 0

				Convey("Messages length is 0", func() {
					So(len(messages), ShouldEqual, 0)
				})

				Convey("Max Id is 0", func() {
					So(maxId, ShouldEqual, 0)
				})

				Convey("Run worker with user", func() {

					testdb.SetQueryWithArgsFunc(func(query string, args []driver.Value) (result driver.Rows, err error) {

						columns := []string{"id", "email"}
						rows := "2,test@test.ru"

						return testdb.RowsFromCSVString(columns, rows), nil
					})

					go worker(&messages, &maxId)

					user := <-messages

					So(user, ShouldResemble, types.User{Id: 2, Email: "test@test.ru"})
					So(maxId, ShouldEqual, 2)
					So(len(messages), ShouldEqual, 0)

					testdb.Reset()

				})

				Convey("Run worker without user", func() {

					testdb.SetQueryWithArgsFunc(func(query string, args []driver.Value) (result driver.Rows, err error) {

						columns := []string{"id", "email"}
						rows := ""

						return testdb.RowsFromCSVString(columns, rows), nil
					})

					go worker(&messages, &maxId)

					So(maxId, ShouldEqual, 0)
					So(len(messages), ShouldEqual, 0)

					testdb.Reset()

				})
			})

			Convey("Check start func", func() {

				Convey("Run with user", func() {

					testdb.SetQueryWithArgsFunc(func(query string, args []driver.Value) (result driver.Rows, err error) {

						columns := []string{"id", "email"}
						rows := "2,test@test.ru"

						return testdb.RowsFromCSVString(columns, rows), nil
					})
					testdb.SetExecWithArgsFunc(func(query string, args []driver.Value) (result driver.Result, err error) {

						return testResult{1, 1}, nil
					})
					backend := testBackend(200, `{"status":200, "socialProfiles":[{"type":"facebook", "url":"http://test.com"}]}`)
					defer backend.Close()

					cfg.Fullcontact.Url = backend.URL

					go start()

					testdb.Reset()

				})

			})

			dbMap.Db.Close()

		})

		Convey("Check config", func() {

			Convey("Empty Config Url", func() {
				configUrl = ""

				err := initCfg()

				So(err, ShouldNotBeNil)
			})

			Convey("Check database config", func() {
				Convey("Driver:", func() {
					So(cfg.Database.Driver, ShouldHaveSameTypeAs, "")
					So(cfg.Database.Driver, ShouldNotBeNil)
				})
				Convey("Database:", func() {
					So(cfg.Database.Database, ShouldHaveSameTypeAs, "")
					So(cfg.Database.Database, ShouldNotBeNil)
				})
				Convey("Username:", func() {
					So(cfg.Database.Username, ShouldHaveSameTypeAs, "")
					So(cfg.Database.Username, ShouldNotBeNil)
				})
				Convey("Password:", func() {
					So(cfg.Database.Password, ShouldHaveSameTypeAs, "")
					So(cfg.Database.Password, ShouldNotBeNil)
				})
				Convey("Port:", func() {
					So(cfg.Database.Port, ShouldHaveSameTypeAs, 1)
					So(cfg.Database.Port, ShouldNotBeNil)
				})
				Convey("Host:", func() {
					So(cfg.Database.Host, ShouldHaveSameTypeAs, "")
					So(cfg.Database.Host, ShouldNotBeNil)
				})
			})

			Convey("Check Fullcontact config", func() {
				Convey("Url:", func() {
					So(cfg.Fullcontact.Url, ShouldHaveSameTypeAs, "")
					So(cfg.Fullcontact.Url, ShouldNotBeNil)
				})
				Convey("ApiKey:", func() {
					So(cfg.Fullcontact.ApiKey, ShouldHaveSameTypeAs, "")
					So(cfg.Fullcontact.ApiKey, ShouldNotBeNil)
				})
			})

			Convey("Test generateDataSourceName()", func() {
				cfg.Database.Driver = "test"
				cfg.Database.Database = "test"
				cfg.Database.Username = "test"
				cfg.Database.Password = "test"
				cfg.Database.Port = 1234
				cfg.Database.Host = "localhost"
				So(generateDataSourceName(), ShouldEqual, "host=localhost port=1234 user=test password=test dbname=test sslmode=disable")
			})

		})

	})
}
