package main

import (
	"log"

	"cms/db"
)

func main() {
	if err := db.Init("cms.db"); err != nil {
		log.Fatal("Failed to connect database:", err)
	}
	defer db.Close()

	if err := db.Migrate(); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// ユーザー作成（必要に応じて編集）
	_, err := db.DB.Exec(`
		INSERT OR IGNORE INTO users (email, password_hash, name)
		VALUES (?, ?, ?)
	`, "kazu@example.com", "", "Kazu")

	if err != nil {
		log.Fatal("Failed to seed user:", err)
	}

	log.Println("Seed completed")
}
