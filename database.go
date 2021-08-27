package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dbfixture"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/driver/sqliteshim"
	"github.com/uptrace/bun/extra/bundebug"
	"gopkg.in/yaml.v2"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
)

var (
	ctx, db            = loadDatabase("config/", databaseName)
	databaseName       = "database.yaml"
	sampleDatabaseName = "sample-database.yaml"
)

// loadDatabase will load a new database from config/database.yaml or ./sample-database.yaml
func loadDatabase(dir string, file string) (context.Context, *bun.DB) {
	newCtx := context.Background()

	sqlite, err := sql.Open(sqliteshim.ShimName, "file::memory:?cache=shared")
	if err != nil {
		panic(err)
	}
	sqlite.SetMaxOpenConns(1)

	newDB := bun.NewDB(sqlite, sqlitedialect.New())
	newDB.AddQueryHook(bundebug.NewQueryHook(bundebug.WithVerbose()))

	// Register models for the fixture.
	newDB.RegisterModel((*User)(nil))

	// Create tables and load initial data.
	fixture := dbfixture.New(newDB, dbfixture.WithRecreateTables())
	if err := fixture.Load(newCtx, os.DirFS(dir), file); err != nil {
		if file != sampleDatabaseName { // prevent recursive loop
			log.Printf("- error loading '%s%s', defaulting to './%s', error: %v", dir, file, sampleDatabaseName, err)
			return loadDatabase("./", sampleDatabaseName)
		} else {
			panic(err)
		}
	}

	return newCtx, newDB
}

// GetUser will get a User with the provided state
func GetUser(state string) *User {
	// Select one user by their state key.
	user := new(User)
	errored := false
	fmt.Printf("user: %v", user)
	if err := db.NewSelect().Model(user).Where("state = ?", state).Scan(ctx); err != nil {
		errored = true
		log.Printf("- Failed to find user with 'state' '%s'", state)
	}

	if !errored {
		return user
	}
	return nil
}

// InsertUser will insert a new User, or overwrite an existing user with a matching id
func InsertUser(user User) error {
	_, err := db.NewInsert().Model(&user).On("CONFLICT (id) DO UPDATE").
		Set("id = EXCLUDED.id").
		Set("username = EXCLUDED.username").
		Set("name = EXCLUDED.name").
		Set("state = EXCLUDED.state").
		Exec(ctx) // We do not update the whitelisted status here because we want it to stay the same
	return err
}

// UpdateUserWhitelist will update the whitelisted status of a User matching id
func UpdateUserWhitelist(id int, whitelisted bool) error {
	user := new(User)
	user.Whitelisted = whitelisted
	_, err := db.NewUpdate().Model(user).Column("whitelisted").Where("id = ?", id).Exec(ctx)
	return err
}

// TODO: Is there really not a proper way to do this?
func saveDatabase() {
	users := make([]User, 0)
	if err := db.NewSelect().Model(&users).OrderExpr("id ASC").Scan(ctx); err != nil {
		panic(err)
	}

	if *debug {
		log.Printf("Users: %v", users)
	}

	formattedData := []fixtureData{{Model: "User", Rows: users}}
	data, err := yaml.Marshal(formattedData)
	if err != nil {
		panic(err)
	}

	err = ioutil.WriteFile("config/"+databaseName, data, fs.FileMode(0700))

	if err != nil {
		panic(err)
	}
}

func (u User) String() string {
	return fmt.Sprintf("User<%v, %s, %s, %s, %v>", u.ID, u.Username, u.Name, u.State, u.Whitelisted)
}

type User struct {
	ID          int
	Username    string
	Name        string
	State       string
	Whitelisted bool
}

type fixtureData struct {
	Model string `yaml:"model"`
	Rows  []User `yaml:"rows"`
}
