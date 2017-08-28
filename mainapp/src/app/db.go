// The purpose of this file is to put all of the interaction with MongoDB in
// one place.  This lets us change backends without having to modify other
// parts of the code, and it also makes it easier to see how the program
// interacts with data, since it's all in one place.
package main

import (
	"encoding/json"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"log"
	"time"
	"io"
	"errors"
	"mime/multipart"
	"bufio"
  _ "image/png"
	_ "image/jpeg"
)

// the database connection
var db *mgo.Database

// a user row from the database looks like this:
type User struct {
	ID       bson.ObjectId `bson:"_id"`
	State    int           `bson:"state"`
	Googleid string        `bson:"googleid"`
	Name     string        `bson:"name"`
	Email    string        `bson:"email"`
	Create   time.Time     `bson:"create"`
}

type Project struct {
	ID            bson.ObjectId `bson:"_id" json:"id,omitempty"`
	UserID        string        `bson:"userid" json:"userid"`
	CodeFileName  string        `bson:"codefilename" json:"codefilename"`
	Images        []string       `bson:"images" json:"images"`
	Create        time.Time     `bson:"create"`
}

// open the database
func openDB() {
	var err error
	log.Println("opening database " + cfg.DbHost)
	m, err := mgo.Dial(cfg.DbHost)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("database open")
	db = m.DB(cfg.DbName)
}

// close the database
// NB: We defer() this from main()
func closeDB() {
	log.Println("closing database")
	db.Session.Close()
}



func deleteProject(projectName string, userId string) {
	// Get collection
	collection := db.C("projects")
	err := collection.Remove(bson.M{ "$and": []bson.M{ bson.M{"codefilename": projectName}, bson.M{"userid": userId}}})
	if err != nil {
		log.Println(err)
	}
}

func writeToGridFile(file multipart.File, gridFile *mgo.GridFile) error {
    reader := bufio.NewReader(file)
    defer func() { file.Close() }()
    // make a buffer to keep chunks that are read
    buf := make([]byte, 1024)
    for {
        // read a chunk
        n, err := reader.Read(buf)
        if err != nil && err != io.EOF {
            return errors.New("Could not read the input file")
        }
        if n == 0 {
            break
        }
        // write a chunk
        if _, err := gridFile.Write(buf[:n]); err != nil {
            return errors.New("Could not write to GridFs for "+ gridFile.Name())
        }
    }
    gridFile.Close()
    return nil
}

func getFileFromGridFS(filename string, bucket string) []byte {
	file, err := db.GridFS(bucket).Open(filename)

	b :=  make([]byte, 261120)
	_, err = file.Read(b)
	if err != nil {
		log.Println("file reading: ", err.Error())
	}
	err = file.Close()
	if err != nil {
		log.Println("file closing: ", err.Error())
	}
	return b
}

// insert a row into the user table
func addNewUser(googleid string, name string, email string, state int) error {
	u := User{
		ID:       bson.NewObjectId(),
		State:    state,
		Googleid: googleid,
		Name:     name,
		Email:    email,
		Create:   time.Now(),
	}
	err := db.C("users").Insert(u)
	if err != nil {
		log.Println(err)
	}
	return err
}

// get a user's record, to make login/register decisions
func getUserById(googleId string) (*User, error) {
	u := User{}
	err := db.C("users").Find(bson.M{"googleid": googleId}).Select(nil).One(&u)
	// NB: Findone returns an error on not found, so we need to
	//     disambiguate between DB errors and not-found errors
	if err != nil {
		if err.Error() == "not found" {
			return nil, nil
		}
		log.Println("Error querying users", err)
		return nil, err
	}
	return &u, nil
}

func createProject(project Project) bool {
	project.ID = bson.NewObjectId()
	project.Images = []string{}
	err := db.C("projects").Insert(project)

	if err != nil {
		log.Println(err)
		return false
	}
	return true
}

// Update a row in the project table
func updateProjectRow(project Project) bool {
	q:= bson.M{"_id": project.ID}
	fields := bson.M{"userid": project.UserID,
		"codefilename": project.CodeFileName, "images": project.Images,
	}

	change := bson.M{"$set": fields}
	err := db.C("projects").Update(q, change)
	if err != nil {
		log.Println("update row error: ", err.Error())
		return false
	}
	log.Println("update succesful")
	return true
}

func getProjectByName(projectName string, userID string) []byte {
	project := Project{}
	log.Println(projectName + " " + userID)
	err := db.C("projects").Find(bson.M{ "$and": []bson.M{ bson.M{"codefilename": projectName}, bson.M{"userid": userID}}}).Select(nil).One(&project)
	if err != nil {
		if err.Error() == "not found" {
			log.Println("not found")
			return []byte("not found")
		}
		log.Println("Error querying projects", err)
		return []byte(err.Error())
	}
	jsonData, err := json.Marshal(project)
	if err != nil {
		log.Println("internal error ")
		return []byte("internal error")
	}
	log.Println("successful");
	return jsonData
}

func getProjectObjectByName(projectName string, userID string) Project {
	project := Project{}
	log.Println(projectName + " " + userID)
	err := db.C("projects").Find(bson.M{ "$and": []bson.M{ bson.M{"codefilename": projectName}, bson.M{"userid": userID}}}).Select(nil).One(&project)
	if err != nil {
		return Project{}
	}
	return project
}

func getAllProjects(userID string) []byte {
	var results []Project
	err := db.C("projects").Find(bson.M{"userid": userID}).All(&results)
	if err != nil {
		log.Fatal(err)
	}

	jsonData, err := json.Marshal(results)
	if err != nil {
		return []byte("error")
	}
	return jsonData
}
