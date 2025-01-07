package main

import (
	"crypto/tls"
	"crypto/x509"
	"embed"
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/cloudlena/adapters/logging"
	"github.com/cloudlena/s3manager/internal/app/s3manager"
	"github.com/gorilla/mux"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/spf13/viper"
)

//go:embed web/template
var templateFS embed.FS

//go:embed web/static
var staticFS embed.FS

type configuration struct {
	Endpoint            string
	UseIam              bool
	IamEndpoint         string
	AccessKeyID         string
	SecretAccessKey     string
	Region              string
	AllowDelete         bool
	ForceDownload       bool
	UseSSL              bool
	SkipSSLVerification bool
	SignatureType       string
	ListRecursive       bool
	Address             string
	Port                string
	Timeout             int32
	SseType             string
	SseKey              string
	SharedBucketsPath   string
	CaCert              string
}

func parseConfiguration() configuration {
	var accessKeyID, secretAccessKey, iamEndpoint string

	viper.SetConfigName("config")
	viper.AddConfigPath(filepath.Join("$HOME", ".s3manager"))
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			log.Fatal(err)
		}
	}

	viper.AutomaticEnv()

	viper.SetDefault("endpoint", "s3.amazonaws.com")
	endpoint := viper.GetString("endpoint")

	useIam := viper.GetBool("use_iam")

	if useIam {
		iamEndpoint = viper.GetString("iam_endpoint")
	} else {
		accessKeyID = viper.GetString("access_key_id")
		if len(accessKeyID) == 0 {
			log.Fatal("please provide ACCESS_KEY_ID")
		}

		secretAccessKey = viper.GetString("secret_access_key")
		if len(secretAccessKey) == 0 {
			log.Fatal("please provide SECRET_ACCESS_KEY")
		}
	}

	region := viper.GetString("region")

	viper.SetDefault("allow_delete", true)
	allowDelete := viper.GetBool("allow_delete")

	viper.SetDefault("force_download", true)
	forceDownload := viper.GetBool("force_download")

	viper.SetDefault("use_ssl", true)
	useSSL := viper.GetBool("use_ssl")

	viper.SetDefault("skip_ssl_verification", false)
	skipSSLVerification := viper.GetBool("skip_ssl_verification")

	viper.SetDefault("signature_type", "V4")
	signatureType := viper.GetString("signature_type")

	listRecursive := viper.GetBool("list_recursive")

	viper.SetDefault("port", "8080")
	port := viper.GetString("port")

	viper.SetDefault("address", "")
	address := viper.GetString("address")

	viper.SetDefault("timeout", 600)
	timeout := viper.GetInt32("timeout")

	viper.SetDefault("sse_type", "")
	sseType := viper.GetString("sse_type")

	viper.SetDefault("sse_key", "")
	sseKey := viper.GetString("sse_key")

	viper.SetDefault("shared_buckets_path", "")
	sharedBucketsPath := viper.GetString("shared_buckets_path")

	viper.SetDefault("ca_cert", "")
	caCert := viper.GetString("ca_cert")

	return configuration{
		Endpoint:            endpoint,
		UseIam:              useIam,
		IamEndpoint:         iamEndpoint,
		AccessKeyID:         accessKeyID,
		SecretAccessKey:     secretAccessKey,
		Region:              region,
		AllowDelete:         allowDelete,
		ForceDownload:       forceDownload,
		UseSSL:              useSSL,
		SkipSSLVerification: skipSSLVerification,
		SignatureType:       signatureType,
		ListRecursive:       listRecursive,
		Port:                port,
		Address:             address,
		Timeout:             timeout,
		SseType:             sseType,
		SseKey:              sseKey,
		SharedBucketsPath:   sharedBucketsPath,
		CaCert:              caCert,
	}
}

