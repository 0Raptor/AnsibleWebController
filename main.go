/*
   Web-application to execute ansible playbooks and monitor results
   Copyright (C) 2022  0Raptor

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU General Public License for more details.

   You should have received a copy of the GNU General Public License
   along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package main

import (
	"encoding/xml"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

//GlOBAL CONST VARS
var confpath = "config/settings.xml"
var commandspath = "config/commands.xml"

// GLOBAL VARS
var port = ""
var logdir = ""
var tasks Tasks

// ------------------------------------- File Handling ------------------------------------- //

// Read conents of file specified as "path" and return it as string
func readFile(path string) string {
	b, err := os.ReadFile(path) // read file to byte array
	if err != nil {             //report error
		log.Fatal(err)
	}
	return string(b)
}

func readXmlFile(path string) []byte {
	// Open xmlFile
	xmlFile, err := os.Open(path)
	// if we os.Open returns an error then handle it
	if err != nil {
		log.Fatal(err)
	}

	// defer the closing of our xmlFile so that we can parse it later on
	defer xmlFile.Close()

	// Read data
	byteValue, _ := ioutil.ReadAll(xmlFile)

	// Return
	return byteValue
}

func writeFile(path string, content string) {
	bytes := []byte(content)               //convert string to bytes
	err := os.WriteFile(path, bytes, 0644) //write string in file
	if err != nil {                        //check for error
		log.Fatal(err)
	}
}

// --------------------------------           ***           -------------------------------- //

// -------------------------------- STRUCTS for XML Parsing -------------------------------- //
// Config
type Config struct {
	XMLName xml.Name `xml:"xml"`
	Port    string   `xml:"port"`
	LogDir  string   `xml:"logdir"`
}

// Tasks
type Tasks struct {
	XMLName xml.Name `xml:"xml"`
	Task    []Task   `xml:"task"`
}
type Task struct {
	XMLName xml.Name `xml:"task"`
	Name    string   `xml:"name"`
	Command string   `xml:"command"`
	Form    Form     `xml:"form"`
}
type Form struct {
	XMLName xml.Name `xml:"form"`
	Input   []Input  `xml:"input"`
}
type Input struct {
	XMLName  xml.Name `xml:"input"`
	Kind     string   `xml:"type"`
	Options  string   `xml:"options"`
	Label    string   `xml:"label"`
	Variable string   `xml:"var"`
}

// --------------------------------           ***           -------------------------------- //

// ---------------------------------- Sanitize User Input ---------------------------------- //

var nonAlphanumericRegex = regexp.MustCompile(`[^a-zA-Z0-9 ]+`)

func clearString(str string) string {
	return nonAlphanumericRegex.ReplaceAllString(str, "")
}

func numericDatetimeMonth(month time.Month) int {
	switch month {
	case time.January:
		return 1
	case time.February:
		return 2
	case time.March:
		return 3
	case time.April:
		return 4
	case time.May:
		return 5
	case time.June:
		return 6
	case time.July:
		return 7
	case time.August:
		return 8
	case time.September:
		return 9
	case time.October:
		return 10
	case time.November:
		return 11
	case time.December:
		return 12
	default:
		return 0
	}
}

// --------------------------------           ***           -------------------------------- //

// Load configuration from file at given path
func loadConfig(path string) {
	// Read XML File
	byteValue := readXmlFile(path)

	// Parse in config struct for further access
	var conf Config
	xml.Unmarshal(byteValue, &conf)

	// Store values in global vars
	port = conf.Port
	logdir = conf.LogDir
}

// Read all tasks from config file and store in global var
func loadTasks(path string) {
	// Read XML File
	byteValue := readXmlFile(path)

	// Parse in tasks struct for further access
	xml.Unmarshal(byteValue, &tasks)
}

func getTaskTable() string {
	replacewith := ""

	for i := 0; i < len(tasks.Task); i++ {
		//get job number ascii encoded
		no := strconv.Itoa(i)

		// add name to first column
		replacewith += "<tr><td align =\"left\" class=\"stop-stretching\"><b>" + no + ": " + tasks.Task[i].Name + "</b></td><td align =\"center\">"

		// create form if required
		if tasks.Task[i].Form.Input != nil {
			// outer form with submit
			replacewith += "<form id=\"taskForm" + no + "\" action=\"/run\"><table>"
			replacewith += "<tr><td></td><td><input type=\"text\" name=\"id\" value=\"" + no + "\" hidden required></td></tr>"

			//add html syntax for each input that is specified in xml
			for j := 0; j < len(tasks.Task[i].Form.Input); j++ {
				if tasks.Task[i].Form.Input[j].Kind != "dropdown" {
					// create inputs other than dropdowns
					replacewith += "<tr><td><label for=\"" + tasks.Task[i].Form.Input[j].Label + "\">" + tasks.Task[i].Form.Input[j].Label + "</label></td>" //label to descripe input
					replacewith += "<td><input type=\"" + tasks.Task[i].Form.Input[j].Kind + "\" name=\"" + tasks.Task[i].Form.Input[j].Variable + "\" id=\"" + tasks.Task[i].Form.Input[j].Variable + "\" required></td></tr>"
				} else { //create dropdown menu
					replacewith += "<tr><td><label for=\"" + tasks.Task[i].Form.Input[j].Variable + "\">" + tasks.Task[i].Form.Input[j].Label + "</label></td>" //label to descripe input

					// outer dropdown
					replacewith += "<td><select name=\"" + tasks.Task[i].Form.Input[j].Variable + "\" id=\"" + tasks.Task[i].Form.Input[j].Variable + "\" required>"

					//extract options
					opts := strings.Split(tasks.Task[i].Form.Input[j].Options, ";")

					//add options nested in select
					for k := 0; k < len(opts); k++ {
						replacewith += "<option>" + opts[k] + "</option>"
					}

					//close outer dropdown
					replacewith += "</select></td></tr>"
				}
			}

			// close form
			replacewith += "</table></form>"

			// create button to fire job (with data from form)
			replacewith += "</td><td align =\"right\" class=\"stop-stretching\"><input type=\"submit\" form=\"taskForm" + no + "\" value=\"Start\"/></td></tr>"
		} else { // else use a normal button to fire job
			replacewith += "</td><td align =\"right\" class=\"stop-stretching\"><a href=\"run?id=" + no + "\" class=\"button\">Start</a></right></td></tr>"
		}
	}

	return replacewith
}

// Find all log files in log dir and return their names
func getLogs() string {
	logs := ""

	//get files in logdir
	files, err := ioutil.ReadDir(logdir)
	if err != nil {
		log.Fatal(err)
	}

	//loop over found files
	for _, file := range files {
		if !file.IsDir() { //if not a directory append to list
			logs += "<tr><td>"
			logs += "<a href=\"show?logfile=" + logdir + "/" + file.Name() + "\" target=\"_blank\">" + strings.Replace(file.Name(), ".log", "", 1) + "</a>"
			logs += "</td></tr>"
		}
	}

	return logs
}

func index(w http.ResponseWriter, r *http.Request) {
	//get data from url (e.g. confirmation a task was started)
	message := ""
	for k, v := range r.URL.Query() {
		//log.Printf("%s: %s\n", k, v)
		if k == "started" {
			message = clearString(v[0])
		}
	}

	// set header for web content
	w.Header().Set("Content-Type", "text/html")

	// load html file
	content := readFile("html/index.html")

	// replace spacers (??xxx??) with actual content
	content = strings.Replace(content, "??taskList??", getTaskTable(), 1)
	content = strings.Replace(content, "??logList??", getLogs(), 1)
	//    if a job was startet note it, else remove
	if message != "" {
		content = strings.Replace(content, "??executionReport??", message, 1)
	} else {
		content = strings.Replace(content, "??executionReport??", "", 1)
	}

	w.Write([]byte(content))
}

func runtask(w http.ResponseWriter, r *http.Request) {
	/*
		//try to get data transmitted via different ways
		switch r.Method {
		case "GET": //strip from url ?abc=def
			for k, v := range r.URL.Query() {
				log.Printf("%s: %s\n", k, v)
			}
			w.Write([]byte("Received a GET request\n"))
		case "POST": //get from body
			reqBody, err := ioutil.ReadAll(r.Body)
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("%s\n", reqBody)
			w.Write([]byte("Received a POST request\n"))
		default: //unknown
			w.WriteHeader(http.StatusNotImplemented)
			w.Write([]byte(http.StatusText(http.StatusNotImplemented)))
		}
	*/

	// get task id from url transmitted data
	taskid := -1
	for k, v := range r.URL.Query() {
		if k == "id" {
			tmp, err := strconv.Atoi(v[0])
			if err == nil {
				taskid = tmp
			} else {
				w.Write([]byte("500: Internal Server Error\nFailed to parse given id\n"))
				return
			}
		}
	}
	if taskid == -1 { // check if something was found
		w.Write([]byte("400: Bad Request\nUnable to obtain information with given arguments\n"))
		return
	}

	// load command
	cmd := tasks.Task[taskid].Command

	// get other data from url and replace command vars
	for k, v := range r.URL.Query() {
		if k != "id" {
			cmd = strings.Replace(cmd, "??"+k+"??", v[0], -1)
		}
	}

	// prepare logging
	lognr, err := strconv.Atoi(readFile(logdir + "/cntr")) // load last job id                                           // get next
	if err != nil {
		w.Write([]byte("500: Internal Server Error\nFailed to parse current job id\n"))
		return
	}
	lognr += 1
	writeFile(logdir+"/cntr", strconv.Itoa(lognr)) // write new job id back

	// apped >> to output to logdir
	now := time.Now() //get current time
	datetime := strconv.Itoa(now.Day()) + "." + strconv.Itoa(numericDatetimeMonth(now.Month())) + "." + strconv.Itoa(now.Year()) + " " +
		strconv.Itoa(now.Hour()) + "." + strconv.Itoa(now.Minute()) + "." + strconv.Itoa(now.Second()) //parse time to string
	logfilename := "Job " + strconv.Itoa(lognr) + " - Task " + strconv.Itoa(taskid) + " - " + datetime + ".log" //create name for logfile
	cmd += " > " + logdir + "/" + logfilename                                                                   // update cmd with new data

	// debug
	log.Printf("Going to execute: %s", cmd)

	// run command in shell
	shell := exec.Command(cmd)
	err = shell.Start()
	if err != nil {
		log.Fatal(err)
		w.Write([]byte("500: Internal Server Error\nSomething went wrong executing the command\n"))
		return
	}

	//navigate back to main
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte("<html><head><meta http-equiv=\"refresh\" content=\"0; URL=?=Executed Task " + strconv.Itoa(taskid) + " as Job " + strconv.Itoa(lognr) + "\"></head>" +
		"<body>If you are not redirected automatically, follow this <a href=\"?started=" + strconv.Itoa(taskid) + " as Job " + strconv.Itoa(lognr) + "\">Link</a></body></html>"))
}

