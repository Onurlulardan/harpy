package iamactorv1

import _ "embed"

//go:embed sql/consume.sql
var consumeStatement string
