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
		var entriesJson EntriesJson
		query := `
			SELECT
				source_id,
				name,
				image,
				content,
				created_at
			FROM
				entry
			WHERE
				id IN (?)
		`
		// クエリパラメータからidを取得
		queryIDs, ok := r.URL.Query()["id"]
		// idが指定されていない場合は全件取得
		if !ok {
			query = `
				SELECT
					source_id,
					name,
					image,
					content,
					created_at
				FROM
					entry
			`
			// 全件取得
			err := db.SelectContext(r.Context(), &entriesJson.Entries, query)
			if err != nil {
				log.Println(fmt.Sprintf("select error: %v", err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			// レスポンスボディに書き込む
			err = json.NewEncoder(w).Encode(&entriesJson)
			if err != nil {
				log.Println(fmt.Sprintf("json encode error: %v", err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		} else if len(queryIDs) == 1 {
			query = `
				SELECT
					source_id,
					name,
					image,
					content,
					created_at
				FROM
					entry
				WHERE
					id = $1
			`
			// 1件取得
			err := db.SelectContext(r.Context(), &entriesJson.Entries, query, queryIDs[0])
			if err != nil {
				log.Println(fmt.Sprintf("select error: %v", err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			// レスポンスボディに書き込む
			err = json.NewEncoder(w).Encode(&entriesJson)
			if err != nil {
				log.Println(fmt.Sprintf("json encode error: %v", err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
		// idの数だけ置換文字を作成
		query, args, err := sqlx.In(query, queryIDs)
		if err != nil {
			log.Println(fmt.Sprintf("db.In error: %v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		// Postgresの場合は置換文字を$1, $2, ...とする必要がある
		query = sqlx.Rebind(len(queryIDs), query)
		// 複数件取得
		err = db.SelectContext(r.Context(), &entriesJson.Entries, query, args...)
		if err != nil {
			log.Println(fmt.Sprintf("select error: %v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		// レスポンスボディに書き込む
		err = json.NewEncoder(w).Encode(&entriesJson)
		if err != nil {
			log.Println(fmt.Sprintf("json encode error: %v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
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
		var entriesJson EntriesJson
		query := `
			UPDATE
				entry
			SET
				source_id = :source_id,
				name = :name,
				image = :image,
				content = :content,
				created_at = :created_at
			WHERE
				id = :id
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
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		}
		if len(entriesJson.Entries) == 0 {
			log.Println("json unexpected error: empty body")
			http.Error(w, "json unexpected error: empty body", http.StatusUnprocessableEntity)
		}
		for _, entry := range entriesJson.Entries {
			// 更新
			_, err = db.NamedExecContext(r.Context(), query, entry)
			if err != nil {
				log.Println(fmt.Sprintf("update error: %v", err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		}
		// レスポンスボディに書き込む
		err = json.NewEncoder(w).Encode(&entriesJson)
		if err != nil {
			log.Println(fmt.Sprintf("json encode error: %v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	case http.MethodDelete:
		var delIDs IDs
		query := `
			DELETE FROM
				entry
			WHERE
				id IN (?)
		`
		// リクエストボディを読み込む
		jsonBytes, err := io.ReadAll(r.Body)
		if err != nil {
			log.Println(fmt.Sprintf("read error: %v", err))
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		err = json.Unmarshal(jsonBytes, &delIDs)
		if err != nil {
			log.Println(fmt.Sprintf("json decode error: %v body:%v", err, r.Body))
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		if len(delIDs.IDs) == 0 {
			return
		// 1件の場合はIN句を使わない
		} else if len(delIDs.IDs) == 1 {
			query = `
				DELETE FROM
					entry
				WHERE
					id = $1
			`
			// 1件削除
			_, err = db.ExecContext(r.Context(), query, delIDs.IDs[0])
			if err != nil {
				log.Println(fmt.Sprintf("delete error: %v", err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			// レスポンスボディに書き込む
			err = json.NewEncoder(w).Encode(&delIDs)
			if err != nil {
				log.Println(fmt.Sprintf("json encode error: %v", err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
		// idの数だけ置換文字を作成
		query, args, err := sqlx.In(query, delIDs.IDs)
		if err != nil {
			log.Println(fmt.Sprintf("db.In error: %v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		// Postgresの場合は置換文字を$1, $2, ...とする必要がある
		query = sqlx.Rebind(len(delIDs.IDs), query)
		_, err = db.ExecContext(r.Context(), query, args...)
		if err != nil {
			log.Println(fmt.Sprintf("delete error: %v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		// レスポンスボディに書き込む
		err = json.NewEncoder(w).Encode(&delIDs)
		if err != nil {
			log.Println(fmt.Sprintf("json encode error: %v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
