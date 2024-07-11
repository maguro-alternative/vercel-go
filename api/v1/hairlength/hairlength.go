package hairlength

import (
	"encoding/json"
	"log"
	"fmt"
	"net/http"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	validation "github.com/go-ozzo/ozzo-validation"
)

type HairLength struct {
	EntryID          int64 `db:"entry_id"`
	HairLengthTypeID int64 `db:"hairlength_type_id"`
}

func (h *HairLength) Validate() error {
	return validation.ValidateStruct(h,
		validation.Field(&h.EntryID, validation.Required),
		validation.Field(&h.HairLengthTypeID, validation.Required),
	)
}

type HairLengthsJson struct {
	HairLengths []HairLength `json:"hairlengths"`
}

func (h *HairLengthsJson) Validate() error {
	return validation.ValidateStruct(h,
		validation.Field(&h.HairLengths, validation.Required),
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
		var hairLengthsJson HairLengthsJson
		query := `
			SELECT
				entry_id,
				hairlength_type_id
			FROM
				hairlength
			WHERE
				entry_id IN (?)
		`
		queryIDs, ok := r.URL.Query()["entry_id"]
		if !ok {
			query = `
				SELECT
					entry_id,
					hairlength_type_id
				FROM
					hairlength
			`
			err := db.SelectContext(r.Context(), &hairLengthsJson.HairLengths, query)
			if err != nil {
				log.Printf(fmt.Sprintf("db error: %v", err))
				http.Error(w, "db error", http.StatusInternalServerError)
			}
			// jsonを返す
			err = json.NewEncoder(w).Encode(&hairLengthsJson)
			if err != nil {
				log.Printf(fmt.Sprintf("json encode error: %v", err))
				http.Error(w, "json encode error", http.StatusInternalServerError)
			}
			return
		} else if len(queryIDs) == 1 {
			query = `
				SELECT
					entry_id,
					hairlength_type_id
				FROM
					hairlength
				WHERE
					entry_id = $1
			`
			err := db.SelectContext(r.Context(), &hairLengthsJson.HairLengths, query, queryIDs[0])
			if err != nil {
				log.Printf(fmt.Sprintf("db error: %v", err))
				http.Error(w, "db error", http.StatusInternalServerError)
			}
			// jsonを返す
			err = json.NewEncoder(w).Encode(&hairLengthsJson)
			if err != nil {
				log.Printf(fmt.Sprintf("json encode error: %v", err))
				http.Error(w, "json encode error", http.StatusInternalServerError)
			}
			return
		}
		// idの数だけ置換文字を作成
		query, args, err := sqlx.In(query, queryIDs)
		if err != nil {
			log.Printf(fmt.Sprintf("db error: %v", err))
			http.Error(w, "db error", http.StatusInternalServerError)
		}
		// Postgresの場合は置換文字を$1, $2, ...とする必要がある
		query = sqlx.Rebind(len(queryIDs), query)
		err = db.SelectContext(r.Context(), &hairLengthsJson.HairLengths, query, args...)
		if err != nil {
			log.Printf(fmt.Sprintf("db error: %v", err))
			http.Error(w, "db error", http.StatusInternalServerError)
		}
		// jsonを返す
		err = json.NewEncoder(w).Encode(&hairLengthsJson)
		if err != nil {
			log.Printf(fmt.Sprintf("json encode error: %v", err))
			http.Error(w, "json encode error", http.StatusInternalServerError)
		}
	case http.MethodPost:
		var hairLengthsJson HairLengthsJson
		query := `
			INSERT INTO hairlength (
				entry_id,
				hairlength_type_id
			) VALUES (
				:entry_id,
				:hairlength_type_id
			)
		`
		if err := json.NewDecoder(r.Body).Decode(&hairLengthsJson); err != nil {
			log.Printf(fmt.Sprintf("json decode error: %v body:%v", err, r.Body))
			http.Error(w, "json decode error", http.StatusBadRequest)
		}
		// jsonバリデーション
		err := hairLengthsJson.Validate()
		if err != nil {
			log.Printf(fmt.Sprintf("validation error: %v", err))
			http.Error(w, "validation error", http.StatusUnprocessableEntity)
		}
		for _, hl := range hairLengthsJson.HairLengths {
			// jsonのバリデーションを通過したデータをDBに登録
			err = hl.Validate()
			if err != nil {
				log.Printf(fmt.Sprintf("validation error: %v", err))
				http.Error(w, "validation error", http.StatusUnprocessableEntity)
			}
			if _, err := db.NamedExecContext(r.Context(), query, hl); err != nil {
				log.Printf(fmt.Sprintf("db error: %v", err))
				http.Error(w, "db error", http.StatusInternalServerError)
			}
		}
		// jsonを返す
		err = json.NewEncoder(w).Encode(&hairLengthsJson)
		if err != nil {
			log.Printf(fmt.Sprintf("json encode error: %v", err))
			http.Error(w, "json encode error", http.StatusInternalServerError)
		}
	case http.MethodPut:
		var hairLengthsJson HairLengthsJson
		query := `
			UPDATE
				hairlength
			SET
				hairlength_type_id = :hairlength_type_id
			WHERE
				entry_id = :entry_id
		`
		// jsonをデコード
		if err := json.NewDecoder(r.Body).Decode(&hairLengthsJson); err != nil {
			log.Printf(fmt.Sprintf("json decode error: %v body:%v", err, r.Body))
			http.Error(w, "json decode error", http.StatusBadRequest)
		}
		// jsonバリデーション
		err := hairLengthsJson.Validate()
		if err != nil {
			log.Printf(fmt.Sprintf("validation error: %v", err))
			http.Error(w, "validation error", http.StatusUnprocessableEntity)
		}
		for _, hl := range hairLengthsJson.HairLengths {
			// jsonのバリデーションを通過したデータをDBに登録
			err = hl.Validate()
			if err != nil {
				log.Printf(fmt.Sprintf("validation error: %v", err))
				http.Error(w, "validation error", http.StatusUnprocessableEntity)
			}
			if _, err := db.NamedExecContext(r.Context(), query, hl); err != nil {
				log.Printf(fmt.Sprintf("db error: %v", err))
				http.Error(w, "db error", http.StatusInternalServerError)
			}
		}
		// jsonを返す
		err = json.NewEncoder(w).Encode(&hairLengthsJson)
		if err != nil {
			log.Printf(fmt.Sprintf("json encode error: %v", err))
			http.Error(w, "json encode error", http.StatusInternalServerError)
		}
	case http.MethodDelete:
		var delIDs IDs
		query := `
			DELETE FROM
				hairlength
			WHERE
				entry_id IN (?)
		`
		// jsonをデコード
		if err := json.NewDecoder(r.Body).Decode(&delIDs); err != nil {
			log.Printf(fmt.Sprintf("json decode error: %v body:%v", err, r.Body))
			http.Error(w, "json decode error", http.StatusBadRequest)
		}
		// jsonバリデーション
		err := delIDs.Validate()
		if err != nil {
			log.Printf(fmt.Sprintf("validation error: %v", err))
			http.Error(w, "validation error", http.StatusUnprocessableEntity)
		}
		if len(delIDs.IDs) == 0 {
			return
		} else if len(delIDs.IDs) == 1 {
			query = `
				DELETE FROM
					hairlength
				WHERE
					entry_id = $1
			`
			_, err := db.ExecContext(r.Context(), query, delIDs.IDs[0])
			if err != nil {
				log.Printf(fmt.Sprintf("db error: %v", err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			err = json.NewEncoder(w).Encode(&delIDs)
			if err != nil {
				log.Printf(fmt.Sprintf("json encode error: %v", err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
		// idの数だけ置換文字を作成
		query, args, err := sqlx.In(query, delIDs.IDs)
		if err != nil {
			log.Printf(fmt.Sprintf("db error: %v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		// Postgresの場合は置換文字を$1, $2, ...とする必要がある
		query = sqlx.Rebind(len(delIDs.IDs), query)
		_, err = db.ExecContext(r.Context(), query, args...)
		if err != nil {
			log.Printf(fmt.Sprintf("db error: %v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		// レスポンスの作成
		err = json.NewEncoder(w).Encode(&delIDs)
		if err != nil {
			log.Printf(fmt.Sprintf("json encode error: %v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}