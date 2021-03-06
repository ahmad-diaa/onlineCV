// Routes related to REST paths for accessing the DATA table
package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"bytes"
	"github.com/gorilla/schema"
	"encoding/base64"
	"mime/multipart"
	"golang.org/x/net/context"

	"google.golang.org/api/option"
	"cloud.google.com/go/storage"
)

var decoder = schema.NewDecoder()

func do403(w http.ResponseWriter) {
	http.Error(w, "Forbidden", http.StatusForbidden)
}

func do404(w http.ResponseWriter) {
	http.Error(w, "Not Found", http.StatusNotFound)
}

func do500(w http.ResponseWriter) {
	http.Error(w, "Internal Server Error", http.StatusInternalServerError)
}

// Helper routine for sending JSON back to the client a bit more cleanly
func jResp(w http.ResponseWriter, data interface{}) {
	payload, err := json.Marshal(data)
	if err != nil {
		log.Println("Internal Server Error:", err.Error())
		do500(w)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write([]byte(string(payload)))
}


func handleGetProjectObject(w http.ResponseWriter, r *http.Request) {
			if !checkLogin(r) {
				do403(w)
				return
			}
			p := strings.Split(r.URL.Path, "/")
			cookie, _ := r.Cookie("id")
			userId := cookie.Value
			projectName:= p[2]

			res :=  getProjectByName(projectName,userId)
			if bytes.Compare(res, []byte("not found")) == 0 {
				do404(w)
				return
			}
			if bytes.Compare(res, []byte("internal error")) == 0 {
				do500(w)
				return
			}
			var project Project
			if err := json.Unmarshal(res, &project); err != nil {
        log.Println("json unmarshal project object: ",err)
    	}
			codefile := project.CodeFileName + userId
			images   := project.Images
			codeByteArray := getFileFromGridFS(codefile, "code")
			jsonData := make(map[string]string)

     	w.Header().Set("Content-Type", "application/json")
			jsonData["codefile"] = base64.StdEncoding.EncodeToString(codeByteArray)
			for _, image := range images {
				imageData := getFileFromGridFS(image, "image")
				encoded := base64.StdEncoding.EncodeToString(imageData)
				jsonData[image] = encoded
			}

			jsonString, err := json.Marshal(jsonData)
      if err != nil {
        log.Println("json error " + err.Error())
        return
      }
      w.Write([]byte(string(jsonString)))
}

func handleGetAllProjects(w http.ResponseWriter, r *http.Request) {
			if !checkLogin(r) {
				do403(w)
				return
			}
			cookie, _ := r.Cookie("id")
			userId := cookie.Value
			w.Write(getAllProjects(userId))
}

func handleGetProject(w http.ResponseWriter, r *http.Request) {

	if !checkLogin(r) {
		do403(w)
		return
	}

	p := strings.Split(r.URL.Path, "/")
	cookie, _ := r.Cookie("id")
	userId := cookie.Value
	projectName:= p[2]

	res :=  getProjectByName(projectName,userId)
	if bytes.Compare(res, []byte("not found")) == 0 {
		do404(w)
		return
	}
	if bytes.Compare(res, []byte("internal error")) == 0 {
		do500(w)
		return
	}
	log.Println("get succesful")
	w.Header().Set("Content-Type", "application/json")
	w.Write(res)
}

func handlePostProject(w http.ResponseWriter, r *http.Request) {

	if !checkLogin(r) {
		do403(w)
		return
	}

	defer r.Body.Close()
	contents, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("Error reading body of POST request")
		jResp(w, "error")
		return
	}

	var project Project
	log.Println("%s", contents)
	err = json.Unmarshal(contents, &project)
	if err != nil {
		log.Println("Error unmarshaling JSON reply", err)
		jResp(w, "error")
		return
	}
	cookie, _ := r.Cookie("id")
	project.UserID = cookie.Value

	// insert the data
	ok := createProject(project)
	if ok {
		jResp(w, "{res: 'ok'}")
	} else {
		jResp(w, "{res: 'error'}")
		do500(w)
		return
	}
}

// var ImageTemplate string = `<!DOCTYPE html>
//     <html lang="en"><head></head>
//     <body><img src="data:image/jpg;base64,{{.Image}}"></body>`
//
// func handleShowImage(w http.ResponseWriter, r *http.Request) {
// 	  img := getImage("24a3f5ae-b0c7-4e5a-b260-22a4b37bb185")
// 		writeImageWithTemplate(w, &img);
// }
//
// // Writeimagewithtemplate encodes an image 'img' in jpeg format and writes it into ResponseWriter using a template.
// func writeImageWithTemplate(w http.ResponseWriter, img *image.Image) {
//
// 	buffer := new(bytes.Buffer)
// 	if err := jpeg.Encode(buffer, *img, nil); err != nil {
// 		log.Fatalln("unable to encode image.")
// 	}
//
// 	str := base64.StdEncoding.EncodeToString(buffer.Bytes())
// 	if tmpl, err := template.New("image").Parse(ImageTemplate); err != nil {
// 		log.Println("unable to parse image template.")
// 	} else {
// 		data := map[string]interface{}{"Image": str}
// 		if err = tmpl.Execute(w, data); err != nil {
// 			log.Println("unable to execute template.")
// 		}
// 	}
// }

