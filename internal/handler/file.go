package handler

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/alexsey-popov/gkeep/internal/auth"
	"github.com/alexsey-popov/gkeep/internal/models"
	"github.com/alexsey-popov/gkeep/internal/repository"
	"github.com/alexsey-popov/gkeep/proto"
	"github.com/alexsey-popov/gkeep/proto/file"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// toProtoFile преобразует модель File в protobuf сообщение File
func toProtoFile(item *models.File) *file.File {
	if item == nil {
		return nil
	}

	return file.File_builder{
		Id:       &item.Id,
		Name:     &item.Name,
		Filename: &item.Filename,
		Size:     &item.Size,
		Checksum: &item.Checksum,
	}.Build()
}

// FileList получение списка файлов пользователя
func (h *Handler) FileList(ctx context.Context, in *proto.ListRequest) (*file.ListResponse, error) {
	var response file.ListResponse

	userID, err := auth.UserIDFromContext(ctx)
	if err != nil || userID == "" {
		return &response, status.Error(codes.Unauthenticated, "пользователь не аутентифицирован")
	}

	name := in.GetName()
	page := in.GetPage()

	items, err := h.rep.FileList(ctx, userID, name, page)
	if err != nil {
		return &response, status.Error(codes.Internal, "ошибка при получении списка файлов")
	}

	pbList := make([]*file.File, 0, len(items))
	for _, item := range items {
		item := item
		pbList = append(pbList, toProtoFile(item))
	}

	response.SetFiles(pbList)

	return &response, nil
}

// FileShow получение информации о файле по идентификатору
func (h *Handler) FileShow(ctx context.Context, in *proto.IDRequest) (*file.Response, error) {
	var response file.Response

	userID, err := auth.UserIDFromContext(ctx)
	if err != nil || userID == "" {
		return &response, status.Error(codes.Unauthenticated, "пользователь не аутентифицирован")
	}

	id := strings.TrimSpace(in.GetId())
	if id == "" {
		return &response, status.Error(codes.InvalidArgument, "id не указан")
	}

	item, err := h.rep.FileShow(ctx, userID, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &response, status.Error(codes.NotFound, "запись не найдена")
		}
		return &response, status.Error(codes.Internal, "ошибка при получении записи")
	}

	response.SetFile(toProtoFile(item))

	return &response, nil
}

// FileDelete удаление записи файла по идентификатору
func (h *Handler) FileDelete(ctx context.Context, in *proto.IDRequest) (*file.Response, error) {
	var response file.Response

	userID, err := auth.UserIDFromContext(ctx)
	if err != nil || userID == "" {
		return &response, status.Error(codes.Unauthenticated, "пользователь не аутентифицирован")
	}

	id := strings.TrimSpace(in.GetId())
	if id == "" {
		return &response, status.Error(codes.InvalidArgument, "id не указан")
	}

	item, err := h.rep.FileDelete(ctx, userID, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &response, status.Error(codes.NotFound, "запись не найдена")
		}
		return &response, status.Error(codes.Internal, "ошибка при удалении записи")
	}

	if item.StoragePath != "" {
		_ = os.Remove(item.StoragePath)
	}

	response.SetFile(toProtoFile(item))

	return &response, nil
}

