package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq" //import postgres driver
)

// ServerConfig is config struct for server
type ServerConfig struct {
	Addr    string `yaml:"listen_addr" envconfig:"LISTEN_ADDR"  default:":8080"`
	Timeout struct {
		// graceful shutdown
		Server time.Duration `yaml:"server" default:10`
		// write operation
		Write time.Duration `yaml:"write" default:10`
		// read operation
		Read time.Duration `yaml:"read" default:10`
		// time until idle session is closed
		Idle time.Duration `yaml:"idle" default:10`
	} `yaml:"timeout"`
}

// DbConfig is configu struct for database
type DbConfig struct {
	User     string `yaml:"user" envconfig:"DB_USER"`
	Password string `yaml:"password" envconfig:"DB_PASSWORD"`
	DbName   string `yaml:"name" envconfig:"DB_NAME"`
}

// GetDatabase creates database connection using postgres driver
func (c *DbConfig) GetDatabase() (*sql.DB, error) {
	dbinfo := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable",
		c.User, c.Password, c.DbName)
	logrus.WithField("dbname", c.DbName).Warn("Connecting to database.")
	return sql.Open("postgres", dbinfo)
}

// Config is struct for both database and server configuration
type Config struct {
	AppName      string
	ServerConfig ServerConfig `yaml:"server"`
	DBConfig     DbConfig     `yaml:"database"`
}

// NewConfig creates a new config from yaml file
// It first reads from config.yaml file. It overrides values from environment values
func NewConfig(configPath string) (*Config, error) {
	logrus.WithField("configPath", configPath).Info("reading from config path")
	config := &Config{}
	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, err
	}

	logrus.Info("Reading environment variables")
	err = envconfig.Process("", config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

// App is the overall representation of the application.
type App struct {
	conf         *Config
	db           *sql.DB
	Router       *mux.Router
	ShutdownHook func()
}

// NewApp creates a new instance of App
func NewApp(conf *Config) *App {
	return &App{
		conf:   conf,
		Router: mux.NewRouter(),
	}
}

// AddRoute adds a route to applicatoin
func (app *App) AddRoute(method string, route string, apiHandler func(w http.ResponseWriter, r *http.Request)) {
	app.Router.HandleFunc(route, apiHandler).Methods(method)
}

// RenderJson creates a json response of the data
func (app *App) RenderJson(writer http.ResponseWriter, status int, data interface{}) {
	writer.Header().Set("Content-Type", "application/json")
	if data != nil {
		jsonData, err := json.Marshal(data)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		writer.Write(jsonData)
	}
}

// RenderError render error cases
func (app *App) RenderError(writer http.ResponseWriter, ex ErrorResponse) {
	logrus.WithError(ex.Error).Error(ex.Message)
	app.RenderJson(writer, ex.Status, ex)
}

// RenderErrorResponse render error response
func (app *App) RenderErrorResponse(writer http.ResponseWriter, httpStatus int, err error, message string) {
	logrus.WithError(err).Error(message)
	app.RenderError(writer, ErrorResponse{Status: httpStatus, Error: err, Message: message})
}

func (app *App) runServer() error {
	var runChan = make(chan os.Signal, 1)
	ctx, cancel := context.WithTimeout(
		context.Background(),
		app.conf.ServerConfig.Timeout.Server,
	)
	defer cancel()

	server := &http.Server{
		Addr:         app.conf.ServerConfig.Addr,
		Handler:      app.Router,
		ReadTimeout:  app.conf.ServerConfig.Timeout.Read * time.Second,
		WriteTimeout: app.conf.ServerConfig.Timeout.Write * time.Second,
		IdleTimeout:  app.conf.ServerConfig.Timeout.Idle * time.Second,
	}

	// Handle ctrl+c/ctrl+x interrupt
	signal.Notify(runChan, os.Interrupt, syscall.SIGTSTP)

	// Alert the user that the server is starting
	logrus.WithField("address", server.Addr).Info("Staring server...")

	// Run the server on a new goroutine
	go func() {
		if err := server.ListenAndServe(); err != nil {
			if err == http.ErrServerClosed {
				// Normal interrupt operation, ignore
			} else {
				logrus.Fatalf("Server failed to start due to err: %v", err)
			}
		}
	}()

	// Block on this channel listeninf for those previously defined syscalls assign
	// to variable so we can let the user know why the server is shutting down
	interrupt := <-runChan

	// If we get one of the pre-prescribed syscalls, gracefully terminate the server
	// while alerting the user
	log.Printf("Server is shutting down due to %+v\n", interrupt)
	app.ShutdownHook()
	if err := server.Shutdown(ctx); err != nil {
		logrus.Fatalf("Server was unable to gracefully shutdown due to err: %+v", err)
	}
	return nil
}

// Run run the server application
func (app *App) Run() error {
	var err error
	app.db, err = app.conf.DBConfig.GetDatabase()
	if err != nil {
		fmt.Printf("Failed to get connection:%v\n", err)
		return fmt.Errorf("failed to get postgres connection: %v", err)
	}

	app.AddRoutes()

	app.ShutdownHook = func() {
		logrus.Info("Closing database connections....")
		if app != nil && app.db != nil {
			app.db.Close()
		}
	}

	logrus.Info("Running server....")
	return app.runServer()
}