func handleDeleteProject(w http.ResponseWriter, r *http.Request) {
	if !checkLogin(r) {
	 do403(w)
	 return
 }
 defer r.Body.Close()
 contents, err := ioutil.ReadAll(r.Body)
 if err != nil {
	 log.Println("Error reading body of DELETE request")
	 jResp(w, "error")
	 return
 }

 var project Project
 err = json.Unmarshal(contents, &project)
 if err != nil {
	 log.Println("Error unmarshaling JSON reply", err)
	 jResp(w, "error")
	 return
 }
 cookie, _ := r.Cookie("id")
 project.UserID = cookie.Value
 temp := getProjectObjectByName(project.CodeFileName,project.UserID)
 if temp.UserID == "" {
	 return
 }
 deleteFileFromGCE(project.UserID + project.CodeFileName, cfg.CodesBucket, r, w)
 for _,imagename := range project.Images {
	 deleteFileFromGCE(project.UserID + imagename, cfg.ImagesBucket, r, w)
 }
	deleteProject(project.CodeFileName, project.UserID)
}

func handleUploadCode(w http.ResponseWriter, r *http.Request) {
	if !checkLogin(r) {
		do403(w)
		return
	}

	err := r.ParseMultipartForm(0)
    if err != nil {
			log.Println(err)
			do500(w)
			return
    }
	bucketName  := cfg.CodesBucket
	projectName := r.PostFormValue("codefilename")
	cookie, _ := r.Cookie("id")
	userId := cookie.Value
  temp := getProjectObjectByName(projectName,userId)
  if temp.UserID == "" {
 	  do404(w)
	  return
	}
	saveFiles(r, bucketName, w)
}

func handleUploadImage(w http.ResponseWriter, r *http.Request) {
	if !checkLogin(r) {
		do403(w)
		return
	}

	var newFiles []string
	err := r.ParseMultipartForm(0)
    if err != nil {
			log.Println(err)
			do500(w)
			return
    }
		for _, fileHeaders := range r.MultipartForm.File {
				 for _, fileHeader := range fileHeaders {
						filename := fileHeader.Filename
						newFiles = append(newFiles, filename)
			}
		}

	bucketName  := cfg.ImagesBucket
	projectName := r.PostFormValue("codefilename")
	cookie, _ := r.Cookie("id")
	userId := cookie.Value
	if !checkUser(r, userId) {
		do403(w)
		return
	}

	log.Println("project: ", projectName)
	 var project Project
	 project.UserID       = userId
	 project.CodeFileName = projectName
	 project.Images       = newFiles

	temp := getProjectObjectByName(project.CodeFileName,project.UserID)
	if temp.UserID == "" {
		do404(w)
		return
	}

	project.ID = temp.ID
	for _, im := range temp.Images {
		if !sliceContains(im, project.Images) {
			project.Images = append(project.Images, im)
		}
	}
	ok := updateProjectRow(project)
	if !ok {
		do500(w)
		return
	}
	saveFiles(r, bucketName, w)
}

func saveFiles (r *http.Request, bucketName string, w http.ResponseWriter) {
	i := 0
	for _, fileHeaders := range r.MultipartForm.File {
			 for _, fileHeader := range fileHeaders {
					file, _ := fileHeader.Open()
					i++
					filename := fileHeader.Filename
					projectName := r.PostFormValue("codefilename")
					cookie, _ := r.Cookie("id")
					userId := cookie.Value
					var contentType string

					if bucketName == cfg.CodesBucket {
						filename =  userId + projectName
						contentType = "text/plain"
					} else {
						filename =  userId + filename
						contentType = "image/jpeg"
					}

					// deleteFileFromGCE(filename, bucketName)
					writeFiletoGCE(bucketName, w, filename, file, r, contentType)
			}
		}
	}

func writeFiletoGCE(bucketName string, w http.ResponseWriter, newFilename string, file multipart.File, r *http.Request, contentType string) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx, option.WithCredentialsFile(cfg.CloudStorageCredentials))
	if err != nil {
		log.Println("failed to create client: %v", err)
		return
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
	d.createFile(newFilename, file, contentType)
	if d.failed {
		w.WriteHeader(http.StatusInternalServerError)
		buf.WriteTo(w)
		log.Println("bucket file save failed")
	} else {
		w.WriteHeader(http.StatusOK)
		buf.WriteTo(w)
		log.Println("bucket file save succeeded")
	}
}

func deleteFileFromGCE(name string, bucketName string, r *http.Request, w http.ResponseWriter) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx, option.WithCredentialsFile(cfg.CloudStorageCredentials))
	if err != nil {
		log.Println("failed to create client: %v", err)
		return
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
	d.cleanUp = append(d.cleanUp, name)
	d.deleteFiles()

	if d.failed {
		w.WriteHeader(http.StatusInternalServerError)
		buf.WriteTo(w)
		log.Println("bucket file delete failed")
	} else {
		w.WriteHeader(http.StatusOK)
		buf.WriteTo(w)
		log.Println("bucket file delete succeeded")
	}
}
