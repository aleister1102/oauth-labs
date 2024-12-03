package main

import (
	"github.com/cyllective/oauth-labs/configure/internal/db"
	"github.com/cyllective/oauth-labs/configure/internal/lab00"
	"github.com/cyllective/oauth-labs/configure/internal/lab01"
	"github.com/cyllective/oauth-labs/configure/internal/lab02"
	"github.com/cyllective/oauth-labs/configure/internal/lab03"
	"github.com/cyllective/oauth-labs/configure/internal/lab04"
	"github.com/cyllective/oauth-labs/configure/internal/lab05"
	"github.com/cyllective/oauth-labs/configure/internal/victim"
)

func main() {
	lab00.Configure()
	lab01.Configure()
	lab02.Configure()
	lab03.Configure()
	lab04.Configure()
	lab05.Configure()
	victim.Configure()
	db.Configure()
}
