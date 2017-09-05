package main
//[START imports]
import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"

	"cloud.google.com/go/storage"
	"golang.org/x/net/context"
	"google.golang.org/appengine/log"
)

//[END imports]

func init() {
	// http.HandleFunc("/", handler)
}

// bucket_struct struct holds information needed to run the various bucket_struct functions.
type bucket_struct struct {
	client     *storage.Client
	bucketName string
	bucket     *storage.BucketHandle

	w   io.Writer
	ctx context.Context
	// cleanUp is a list of filenames that need cleaning up at the end of the bucket_struct.
	cleanUp []string
	// failed indicates that one or more of the bucket_struct steps failed.
	failed bool
}

func (d *bucket_struct) errorf(format string, args ...interface{}) {
	d.failed = true
	fmt.Fprintln(d.w, fmt.Sprintf(format, args...))
	log.Errorf(d.ctx, format, args...)
}

//[START write]
// createFile creates a file in Google Cloud Storage.
func (d *bucket_struct) createFile(fileName string, file []byte, ContentType string) {
	fmt.Fprintf(d.w, "Creating file /%v/%v\n", d.bucketName, fileName)

	wc := d.bucket.Object(fileName).NewWriter(d.ctx)
	wc.ContentType = ContentType
	if _, err := wc.Write(file); err != nil {
		d.errorf("createFile: unable to write data to bucket %q, file %q: %v", d.bucketName, fileName, err)
		return
	}
	if err := wc.Close(); err != nil {
		d.errorf("createFile: unable to close bucket %q, file %q: %v", d.bucketName, fileName, err)
		return
	}
}

//[END write]

//[START read]
// readFile reads the named file in Google Cloud Storage.
func (d *bucket_struct) readFile(fileName string) ([]byte) {
	io.WriteString(d.w, "\nAbbreviated file content (first line and last 1K):\n")

	rc, err := d.bucket.Object(fileName).NewReader(d.ctx)
	if err != nil {
		d.errorf("readFile: unable to open file from bucket %q, file %q: %v", d.bucketName, fileName, err)
		return []byte{}
	}
	defer rc.Close()
	slurp, err := ioutil.ReadAll(rc)
	if err != nil {
		d.errorf("readFile: unable to read data from bucket %q, file %q: %v", d.bucketName, fileName, err)
		return []byte{}
	}

	fmt.Fprintf(d.w, "%s\n", bytes.SplitN(slurp, []byte("\n"), 2)[0])
	if len(slurp) > 1024 {
		fmt.Fprintf(d.w, "...%s\n", slurp[len(slurp)-1024:])
	} else {
		fmt.Fprintf(d.w, "%s\n", slurp)
	}
	return slurp
}

// deleteFiles deletes all the temporary files from a bucket created by this bucket_struct.
func (d *bucket_struct) deleteFiles() {
	io.WriteString(d.w, "\nDeleting files...\n")
	for _, v := range d.cleanUp {
		fmt.Fprintf(d.w, "Deleting file %v\n", v)
		if err := d.bucket.Object(v).Delete(d.ctx); err != nil {
			d.errorf("deleteFiles: unable to delete bucket %q, file %q: %v", d.bucketName, v, err)
			return
		}
	}
}
