package template

import _ "embed"

//go:embed queries/get_all.sql
var queryGetAll string

//go:embed queries/get_by_name.sql
var queryGetByName string

//go:embed queries/upsert.sql
var queryUpsert string

