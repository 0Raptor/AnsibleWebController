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
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

// GLOBAL VARS
var port = ""
var logdir = ""
var tasks Tasks

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
	log.Printf("Found %d tasks", len(tasks.Task))

	for i := 0; i < len(tasks.Task); i++ {
		//get job number ascii encoded
		no := strconv.Itoa(i)

		// add name to first column
		replacewith += "<tr><td><b>" + no + ": " + tasks.Task[i].Name + "</b><br />"

		// create form if required
		if tasks.Task[i].Form.Input != nil {
			// outer form with submit
			replacewith += "<form id=\"taskForm" + no + "\" action=\"/run\">"
			replacewith += "<input type=\"text\" name=\"id\" value=\"" + no + "\" hidden required><br>"

			//add html syntax for each input that is specified in xml
			for j := 0; j < len(tasks.Task[i].Form.Input); j++ {
				if tasks.Task[i].Form.Input[j].Kind != "dropdown" {
					// create inputs other than dropdowns
					replacewith += "<input type=\"" + tasks.Task[i].Form.Input[j].Kind + "\" name=\"" + tasks.Task[i].Form.Input[j].Variable + "\" required><br>"
				} else { //create dropdown menu
					// outer dropdown
					replacewith += "<select name=\"" + tasks.Task[i].Form.Input[j].Variable + "\" required>"

					//extract options
					opts := strings.Split(tasks.Task[i].Form.Input[j].Options, ";")

					//add options nested in select
					for k := 0; k < len(opts); k++ {
						replacewith += "<option>" + opts[k] + "</option>"
					}

					//close outer dropdown
					replacewith += "</select>"
				}
			}

			// close form
			replacewith += "</form>"

			// create button to fire job (with data from form)
			replacewith += "</td><td><input type=\"submit\" form=\"taskForm" + no + "\" value=\"Start\"/></td></tr>"
		} else { // else use a normal button to fire job
			replacewith += "</td><td>" + "Start Button" + "</td></tr>"
		}
	}

	return replacewith
}

// Find all log files in log dir and return their names
func getLogs() string {
	logs := ""
	return logs
}

func index(w http.ResponseWriter, r *http.Request) {
	// set header for web content
	w.Header().Set("Content-Type", "text/html")

	// load html file
	content := readFile("html/index.html")

	// replace spacers (??xxx??) with actual content
	content = strings.Replace(content, "??taskList??", getTaskTable(), 1)
	content = strings.Replace(content, "??logList??", getLogs(), 1)
	//    if a job was startet note it, else remove
	if 1 == 1 {
		content = strings.Replace(content, "??executionReport??", "Started Task", 1)
	} else {
		content = strings.Replace(content, "??executionReport??", "", 1)
	}

	w.Write([]byte(content))
}

func reload(w http.ResponseWriter, r *http.Request) {
	log.Print("Reloading AWC configuration...")
	loadConfig("config/settings.xml")
	log.Print("Configuration loaded.")
	loadTasks("config/commands.xml")
	log.Print("Commands loaded.")
	log.Print("DONE. Port change will be ignored.")
}

func main() {
	// Configure backgound listener
	http.HandleFunc("/", index)
	http.HandleFunc("/reloadc", reload)

	// Greet user
	log.Print("Ansible Web Controller AWC - 2022  0Raptor - GNU GENERAL PUBLIC LICENSE v3")
	log.Print("This program comes with ABSOLUTELY NO WARRANTY.")
	log.Print("")

	// Load config
	log.Print("Preparing AWC...")
	loadConfig("config/settings.xml")
	log.Print("Configuration loaded.")
	loadTasks("config/commands.xml")
	log.Print("Commands loaded.")

	// Start http listening to port 8080 - in case of crash, log it to console
	log.Printf("AWC starts listening on %s...", port)
	log.Fatal(http.ListenAndServe(port, nil))

	// Exit
	log.Print("AWC was terminated. Exiting...")
}
