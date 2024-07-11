package hairlengthtype

import (
	"encoding/json"
	"log"
	"fmt"
	"net/http"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	validation "github.com/go-ozzo/ozzo-validation"
)

type HairLengthType struct {
	ID     int64 `db:"id"`
	Length string `db:"length"`
}

func (h *HairLengthType) Validate() error {
	return validation.ValidateStruct(h,
		validation.Field(&h.Length, validation.Required),
	)
}

type HairLengthTypesJson struct {
	HairLengthTypes []HairLengthType `json:"hairlength_types"`
}

func (h *HairLengthTypesJson) Validate() error {
	return validation.ValidateStruct(h,
		validation.Field(&h.HairLengthTypes, validation.Required),
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
		var hairLengthTypesJson HairLengthTypesJson
		query := `
			SELECT
				id,
				length
			FROM
				hairlength_type
			WHERE
				id IN (?)
		`
		queryIDs, ok := r.URL.Query()["id"]
		if !ok {
			query = `
				SELECT
					id,
					length
				FROM
					hairlength_type
			`
			err := db.SelectContext(r.Context(), &hairLengthTypesJson.HairLengthTypes, query)
			if err != nil {
				log.Printf(fmt.Sprintf("db error: %v", err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			// json書き込み
			err = json.NewEncoder(w).Encode(&hairLengthTypesJson)
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
					length
				FROM
					hairlength_type
				WHERE
					id = $1
			`
			err := db.SelectContext(r.Context(), &hairLengthTypesJson.HairLengthTypes, query, queryIDs[0])
			if err != nil {
				log.Printf(fmt.Sprintf("db error: %v", err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			// json書き込み
			err = json.NewEncoder(w).Encode(&hairLengthTypesJson)
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
		err = db.SelectContext(r.Context(), &hairLengthTypesJson.HairLengthTypes, query, args...)
		if err != nil {
			log.Printf(fmt.Sprintf("db error: %v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// json書き込み
		err = json.NewEncoder(w).Encode(&hairLengthTypesJson)
		if err != nil {
			log.Printf(fmt.Sprintf("json encode error: %v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	case http.MethodPost:
		var hairLengthTypesJson HairLengthTypesJson
		query := `
			INSERT INTO hairlength_type (
				length
			) VALUES (
				:length
			)
		`
		// json読み込み
		err := json.NewDecoder(r.Body).Decode(&hairLengthTypesJson)
		if err != nil {
			log.Printf(fmt.Sprintf("json decode error: %v body:%v", err, r.Body))
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		// jsonバリデーション
		err = hairLengthTypesJson.Validate()
		if err != nil {
			log.Printf(fmt.Sprintf("json validate error: %v", err))
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}
		for _, hairLengthType := range hairLengthTypesJson.HairLengthTypes {
			// jsonバリデーション
			err = hairLengthType.Validate()
			if err != nil {
				log.Printf(fmt.Sprintf("json validate error: %v", err))
				http.Error(w, err.Error(), http.StatusUnprocessableEntity)
				return
			}
			_, err = db.NamedExecContext(r.Context(), query, hairLengthType)
			if err != nil {
				log.Printf(fmt.Sprintf("db error: %v", err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		// json書き込み
		err = json.NewEncoder(w).Encode(&hairLengthTypesJson)
		if err != nil {
			log.Printf(fmt.Sprintf("json encode error: %v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	case http.MethodPut:
		var hairLengthTypesJson HairLengthTypesJson
		query := `
			UPDATE
				hairlength_type
			SET
				length = :length
			WHERE
				id = :id
		`
		// json読み込み
		err := json.NewDecoder(r.Body).Decode(&hairLengthTypesJson)
		if err != nil {
			log.Printf(fmt.Sprintf("json decode error: %v body:%v", err, r.Body))
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		// jsonバリデーション
		err = hairLengthTypesJson.Validate()
		if err != nil {
			log.Printf(fmt.Sprintf("json validate error: %v", err))
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}
		for _, hairLengthType := range hairLengthTypesJson.HairLengthTypes {
			// jsonバリデーション
			err = hairLengthType.Validate()
			if err != nil {
				log.Printf(fmt.Sprintf("json validate error: %v", err))
				http.Error(w, err.Error(), http.StatusUnprocessableEntity)
				return
			}
			_, err = db.NamedExecContext(r.Context(), query, hairLengthType)
			if err != nil {
				log.Printf(fmt.Sprintf("db error: %v", err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		// json書き込み
		err = json.NewEncoder(w).Encode(&hairLengthTypesJson)
		if err != nil {
			log.Printf(fmt.Sprintf("json encode error: %v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	case http.MethodDelete:
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

