package app

import (
  "appengine"
  "net/http"
  "appengine/urlfetch"
  "time"
  "io/ioutil"

  "fmt"
  "appengine/datastore"
  "encoding/json"
)

func init() {
  http.HandleFunc("/task/fetching", fetchingTask)
}

func fetchingTask(w http.ResponseWriter, r *http.Request) {
  var (
    c = appengine.NewContext(r)
    asin = r.FormValue("asin")
  )

  defer func() {
    if e := recover(); e != nil {
      c.Criticalf("panic: %v", e)
    }
  }()

  html, err := fetch("http://www.amazon.com/gp/offer-listing/" + asin + "/sr=/qid=/ref=olp_tab_new?ie=UTF8&coliid=&me=&qid=&sr=&seller=&colid=&condition=new", r)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
  rs_listing := parseListingPage(string(html))

  html, err = fetch("http://www.amazon.com/gp/product/" + asin + "/ref=olp_product_details?ie=UTF8&me=&seller=", r)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
  rs_buybox := parseBuyboxPage(string(html), rs_listing)
  fmt.Println(rs_buybox)

  content, err := json.Marshal(rs_listing)
  content1, err := json.Marshal(rs_buybox)

  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }

  s := Snapshot {
    Asin: asin,
    Time: time.Now(),
    Ranking: content,
    Buybox: content1,
  }

  datastore.Put(c, datastore.NewIncompleteKey(c, "snapshot", nil), &s)
}

func fetch(url string, r *http.Request) ([]byte, error) {

  c := appengine.NewContext(r)
  client := urlfetch.Client(c)
  resp, err := client.Get(url)
  defer resp.Body.Close()

  if err != nil {
    return nil, err
  }

  bs, err := ioutil.ReadAll(resp.Body)
  /*bs, err := ioutil.ReadFile("app/html.html")*/
  if err != nil {
    return nil, err
  }

  return bs, nil
}
