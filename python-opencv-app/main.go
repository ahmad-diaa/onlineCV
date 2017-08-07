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
  "encoding/base64"
  "io/ioutil"
)

var pythonFile string
var requestFilenames []string

func contains(target string, filenames []string) bool{
  for _, value := range filenames {
      if value == target {
        return true
      }
  }
  return false
}
func saveFiles(r *http.Request) error{
  decoder := json.NewDecoder(r.Body)
  var jsonData map[string]string
  err := decoder.Decode(&jsonData)
   if err != nil {
       log.Println("json decode error: ", err.Error())
   }
    for k, v := range jsonData {
        filename := k
        if filename == "" {
          continue
        }
        requestFilenames = append(requestFilenames, filename)
        if filename == "codefile" {
          filename   = "python.py"
          pythonFile = filename
        }
        out, err := os.Create(filename)
        if err != nil  {
          return err
        }
        defer out.Close()
        sDec, _  := base64.StdEncoding.DecodeString(v)
        _, err = io.Copy(out, bytes.NewReader(sDec))
        if err != nil  {
          return err
        }
    }
    return nil
}

func isImage(filename string) bool {
  if strings.HasSuffix(filename, "jpg") {
    return true
  }
  if strings.HasSuffix(filename, "png") {
    return true
  }
  if strings.HasSuffix(filename, "jpeg") {
      return true
  }
  return false
}

func runCodeHandler(w http.ResponseWriter, r *http.Request) {
     pwd, _ := os.Getwd()
     saveFiles(r)
     jsonData := make(map[string]string)
     w.Header().Set("Content-Type", "application/json")
     cmd := exec.Command("sh", "-c", "python " + pythonFile)
     cmd.Env = os.Environ()
     var outb, errb bytes.Buffer
      cmd.Stdout = &outb
      cmd.Stderr = &errb
      err := cmd.Run()
      if err != nil {
         jsonData["output"] = errb.String()
         log.Println("running error: ", err.Error())
      } else {
        files, _ := ioutil.ReadDir(pwd)
        for _, fileinfo := range files {
          filename := string(fileinfo.Name())
          log.Println("FFF: " , filename)
          if !isImage(filename) {
            continue
          }

          if contains(filename,requestFilenames) {
            continue
          }
          bytes, err := ioutil.ReadFile(filename)
          if err != nil {
              log.Println(err)
          }
          encoded := base64.StdEncoding.EncodeToString(bytes)
          jsonData[filename] = encoded
        }
        jsonData["output"] = outb.String()
      }
      jsonString, err := json.Marshal(jsonData)
      if err != nil {
        log.Println("json error " + err.Error())
        return
      }
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
