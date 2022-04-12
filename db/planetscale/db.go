package planetscale

import (
	"database/sql"
	"fmt"
	db2 "github.com/navbryce/next-dorm-be/db"
	"github.com/upper/db/v4"
	"github.com/upper/db/v4/adapter/mysql"
	"os"
)

type PlanetScaleDB struct {
	*PostDB
	*SubscriptionDB
	*UserDB
	sess  db.Session
	sqlDB *sql.DB
}

func GetDatabase() (db2.Database, error) {
	// TODO: MOVE CONFIG PARSING AND VALIDATINO TO SEPARATE MODULE
	db, err := sql.Open("mysql",
		fmt.Sprintf("%s:%s@tcp(%s)/next-dorm?tls=true&parseTime=true",
			os.Getenv("DB_USER"), os.Getenv("DB_PASS"), os.Getenv("DB_HOST")))
	if err != nil {
		return nil, err
	}

	// TODO: Move to config
	db.SetMaxIdleConns(50)
	db.SetMaxOpenConns(50)
	db.SetConnMaxIdleTime(0)

	sess, err := mysql.New(db)
	if err != nil {
		return nil, err
	}

	return &PlanetScaleDB{
		PostDB:         getPostDB(sess),
		SubscriptionDB: getSubscriptionDB(sess),
		UserDB:         getUserDB(sess),
		sess:           sess,
		sqlDB:          db,
	}, nil
}

func (psdb *PlanetScaleDB) GetSQLDB() *sql.DB {
	return psdb.sqlDB
}

func (psdb *PlanetScaleDB) Close() error {
	return psdb.sess.Close()
}
