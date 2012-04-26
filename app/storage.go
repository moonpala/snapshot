package app

import (
  "time"
)

type Asin struct {
  Name string
}

type Snapshot struct {
  Asin string
  Time time.Time
  Ranking []byte
  Buybox []byte
}

