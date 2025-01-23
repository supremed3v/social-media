package main

import (
	"expvar"
	"log"
	"runtime"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/supremed3v/social-media/internal/auth"
	"github.com/supremed3v/social-media/internal/cloudinary"
	"github.com/supremed3v/social-media/internal/db"
	"github.com/supremed3v/social-media/internal/env"
	"github.com/supremed3v/social-media/internal/mailer"
	"github.com/supremed3v/social-media/internal/ratelimiter"
	"github.com/supremed3v/social-media/internal/store"
	"github.com/supremed3v/social-media/internal/store/cache"
	"go.uber.org/zap"
)

const version = "1.1.0"

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
		addr:            env.GetString("ADDR", ":8080"),
		apiURL:          env.GetString("EXTERNAL_URL", "localhost:8080"),
		frontendURL:     env.GetString("FRONTEND_URL", "http:localhost:4000"),
		env:             "development",
		maxMultipartMem: 10 << 20, // 10 MB
		db: dbConfig{
			addr:         env.GetString("DB_ADDR", "postgres://admin:adminpassword@localhost/social?sslmode=disable"),
			maxOpenConns: env.GetInt("DB_MAX_OPEN_CONNS", 30),
			maxIdleConns: env.GetInt("DB_MAX_IDLE_CONNS", 30),
			maxIdleTime:  env.GetString("DB_MAX_IDLE_TIME", "15m"),
		},
		redisCfg: redisConfig{
			addr:    env.GetString("REDIS_ADDR", "localhost:6379"),
			pw:      env.GetString("REDIS_PW", ""),
			db:      env.GetInt("REDIS_DB", 0),
			enabled: env.GetBool("REDIS_ENABLED", true),
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
			token: tokenConfig{
				secret: env.GetString("AUTH", "example"),
				exp:    time.Hour * 24 * 3, // 3 days
				issuer: "socialmedia",
			},
		},
		rateLimiter: ratelimiter.Config{
			RequestsPerTimeFrame: env.GetInt("RATELIMITER_REQUEST_COUNT", 20),
			TimeFrame:            time.Second * 5,
			Enabled:              env.GetBool("RATE_LIMITER_ENABLED", true),
		},
		cloudinary: cloudinary.CloudinaryConfig{
			CloudName: env.GetString("CLOUDINARY_CLOUD_NAME", ""),
			APIKey:    env.GetString("CLOUDINARY_API_KEY", ""),
			APISecret: env.GetString("CLOUDINARY_API_SECRET", ""),
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

	// Cache
	var rdb *redis.Client
	if cfg.redisCfg.enabled {
		rdb = cache.NewRedisClient(cfg.redisCfg.addr, cfg.redisCfg.pw, cfg.redisCfg.db)
		logger.Info("redis cached connection established")

		defer rdb.Close()
	}

	// Rate Limiter

	rateLimiter := ratelimiter.NewFixedWindowLimiter(
		cfg.rateLimiter.RequestsPerTimeFrame,
		cfg.rateLimiter.TimeFrame,
	)

	// Cloudinary

	cloudinaryService, _ := cloudinary.NewCloudinary(cfg.cloudinary)

	logger.Info("db connected")

	store := store.NewStorage(db)
	cacheStorage := cache.NewRedisStorage(rdb)

	mailer := mailer.NewSendgrid(cfg.mail.sendGrid.apiKey, cfg.mail.fromEmail)

	jwtAuthenticator := auth.NewJWTAuthenticator(cfg.auth.token.secret, cfg.auth.token.issuer, cfg.auth.token.issuer)

	app := &application{
		config:        cfg,
		store:         store,
		cacheStorage:  cacheStorage,
		logger:        logger,
		mailer:        mailer,
		authenticator: jwtAuthenticator,
		rateLimiter:   rateLimiter,
		cloudinary:    cloudinaryService,
	}

	// Metrics collected

	expvar.NewString("version").Set(version)
	expvar.Publish("database", expvar.Func(func() any {
		return db.Stats()
	}))
	expvar.Publish("goroutines", expvar.Func(func() any {
		return runtime.NumGoroutine()
	}))

	mux := app.mount()

	logger.Fatal(app.run(mux))
}
