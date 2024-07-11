package haircolortype

import (
	"encoding/json"
	"log"
	"fmt"
	"net/http"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	validation "github.com/go-ozzo/ozzo-validation"
)

type HairColorType struct {
	ID    int64 `db:"id"`
	Color string `db:"color"`
}

func (h *HairColorType) Validate() error {
	return validation.ValidateStruct(h,
		validation.Field(&h.Color, validation.Required),
	)
}

type HairColorTypesJson struct {
	HairColorTypes []HairColorType `json:"haircolor_types"`
}

func (h *HairColorTypesJson) Validate() error {
	return validation.ValidateStruct(h,
		validation.Field(&h.HairColorTypes, validation.Required),
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
		var hairColorTypesJson HairColorTypesJson
		query := `
			SELECT
				id,
				color
			FROM
				haircolor_type
			WHERE
				id IN (?)
		`
		queryIDs, ok := r.URL.Query()["id"]
		// idが指定されていない場合は全件取得
		if !ok {
			query = `
				SELECT
					id,
					color
				FROM
					haircolor_type
			`
			err := db.SelectContext(r.Context(), &hairColorTypesJson.HairColorTypes, query)
			if err != nil {
				log.Printf(fmt.Sprintf("db error: %v", err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			// レスポンスの作成
			err = json.NewEncoder(w).Encode(&hairColorTypesJson)
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
					haircolor_type
				WHERE
					id = $1
			`
			err := db.SelectContext(r.Context(), &hairColorTypesJson.HairColorTypes, query, queryIDs[0])
			if err != nil {
				log.Printf(fmt.Sprintf("db error: %v", err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			// レスポンスの作成
			err = json.NewEncoder(w).Encode(&hairColorTypesJson)
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
		err = db.SelectContext(r.Context(), &hairColorTypesJson.HairColorTypes, query, args...)
		if err != nil {
			log.Printf(fmt.Sprintf("db error: %v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// レスポンスの作成
		err = json.NewEncoder(w).Encode(&hairColorTypesJson)
		if err != nil {
			log.Printf(fmt.Sprintf("json encode error: %v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	case http.MethodPost:
		var hairColorTypesJson HairColorTypesJson
		query := `
			INSERT INTO haircolor_type (
				color
			) VALUES (
				:color
			)
		`
		err := json.NewDecoder(r.Body).Decode(&hairColorTypesJson)
		if err != nil {
			log.Printf(fmt.Sprintf("json decode error: %v body:%v", err, r.Body))
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		// jsonバリデーション
		err = hairColorTypesJson.Validate()
		if err != nil {
			log.Printf(fmt.Sprintf("validation error: %v", err))
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}
		for _, hairColorType := range hairColorTypesJson.HairColorTypes {
			// jsonのバリデーションを通過したデータをDBに登録
			err = hairColorType.Validate()
			if err != nil {
				log.Printf(fmt.Sprintf("validation error: %v", err))
				http.Error(w, err.Error(), http.StatusUnprocessableEntity)
				return
			}
			_, err = db.NamedExecContext(r.Context(), query, hairColorType)
			if err != nil {
				log.Printf(fmt.Sprintf("db error: %v", err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		// レスポンスの作成
		err = json.NewEncoder(w).Encode(&hairColorTypesJson)
		if err != nil {
			log.Printf(fmt.Sprintf("json encode error: %v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	case http.MethodPut:
		var hairColorTypesJson HairColorTypesJson
		query := `
			UPDATE
				haircolor_type
			SET
				color = :color
			WHERE
				id = :id
		`
		// jsonのデコード
		err := json.NewDecoder(r.Body).Decode(&hairColorTypesJson)
		if err != nil {
			log.Printf(fmt.Sprintf("json decode error: %v body:%v", err, r.Body))
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		// jsonバリデーション
		err = hairColorTypesJson.Validate()
		if err != nil {
			log.Printf(fmt.Sprintf("validation error: %v", err))
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}
		for _, hairColorType := range hairColorTypesJson.HairColorTypes {
			// jsonのバリデーションを通過したデータをDBに登録
			err = hairColorType.Validate()
			if err != nil {
				log.Printf(fmt.Sprintf("validation error: %v", err))
				http.Error(w, err.Error(), http.StatusUnprocessableEntity)
				return
			}
			_, err = db.NamedExecContext(r.Context(), query, hairColorType)
			if err != nil {
				log.Printf(fmt.Sprintf("db error: %v", err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		err = json.NewEncoder(w).Encode(&hairColorTypesJson)
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
