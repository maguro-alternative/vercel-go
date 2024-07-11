package bwh

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

type BWH struct {
	EntryID int64  `db:"entry_id" json:"entry_id"`
	Bust    int64  `db:"bust" json:"bust"`
	Waist   int64  `db:"waist" json:"waist"`
	Hip     int64  `db:"hip" json:"hip"`
	Height  *int64 `db:"height" json:"height"`
	Weight  *int64 `db:"weight" json:"weight"`
}

func (b *BWH) Validate() error {
	return validation.ValidateStruct(b,
		validation.Field(&b.EntryID, validation.Required),
		validation.Field(&b.Bust, validation.Required),
		validation.Field(&b.Waist, validation.Required),
		validation.Field(&b.Hip, validation.Required),
	)
}

type BWHsJson struct {
	BWHs []BWH `json:"bwhs"`
}

func (b *BWHsJson) Validate() error {
	return validation.ValidateStruct(b,
		validation.Field(&b.BWHs, validation.Required),
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
		var bwhsJson BWHsJson
		query := `
			SELECT
				entry_id,
				bust,
				waist,
				hip,
				height,
				weight
			FROM
				bwh
			WHERE
				entry_id IN (?)
		`
		// クエリパラメータからentry_idを取得
		queryIDs, ok := r.URL.Query()["entry_id"]
		// クエリパラメータにentry_idがない場合は全件取得
		if !ok {
			query := `
				SELECT
					entry_id,
					bust,
					waist,
					hip,
					height,
					weight
				FROM
					bwh
			`
			err := db.SelectContext(r.Context(), &bwhsJson.BWHs, query)
			if err != nil {
				log.Println(fmt.Sprintf("select error: %v", err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			// json返却
			err = json.NewEncoder(w).Encode(&bwhsJson)
			if err != nil {
				log.Println(fmt.Sprintf("json encode error: %v", err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		// クエリパラメータのentry_idが1つの場合はそのentry_idのみ取得
		} else if len(queryIDs) == 1 {
			query := `
				SELECT
					entry_id,
					bust,
					waist,
					hip,
					height,
					weight
				FROM
					bwh
				WHERE
					entry_id = $1
			`
			err := db.SelectContext(r.Context(), &bwhsJson.BWHs, query, queryIDs[0])
			if err != nil {
				log.Println(fmt.Sprintf("select error: %v", err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			// json返却
			err = json.NewEncoder(w).Encode(&bwhsJson)
			if err != nil {
				log.Println(fmt.Sprintf("json encode error: %v", err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
		// idの数だけ置換文字を作成
		query, args, err := sqlx.In(query, queryIDs)
		if err != nil {
			log.Println(fmt.Sprintf("in error: %v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		// Postgresの場合は置換文字を$1, $2, ...とする必要がある
		query = sqlx.Rebind(len(queryIDs), query)
		err = db.SelectContext(r.Context(), &bwhsJson.BWHs, query, args...)
		if err != nil {
			log.Println(fmt.Sprintf("select error: %v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		// json返却
		err = json.NewEncoder(w).Encode(&bwhsJson)
		if err != nil {
			log.Println(fmt.Sprintf("json encode error: %v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	case http.MethodPost:
		var bwhsJson BWHsJson
		query := `
			INSERT INTO bwh (
				entry_id,
				bust,
				waist,
				hip,
				height,
				weight
			) VALUES (
				:entry_id,
				:bust,
				:waist,
				:hip,
				:height,
				:weight
			)
		`
		// json読み込み
		jsonBytes, err := io.ReadAll(r.Body)
		if err != nil {
			log.Println(fmt.Sprintf("read error: %v", err))
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		err = json.Unmarshal(jsonBytes, &bwhsJson)
		if err != nil {
			log.Println(fmt.Sprintf("json unmarshal error: %v", err))
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		if len(bwhsJson.BWHs) == 0 {
			log.Println("json unexpected error: empty body")
			http.Error(w, "json unexpected error: empty body", http.StatusBadRequest)
		}
		for _, bwh := range bwhsJson.BWHs {
			// DBへの登録
			_, err = db.NamedExecContext(r.Context(), query, bwh)
			if err != nil {
				log.Println(fmt.Sprintf("insert error: %v", err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		}
		// json返却
		err = json.NewEncoder(w).Encode(&bwhsJson)
		if err != nil {
			log.Println(fmt.Sprintf("json encode error: %v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	case http.MethodPut:
		var bwhsJson BWHsJson
		query := `
			UPDATE
				bwh
			SET
				bust = :bust,
				waist = :waist,
				hip = :hip,
				height = :height,
				weight = :weight
			WHERE
				entry_id = :entry_id
		`
		// json読み込み
		jsonBytes, err := io.ReadAll(r.Body)
		if err != nil {
			log.Println(fmt.Sprintf("read error: %v", err))
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		err = json.Unmarshal(jsonBytes, &bwhsJson)
		if err != nil {
			log.Println(fmt.Sprintf("json unmarshal error: %v", err))
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		}
		for _, bwh := range bwhsJson.BWHs {
			_, err = db.NamedExecContext(r.Context(), query, bwh)
			if err != nil {
				log.Println(fmt.Sprintf("update error: %v", err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		}
		// json返却
		err = json.NewEncoder(w).Encode(&bwhsJson)
		if err != nil {
			log.Println(fmt.Sprintf("json encode error: %v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	case http.MethodDelete:
		var delIDs IDs
		query := `
			DELETE FROM
				bwh
			WHERE
				entry_id IN (?)
		`
		// json読み込み
		jsonBytes, err := io.ReadAll(r.Body)
		if err != nil {
			log.Println(fmt.Sprintf("read error: %v", err))
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		err = json.Unmarshal(jsonBytes, &delIDs)
		if err != nil {
			log.Println(fmt.Sprintf("json unmarshal error: %v", err))
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		// idが0の場合は何もしない
		if len(delIDs.IDs) == 0 {
			return
		// idが1の場合はそのidのみ削除
		} else if len(delIDs.IDs) == 1 {
			query := `
				DELETE FROM
					bwh
				WHERE
					entry_id = $1
			`
			_, err = db.ExecContext(r.Context(), query, delIDs.IDs[0])
			if err != nil {
				log.Println(fmt.Sprintf("delete error: %v", err))
				http.Error(w, fmt.Sprintf("delete error: %v", err), http.StatusInternalServerError)
			}
			err = json.NewEncoder(w).Encode(&delIDs)
			if err != nil {
				log.Println(fmt.Sprintf("json encode error: %v", err))
				http.Error(w, fmt.Sprintf("json encode error: %v", err), http.StatusInternalServerError)
			}
			return
		}
		// idの数だけ置換文字を作成
		query, args, err := sqlx.In(query, delIDs.IDs)
		if err != nil {
			log.Println(fmt.Sprintf("in error: %v", err))
			http.Error(w, fmt.Sprintf("in error: %v", err), http.StatusInternalServerError)
		}
		// Postgresの場合は置換文字を$1, $2, ...とする必要がある
		query = sqlx.Rebind(len(delIDs.IDs), query)
		_, err = db.ExecContext(r.Context(), query, args...)
		if err != nil {
			log.Println(fmt.Sprintf("delete error: %v", err))
			http.Error(w, fmt.Sprintf("delete error: %v", err), http.StatusInternalServerError)
		}
		// json返却
		err = json.NewEncoder(w).Encode(&delIDs)
		if err != nil {
			log.Println(fmt.Sprintf("json encode error: %v", err))
			http.Error(w, fmt.Sprintf("json encode error: %v", err), http.StatusInternalServerError)
		}
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
