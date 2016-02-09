package main

import (
	"database/sql"
	"fbs.com/social-collector/providers"
	"fbs.com/social-collector/types"
	"flag"
	"fmt"
	"github.com/go-gorp/gorp"
	_ "github.com/lib/pq"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"time"
)

var (
	configUrl string
	Cfg       types.Config
	DbMap     *gorp.DbMap
)

func main() {
	flag.Parse()

	err := initCfg()
	if err != nil {
		panic(err)
	}
	err = initDb()
	if err != nil {
		panic(err)
	}

	for {
		start()
	}

}
func start() {

	var messages = make(chan types.User, 1)

	go workerLoop(&messages)

	listenLoop(&messages)

}

func init() {
	flag.StringVar(&configUrl, "config", "cfg/config.yml", "a string")
}

func initCfg() (err error) {

	data, err := ioutil.ReadFile(configUrl)
	if err != nil {
		return
	}

	err = yaml.Unmarshal(data, &Cfg)
	return
}

func initDb() (err error) {

	db, err := sql.Open(Cfg.Database.Driver, generateDataSourceName())
	if err != nil || db.Ping() != nil {
		return
	}
	DbMap = &gorp.DbMap{Db: db, Dialect: gorp.PostgresDialect{}}
	DbMap.TraceOn("[gorp]", log.New(os.Stdout, "myapp:", log.Lmicroseconds))
	DbMap.AddTableWithName(types.Social{}, `social"."users`)
	return
}
func listenLoop(messages *chan types.User) {

	defer func() {
		if r := recover(); r != nil {
			listenLoop(messages)
		}
	}()

	var provider = providers.Fullcontact{Url: Cfg.Fullcontact.Url, ApiKey: Cfg.Fullcontact.ApiKey}

	for {
		user := <-*messages
		err := search(user, &provider)
		if err != nil {
			log.Printf("%s", err)
		}
	}
}

func search(user types.User, provider *providers.Fullcontact) (err error) {

	social, err := provider.Request(user)

	if err == nil && social.IsValid() {
		err = DbMap.Insert(&social)
	}
	return
}

func workerLoop(messages *chan types.User) {

	defer func() {
		if r := recover(); r != nil {
			workerLoop(messages)
		}
	}()

	maxId := 0

	for {
		worker(*messages, &maxId)
	}

}
func worker(messages chan<- types.User, maxId *int) {
	var users []types.User

	_, err := DbMap.Select(&users, "select u.id, u.email from personal_area.user as u left join social.users as su on su.user_id = u.id where u.email is not null  and su.user_id is null and u.id > :maxId order by u.id limit 100", map[string]interface{}{
		"maxId": *maxId,
	})

	if err != nil {
		log.Printf("Select from db:%s", err)
		return
	}

	if len(users) > 0 {

		*maxId = users[len(users)-1].Id

		for _, user := range users {
			messages <- user
		}

	} else {
		*maxId = 0
	}

	time.Sleep(time.Second * 10)

}

func generateDataSourceName() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", Cfg.Database.Host, Cfg.Database.Port, Cfg.Database.Username, Cfg.Database.Password, Cfg.Database.Database)
}
