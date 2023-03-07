package redis

import (
	"context"
	"strconv"
	"time"

	// MySQL Driver
	// _ "github.com/go-sql-driver/mysql"
	"github.com/gomodule/redigo/redis"
	_ "github.com/gomodule/redigo/redis"
	"golang.org/x/exp/slices"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/config"
	"github.com/p4gefau1t/trojan-go/log"
	"github.com/p4gefau1t/trojan-go/statistic"
	"github.com/p4gefau1t/trojan-go/statistic/memory"
)

const Name = "REDIS"

type Authenticator struct {
	*memory.Authenticator
	pool           *redis.Pool
	reportpool     *redis.Pool
	updateDuration time.Duration
	ctx            context.Context
}

func (a *Authenticator) updater() {
	conn := a.pool.Get()
	connreport := a.reportpool.Get()
	for {
		// update memory
		keys, err := redis.Strings(conn.Do("keys", "userkey:*"))
		if err != nil {
			log.Error(common.NewError("failed to pull data from the database").Base(err))
			time.Sleep(a.updateDuration)
			continue
		}

		for _, user := range a.ListUsers() {
			// swap upload and download for users
			hash := user.Hash()
			sent, recv := user.ResetTraffic()
			if recv > 0 {
				connreport.Do("hincrby", "info:"+hash, "upload", recv)
			}
			if sent > 0 {
				connreport.Do("hincrby", "info:"+hash, "download", sent)
			}
			if !slices.Contains(keys, hash) {
				a.DelUser(hash)
			}
			// s, err := a.db.Exec("UPDATE `users` SET `upload`=`upload`+?, `download`=`download`+? WHERE `password`=?;", recv, sent, hash)
			// if err != nil {
			// 	log.Error(common.NewError("failed to update data to user table").Base(err))
			// 	continue
			// }
			// if r, err := s.RowsAffected(); err != nil {
			// 	if r == 0 {
			// 		a.DelUser(hash)
			// 	}
			// }

		}

		for _, key := range keys {
			// var hash string
			// var quota, download, upload int64
			hash := key[8:]
			a.AddUser(hash)
		}
		// for rows.Next() {
		// 	var hash string
		// 	var quota, download, upload int64
		// 	err := rows.Scan(&hash, &quota, &download, &upload)
		// 	if err != nil {
		// 		log.Error(common.NewError("failed to obtain data from the query result").Base(err))
		// 		break
		// 	}
		// 	if download+upload < quota || quota < 0 {
		// 		a.AddUser(hash)
		// 	} else {
		// 		a.DelUser(hash)
		// 	}
		// }

		select {
		case <-time.After(a.updateDuration):
		case <-a.ctx.Done():
			log.Debug("Redis daemon exiting...")
			return
		}
	}
}

func newPool(server string, auth string) *redis.Pool {
	return &redis.Pool{
		// Maximum number of idle connections in the pool.
		MaxIdle: 2,
		// max number of connections
		MaxActive: 10,
		// Dial is an application supplied function for creating and
		// configuring a connection.
		Dial: func() (redis.Conn, error) {
			if auth == "" {
				c, err := redis.Dial("tcp", server)
				if err != nil {
					panic(err.Error())
				}
				return c, err
			} else {
				c, err := redis.Dial("tcp", server, redis.DialPassword(auth))
				if err != nil {
					panic(err.Error())
				}
				return c, err
			}
		},
	}
}

// func connectDatabase(driverName, username, password, ip string, port int, dbName string) (*sql.DB, error) {
// 	path := strings.Join([]string{username, ":", password, "@tcp(", ip, ":", fmt.Sprintf("%d", port), ")/", dbName, "?charset=utf8"}, "")
// 	return sql.Open(driverName, path)
// }

func NewAuthenticator(ctx context.Context) (statistic.Authenticator, error) {
	log.Debug("redis authenticator start")
	cfg := config.FromContext(ctx, Name).(*Config)
	pool := newPool(":6379", "")
	reportpool := newPool(cfg.Redis.ServerHost+":"+strconv.Itoa(cfg.Redis.ServerPort), "0e44ae02b1018ba9a00a378fa069fc2e1e626bb4dc78d70fd8a035e0fcc541fa")
	// db, err := connectDatabase(
	// 	"mysql",
	// 	cfg.MySQL.Username,
	// 	cfg.MySQL.Password,
	// 	cfg.MySQL.ServerHost,
	// 	cfg.MySQL.ServerPort,
	// 	cfg.MySQL.Database,
	// )
	// if err != nil {
	// 	return nil, common.NewError("Failed to connect to database server").Base(err)
	// }
	memoryAuth, err := memory.NewAuthenticator(ctx)
	if err != nil {
		return nil, err
	}
	a := &Authenticator{
		pool:           pool,
		reportpool:     reportpool,
		ctx:            ctx,
		updateDuration: time.Duration(cfg.Redis.CheckRate) * time.Second,
		Authenticator:  memoryAuth.(*memory.Authenticator),
	}
	go a.updater()
	log.Debug("redis authenticator created")
	return a, nil
}

func init() {
	statistic.RegisterAuthenticatorCreator(Name, NewAuthenticator)
}
