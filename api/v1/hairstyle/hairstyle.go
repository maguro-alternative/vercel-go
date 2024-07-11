package hairstyle

import (
	"encoding/json"
	"log"
	"fmt"
	"net/http"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	validation "github.com/go-ozzo/ozzo-validation"
)

type HairStyle struct {
	EntryID int64 `db:"entry_id"`
	StyleID int64 `db:"style_id"`
}

func (h *HairStyle) Validate() error {
	return validation.ValidateStruct(h,
		validation.Field(&h.EntryID, validation.Required),
		validation.Field(&h.StyleID, validation.Required),
	)
}

type HairStylesJson struct {
	HairStyles []HairStyle `json:"hair_styles"`
}

func (h *HairStylesJson) Validate() error {
	return validation.ValidateStruct(h,
		validation.Field(&h.HairStyles, validation.Required),
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
		var hairStylesJson HairStylesJson
		query := `
			SELECT
				entry_id,
				style_id
			FROM
				hairstyle
			WHERE
				entry_id IN (?)
		`
		queryIDs, ok := r.URL.Query()["entry_id"]
		if !ok {
			query = `
				SELECT
					entry_id,
					style_id
				FROM
					hairstyle
			`
			err := db.SelectContext(r.Context(), &hairStylesJson.HairStyles, query)
			if err != nil {
				log.Printf(fmt.Sprintf("db error: %v", err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			// json書き込み
			err = json.NewEncoder(w).Encode(&hairStylesJson)
			if err != nil {
				log.Printf(fmt.Sprintf("json encode error: %v", err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		} else if len(queryIDs) == 1 {
			query = `
				SELECT
					entry_id,
					style_id
				FROM
					hairstyle
				WHERE
					entry_id = $1
			`
			err := db.SelectContext(r.Context(), &hairStylesJson.HairStyles, query, queryIDs[0])
			if err != nil {
				log.Printf(fmt.Sprintf("db error: %v", err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			// json書き込み
			err = json.NewEncoder(w).Encode(&hairStylesJson)
			if err != nil {
				log.Printf(fmt.Sprintf("json encode error: %v", err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
		// idの数だけ置換文字を作成
		query, args, err := sqlx.In(query, queryIDs)
		if err != nil {
			log.Printf(fmt.Sprintf("db error: %v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		// Postgresの場合は置換文字を$1, $2, ...とする必要がある
		query = sqlx.Rebind(len(queryIDs), query)
		err = db.SelectContext(r.Context(), &hairStylesJson.HairStyles, query, args...)
		if err != nil {
			log.Printf(fmt.Sprintf("db error: %v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		// json書き込み
		err = json.NewEncoder(w).Encode(&hairStylesJson)
		if err != nil {
			log.Printf(fmt.Sprintf("json encode error: %v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	case http.MethodPost:
		var hairStylesJson HairStylesJson
		query := `
			INSERT INTO hairstyle (
				entry_id,
				style_id
			) VALUES (
				:entry_id,
				:style_id
			)
		`
		// json読み込み
		if err := json.NewDecoder(r.Body).Decode(&hairStylesJson); err != nil {
			log.Printf(fmt.Sprintf("json decode error: %v body:%v", err, r.Body))
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		// jsonバリデーション
		err := hairStylesJson.Validate()
		if err != nil {
			log.Printf(fmt.Sprintf("json validate error: %v", err))
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		}
		for _, hs := range hairStylesJson.HairStyles {
			// jsonバリデーション
			err = hs.Validate()
			if err != nil {
				log.Printf(fmt.Sprintf("json validate error: %v", err))
				http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			}
			if _, err := db.NamedExecContext(r.Context(), query, hs); err != nil {
				log.Printf(fmt.Sprintf("db error: %v", err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		}
		// json書き込み
		err = json.NewEncoder(w).Encode(&hairStylesJson)
		if err != nil {
			log.Printf(fmt.Sprintf("json encode error: %v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	case http.MethodPut:
	case http.MethodDelete:
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