// FileUpload загрузка файла по чанкам (Client Streaming)
func (h *Handler) FileUpload(stream grpc.ClientStreamingServer[file.UploadFileRequest, file.UploadFileResponse]) error {
	ctx := stream.Context()
	userID, err := auth.UserIDFromContext(ctx)
	if err != nil || userID == "" {
		return status.Error(codes.Unauthenticated, "пользователь не аутентифицирован")
	}

	// 1. Читаем первое сообщение с метаданными
	req, err := stream.Recv()
	if err != nil {
		return status.Error(codes.InvalidArgument, "ошибка чтения метаданных файла")
	}

	meta := req.GetMetadata()
	if meta == nil || strings.TrimSpace(meta.GetName()) == "" {
		return status.Error(codes.InvalidArgument, "первым сообщением должны быть метаданные файла")
	}

	// 2. Создаем директорию attachments/{userId}/{date} и файл для накопления чанков
	dateStr := time.Now().Format("2006-01-02")
	dir := filepath.Join("attachments", userID, dateStr)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return status.Error(codes.Internal, "ошибка создания директории для сохранения файла")
	}

	tempFile, err := os.CreateTemp(dir, "upload-*")
	if err != nil {
		return status.Error(codes.Internal, "ошибка создания временного файла")
	}
	tempPath := tempFile.Name()

	var writer io.Writer = tempFile
	if h.cipher != nil {
		encWriter, err := h.cipher.NewEncryptWriter(tempFile)
		if err != nil {
			_ = tempFile.Close()
			_ = os.Remove(tempPath)
			return status.Error(codes.Internal, "ошибка инициализации шифрования файла")
		}
		writer = encWriter
	}

	hasher := sha256.New()
	var totalBytes int64

	var success bool
	defer func() {
		tempFile.Close()
		if !success {
			_ = os.Remove(tempPath)
		}
	}()

	// 3. Потоковое чтение чанков
	for {
		req, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return status.Errorf(codes.DataLoss, "обрыв соединения при передаче чанков: %v", err)
		}

		chunk := req.GetChunk()
		if len(chunk) > 0 {
			if _, err := writer.Write(chunk); err != nil {
				return status.Error(codes.Internal, "ошибка записи чанка на диск")
			}
			hasher.Write(chunk)
			totalBytes += int64(len(chunk))
		}
	}

	// 4. Проверка целостности
	calculatedChecksum := hex.EncodeToString(hasher.Sum(nil))
	if meta.GetChecksum() != "" && meta.GetChecksum() != calculatedChecksum {
		return status.Error(codes.InvalidArgument, "несовпадение контрольной суммы файла")
	}

	// 5. Сохранение метаданных в репозитории
	fileRecord, err := h.rep.FileCreate(ctx, userID, strings.TrimSpace(meta.GetName()), strings.TrimSpace(meta.GetFilename()), tempPath, totalBytes, calculatedChecksum)
	if err != nil {
		if errors.Is(err, repository.ErrConflictUnique) {
			return status.Error(codes.AlreadyExists, "файл с таким названием уже существует")
		}
		return status.Error(codes.Internal, "ошибка сохранения информации о файле")
	}

	success = true

	var resp file.UploadFileResponse
	resp.SetFile(toProtoFile(fileRecord))
	return stream.SendAndClose(&resp)
}

// FileDownload скачивание файла по чанкам (Server Streaming)
func (h *Handler) FileDownload(in *file.DownloadFileRequest, stream grpc.ServerStreamingServer[file.DownloadFileResponse]) error {
	ctx := stream.Context()
	userID, err := auth.UserIDFromContext(ctx)
	if err != nil || userID == "" {
		return status.Error(codes.Unauthenticated, "пользователь не аутентифицирован")
	}

	id := strings.TrimSpace(in.GetId())
	if id == "" {
		return status.Error(codes.InvalidArgument, "id не указан")
	}

	item, err := h.rep.FileShow(ctx, userID, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return status.Error(codes.NotFound, "файл не найден")
		}
		return status.Error(codes.Internal, "ошибка при получении информации о файле")
	}

	f, err := os.Open(item.StoragePath)
	if err != nil {
		return status.Error(codes.Internal, "ошибка открытия файла")
	}
	defer f.Close()

	var reader io.Reader = f
	if h.cipher != nil {
		decReader, err := h.cipher.NewDecryptReader(f)
		if err != nil {
			return status.Error(codes.Internal, "ошибка инициализации расшифровывания файла")
		}
		reader = decReader
	}

	// 1. Отправляем метаданные
	metaResp := file.DownloadFileResponse_builder{
		Metadata: file.FileMetadata_builder{
			Name:     &item.Name,
			Filename: &item.Filename,
			Size:     &item.Size,
			Checksum: &item.Checksum,
		}.Build(),
	}.Build()
	if err := stream.Send(metaResp); err != nil {
		return status.Errorf(codes.Internal, "ошибка отправки метаданных файла: %v", err)
	}

	// 2. Отправляем бинарные чанки
	buf := make([]byte, 64*1024)
	for {
		n, err := reader.Read(buf)
		if n > 0 {
			chunkResp := &file.DownloadFileResponse{}
			chunkResp.SetChunk(buf[:n])
			if err := stream.Send(chunkResp); err != nil {
				return status.Errorf(codes.DataLoss, "ошибка передачи чанка: %v", err)
			}
		}
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return status.Errorf(codes.Internal, "ошибка чтения файла: %v", err)
		}
	}

	return nil
}
