package main

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestRecomputePriceCache(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectExec("INSERT INTO gym_price_cache").
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := recomputePriceCache(context.Background(), db); err != nil {
		t.Fatalf("recomputePriceCache error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
