package haircolor

import (
	"encoding/json"
	"log"
	"fmt"
	"net/http"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	validation "github.com/go-ozzo/ozzo-validation"
)

type HairColor struct {
	EntryID int64 `db:"entry_id"`
	ColorID int64 `db:"color_id"`
}

func (h *HairColor) Validate() error {
	return validation.ValidateStruct(h,
		validation.Field(&h.EntryID, validation.Required),
		validation.Field(&h.ColorID, validation.Required),
	)
}

type HairColorsJson struct {
	HairColors []HairColor `json:"haircolors"`
}

func (h *HairColorsJson) Validate() error {
	return validation.ValidateStruct(h,
		validation.Field(&h.HairColors, validation.Required),
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
		var hairColorsJson HairColorsJson
		query := `
			SELECT
				entry_id,
				color_id
			FROM
				haircolor
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
					haircolor
			`
			// DBから全件取得
			err := db.SelectContext(r.Context(), &hairColorsJson.HairColors, query)
			if err != nil {
				log.Printf(fmt.Sprintf("db error: %v", err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			// json書き込み
			err = json.NewEncoder(w).Encode(&hairColorsJson)
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
					haircolor
				WHERE
					entry_id = $1
			`
			// DBから1件取得
			err := db.SelectContext(r.Context(), &hairColorsJson.HairColors, query, queryIDs[0])
			if err != nil {
				log.Printf(fmt.Sprintf("db error: %v", err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			// json書き込み
			err = json.NewEncoder(w).Encode(&hairColorsJson)
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
		err = db.SelectContext(r.Context(), &hairColorsJson.HairColors, query, args...)
		if err != nil {
			log.Printf(fmt.Sprintf("db error: %v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		// json書き込み
		err = json.NewEncoder(w).Encode(&hairColorsJson)
		if err != nil {
			log.Printf(fmt.Sprintf("json encode error: %v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	case http.MethodPost:
		var hairColorsJson HairColorsJson
		query := `
			INSERT INTO haircolor (
				entry_id,
				color_id
			) VALUES (
				:entry_id,
				:color_id
			)
		`
		// json読み込み
		err := json.NewDecoder(r.Body).Decode(&hairColorsJson)
		if err != nil {
			log.Printf(fmt.Sprintf("json decode error: %v body:%v", err, r.Body))
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		// jsonバリデーション
		err = hairColorsJson.Validate()
		if err != nil {
			log.Printf(fmt.Sprintf("validation error: %v", err))
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		}
		for _, hc := range hairColorsJson.HairColors {
			// jsonバリデーション
			err = hc.Validate()
			if err != nil {
				log.Printf(fmt.Sprintf("validation error: %v", err))
				http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			}
			if _, err := db.NamedExecContext(r.Context(), query, hc); err != nil {
				log.Printf(fmt.Sprintf("db error: %v", err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		}
		// json書き込み
		err = json.NewEncoder(w).Encode(&hairColorsJson)
		if err != nil {
			log.Printf(fmt.Sprintf("json encode error: %v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	case http.MethodPut:
		var hairColorsJson HairColorsJson
		query := `
			UPDATE
				haircolor
			SET
				color_id = :color_id
			WHERE
				entry_id = :entry_id
		`
		if err := json.NewDecoder(r.Body).Decode(&hairColorsJson); err != nil {
			log.Printf(fmt.Sprintf("json decode error: %v body:%v", err, r.Body))
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		// jsonバリデーション
		if err := hairColorsJson.Validate(); err != nil {
			log.Printf(fmt.Sprintf("validation error: %v", err))
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		}
		for _, hc := range hairColorsJson.HairColors {
			if err := hc.Validate(); err != nil {
				log.Printf(fmt.Sprintf("validation error: %v", err))
				http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			}
			if _, err := db.NamedExecContext(r.Context(), query, hc); err != nil {
				log.Printf(fmt.Sprintf("db error: %v", err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		}
		// json書き込み
		err := json.NewEncoder(w).Encode(&hairColorsJson)
		if err != nil {
			log.Printf(fmt.Sprintf("json encode error: %v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	case http.MethodDelete:
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
