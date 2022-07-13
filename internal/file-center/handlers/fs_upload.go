package handlers

import (
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/sbasestarter/file-center/internal/config"
	"github.com/sgostarter/libfs"
)

type fsUploadResponse struct {
	ErrCode int    `json:"err_code"`
	ErrMsg  string `json:"err_msg,omitempty"`
	FileURL string `json:"file_url,omitempty"`
}

// HandleFsUpload function
func HandleFsUpload(cfg *config.Config) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		err := handleFsUpload(w, r, cfg)
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

func handleFsUpload(w http.ResponseWriter, r *http.Request, cfg *config.Config) error {
	kvs := r.URL.Query()
	fileName := kvs.Get("file_name")
	fileSize, _ := strconv.ParseUint(kvs.Get("file_size"), 10, 64)
	fileMd5 := kvs.Get("file_md5")

	// 这里有个BUG，如果考虑权限验证，可以通过md5和size来构建假数据获取服务端已经存在的文件的权限
	if fileSize > 0 && fileMd5 != "" && fileName != "" {
		item, err := libfs.NewSFSItemByInfo(fileMd5, fileSize, fileName, cfg.StgRoot, cfg.StgTmpRoot)
		if err != nil {
			cfg.ContextLogger.Error(r.Context(), err)
		} else {
			dExists, fExists, err := item.ExistsInStorage()
			if err != nil {
				cfg.ContextLogger.Error(r.Context(), err)
			} else {
				if fExists {
					fileID, err := item.GetFileID()
					if err != nil {
						return err
					}
					return doJSONObjectResponse(r.Context(), w, &fsUploadResponse{
						FileURL: fileID,
					})
				}
				if dExists {
					err = item.WriteFileRecord()
					if err != nil {
						return err
					}
				}
			}
		}
	}

	v1 := false
	file, head, err := r.FormFile("file")
	if err != nil {
		file, head, err = r.FormFile("uploadfile")
		if err != nil {
			return err
		}
		v1 = true
	}

	defer func() {
		err := file.Close()
		if err != nil {
			cfg.ContextLogger.Error(r.Context(), err)
		}
	}()

	if fileName == "" {
		fileName = head.Filename
	}

	item, err := libfs.NewSFSItem(fileName, cfg.StgRoot, cfg.StgTmpRoot)
	if err != nil {
		return err
	}
	err = item.WriteFile(file)
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

	if v1 {
		_, _ = io.WriteString(w, fmt.Sprintf("{\"is_ok\":true, \"url\":\"%s\"}", "/download/"+fileID))
		return nil
	}
	return doJSONObjectResponse(r.Context(), w, &fsUploadResponse{
		FileURL: fileID,
	})
}
