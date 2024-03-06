package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"firebase.google.com/go/v4/db"
	_ "github.com/lib/pq"
	"google.golang.org/api/option"
)

type FirebaseCredentials struct {
	Type                    string `json:"type"`
	ProjectID               string `json:"project_id"`
	PrivateKeyID            string `json:"private_key_id"`
	PrivateKey              string `json:"private_key"`
	ClientEmail             string `json:"client_email"`
	ClientID                string `json:"client_id"`
	AuthURI                 string `json:"auth_uri"`
	TokenURI                string `json:"token_uri"`
	AuthProviderX509CertURL string `json:"auth_provider_x509_cert_url"`
	ClientX509CertURL       string `json:"client_x509_cert_url"`
}

var AuthClient *auth.Client // <-- add this
var (
	postgresPassword string
	PgDb             *sql.DB
	FbDb             *db.Client
)

var DbClient *db.Client

func Init() {
	postgresPassword = os.Getenv("POSTGRES_PASSWORD")
	if postgresPassword == "" {
		log.Fatalf("Failed to get POSTGRES_PASSWORD environment variable")
	}

	psqlInfo := fmt.Sprintf("postgresql://postgres:%s@roundhouse.proxy.rlwy.net:20318/railway", postgresPassword)
	var err error
	PgDb, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}

	var jsonVal []byte
	err = PgDb.QueryRow("SELECT credentials FROM firebase").Scan(&jsonVal)
	if err != nil {
		log.Fatalf("Failed to retrieve Firebase credentials: %v", err)
	}

	var credentials FirebaseCredentials
	err = json.Unmarshal(jsonVal, &credentials)
	if err != nil {
		log.Fatalf("Failed to unmarshal Firebase credentials: %v", err)
	}

	optBytes, err := json.Marshal(credentials)
	if err != nil {
		log.Fatalf("Failed to marshal Firebase credentials: %v", err)
	}

	opt := option.WithCredentialsJSON(optBytes)

	fbApp, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		log.Fatalf("Failed to initialize Firebase app: %v", err)
	}

	DbClient, err = fbApp.DatabaseWithURL(context.Background(), "https://lcr-webapp-default-rtdb.firebaseio.com/")
	if err != nil {
		log.Fatalf("Failed to initialize Firebase RTDB client: %v", err)
	}

	AuthClient, err = fbApp.Auth(context.Background())
	if err != nil {
		log.Fatalf("Failed to initialize Firebase Auth client: %v", err)
	}
}

// CloseDbConnections - This function should be deferred in your main function to properly close database connections when your application stops
func ClosePgConnection() {
	PgDb.Close()
}
