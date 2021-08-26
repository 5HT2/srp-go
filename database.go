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
	ctx, db            = loadDatabase(customDatabasePath)
	sampleDatabasePath = "sample/database.yaml"
	customDatabasePath = "config/database.yaml"
)

// loadDatabase will load a new database from config/fixture.yaml
func loadDatabase(path string) (context.Context, *bun.DB) {
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
	if err := fixture.Load(newCtx, os.DirFS("config"), "fixture.yaml"); err != nil {
		if path != sampleDatabasePath { // prevent recursive loop
			return loadDatabase(sampleDatabasePath)
		} else {
			panic(err)
		}
	}

	return newCtx, newDB
}

// GetUser will get a User with the provided state
func GetUser(state string) *User {
	// Select one user by their state key.
	user1 := new(User)
	errored := false
	fmt.Printf("user: %v", user1)
	if err := db.NewSelect().Model(user1).Where("state = ?", state).Scan(ctx); err != nil {
		errored = true
		log.Printf("- Failed to find user with 'state' '%s'", state)
	}

	if !errored {
		return user1
	}
	return nil
}

// InsertUser will insert a new User, or overwrite an existing user with a matching id
func InsertUser(user User) error {
	_, err := db.NewInsert().Model(&user).On("CONFLICT (id) DO UPDATE").
		Set("id = EXCLUDED.id").
		Set("state = EXCLUDED.state").
		Set("whitelisted = EXCLUDED.whitelisted").
		Exec(ctx)
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

	err = ioutil.WriteFile(customDatabasePath, data, fs.FileMode(0700))

	if err != nil {
		panic(err)
	}
}

func (u User) String() string {
	return fmt.Sprintf("User<%v, %s, %v>", u.ID, u.State, u.Whitelisted)
}

type User struct {
	ID          int
	State       string
	Whitelisted bool
}

type fixtureData struct {
	Model string `yaml:"model"`
	Rows  []User `yaml:"rows"`
}
