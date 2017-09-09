// A demo web application to show how to use OAuth 2.0 (Google+ Provider) and
// MongoDB from Go.
package main

import (
	"flag"
	"log"
	"net/http"
)

// The main function just configures resources and starts listening for new
// web requests
func main() {
	// parse command line options
	configPath := flag.String("configfile", "config.json", "Path to the configuration (JSON) file")
	flag.Parse()

	// load the JSON config file
	loadConfig(*configPath)

	// open the database
	openDB()
	defer closeDB()

	connectToRedis()


	// set up the routes... it's good to have these all in one place,
	// since we need to be cautious about orders when there is a common
	// prefix
	router := new(Router)

	router.Register("/projectobject/[0-9a-zA-z]+$", "GET", handleGetProjectObject)
	router.Register("/projects/[0-9a-zA-z]+$", "GET", handleGetProject)
	router.Register("/projects", "GET", handleGetAllProjects)
	router.Register("/projects", "POST", handlePostProject)
	router.Register("/projects", "DELETE", handleDeleteProject)

	router.Register("/python", "POST", handleRunPython)

	// uploading files either for code file or images that
	// will be used by this project
	router.Register("/uploadcode",   "POST",   handleUploadCode)
	router.Register("/uploadimages", "POST", handleUploadImage)

	// OAuth and login/out routes
	router.Register("/auth/google/callback$", "GET", handleGoogleCallback)
	router.Register("/register", "GET", handleGoogleRegister)
	router.Register("/logout", "GET", handleLogout)
	router.Register("/login", "GET", handleGoogleLogin)



	// Static files
	// router.Register("/public/", "GET", handlePublicFile)   // NB: regexp
	// router.Register("/private/", "GET", handlePrivateFile) // NB: regexp
	// The logged-in main page
	// router.Register("/app", "GET", handleApp)
	// The not-logged-in main page
	// router.Register("/", "GET", handleMain)

	// print a diagnostic message and start the server
	log.Println("Server running on port " + cfg.AppPort)
	http.ListenAndServe(":"+cfg.AppPort, router)
}
