package app

import (
  "regexp"
  /*"encoding/json"*/
  /*"fmt"*/
)

type Entity struct {
  Rank int
  Sales string
  Shipping string
  Seller string
  Iffba bool
}

type Buybox struct {
  Sales string
  Shipping string
  Seller string
  Iffba bool
}

func parseListingPage(html string) map[string]Entity{
  r := regexp.MustCompile(`(?ms)<tbody class="result">.*?(?:<span class="price">\$([\d\.]+)<\/span>.*?)?(?:<span class="price_shipping">\+ \$([^<>]*)<\/span>.*?)?(?:width="[120|90]+" alt="([\w\s\-\,\.]+)?".*?)?(?:<a href="[^"]+"><b>([\w\s\-\,\.]*)<\/b><\/a>.*?)?(?:(Fulfillment) by Amazon.*?)?<\/tbody>`)
  result := r.FindAllStringSubmatch(string(html), -1)
  content := make(map[string]Entity)
  for i := 0; i < len(result); i++ {
    content[string(i)] = Entity{i+1, result[i][1], result[i][2], result[i][3] + result[i][4], result[i][5] == "Fulfillment"}
  }
  return content
}

func parseBuyboxPage(html string, listing map[string]Entity) Buybox{
  r := regexp.MustCompile(`(?ms)<b class="priceLarge">\$([^<>]+)<\/b>.*?<div class="buying" style="[^"]+">.*?<span class="avail[^"]+">In Stock.*?<\/span><br \/>.*?(?:Ships from and sold by <b><a href="[^"]+">(.*?)<\/a>.*?)?(?:Sold by <b><a href="[^"]+">(.*?)<\/a><\/b> and.*?)?(?:<a href="[^"]+" id="[^"]+"><strong>(.*?)<\/strong>.*?)?<\/[b|a]?>\.`)
  result := r.FindAllStringSubmatch(string(html), -1)
  return Buybox{result[0][1], result[0][2], result[0][3], result[0][4] == "Fulfilled by Amazon"}
}
