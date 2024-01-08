-- Below is an example of connection pooling in golang
```go
type mysqlpool struct {
	db *sql.DB
}

func getLatestPool(server string, port int64, user string, pwd string, dbname string, connLifeTimeSecs time.Duration, maxIdleConns int64, maxOpenConns int64) (*mysqlpool, error) {
	pool := &mysqlpool{}
	
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", user, pwd, server, port, dbname))
	
	if err != nil {
		return nil, err
	}

	if pe := db.Ping(); pe != nil {
		return nil, pe
	}

	pool.db = db
	pool.db.SetConnMaxLifetime(time.Second * connLifeTimeSecs)
	pool.db.SetMaxIdleConns(int(maxIdleConns))
	pool.db.SetMaxOpenConns(int(maxOpenConns))

	return pool, nil
}

```
