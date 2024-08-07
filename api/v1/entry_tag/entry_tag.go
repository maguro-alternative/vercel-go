package entrytag

import (
	"encoding/json"
	"log"
	"io"
	"fmt"
	"time"
	"net/http"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	validation "github.com/go-ozzo/ozzo-validation"
)

type EntryTag struct {
	ID      int64 `db:"id" json:"id"`
	EntryID int64 `db:"entry_id" json:"entry_id"`
	TagID   int64 `db:"tag_id" json:"tag_id"`
}

func (e *EntryTag) Validate() error {
	return validation.ValidateStruct(e,
		validation.Field(&e.EntryID, validation.Required),
		validation.Field(&e.TagID, validation.Required),
	)
}

type EntryTagsJson struct {
	EntryTags []EntryTag `json:"entry_tags"`
}

func (e *EntryTagsJson) Validate() error {
	return validation.ValidateStruct(e,
		validation.Field(&e.EntryTags, validation.Required),
	)
}

type Source struct {
	ID      int64 `db:"id" json:"id"`
	Name    string `db:"name" json:"name"`
	Url     string `db:"url" json:"url"`
	Type    string `db:"type" json:"type"`
}

type Entry struct {
	ID        int64    `db:"id" json:"id"`
	SourceID  int64     `db:"source_id" json:"source_id"`
	Name      string    `db:"name" json:"name"`
	Image     string    `db:"image" json:"image"`
	Content   string    `db:"content" json:"content"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type Tag struct {
	ID        int64    `db:"id" json:"id"`
	Name      string    `db:"name" json:"name"`
}

type IDs struct {
	IDs []int64 `json:"ids"`
}

func (i *IDs) Validate() error {
	return validation.ValidateStruct(i,
		validation.Field(&i.IDs, validation.Required),
	)
}

func Handler(w http.ResponseWriter, r *http.Request) {
	db, err := sqlx.Open("postgres", "")
	if err != nil {
		log.Printf("sql.Open error %s", err)
	}
	defer db.Close()
	switch r.Method {
	case http.MethodGet:
		var entryTagsJson EntryTagsJson
		query := `
			SELECT
				id,
				entry_id,
				tag_id
			FROM
				entry_tag
			WHERE
				id IN (?)
		`
		// idが指定されていない場合は全件取得
		queryIDs, ok := r.URL.Query()["id"]
		if !ok {
			query = `
				SELECT
					id,
					entry_id,
					tag_id
				FROM
					entry_tag
			`
			err := db.SelectContext(r.Context(), &entryTagsJson.EntryTags, query)
			if err != nil {
				log.Println(fmt.Sprintf("select error: %v", err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			// json返却
			err = json.NewEncoder(w).Encode(&entryTagsJson)
			if err != nil {
				log.Println(fmt.Sprintf("json encode error: %v", err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		// idが指定されている場合は指定されたidのみ取得
		} else if len(queryIDs) == 1 {
			query = `
				SELECT
					id,
					entry_id,
					tag_id
				FROM
					entry_tag
				WHERE
					id = $1
			`
			err := db.SelectContext(r.Context(), &entryTagsJson.EntryTags, query, queryIDs[0])
			if err != nil {
				log.Println(fmt.Sprintf("select error: %v", err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			// json返却
			err = json.NewEncoder(w).Encode(&entryTagsJson)
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
		err = db.SelectContext(r.Context(), &entryTagsJson.EntryTags, query, args...)
		if err != nil {
			log.Println(fmt.Sprintf("select error: %v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		// json返却
		err = json.NewEncoder(w).Encode(&entryTagsJson)
		if err != nil {
			log.Println(fmt.Sprintf("json encode error: %v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	case http.MethodPost:
		var entryTagsJson EntryTagsJson
		query := `
			INSERT INTO entry_tag (
				entry_id,
				tag_id
			) VALUES (
				:entry_id,
				:tag_id
			)
		`
		// json読み込み
		jsonBytes, err := io.ReadAll(r.Body)
		if err != nil {
			log.Println(fmt.Sprintf("read error: %v", err))
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		err = json.Unmarshal(jsonBytes, &entryTagsJson)
		if err != nil {
			log.Println(fmt.Sprintf("json unmarshal error: %v", err))
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		}
		// jsonバリデーション
		err = entryTagsJson.Validate()
		if err != nil {
			log.Println(fmt.Sprintf("validation error: %v", err))
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		}
		for _, entryTag := range entryTagsJson.EntryTags {
			// jsonバリデーション
			err = entryTag.Validate()
			if err != nil {
				log.Println(fmt.Sprintf("validation error: %v", err))
				http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			}
			_, err = db.NamedExecContext(r.Context(), query, entryTag)
			if err != nil {
				log.Println(fmt.Sprintf("insert error: %v", err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		}
		// json返却
		err = json.NewEncoder(w).Encode(&entryTagsJson)
		if err != nil {
			log.Println(fmt.Sprintf("json encode error: %v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	case http.MethodPut:
		var entryTagsJson EntryTagsJson
		query := `
			UPDATE
				entry_tag
			SET
				entry_id = :entry_id,
				tag_id = :tag_id
			WHERE
				id = :id
		`
		jsonBytes, err := io.ReadAll(r.Body)
		if err != nil {
			log.Println(fmt.Sprintf("read error: %v", err))
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		err = json.Unmarshal(jsonBytes, &entryTagsJson)
		if err != nil {
			log.Println(fmt.Sprintf("json unmarshal error: %v", err))
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		}
		// jsonバリデーション
		err = entryTagsJson.Validate()
		if err != nil {
			log.Println(fmt.Sprintf("validation error: %v", err))
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		}
		for _, entryTag := range entryTagsJson.EntryTags {
			// jsonバリデーション
			err = entryTag.Validate()
			if err != nil {
				log.Println(fmt.Sprintf("validation error: %v", err))
				http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			}
			_, err := db.NamedExecContext(r.Context(), query, entryTag)
			if err != nil {
				log.Println(fmt.Sprintf("update error: %v", err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		}
		// json返却
		err = json.NewEncoder(w).Encode(&entryTagsJson)
		if err != nil {
			log.Println(fmt.Sprintf("json encode error: %v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	case http.MethodDelete:
		var delIDs IDs
		query := `
			DELETE FROM
				entry_tag
			WHERE
				id IN (?)
		`
		jsonBytes, err := io.ReadAll(r.Body)
		if err != nil {
			log.Println(fmt.Sprintf("read error: %v", err))
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		err = json.Unmarshal(jsonBytes, &delIDs)
		if err != nil {
			log.Println(fmt.Sprintf("json unmarshal error: %v", err))
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		}
		// jsonバリデーション
		err = delIDs.Validate()
		if err != nil {
			log.Println(fmt.Sprintf("validation error: %v", err))
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		}
		if len(delIDs.IDs) == 0 {
			return
		} else if len(delIDs.IDs) == 1 {
			query = `
				DELETE FROM
					entry_tag
				WHERE
					id = $1
			`
			_, err = db.ExecContext(r.Context(), query, delIDs.IDs[0])
			if err != nil {
				log.Println(fmt.Sprintf("delete error: %v", err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			// json返却
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
		query = sqlx.Rebind(len(delIDs.IDs),query)
		_, err = db.ExecContext(r.Context(), query, args...)
		if err != nil {
			log.Println(fmt.Sprintf("delete error: %v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		// json返却
		err = json.NewEncoder(w).Encode(&delIDs)
		if err != nil {
			log.Println(fmt.Sprintf("json encode error: %v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
