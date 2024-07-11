package eyescolortype

import (
	"encoding/json"
	"log"
	"io"
	"fmt"
	"net/http"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	validation "github.com/go-ozzo/ozzo-validation"
)

type EyeColorType struct {
	ID    int64  `db:"id"`
	Color string `db:"color"`
}

func (e *EyeColorType) Validate() error {
	return validation.ValidateStruct(e,
		validation.Field(&e.Color, validation.Required),
	)
}

type EyeColorTypesJson struct {
	EyeColorTypes []EyeColorType `json:"eyecolor_types"`
}

func (e *EyeColorTypesJson) Validate() error {
	return validation.ValidateStruct(e,
		validation.Field(&e.EyeColorTypes, validation.Required),
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
		var eyeColorTypesJson EyeColorTypesJson
		query := `
			SELECT
				id,
				color
			FROM
				eyecolor_type
			WHERE
				id IN (?)
		`
		queryIDs, ok := r.URL.Query()["id"]
		if !ok {
			query = `
				SELECT
					id,
					color
				FROM
					eyecolor_type
			`
			err := db.SelectContext(r.Context(), &eyeColorTypesJson.EyeColorTypes, query)
			if err != nil {
				log.Printf(fmt.Sprintf("db error: %v", err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			// jsonを返す
			err = json.NewEncoder(w).Encode(&eyeColorTypesJson)
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
					color
				FROM
					eyecolor_type
				WHERE
					id = $1
			`
			err := db.SelectContext(r.Context(), &eyeColorTypesJson.EyeColorTypes, query, queryIDs[0])
			if err != nil {
				log.Printf(fmt.Sprintf("db error: %v", err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			// jsonを返す
			err = json.NewEncoder(w).Encode(&eyeColorTypesJson)
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
		err = db.SelectContext(r.Context(), &eyeColorTypesJson.EyeColorTypes, query, args...)
		if err != nil {
			log.Printf(fmt.Sprintf("db error: %v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// jsonを返す
		err = json.NewEncoder(w).Encode(&eyeColorTypesJson)
		if err != nil {
			log.Printf(fmt.Sprintf("json encode error: %v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	case http.MethodPost:
		var eyeColorTypesJson EyeColorTypesJson
		query := `
			INSERT INTO eyecolor_type (
				color
			) VALUES (
				:color
			)
		`
		jsonBytes, err := io.ReadAll(r.Body)
		if err != nil {
			log.Println(fmt.Sprintf("read error: %v", err))
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		err = json.Unmarshal(jsonBytes, &eyeColorTypesJson)
		if err != nil {
			log.Printf(fmt.Sprintf("json decode error: %v body:%v", err, r.Body))
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}
		// jsonバリデーション
		err = eyeColorTypesJson.Validate()
		if err != nil {
			log.Printf(fmt.Sprintf("validation error: %v", err))
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}
		for _, eyeColorType := range eyeColorTypesJson.EyeColorTypes {
			// jsonバリデーション
			err = eyeColorType.Validate()
			if err != nil {
				log.Printf(fmt.Sprintf("validation error: %v", err))
				http.Error(w, err.Error(), http.StatusUnprocessableEntity)
				return
			}
			_, err = db.NamedExecContext(r.Context(), query, eyeColorType)
			if err != nil {
				log.Printf(fmt.Sprintf("db error: %v", err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		// jsonを返す
		err = json.NewEncoder(w).Encode(&eyeColorTypesJson)
		if err != nil {
			log.Printf(fmt.Sprintf("json encode error: %v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	case http.MethodPut:
		var eyeColorTypesJson EyeColorTypesJson
		query := `
			UPDATE
				eyecolor_type
			SET
				color = :color
			WHERE
				id = :id
		`
		jsonBytes, err := io.ReadAll(r.Body)
		if err != nil {
			log.Println(fmt.Sprintf("read error: %v", err))
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		err = json.Unmarshal(jsonBytes, &eyeColorTypesJson)
		if err != nil {
			log.Printf(fmt.Sprintf("json decode error: %v body:%v", err, r.Body))
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}
		// jsonバリデーション
		err = eyeColorTypesJson.Validate()
		if err != nil {
			log.Printf(fmt.Sprintf("validation error: %v", err))
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}
		for _, eyeColorType := range eyeColorTypesJson.EyeColorTypes {
			err = eyeColorType.Validate()
			if err != nil {
				log.Printf(fmt.Sprintf("validation error: %v", err))
				http.Error(w, err.Error(), http.StatusUnprocessableEntity)
				return
			}
			_, err = db.NamedExecContext(r.Context(), query, eyeColorType)
			if err != nil {
				log.Printf(fmt.Sprintf("db error: %v", err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		// jsonを返す
		err = json.NewEncoder(w).Encode(&eyeColorTypesJson)
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
