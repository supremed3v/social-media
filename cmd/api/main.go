package main

import (
	"log"
	"time"

	"github.com/supremed3v/social-media/internal/db"
	"github.com/supremed3v/social-media/internal/env"
	"github.com/supremed3v/social-media/internal/mailer"
	"github.com/supremed3v/social-media/internal/store"
	"go.uber.org/zap"
)

const version = "0.0.1"

//	@title			Go Social Media
//	@description	API for Social Media, social network for GO devs
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	API Support
//	@contact.url	http://www.swagger.io/support
//	@contact.email	support@swagger.io

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

//	@BasePath					/v1
//
//	@securityDefinitions.apikey	ApiKeyAuth
//	@in							header
//	@name						Authorization
//	@description

func main() {

	cfg := config{
		addr:        env.GetString("ADDR", ":8080"),
		apiURL:      env.GetString("EXTERNAL_URL", "localhost:8080"),
		frontendURL: env.GetString("FRONTEND_URL", "http:localhost:4000"),
		env:         "development",
		db: dbConfig{
			addr:         env.GetString("DB_ADDR", "postgres://admin:adminpassword@localhost/social?sslmode=disable"),
			maxOpenConns: env.GetInt("DB_MAX_OPEN_CONNS", 30),
			maxIdleConns: env.GetInt("DB_MAX_IDLE_CONNS", 30),
			maxIdleTime:  env.GetString("DB_MAX_IDLE_TIME", "15m"),
		},
		mail: mailConfig{
			exp: time.Hour * 24 * 3, //3 days
			sendGrid: sendGridConfig{
				apiKey: env.GetString("SENDGRID_API_KEYS", ""),
			},
			fromEmail: env.GetString("FROM_EMAIL", ""),
		},
		auth: authConfig{
			basic: basicConfig{
				user: "admin",
				pass: "admin",
			},
		},
	}

	// Logger
	logger := zap.Must(zap.NewProduction()).Sugar()
	defer logger.Sync()

	// Database
	db, err := db.New(cfg.db.addr, cfg.db.maxOpenConns, cfg.db.maxIdleConns, cfg.db.maxIdleTime)

	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	logger.Info("db connected")

	store := store.NewStorage(db)

	mailer := mailer.NewSendgrid(cfg.mail.sendGrid.apiKey, cfg.mail.fromEmail)

	app := &application{
		config: cfg,
		store:  store,
		logger: logger,
		mailer: mailer,
	}

	mux := app.mount()

	logger.Fatal(app.run(mux))
}
