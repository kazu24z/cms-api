package category

import "embed"

//go:embed queries/*.sql
var queryFS embed.FS

func loadQuery(name string) string {
	data, err := queryFS.ReadFile("queries/" + name)
	if err != nil {
		panic("failed to load query: " + name)
	}
	return string(data)
}

var (
	queryGetAll  = loadQuery("get_all.sql")
	queryGetByID = loadQuery("get_by_id.sql")
	queryCreate  = loadQuery("create.sql")
	queryUpdate  = loadQuery("update.sql")
	queryDelete  = loadQuery("delete.sql")
)
