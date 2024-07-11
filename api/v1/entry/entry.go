package entry

import (
	"encoding/json"
	"log"
	"io"
	"fmt"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type Entry struct {
	ID        int64     `db:"id" json:"id"`
	SourceID  int64     `db:"source_id" json:"source_id"`
	Name      string    `db:"name" json:"name"`
	Image     string    `db:"image" json:"image"`
	Content   string    `db:"content" json:"content"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type EntriesJson struct {
	Entries []Entry `json:"entries"`
}

type Source struct {
	ID   *int64 `db:"id" json:"id"`
	Name string `db:"name" json:"name"`
	Url  string `db:"url" json:"url"`
	Type string `db:"type" json:"type"`
}

type IDs struct {
	IDs []int64 `json:"ids"`
}

func Handler(w http.ResponseWriter, r *http.Request) {
	db, err := sqlx.Open("postgres", "")
	if err != nil {
		log.Printf("sql.Open error %s", err)
	}
	defer db.Close()
	switch r.Method {
	case http.MethodGet:
	case http.MethodPost:
		var entriesJson EntriesJson
		query := `
			INSERT INTO entry (
				source_id,
				name,
				image,
				content,
				created_at
			) VALUES (
				:source_id,
				:name,
				:image,
				:content,
				:created_at
			)
		`
		// リクエストボディを読み込む
		jsonBytes, err := io.ReadAll(r.Body)
		if err != nil {
			log.Println(fmt.Sprintf("read error: %v", err))
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		err = json.Unmarshal(jsonBytes, &entriesJson)
		if err != nil {
			log.Println(fmt.Sprintf("json decode error: %v body:%v", err, r.Body))
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		if len(entriesJson.Entries) == 0 {
			log.Println("json unexpected error: empty body")
			http.Error(w, "json unexpected error: empty body", http.StatusBadRequest)
		}
		for _, entry := range entriesJson.Entries {
			_, err = db.NamedExecContext(r.Context(), query, entry)
			if err != nil {
				log.Println(fmt.Sprintf("insert error: %v", err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		}
		// レスポンスボディに書き込む
		err = json.NewEncoder(w).Encode(&entriesJson)
		if err != nil {
			log.Println(fmt.Sprintf("json encode error: %v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	case http.MethodPut:
	case http.MethodDelete:
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
