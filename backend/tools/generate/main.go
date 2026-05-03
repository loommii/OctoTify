package main

import (
	"fmt"

	"gorm.io/driver/sqlite"
	"gorm.io/gen"
	"gorm.io/gorm"

	"octotify/internal/model"
)

func main() {
	g := gen.NewGenerator(gen.Config{
		OutPath:      "./internal/query",
		ModelPkgPath: "./internal/model",
		Mode:         gen.WithoutContext | gen.WithDefaultQuery | gen.WithQueryInterface,
	})

	db, err := gorm.Open(sqlite.Open("./data/octotify.db"))
	if err != nil {
		panic(fmt.Errorf("connect database failed: %w", err))
	}

	g.UseDB(db)

	g.ApplyBasic(
		model.User{},
		model.Source{},
		model.Channel{},
		model.Message{},
		model.SourceChannel{},
		model.RefreshToken{},
	)

	g.Execute()
}
