package postgres

import (
	"context"
	"testing"

	"github.com/presnalex/go-micro/v3/service"
)

func TestPostgres(t *testing.T) {
	type Entity struct {
		Name string `db:"name"`
		ID   int64  `db:"id"`
	}
	db, err := Connect(&service.PostgresConfig{
		Addr:            "127.0.0.1:5432",
		Login:           "postgres",
		Passw:           "password",
		DBName:          "postgres",
		AppName:         "test",
		ConnMax:         5,
		ConnLifetime:    0,
		ConnMaxIdleTime: 0,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	dst := &Entity{}

	/*
		if _, err = db.ExecContext(context.TODO(), `insert into test (name) values ($1);`, "vtolstov"); err != nil {
			t.Fatal(err)
		}
	*/

	query := `select * from test where name = $1`
	args := []interface{}{`vtolstov`, `vtolstov' and (select 1 from (select pg_sleep(20))x)=1 -`}
	if err = db.GetContext(context.TODO(), dst, query, args[0]); err != nil {
		t.Fatal(err)
	}

	t.Logf("dst %#+v", dst)

	if err = db.GetContext(context.TODO(), dst, query, args[1]); err != nil {
		t.Fatal(err)
	}

	t.Logf("dst %#+v", dst)
}
