package app

import (
  "net/http"
  "fmt"
  "strings"

  "appengine"
  "appengine/datastore"
  "appengine/taskqueue"

  "text/template"
  "net/url"
  "encoding/json"
)

const backendName = "fetching"

func init() {
  http.HandleFunc("/", root)
  http.HandleFunc("/add", add)
  http.HandleFunc("/cron", cron)
  http.HandleFunc("/view", view)
}

func add(w http.ResponseWriter, r *http.Request) {
  c := appengine.NewContext(r)
  asins := strings.Split(r.FormValue("asin"), "|");
  for i:=0;i<len(asins);i++ {
    asin := Asin{asins[i]}

    if _, err := datastore.Put(c, datastore.NewKey(c, "asin", asins[i], 0, nil), &asin); err != nil {
      http.Error(w, err.Error(), http.StatusInternalServerError)
      return
    }
  }
}

func root(w http.ResponseWriter, r *http.Request) {
  c := appengine.NewContext(r)
  q := datastore.NewQuery("asin")

  var asins []*Asin
  _, err := q.GetAll(c, &asins)
  if err != nil {
    return
  }

  if err := template.Must(template.ParseFiles("app/root.html")).Execute(w, asins); err != nil {
    c.Errorf("root: %v", err)
  }
}

func view(w http.ResponseWriter, r *http.Request) {
  c := appengine.NewContext(r)
  q := datastore.NewQuery("snapshot").
    Filter("Asin=", r.FormValue("asin")).
    Order("-Time")

  var ss []*Snapshot
  if _, err := q.GetAll(c, &ss); err != nil {
    return
  }

  /*var Listing struct{listing map[string]Entity}*/
  /*var History struct{history []snap}*/
  type Snapcurrent struct {
    rank map[string]Entity
    buybox Buybox
  }
  data := make([]Snapcurrent, len(ss))
  i:=0

  for _, snap := range ss {
    var entities map[string]Entity
    if err := json.Unmarshal(snap.Ranking, &entities); err != nil {
      fmt.Println("error:", err)
      return
    }

    var buybox Buybox
    if err := json.Unmarshal(snap.Buybox, &buybox); err != nil {
      fmt.Println("error:", err)
      return
    }

    data[i] = Snapcurrent{entities, buybox}
    i++
  }
  /*fmt.Printf("%+v", data)*/

  if err := template.Must(template.ParseFiles("app/view.html")).
    Execute(w, data); err != nil {
    c.Errorf("root: %v", err)
  }
}

func cron(w http.ResponseWriter, r *http.Request) {
  c := appengine.NewContext(r)
  q := datastore.NewQuery("asin")

  for t := q.Run(c);; {
    var x Asin
    _, err := t.Next(&x)
    if err == datastore.Done {
      break
    }
    if err != nil  {
      return
    }

    task := taskqueue.NewPOSTTask("/task/fetching", url.Values{
      "asin": {x.Name},
    })

    if !appengine.IsDevAppServer() {
      host := backendName+"."+appengine.DefaultVersionHostname(c)
      task.Header.Set("Host", host)
    }

    if _, err := taskqueue.Add(c, task, ""); err != nil {
      c.Errorf("add fetching task: %v", err)
      http.Error(w, "Error: couldn't schedule fetching task", 500)
      return
    }

    fmt.Fprintf(w, "OK")
  }
}
