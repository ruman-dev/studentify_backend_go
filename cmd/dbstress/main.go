package main

import (
	"context"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"softixa-solutions.com/studentify/internal/database"
)

func main() {
	db, err := database.New(os.Getenv("DATABASE_URL"))
	if err != nil {
		panic(err)
	}
	defer db.Close()

	var ok, fail atomic.Int64
	var wg sync.WaitGroup
	queries := []string{
		`SELECT COUNT(*) FROM assignments`,
		`SELECT COUNT(*) FROM events`,
		`SELECT COUNT(*) FROM subjects`,
		`SELECT COUNT(*) FROM teachers`,
		`SELECT COUNT(*) FROM users`,
		`SELECT COUNT(*) FROM attendance_records`,
		`SELECT id, name, credit_hours FROM subjects LIMIT 5`,
	}

	for i := 0; i < 40; i++ {
		for _, q := range queries {
			wg.Add(1)
			go func(sql string) {
				defer wg.Done()
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer cancel()
				rows, err := db.QueryContext(ctx, sql)
				if err != nil {
					fail.Add(1)
					fmt.Println("FAIL:", err)
					return
				}
				for rows.Next() {
				}
				if err := rows.Err(); err != nil {
					fail.Add(1)
					fmt.Println("FAIL rows:", err)
					_ = rows.Close()
					return
				}
				_ = rows.Close()
				ok.Add(1)
			}(q)
		}
	}
	wg.Wait()
	fmt.Printf("ok=%d fail=%d\n", ok.Load(), fail.Load())
	if fail.Load() > 0 {
		os.Exit(1)
	}
}
