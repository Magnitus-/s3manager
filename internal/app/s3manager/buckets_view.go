package s3manager

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"io/fs"
	"net/http"
	"strings"
	"time"

	yaml "gopkg.in/yaml.v2"
	"github.com/minio/minio-go/v7"
)



// HandleBucketsView renders all buckets on an HTML page.
func HandleBucketsView(s3 S3, templates fs.FS, allowDelete bool, sharedBucketsPath string) http.HandlerFunc {
	type pageData struct {
		Buckets     []minio.BucketInfo
		AllowDelete bool
	}

	return func(w http.ResponseWriter, r *http.Request) {
		buckets, err := s3.ListBuckets(r.Context())
		if err != nil {
			handleHTTPError(w, fmt.Errorf("error listing buckets: %w", err))
			return
		}

		if sharedBucketsPath != "" {
			matches := strings.SplitN(sharedBucketsPath, "/", 2)
			if len(matches) != 2 {
				handleHTTPError(w, fmt.Errorf("error getting shared buckets object: Shared path is not valid"))
				return
			}
			bucketName := matches[0]
			path := matches[1]

			object, err := s3.GetObject(r.Context(), bucketName, path, minio.GetObjectOptions{})
			if err != nil {
				handleHTTPError(w, fmt.Errorf("error getting shared buckets object: %w", err))
				return
			}

			bs, bErr := ioutil.ReadAll(object)
			if bErr != nil {
				handleHTTPError(w, fmt.Errorf("error getting shared buckets object: %w", bErr))
				return
			}

			additionalBuckets := []string{}
			err = yaml.Unmarshal(bs, &additionalBuckets)
			if err != nil {
				handleHTTPError(w, fmt.Errorf("error getting shared buckets object: %w", err))
				return
			}

			for _, name := range additionalBuckets {
				buckets = append(buckets, minio.BucketInfo{Name: name})
			}
		}

		data := pageData{
			Buckets:     buckets,
			AllowDelete: allowDelete,
		}

		t, err := template.New("").Funcs(template.FuncMap{
			"time_defined": func(t time.Time) bool {
				return !t.IsZero()
			},
		}).ParseFS(templates, "layout.html.tmpl", "buckets.html.tmpl")
		if err != nil {
			handleHTTPError(w, fmt.Errorf("error parsing template files: %w", err))
			return
		}
		err = t.ExecuteTemplate(w, "layout", data)
		if err != nil {
			handleHTTPError(w, fmt.Errorf("error executing template: %w", err))
			return
		}
	}
}
