package hairstyletype

import (
	"encoding/json"
	"log"
	"fmt"
	"net/http"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	validation "github.com/go-ozzo/ozzo-validation"
)

type HairStyleType struct {
	ID    int64  `db:"id"`
	Style string `db:"style"`
}

func (h *HairStyleType) Validate() error {
	return validation.ValidateStruct(h,
		validation.Field(&h.Style, validation.Required),
	)
}

type HairStyleTypesJson struct {
	HairStyleTypes []HairStyleType `json:"hairstyle_types"`
}

func (h *HairStyleTypesJson) Validate() error {
	return validation.ValidateStruct(h,
		validation.Field(&h.HairStyleTypes, validation.Required),
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
		var hairStyleTypesJson HairStyleTypesJson
		query := `
			SELECT
				id,
				style
			FROM
				hairstyle_type
			WHERE
				id IN (?)
		`
		queryIDs, ok := r.URL.Query()["id"]
		if !ok {
			query = `
				SELECT
					id,
					style
				FROM
					hairstyle_type
			`
			err := db.SelectContext(r.Context(), &hairStyleTypesJson.HairStyleTypes, query)
			if err != nil {
				log.Printf(fmt.Sprintf("db error: %v", err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			// レスポンスの作成
			err = json.NewEncoder(w).Encode(&hairStyleTypesJson)
			if err != nil {
				log.Printf(fmt.Sprintf("json encode error: %v", err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			return
		} else if len(queryIDs) == 1 {
			query = `
				SELECT
					id,
					style
				FROM
					hairstyle_type
				WHERE
					id = $1
			`
			err := db.SelectContext(r.Context(), &hairStyleTypesJson.HairStyleTypes, query, queryIDs[0])
			if err != nil {
				log.Printf(fmt.Sprintf("db error: %v", err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			// レスポンスの作成
			err = json.NewEncoder(w).Encode(&hairStyleTypesJson)
			if err != nil {
				log.Printf(fmt.Sprintf("json encode error: %v", err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			return
		}
		// idの数だけ置換文字を作成
		query, args, err := sqlx.In(query, queryIDs)
		if err != nil {
			log.Printf(fmt.Sprintf("db error: %v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// Postgresの場合は置換文字を$1, $2, ...とする必要がある
		query = sqlx.Rebind(len(queryIDs), query)
		err = db.SelectContext(r.Context(), &hairStyleTypesJson.HairStyleTypes, query, args...)
		if err != nil {
			log.Printf(fmt.Sprintf("db error: %v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// レスポンスの作成
		err = json.NewEncoder(w).Encode(&hairStyleTypesJson)
		if err != nil {
			log.Printf(fmt.Sprintf("json encode error: %v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	case http.MethodPost:
		var hairStyleTypesJson HairStyleTypesJson
		query := `
			INSERT INTO hairstyle_type (
				style
			) VALUES (
				:style
			)
		`
		err := json.NewDecoder(r.Body).Decode(&hairStyleTypesJson)
		if err != nil {
			log.Printf(fmt.Sprintf("json decode error: %v body:%v", err, r.Body))
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		// リクエストボディのバリデーション
		err = hairStyleTypesJson.Validate()
		if err != nil {
			log.Printf(fmt.Sprintf("json validate error: %v", err))
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}
		for _, hairStyleType := range hairStyleTypesJson.HairStyleTypes {
			// リクエストボディのバリデーション
			err = hairStyleType.Validate()
			if err != nil {
				log.Printf(fmt.Sprintf("json validate error: %v", err))
				http.Error(w, err.Error(), http.StatusUnprocessableEntity)
				return
			}
			_, err = db.NamedExecContext(r.Context(), query, hairStyleType)
			if err != nil {
				log.Printf(fmt.Sprintf("db error: %v", err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		// レスポンスの作成
		err = json.NewEncoder(w).Encode(&hairStyleTypesJson)
		if err != nil {
			log.Printf(fmt.Sprintf("json encode error: %v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	case http.MethodPut:
		var hairStyleTypesJson HairStyleTypesJson
		query := `
			UPDATE
				hairstyle_type
			SET
				style = :style
			WHERE
				id = :id
		`
		err := json.NewDecoder(r.Body).Decode(&hairStyleTypesJson)
		if err != nil {
			log.Printf(fmt.Sprintf("json decode error: %v body:%v", err, r.Body))
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		// リクエストボディのバリデーション
		err = hairStyleTypesJson.Validate()
		if err != nil {
			log.Printf(fmt.Sprintf("json validate error: %v", err))
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}
		for _, hairStyleType := range hairStyleTypesJson.HairStyleTypes {
			// リクエストボディのバリデーション
			err = hairStyleType.Validate()
			if err != nil {
				log.Printf(fmt.Sprintf("json validate error: %v", err))
				http.Error(w, err.Error(), http.StatusUnprocessableEntity)
				return
			}
			_, err = db.NamedExecContext(r.Context(), query, hairStyleType)
			if err != nil {
				log.Printf(fmt.Sprintf("db error: %v", err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		// レスポンスの作成
		err = json.NewEncoder(w).Encode(&hairStyleTypesJson)
		if err != nil {
			log.Printf(fmt.Sprintf("json encode error: %v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	case http.MethodDelete:
		var delIDs IDs
		query := `
			DELETE FROM
				hairstyle_type
			WHERE
				id IN (?)
		`
		err := json.NewDecoder(r.Body).Decode(&delIDs)
		if err != nil {
			log.Printf(fmt.Sprintf("json decode error: %v body:%v", err, r.Body))
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		// リクエストボディのバリデーション
		err = delIDs.Validate()
		if err != nil {
			log.Printf(fmt.Sprintf("json validate error: %v", err))
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}
		if len(delIDs.IDs) == 0 {
			return
		} else if len(delIDs.IDs) == 1 {
			query = `
				DELETE FROM
					hairstyle_type
				WHERE
					id = $1
			`
			_, err = db.ExecContext(r.Context(), query, delIDs.IDs[0])
			if err != nil {
				log.Printf(fmt.Sprintf("db error: %v", err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			return
		}
		// idの数だけ置換文字を作成
		query, args, err := sqlx.In(query, delIDs.IDs)
		if err != nil {
			log.Printf(fmt.Sprintf("db error: %v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// Postgresの場合は置換文字を$1, $2, ...とする必要がある
		query = sqlx.Rebind(len(delIDs.IDs), query)
		_, err = db.ExecContext(r.Context(), query, args...)
		if err != nil {
			log.Printf(fmt.Sprintf("db error: %v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// レスポンスの作成
		err = json.NewEncoder(w).Encode(&delIDs)
		if err != nil {
			log.Printf(fmt.Sprintf("json encode error: %v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
