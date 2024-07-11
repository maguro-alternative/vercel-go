package link

import (
	"encoding/json"
	"log"
	"fmt"
	"net/http"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	validation "github.com/go-ozzo/ozzo-validation"
)

type Link struct {
	ID       int64  `db:"id"`
	EntryID  int64  `db:"entry_id"`
	Type     string `db:"type"`
	URL      string `db:"url"`
	Nsfw     bool   `db:"nsfw"`
	Darkness bool   `db:"darkness"`
}

func (l *Link) Validate() error {
	return validation.ValidateStruct(l,
		validation.Field(&l.EntryID, validation.Required),
		validation.Field(&l.Type, validation.Required),
		validation.Field(&l.URL, validation.Required),
		validation.Field(&l.Nsfw, validation.In(false, true)),
		validation.Field(&l.Darkness, validation.In(false, true)),
	)
}

type LinksJson struct {
	Links []Link `json:"links"`
}

func (l *LinksJson) Validate() error {
	return validation.ValidateStruct(l,
		validation.Field(&l.Links, validation.Required),
	)
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
		var linksJson LinksJson
		query := `
			SELECT
				entry_id,
				type,
				url,
				nsfw,
				darkness
			FROM
				link
			WHERE
				id IN (?)
		`
		queryIDs, ok := r.URL.Query()["id"]
		if !ok {
			query = `
				SELECT
					entry_id,
					type,
					url,
					nsfw,
					darkness
				FROM
					link
			`
			err := db.SelectContext(r.Context(), &linksJson.Links, query)
			if err != nil {
				log.Printf(fmt.Sprintf("db error: %v", err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			// レスポンスボディをJSONに変換
			err = json.NewEncoder(w).Encode(&linksJson)
			if err != nil {
				log.Printf(fmt.Sprintf("json encode error: %v", err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		} else if len(queryIDs) == 1 {
			query = `
				SELECT
					entry_id,
					type,
					url,
					nsfw,
					darkness
				FROM
					link
				WHERE
					id = $1
			`
			err := db.SelectContext(r.Context(), &linksJson.Links, query, queryIDs[0])
			if err != nil {
				log.Printf(fmt.Sprintf("select error: %v", err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			// レスポンスボディをJSONに変換
			err = json.NewEncoder(w).Encode(&linksJson)
			if err != nil {
				log.Printf(fmt.Sprintf("json encode error: %v", err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
		// idの数だけ置換文字を作成
		query, args, err := sqlx.In(query, queryIDs)
		if err != nil {
			log.Printf(fmt.Sprintf("in query error: %v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		// Postgresの場合は置換文字を$1, $2, ...とする必要がある
		query = sqlx.Rebind(len(queryIDs), query)
		err = db.SelectContext(r.Context(), &linksJson.Links, query, args...)
		if err != nil {
			log.Printf(fmt.Sprintf("select error: %v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		// レスポンスボディをJSONに変換
		err = json.NewEncoder(w).Encode(&linksJson)
		if err != nil {
			log.Printf(fmt.Sprintf("json encode error: %v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	case http.MethodPost:
		var linksJson LinksJson
		query := `
			INSERT INTO link (
				entry_id,
				type,
				url,
				nsfw,
				darkness
			) VALUES (
				:entry_id,
				:type,
				:url,
				:nsfw,
				:darkness
			)
		`
		// リクエストボディをJSONに変換
		err := json.NewDecoder(r.Body).Decode(&linksJson)
		if err != nil {
			log.Printf(fmt.Sprintf("json decode error: %v body:%v", err, r.Body))
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		// JSONのバリデーション
		err = linksJson.Validate()
		if err != nil {
			log.Printf(fmt.Sprintf("validation error: %v", err))
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}
		for _, link := range linksJson.Links {
			// リンクのバリデーション
			err = link.Validate()
			if err != nil {
				log.Printf(fmt.Sprintf("validation error: %v", err))
				http.Error(w, err.Error(), http.StatusUnprocessableEntity)
				return
			}
			_, err = db.NamedExecContext(r.Context(), query, link)
			if err != nil {
				log.Printf(fmt.Sprintf("db error: %v", err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		// レスポンスボディをJSONに変換
		err = json.NewEncoder(w).Encode(&linksJson)
		if err != nil {
			log.Printf(fmt.Sprintf("json encode error: %v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	case http.MethodPut:
		var linksJson LinksJson
		query := `
			UPDATE
				link
			SET
				entry_id = :entry_id,
				type = :type,
				url = :url,
				nsfw = :nsfw,
				darkness = :darkness
			WHERE
				id = :id
		`
		// リクエストボディをJSONに変換
		err := json.NewDecoder(r.Body).Decode(&linksJson)
		if err != nil {
			log.Printf(fmt.Sprintf("json decode error: %v body:%v", err, r.Body))
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		// JSONのバリデーション
		err = linksJson.Validate()
		if err != nil {
			log.Printf(fmt.Sprintf("validation error: %v", err))
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		}
		for _, link := range linksJson.Links {
			// リンクのバリデーション
			err = link.Validate()
			if err != nil {
				log.Printf(fmt.Sprintf("validation error: %v", err))
				http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			}
			_, err = db.NamedExecContext(r.Context(), query, link)
			if err != nil {
				log.Printf(fmt.Sprintf("update error: %v", err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		}
		// レスポンスボディをJSONに変換
		err = json.NewEncoder(w).Encode(&linksJson)
		if err != nil {
			log.Printf(fmt.Sprintf("json encode error: %v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	case http.MethodDelete:
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
