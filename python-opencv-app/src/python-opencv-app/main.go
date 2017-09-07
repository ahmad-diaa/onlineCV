package main

import (
  "net/http"
  "os"
  "io"
  "strings"
  "os/exec"
  "log"
  "bytes"
  "encoding/json"
  "io/ioutil"
  "gopkg.in/mgo.v2/bson"
  "time"
  "golang.org/x/net/context"

  "cloud.google.com/go/storage"
  "google.golang.org/api/option"
)

type Project struct {
	ID            bson.ObjectId `bson:"_id" json:"id,omitempty"`
	UserID        string        `bson:"userid" json:"userid"`
	CodeFileName  string        `bson:"codefilename" json:"codefilename"`
	Images        []string       `bson:"images" json:"images"`
	Create        time.Time     `bson:"create"`
}

type Result struct {
	Output        string         `json:"output"`
	Images        []string       `json:"images"`
}

var pythonFile string
var requestFilenames []string
var userID string


func saveFiles(r *http.Request, w http.ResponseWriter) error{
    b, err := ioutil.ReadAll(r.Body)
  	defer r.Body.Close()
  	if err != nil {
      log.Println("read request error: ", err)
  		http.Error(w, err.Error(), 500)
  		return err
  	}
  	var project Project
  	err = json.Unmarshal(b, &project)
  	if err != nil {
      log.Println("UNmarshal request error: ", err)
  		http.Error(w, err.Error(), 500)
  		return err
  	}
    userID = project.UserID
    project.Images = append(project.Images, project.CodeFileName)
    for _, filename := range project.Images {
        if filename == "" {
          continue
        }

        requestFilenames = append(requestFilenames, filename)
        if filename == project.CodeFileName {
          filename   = "python.py"
        }
        out, err := os.Create(filename)
        if err != nil  {
          return err
        }
        defer out.Close()
        var newFilename string
        newFilename = project.UserID + filename

        log.Println("new filename: ", newFilename)
        ctx := context.Background()
        client, err := storage.NewClient(ctx, option.WithCredentialsFile("onlineCV-b3ad0190adc1.json"))

        if err != nil {
      		log.Println("failed to create client: %v", err)
      		return err
      	}
        var bucketName string
        if filename == "python.py" {
          bucketName = "onlinecv-codes"
          newFilename =  project.UserID + project.CodeFileName
        } else
        {
          bucketName = "onlinecv-images"
        }
      	defer client.Close()
      	buf := &bytes.Buffer{}
      	d := &bucket_struct{
      		w:          buf,
      		ctx:        ctx,
      		client:     client,
      		bucket:     client.Bucket(bucketName),
      		bucketName: bucketName,
      		cleanUp:    []string{},
      	}
      	// d.cleanUp = append(d.cleanUp, newFilename)
      	// d.deleteFiles()
      	sDec:= d.readFile(newFilename)
      	if d.failed {
      		w.WriteHeader(http.StatusInternalServerError)
      		log.Println("bucket file read failed")
          return err
      	} else {
      		w.WriteHeader(http.StatusOK)
      		log.Println("bucket file read succeeded")
      	}
        _, err = io.Copy(out, bytes.NewReader(sDec))
        if err != nil  {
          return err
        }
    }
    return nil
}

func runCodeHandler(w http.ResponseWriter, r *http.Request) {
     pwd, _ := os.Getwd()
     saveFiles(r, w)
     var jsonData Result
     w.Header().Set("Content-Type", "application/json")
     cmd := exec.Command("sh", "-c", "python3 python.py")
     cmd.Env = os.Environ()
     var outb, errb bytes.Buffer
      cmd.Stdout = &outb
      cmd.Stderr = &errb
      err := cmd.Run()
      if err != nil {
         jsonData.Output = errb.String()
         log.Println("running error: ", err.Error())
      } else {
        files, _ := ioutil.ReadDir(pwd)
        for _, fileinfo := range files {
          filename := string(fileinfo.Name())
          if !isImage(filename) {
            continue
          }

          if contains(filename,requestFilenames) {
            continue
          }
          ctx := context.Background()
          client, err := storage.NewClient(ctx, option.WithCredentialsFile("onlineCV-b3ad0190adc1.json"))

        	if err != nil {
        		log.Println("failed to create client: %v", err)
        		return
        	}
          bucketName := "onlinecv-images"
        	defer client.Close()
          buf := &bytes.Buffer{}
        	d := &bucket_struct{
        		w:          buf,
        		ctx:        ctx,
        		client:     client,
        		bucket:     client.Bucket(bucketName),
        		bucketName: bucketName,
        		cleanUp:    []string{},
        	}
          bytes, err := ioutil.ReadFile(filename)
          if err != nil {
              log.Println(err)
          }
          filename = userID + filename


        	d.createFile(filename, bytes, "image/jpeg")
        	if d.failed {
        		w.WriteHeader(http.StatusInternalServerError)
        		log.Println("bucket file read failed")
            return
        	} else {
        		w.WriteHeader(http.StatusOK)
        		log.Println("bucket file read succeeded")
        	}
          jsonData.Images = append(jsonData.Images, filename)
        }
        jsonData.Output = outb.String()
        log.Println(outb.String())
      }
      jsonString, err := json.Marshal(jsonData)
      if err != nil {
        log.Println("json error " + err.Error())
        return
      }
      log.Println("DONE!!")
      w.Write([]byte(string(jsonString)))
      defer deleteUsedFiles()
}

func deleteUsedFiles() {
  pwd, _ := os.Getwd()
  files, _ := ioutil.ReadDir(pwd)
  for _, fileinfo := range files {
    filename := string(fileinfo.Name())
    if !(filename == "python.py" ||isImage(filename))  {
      continue
    }
    err := os.Remove(filename)
    if err != nil {
      log.Println("Remove file error: ", err.Error())
    }
  }
  requestFilenames = []string{}
}

func main() {
    http.HandleFunc("/", runCodeHandler)
    http.ListenAndServe(":8000", nil)
}
