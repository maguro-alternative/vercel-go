package eyescolor

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

type EyeColor struct {
	EntryID int64 `db:"entry_id"`
	ColorID int64 `db:"color_id"`
}

func (e *EyeColor) Validate() error {
	return validation.ValidateStruct(e,
		validation.Field(&e.EntryID, validation.Required),
		validation.Field(&e.ColorID, validation.Required),
	)
}

type EyeColorsJson struct {
	EyeColors []EyeColor `json:"eyecolors"`
}

func (e *EyeColorsJson) Validate() error {
	return validation.ValidateStruct(e,
		validation.Field(&e.EyeColors, validation.Required),
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
		var eyeColorsJson EyeColorsJson
		query := `
			SELECT
				entry_id,
				color_id
			FROM
				eyecolor
			WHERE
				entry_id IN (?)
		`
		queryIDs, ok := r.URL.Query()["entry_id"]
		if !ok {
			query = `
				SELECT
					entry_id,
					color_id
				FROM
					eyecolor
			`
			err := db.SelectContext(r.Context(), &eyeColorsJson.EyeColors, query)
			if err != nil {
				log.Printf(fmt.Sprintf("db error: %v", err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			// json返却
			err = json.NewEncoder(w).Encode(&eyeColorsJson)
			if err != nil {
				log.Printf(fmt.Sprintf("json encode error: %v", err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		} else if len(queryIDs) == 1 {
			query = `
				SELECT
					entry_id,
					color_id
				FROM
					eyecolor
				WHERE
					entry_id = $1
			`
			err := db.SelectContext(r.Context(), &eyeColorsJson.EyeColors, query, queryIDs[0])
			if err != nil {
				log.Printf(fmt.Sprintf("db error: %v", err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			// json返却
			err = json.NewEncoder(w).Encode(&eyeColorsJson)
			if err != nil {
				log.Printf(fmt.Sprintf("json encode error: %v", err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
		// entry_idの数だけ置換文字を作成
		query, args, err := sqlx.In(query, queryIDs)
		if err != nil {
			log.Printf(fmt.Sprintf("db error: %v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		// Postgresの場合は置換文字を$1, $2, ...とする必要がある
		query = sqlx.Rebind(len(queryIDs), query)
		err = db.SelectContext(r.Context(), &eyeColorsJson.EyeColors, query, args...)
		if err != nil {
			log.Printf(fmt.Sprintf("db error: %v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		// json返却
		err = json.NewEncoder(w).Encode(&eyeColorsJson)
		if err != nil {
			log.Printf(fmt.Sprintf("json encode error: %v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	case http.MethodPost:
		var eyeColorsJson EyeColorsJson
		query := `
			INSERT INTO eyecolor (
				entry_id,
				color_id
			) VALUES (
				:entry_id,
				:color_id
			)
		`
		// json読み込み
		jsonBytes, err := io.ReadAll(r.Body)
		if err != nil {
			log.Println(fmt.Sprintf("read error: %v", err))
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		err = json.Unmarshal(jsonBytes, &eyeColorsJson)
		if err != nil {
			log.Printf(fmt.Sprintf("json decode error: %v body:%v", err, r.Body))
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}
		// jsonバリデーション
		err = eyeColorsJson.Validate()
		if err != nil {
			log.Printf(fmt.Sprintf("validation error: %v", err))
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}
		for _, eyeColor := range eyeColorsJson.EyeColors {
			// jsonバリデーション
			err = eyeColor.Validate()
			if err != nil {
				log.Printf(fmt.Sprintf("validation error: %v", err))
				http.Error(w, err.Error(), http.StatusUnprocessableEntity)
				return
			}
			_, err = db.NamedExecContext(r.Context(), query, eyeColor)
			if err != nil {
				log.Printf(fmt.Sprintf("db error: %v", err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		// json返却
		err = json.NewEncoder(w).Encode(&eyeColorsJson)
		if err != nil {
			log.Printf(fmt.Sprintf("json encode error: %v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	case http.MethodPut:
		var eyeColorsJson EyeColorsJson
		query := `
			UPDATE
				eyecolor
			SET
				color_id = :color_id
			WHERE
				entry_id = :entry_id
		`
		jsonBytes, err := io.ReadAll(r.Body)
		if err != nil {
			log.Println(fmt.Sprintf("read error: %v", err))
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		err = json.Unmarshal(jsonBytes, &eyeColorsJson)
		if err != nil {
			log.Printf(fmt.Sprintf("json decode error: %v body:%v", err, r.Body))
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}
		// jsonバリデーション
		err = eyeColorsJson.Validate()
		if err != nil {
			log.Printf(fmt.Sprintf("validation error: %v", err))
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}
		for _, eyeColor := range eyeColorsJson.EyeColors {
			// jsonバリデーション
			err = eyeColor.Validate()
			if err != nil {
				log.Printf(fmt.Sprintf("validation error: %v", err))
				http.Error(w, err.Error(), http.StatusUnprocessableEntity)
				return
			}
			_, err = db.NamedExecContext(r.Context(), query, eyeColor)
			if err != nil {
				log.Printf(fmt.Sprintf("db error: %v", err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		// json返却
		err = json.NewEncoder(w).Encode(&eyeColorsJson)
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
