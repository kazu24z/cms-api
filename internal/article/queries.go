package article

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
	queryGetAll       = loadQuery("get_all.sql")
	queryGetByID      = loadQuery("get_by_id.sql")
	queryGetPublished = loadQuery("get_published.sql")
	queryCreate       = loadQuery("create.sql")
	queryUpdate       = loadQuery("update.sql")
	queryPublish      = loadQuery("publish.sql")
	queryDelete       = loadQuery("delete.sql")
	queryDeleteTags   = loadQuery("delete_tags.sql")
	queryInsertTag    = loadQuery("insert_tag.sql")
)

