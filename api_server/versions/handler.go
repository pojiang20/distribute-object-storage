package versions

import (
	"encoding/json"
	"github.com/pojiang20/distribute-object-storage/src/es"
	"github.com/pojiang20/distribute-object-storage/src/utils"
	"log"
	"net/http"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	m := r.Method
	if m != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	//from是ES搜索的起始页的数据序号，表示从头开始不跳过任何一条数据，size表示每一页的数据规模
	from, size := 0, 1000
	name := utils.GetObjectName(r.URL.EscapedPath())
	for {
		metas, err := es.SearchAllVersions(name, from, size)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		for i := range metas {
			b, _ := json.Marshal(metas[i])
			w.Write(b)
			w.Write([]byte("\n"))
		}
		//from和size能保证如果一次查询没处理完，能够再次从from又查询size大小的数据继续操作
		if len(metas) != size {
			return
		}
		from += size
	}
}
