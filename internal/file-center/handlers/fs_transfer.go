package handlers

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"

	"github.com/sbasestarter/file-center/internal/config"
	"github.com/sgostarter/libfs"
)

func HandleFsTransfer(cfg *config.Config) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		err := handleFsTransfer(w, r, cfg)
		if err != nil {
			err = doJSONObjectResponse(r.Context(), w, &fsUploadResponse{
				ErrCode: -1,
				ErrMsg:  err.Error(),
			})

			if err != nil {
				cfg.ContextLogger.Error(r.Context(), err)
			}
		}
	}
}

func handleFsTransfer(w http.ResponseWriter, r *http.Request, cfg *config.Config) error {
	u := r.FormValue("url")

	uri, err := url.ParseRequestURI(u)
	if err != nil {
		return err
	}

	fileName := path.Base(uri.Path)
	cfg.ContextLogger.Info(r.Context(), "[*] Filename "+fileName)

	// nolint: gosec, noctx
	resp, err := http.Get(u)
	if err != nil {
		return err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	item, err := libfs.NewSFSItem(fileName, cfg.StgRoot, cfg.StgTmpRoot)
	if err != nil {
		return err
	}

	err = item.WriteFile(bytes.NewReader(body))
	if err != nil {
		return err
	}

	err = item.WriteFileRecord()
	if err != nil {
		return err
	}

	fileID, err := item.GetFileID()
	if err != nil {
		return err
	}

	return doJSONObjectResponse(r.Context(), w, &fsUploadResponse{
		FileURL: fileID,
	})
}
