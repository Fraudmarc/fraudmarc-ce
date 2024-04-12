package lib

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/mgutz/dat/dat"
	runner "github.com/mgutz/dat/sqlx-runner"
)

type DBConnectInfo struct {
	dbname  string
	dbuser  string
	dbpass  string
	dbhost  string
	dbssl   string
	maxLife string // Seconds
}
type DBHandle struct {
	DB                 *sql.DB
	runner             *runner.DB
	defaultIdle        int
	defaultConnections int
}

var (
	DB       *sql.DB
	ARRTable = os.Getenv("ARRTable")
	ARTable  = os.Getenv("ARTable")

	sharedDBs    = make(map[string]DBHandle)
	sharedDBLock = sync.Mutex{}

	// variables that are settable from outside routines:
	// SetMaxIdleConns maximum number of idle connections to DB
	SetMaxIdleConns = 4

	// SetMaxOpenConns maximum number of open connections to DB
	SetMaxOpenConns = 16

	AppName string

	reportDB *runner.DB

	DBreporting *runner.DB

	SessionReady = false
	Sess         *session.Session
	SvcSNS       *sns.SNS
	SvcLambda    *lambda.Lambda

	SnsRegion    = os.Getenv("AWS_REGION")
	LambdaRegion = os.Getenv("AWS_REGION")
	DreTable     = "dmarc_reporting_entries"
	RecordChunk  = 10000
	AwsConfig    = aws.Config{
		Credentials: credentials.NewChainCredentials(
			[]credentials.Provider{
				&credentials.EnvProvider{},
				&ec2rolecreds.EC2RoleProvider{
					ExpiryWindow: 5 * time.Minute,
				},
			},
		),
	}
)

const MAX_ROWS = 2500

func init() {
	// create a normal database connection through database/sql
	var err error
	connStr := fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s",
		os.Getenv("REPORTING_DB_HOST"),
		"5432",
		os.Getenv("REPORTING_DB_USER"),
		os.Getenv("REPORTING_DB_NAME"),
		os.Getenv("REPORTING_DB_PASSWORD"))
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}

	// set to reasonable values for production
	DB.SetMaxIdleConns(6)
	DB.SetMaxOpenConns(16)

	if ARRTable == "" {
		ARRTable = "AggregateReportRecord"
	}
	if ARTable == "" {
		ARTable = "AggregateReport"
	}

	DBreporting = GetReportingDB()
}

func GetReportingDB() *runner.DB {
	if reportDB == nil {
		prefix := "REPORTING"
		reportDB = InitDBRunner(prefix)
	}
	return reportDB
}

func InitDBRunner(prefix string) (db *runner.DB) {
	// set this to enable interpolation
	dat.EnableInterpolation = false

	// set to check things like sessions closing.
	// Should be disabled in production/release builds.
	dat.Strict = false

	// Log any query over 10ms as warnings. (optional)
	//runner.LogQueriesThreshold = 1000 * time.Millisecond

	db = GetTheRunner(prefix)

	return
}

func GetTheRunner(prefix string) (theRunner *runner.DB) {
	sharedDBLock.Lock()
	defer sharedDBLock.Unlock()

	shared, exists := sharedDBs[prefix]
	if exists && shared.runner != nil && shared.runner.DB.DB.Ping() == nil {
		theRunner = shared.runner
	} else {
		if shared.DB == nil || shared.DB.Ping() != nil {
			newSharedDB(prefix, &shared)
		}
		shared.runner = runner.NewDB(shared.DB, "postgres")
		theRunner = shared.runner
		sharedDBs[prefix] = shared
	}
	return
}

func newSharedDB(prefix string, shared *DBHandle) *sql.DB {
	sqlDB := InitSQLDB(prefix)
	shared.DB = sqlDB

	idle := os.Getenv(prefix + "_DB_MAX_IDLE_CONN")
	max := os.Getenv(prefix + "_DB_MAX_CONN")

	if idle != "" {
		fmt.Sscanf(idle, "%d", &shared.defaultIdle)
	}

	if max != "" {
		fmt.Sscanf(max, "%d", &shared.defaultConnections)
	}

	shared.DB.SetMaxIdleConns(shared.defaultIdle)

	if shared.defaultConnections != 0 {
		shared.DB.SetMaxOpenConns(shared.defaultConnections)
	}

	return sqlDB
}

// InitSQLDB is a non-runner DB init for migrating code away from runner.DB
func InitSQLDB(prefix string) (sqlDB *sql.DB) {

	info := getDBVars(prefix)

	// create a normal database connection through database/sql
	var errOpen error
	AppName = os.Getenv("LAMBDA_FUNCTION_NAME")
	ostr := fmt.Sprintf("dbname=%s user=%s password=%s host=%s sslmode=%s application_name=%s",
		info.dbname, info.dbuser, info.dbpass, info.dbhost, info.dbssl, AppName)
	//Name, User, Password, Host, SSL)
	sqlDB, errOpen = sql.Open("postgres", ostr)

	if errOpen != nil {
		log.Println("errOpen   :  ", errOpen)
		panic(errOpen)
	}

	// ensures the database can be pinged with an exponential backoff (15 min)
	errPing := sqlDB.Ping()
	if errPing != nil {
		log.Println("errPing   :  ", errPing)
		return
	}

	// set to reasonable values for production
	sqlDB.SetMaxIdleConns(SetMaxIdleConns)
	sqlDB.SetMaxOpenConns(SetMaxOpenConns)
	maxTime, err := time.ParseDuration(info.maxLife)
	if err == nil {
		log.Println("Shared DB for ", prefix, " max conn life set to ", info.maxLife)
		sqlDB.SetConnMaxLifetime(maxTime)
	} else {
		log.Println("Error parsing max connection life time")
	}

	return

} // InitSQLDB

func getDBVars(prefix string) (info DBConnectInfo) {
	info.dbname = os.Getenv(prefix + "_DB_NAME")
	info.dbuser = os.Getenv(prefix + "_DB_USER")
	info.dbpass = os.Getenv(prefix + "_DB_PASSWORD")
	info.dbhost = os.Getenv(prefix + "_DB_HOST")
	info.dbssl = os.Getenv(prefix + "_DB_SSL")
	info.maxLife = os.Getenv(prefix + "_DB_MAX_TIME")
	return
}
