// SpaceTraders dashboard server
package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"strings"
)

var agentName string

func main() {
	runServer()
}

func runServer() {
	temIndex := template.Must(template.New("dashboard.html").ParseFiles("dashboard.html"))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("/")
		err := temIndex.Execute(w, agentName)
		if err != nil {
			log.Fatal(err)
		}
	})

	http.HandleFunc("/htmx.min.js", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "htmx.min.js")
	})
	http.HandleFunc("/style.css", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "style.css")
	})
	http.HandleFunc("/about", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "about.html")
	})

	http.HandleFunc("/server-check", func(w http.ResponseWriter, r *http.Request) {
		resp, err := http.Get("https://api.spacetraders.io/v2")
		if err != nil {
			log.Fatal(err)
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		body = append([]byte{uint8('<'), uint8('p'), uint8('>')}, body...)
		body = append(body, []byte{uint8('<'), uint8('/'), uint8('p'), uint8('>')}...)
		_, err = w.Write(body)
		if err != nil {
			log.Fatal(err)
		}
		resp.Body.Close()
	})

	temRegister := template.Must(template.New("register.html").ParseFiles("register.html"))
	http.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "PUT" {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				log.Fatal(err)
			}
			requestData := strings.Split(string(body), "&")
			for i := range requestData {
				attrValue := strings.Split(requestData[i], "=")
				if attrValue[0] == "agent" {
					agentName = attrValue[1]
					fmt.Println(attrValue)
				}
			}
			err = temRegister.Execute(w, agentName)
			if err != nil {
				log.Fatal(err)
			}
			return
		}
		err := temRegister.Execute(w, agentName)
		if err != nil {
			log.Fatal(err)
		}
		/*_, err := w.Write([]byte(`<!DOCTYPE html><div hx-target="this" hx-swap="outerHTML">
		            <div><label>Name</label>: Placeholder</div>
		            <button hx-get="/register-new">Edit agent</button>
		        </div>`,
				))*/
		/*if err != nil {
			log.Fatal(err)
		}*/
	})

	http.HandleFunc("/register-new", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "register-new.html")
		fmt.Println("register-new")
		/*_, err := w.Write([]byte(`<!DOCTYPE html><form hx-put="/register" hx-target="this" hx-swap="outerHTML">
		        <div>
		            <label>Agent</label>
		            <input type="text" name="agent" value="New agent name">
		        </div>
		        <button>Submit</button>
		        <button hx-get="/register">Cancel</button>
		        </form>`,
				))
				if err != nil {
					log.Fatal(err)
				}*/
	})

	fmt.Println("Server listening on 8080")
	http.ListenAndServe(":8080", nil)
}
