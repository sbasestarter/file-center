package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/jiuzhou-zhao/go-fundamental/loge"
	"github.com/sbasestarter/file-center/internal/config"
	"github.com/sgostarter/libfs"
)

type fsQueryResponse struct {
	ErrCode int    `json:"err_code"`
	ErrMsg  string `json:"err_msg,omitempty"`
	Exists  bool   `json:"exists"`
}

func doJSONObjectResponse(ctx context.Context, w http.ResponseWriter, jsonobj interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	bytes, err := json.Marshal(jsonobj)
	if err != nil {
		loge.Error(ctx, err)
		return err
	}
	_, err = fmt.Fprint(w, string(bytes))
	if err != nil {
		loge.Error(ctx, err)
		return err
	}

	loge.Infof(ctx, "%+v", jsonobj)

	return nil
}

// HandleFsQuery function
func HandleFsQuery(cfg *config.Config) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		err := handleFsQuery(w, r, cfg)
		if err != nil {
			err = doJSONObjectResponse(r.Context(), w, &fsQueryResponse{
				ErrCode: -1,
				ErrMsg:  err.Error(),
			})
			if err != nil {
				loge.Error(r.Context(), err)
			}
		}
	}
}

func handleFsQuery(w http.ResponseWriter, r *http.Request, cfg *config.Config) error {
	fileSize, err := strconv.ParseUint(r.FormValue("file_size"), 10, 64)
	if err != nil {
		return err
	}
	fileMd5 := r.FormValue("file_md5")
	if fileSize <= 0 {
		return fmt.Errorf("invalid paramters: %v - %v", fileSize, fileMd5)
	}

	var exists bool
	if fileMd5 != "" {
		exists, err = libfs.IsSizeMD5ExistsInStorage(fileSize, fileMd5, cfg.StgRoot)
	} else {
		exists, err = libfs.IsSizeExistsInStorage(fileSize, cfg.StgRoot)
	}
	if err != nil {
		return err
	}

	err = doJSONObjectResponse(r.Context(), w, &fsQueryResponse{
		Exists: exists,
	})
	if err != nil {
		return err
	}

	return err
}
