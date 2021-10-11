package web

import (
	"embed"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/trento-project/trento/internal/consul"
	"github.com/trento-project/trento/web/models"
	"github.com/trento-project/trento/web/services"
	"github.com/trento-project/trento/web/services/ara"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/trento-project/trento/docs" // docs is generated by Swag CLI, you have to import it.
)

//go:embed frontend/assets
var assetsFS embed.FS

//go:embed templates
var templatesFS embed.FS

type App struct {
	host string
	port int
	Dependencies
}

type Dependencies struct {
	consul               consul.Client
	engine               *gin.Engine
	store                cookie.Store
	checksService        services.ChecksService
	subscriptionsService services.SubscriptionsService
	hostsService         services.HostsService
	sapSystemsService    services.SAPSystemsService
	tagsService          services.TagsService
}

func DefaultDependencies() Dependencies {
	consulClient, _ := consul.DefaultClient()
	engine := gin.Default()
	store := cookie.NewStore([]byte("secret"))
	mode := os.Getenv(gin.EnvGinMode)

	gin.SetMode(mode)

	db, err := InitDB()
	if err != nil {
		log.Fatalf("failed to connect database: %s", err)
	}

	if err := MigrateDB(db); err != nil {
		log.Fatalf("failed to migrate database: %s", err)
	}

	tagsService := services.NewTagsService(db)
	araService := ara.NewAraService(viper.GetString("ara-addr"))
	checksService := services.NewChecksService(araService)
	subscriptionsService := services.NewSubscriptionsService(consulClient)
	hostsService := services.NewHostsService(consulClient)
	sapSystemsService := services.NewSAPSystemsService(consulClient)

	return Dependencies{
		consulClient, engine, store,
		checksService, subscriptionsService, hostsService, sapSystemsService, tagsService,
	}
}

func InitDB() (*gorm.DB, error) {
	// TODO: refactor this in a common infrastructure init package
	host := viper.GetString("db-host")
	port := viper.GetString("db-port")
	user := viper.GetString("db-user")
	password := viper.GetString("db-password")
	dbName := viper.GetString("db-name")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbName)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return db, nil
}

func MigrateDB(db *gorm.DB) error {
	err := db.AutoMigrate(models.Tag{})
	if err != nil {
		return err
	}

	return nil
}

// shortcut to use default dependencies
func NewApp(host string, port int) (*App, error) {
	return NewAppWithDeps(host, port, DefaultDependencies())
}

// @title Trento API
// @version 1.0
// @description Trento API

// @contact.name Trento Project
// @contact.url https://www.trento-project.io
// @contact.email  trento-project@suse.com
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /
// @schemes http
func NewAppWithDeps(host string, port int, deps Dependencies) (*App, error) {
	app := &App{
		Dependencies: deps,
		host:         host,
		port:         port,
	}

	InitAlerts()
	engine := deps.engine
	engine.HTMLRender = NewLayoutRender(templatesFS, "templates/*.tmpl")
	engine.Use(ErrorHandler)
	engine.Use(sessions.Sessions("session", deps.store))
	engine.StaticFS("/static", http.FS(assetsFS))
	engine.GET("/", HomeHandler)
	engine.GET("/about", NewAboutHandler(deps.subscriptionsService))
	engine.GET("/hosts", NewHostListHandler(deps.consul, deps.tagsService))
	engine.GET("/hosts/:name", NewHostHandler(deps.consul, deps.subscriptionsService))
	engine.GET("/catalog", NewChecksCatalogHandler(deps.checksService))
	engine.GET("/clusters", NewClusterListHandler(deps.consul, deps.checksService, deps.tagsService))
	engine.GET("/clusters/:id", NewClusterHandler(deps.consul, deps.checksService))
	engine.POST("/clusters/:id/settings", NewSaveClusterSettingsHandler(deps.consul))
	engine.GET("/sapsystems", NewSAPSystemListHandler(deps.consul, deps.hostsService, deps.sapSystemsService, deps.tagsService))
	engine.GET("/sapsystems/:id", NewSAPResourceHandler(deps.hostsService, deps.sapSystemsService))
	engine.GET("/databases", NewHanaDatabaseListHandler(deps.consul, deps.hostsService, deps.sapSystemsService, deps.tagsService))
	engine.GET("/databases/:id", NewSAPResourceHandler(deps.hostsService, deps.sapSystemsService))
	engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	apiGroup := engine.Group("/api")
	{
		apiGroup.GET("/ping", ApiPingHandler)

		apiGroup.GET("/tags", ApiListTag(deps.tagsService))
		apiGroup.POST("/hosts/:name/tags", ApiHostCreateTagHandler(deps.consul, deps.tagsService))
		apiGroup.DELETE("/hosts/:name/tags/:tag", ApiHostDeleteTagHandler(deps.consul, deps.tagsService))
		apiGroup.POST("/clusters/:id/tags", ApiClusterCreateTagHandler(deps.consul, deps.tagsService))
		apiGroup.DELETE("/clusters/:id/tags/:tag", ApiClusterDeleteTagHandler(deps.consul, deps.tagsService))
		apiGroup.GET("/clusters/:cluster_id/results", ApiClusterCheckResultsHandler(deps.consul, deps.checksService))
		apiGroup.POST("/sapsystems/:id/tags", ApiSAPSystemCreateTagHandler(deps.sapSystemsService, deps.tagsService))
		apiGroup.DELETE("/sapsystems/:id/tags/:tag", ApiSAPSystemDeleteTagHandler(deps.sapSystemsService, deps.tagsService))
		apiGroup.POST("/databases/:id/tags", ApiDatabaseCreateTagHandler(deps.sapSystemsService, deps.tagsService))
		apiGroup.DELETE("/databases/:id/tags/:tag", ApiDatabaseDeleteTagHandler(deps.sapSystemsService, deps.tagsService))
	}

	return app, nil
}

func (a *App) Start() error {
	s := &http.Server{
		Addr:           fmt.Sprintf("%s:%d", a.host, a.port),
		Handler:        a,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	return s.ListenAndServe()
}

func (a *App) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	a.engine.ServeHTTP(w, req)
}