func main() {
	configuration := parseConfiguration()

	sseType := s3manager.SSEType{Type: configuration.SseType, Key: configuration.SseKey}
	serverTimeout := time.Duration(configuration.Timeout) * time.Second

	// Set up templates
	templates, err := fs.Sub(templateFS, "web/template")
	if err != nil {
		log.Fatal(err)
	}
	// Set up statics
	statics, err := fs.Sub(staticFS, "web/static")
	if err != nil {
		log.Fatal(err)
	}

	// Set up S3 client
	opts := &minio.Options{
		Secure: configuration.UseSSL,
	}
	if configuration.UseIam {
		opts.Creds = credentials.NewIAM(configuration.IamEndpoint)
	} else {
		var signatureType credentials.SignatureType

		switch configuration.SignatureType {
		case "V2":
			signatureType = credentials.SignatureV2
		case "V4":
			signatureType = credentials.SignatureV4
		case "V4Streaming":
			signatureType = credentials.SignatureV4Streaming
		case "Anonymous":
			signatureType = credentials.SignatureAnonymous
		default:
			log.Fatalf("Invalid SIGNATURE_TYPE: %s", configuration.SignatureType)
		}

		opts.Creds = credentials.NewStatic(configuration.AccessKeyID, configuration.SecretAccessKey, "", signatureType)
	}

	if configuration.Region != "" {
		opts.Region = configuration.Region
	}

	if configuration.UseSSL {
		tlsConf := tls.Config{
			InsecureSkipVerify: configuration.SkipSSLVerification,
		}

		if configuration.CaCert != "" {
			caCertContent, err := ioutil.ReadFile(configuration.CaCert)
			if err != nil {
				log.Fatalln(fmt.Errorf("Failed to read CA certificate file: %s", err))
			}
			roots := x509.NewCertPool()
			ok := roots.AppendCertsFromPEM(caCertContent)
			if !ok {
				log.Fatalln(fmt.Errorf("Failed to parse CA certificate: %w", err))
			}
			tlsConf.RootCAs = roots
		}

		opts.Transport = &http.Transport{
			TLSClientConfig: &tlsConf,
		}
	}

	s3, err := minio.New(configuration.Endpoint, opts)
	if err != nil {
		log.Fatalln(fmt.Errorf("error creating s3 client: %w", err))
	}

	// Set up router
	r := mux.NewRouter()
	r.Handle("/", http.RedirectHandler("/buckets", http.StatusPermanentRedirect)).Methods(http.MethodGet)
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.FS(statics)))).Methods(http.MethodGet)
	r.Handle("/buckets", s3manager.HandleBucketsView(s3, templates, configuration.AllowDelete, configuration.SharedBucketsPath)).Methods(http.MethodGet)
	r.PathPrefix("/buckets/").Handler(s3manager.HandleBucketView(s3, templates, configuration.AllowDelete, configuration.ListRecursive)).Methods(http.MethodGet)
	r.Handle("/api/buckets", s3manager.HandleCreateBucket(s3)).Methods(http.MethodPost)
	if configuration.AllowDelete {
		r.Handle("/api/buckets/{bucketName}", s3manager.HandleDeleteBucket(s3)).Methods(http.MethodDelete)
	}
	r.Handle("/api/buckets/{bucketName}/objects", s3manager.HandleCreateObject(s3, sseType)).Methods(http.MethodPost)
	r.Handle("/api/buckets/{bucketName}/objects/{objectName:.*}/url", s3manager.HandleGenerateUrl(s3)).Methods(http.MethodGet)
	r.Handle("/api/buckets/{bucketName}/objects/{objectName:.*}", s3manager.HandleGetObject(s3, configuration.ForceDownload)).Methods(http.MethodGet)
	if configuration.AllowDelete {
		r.Handle("/api/buckets/{bucketName}/objects/{objectName:.*}", s3manager.HandleDeleteObject(s3)).Methods(http.MethodDelete)
	}

	lr := logging.Handler(os.Stdout)(r)
	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%s", configuration.Address, configuration.Port),
		Handler:      lr,
		ReadTimeout:  serverTimeout,
		WriteTimeout: serverTimeout,
	}
	log.Fatal(srv.ListenAndServe())
}