func showlog(w http.ResponseWriter, r *http.Request) {
	//get data from url
	path := ""
	for k, v := range r.URL.Query() {
		//log.Printf("%s: %s\n", k, v)
		if k == "logfile" {
			path = v[0]
		}
	}

	//check input before loading file
	if path == "" { //no data given
		w.Write([]byte("400: Bad Request\nNo argument given\n"))
		return
	} else if !strings.HasPrefix(path, logdir) { //requested file outside of logdir
		w.Write([]byte("403: Forbidden\nRequested ressource not in approved directories\n"))
		return
	} else {
		if _, err := os.Stat(path); err == nil { //file exists
			// set header for web content
			w.Header().Set("Content-Type", "text/html")

			// load html file
			content := readFile("html/read_log.html")

			// replace spacers (??xxx??) with actual content
			pathparts := strings.Split(path, "/")
			content = strings.Replace(content, "??title??", strings.Replace(pathparts[len(pathparts)-1], ".log", "", 1), 1)
			log := readFile(path)
			log = strings.Replace(log, "\r\n", "<br />", -1) //html cannot display normal line breaks
			log = strings.Replace(log, "\n", "<br />", -1)
			content = strings.Replace(content, "??content??", log, 1)

			//display
			w.Write([]byte(content))
			return
		} else if errors.Is(err, os.ErrNotExist) { //file DOES NOT exists
			w.Write([]byte("404: Not Found\nRessource not available on the server\n"))
			return
		} else { //unknown status, but file will not be openable
			w.Write([]byte("500: Internal Server Error\nStatus of requested ressource is unknown\n"))
			return
		}
	}
}

func reload(w http.ResponseWriter, r *http.Request) {
	log.Print("Reloading AWC configuration...")
	loadConfig(confpath)
	log.Print("Configuration loaded.")
	loadTasks(commandspath)
	log.Print("Commands loaded.")
	log.Print("DONE. Port change will be ignored.")
}

func main() {
	// Configure backgound listener
	http.HandleFunc("/", index)
	http.HandleFunc("/reload", reload)
	http.HandleFunc("/run", runtask)
	http.HandleFunc("/show", showlog)

	// Greet user
	log.Print("Ansible Web Controller AWC - 2022  0Raptor - GNU GENERAL PUBLIC LICENSE v3")
	log.Print("This program comes with ABSOLUTELY NO WARRANTY.")
	log.Print("")

	// Load config
	log.Print("Preparing AWC...")
	loadConfig(confpath)
	log.Print("Configuration loaded.")
	loadTasks(commandspath)
	log.Print("Commands loaded.")

	// Start http listening to port 8080 - in case of crash, log it to console
	log.Printf("AWC starts listening on %s...", port)
	log.Fatal(http.ListenAndServe(port, nil))

	// Exit
	log.Print("AWC was terminated. Exiting...")
}
