package server

import (
	"bytes"
	"context"
	"strings"

	"github.com/gorilla/mux"
	"github.com/satori/go.uuid"
	"github.com/sbasestarter/file-center/internal/config"
	"github.com/sbasestarter/file-center/internal/file-center/handlers"
	"github.com/sbasestarter/proto-repo/gen/protorepo-file-center-go"
	"github.com/sgostarter/i/l"
	"github.com/sgostarter/libfs"
)

const (
	APIVer = "/r/v1"
)

type Server struct {
	cfg    *config.Config
	logger l.WrapperWithContext
}

func NewServer(ctx context.Context, cfg *config.Config) *Server {
	if cfg.Logger == nil {
		cfg.Logger = l.NewNopLoggerWrapper()
	}

	return &Server{
		cfg:    cfg,
		logger: cfg.Logger.WithFields(l.StringField(l.ClsKey, "Server")).GetWrapperWithContext(),
	}
}

func (s *Server) saveFile(fileName string, content []byte) (fileID string, err error) {
	item, err := libfs.NewSFSItem(fileName, s.cfg.StgRoot, s.cfg.StgTmpRoot)
	if err != nil {
		return
	}

	err = item.WriteFile(bytes.NewBuffer(content))
	if err != nil {
		return
	}

	err = item.WriteFileRecord()
	if err != nil {
		return
	}

	return item.GetFileID()
}

func (s *Server) UpdateFile(ctx context.Context, r *filecenterpb.UpdateFileRequest) (*filecenterpb.UpdateFileResponse, error) {
	fr := func(msg string) (*filecenterpb.UpdateFileResponse, error) {
		return &filecenterpb.UpdateFileResponse{
			Status: &filecenterpb.ServerStatus{
				Status: filecenterpb.FileCenterStatus_FCS_FAILED,
				Msg:    msg,
			},
		}, nil
	}

	if r.FileName == "" || strings.Contains(r.FileName, "\\/") {
		s.logger.Warnf(ctx, "null or invalid file name: %v", r.FileName)
		r.FileName = uuid.NewV4().String()
	}

	fileID, err := s.saveFile(r.FileName, r.Content)
	if err != nil {
		return fr(err.Error())
	}

	return &filecenterpb.UpdateFileResponse{
		Status: &filecenterpb.ServerStatus{
			Status: filecenterpb.FileCenterStatus_FCS_SUCCESS,
		},
		FileUrl: fileID,
	}, nil
}

func (s *Server) HTTPRegister(r *mux.Router) {
	r.HandleFunc(APIVer+"/fs/query", handlers.HandleFsQuery(s.cfg))
	r.HandleFunc(APIVer+"/fs/upload", handlers.HandleFsUpload(s.cfg))
	r.HandleFunc(APIVer+"/fs/download/{file_id}", handlers.HandleFsDownload(s.cfg))
	r.HandleFunc(APIVer+"/fs/transfer", handlers.HandleFsTransfer(s.cfg))
	r.HandleFunc(APIVer+"/fs/list", handlers.HandleFsList(s.cfg))
	r.HandleFunc("/download/{file_id}", handlers.HandleFsDownloadV1(s.cfg))
	r.HandleFunc("/thumbnail/{width}/{height}/{file_id}", handlers.HandleFsImageCacheV1(s.cfg))
}
