package handler

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/alexsey-popov/gkeep/internal/auth"
	"github.com/alexsey-popov/gkeep/internal/crypto"
	"github.com/alexsey-popov/gkeep/internal/models"
	"github.com/alexsey-popov/gkeep/internal/validator"
	api "github.com/alexsey-popov/gkeep/proto"
	"github.com/alexsey-popov/gkeep/proto/file"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

func TestHandler_FileUnitMethods(t *testing.T) {
	h, rep, _, authCtx, userID := setupTest(t)

	// Создаем тестовый файл в репозитории
	tmp, err := os.CreateTemp("", "test-file-*")
	require.NoError(t, err)
	defer os.Remove(tmp.Name())
	_, err = tmp.WriteString("hello world")
	require.NoError(t, err)
	tmp.Close()

	createdFile := &models.File{
		UserId:      userID,
		Name:        "Мой документ",
		Filename:    "document.txt",
		StoragePath: tmp.Name(),
		Size:        11,
		Checksum:    "checksum-abc",
	}
	err = rep.FileCreate(authCtx, createdFile)
	require.NoError(t, err)

	t.Run("FileList - unauthenticated", func(t *testing.T) {
		req := &api.ListRequest{}
		resp, err := h.FileList(context.Background(), req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, st.Code())
		assert.Nil(t, resp.GetFiles())
	})

	t.Run("FileList - authenticated success", func(t *testing.T) {
		req := &api.ListRequest{}
		resp, err := h.FileList(authCtx, req)
		require.NoError(t, err)
		require.Len(t, resp.GetFiles(), 1)
		assert.Equal(t, "Мой документ", resp.GetFiles()[0].GetName())
		assert.Equal(t, "document.txt", resp.GetFiles()[0].GetFilename())
		assert.Equal(t, int64(11), resp.GetFiles()[0].GetSize())
		assert.Equal(t, "checksum-abc", resp.GetFiles()[0].GetChecksum())
	})

	t.Run("FileList - repository error", func(t *testing.T) {
		hErr, repErr, _, authCtxErr, _ := setupTest(t)
		repErr.fileListErr = errors.New("db error")

		req := &api.ListRequest{}
		resp, err := hErr.FileList(authCtxErr, req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.Internal, st.Code())
		assert.Nil(t, resp.GetFiles())
	})

	t.Run("FileShow - unauthenticated", func(t *testing.T) {
		req := api.IDRequest_builder{
			Id: &createdFile.Id,
		}.Build()
		resp, err := h.FileShow(context.Background(), req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, st.Code())
		assert.Nil(t, resp.GetFile())
	})

	t.Run("FileShow - empty id", func(t *testing.T) {
		req := api.IDRequest_builder{
			Id: stringPtr("   "),
		}.Build()
		resp, err := h.FileShow(authCtx, req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
		assert.Nil(t, resp.GetFile())
	})

	t.Run("FileShow - not found", func(t *testing.T) {
		req := api.IDRequest_builder{
			Id: stringPtr("unknown-id"),
		}.Build()
		resp, err := h.FileShow(authCtx, req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.NotFound, st.Code())
		assert.Nil(t, resp.GetFile())
	})

	t.Run("FileShow - repository error", func(t *testing.T) {
		hErr, repErr, _, authCtxErr, _ := setupTest(t)
		repErr.fileShowErr = errors.New("db error")

		req := api.IDRequest_builder{
			Id: stringPtr("file-1"),
		}.Build()
		resp, err := hErr.FileShow(authCtxErr, req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.Internal, st.Code())
		assert.Nil(t, resp.GetFile())
	})

	t.Run("FileShow - authenticated success", func(t *testing.T) {
		req := api.IDRequest_builder{
			Id: &createdFile.Id,
		}.Build()
		resp, err := h.FileShow(authCtx, req)
		require.NoError(t, err)
		require.NotNil(t, resp.GetFile())
		assert.Equal(t, createdFile.Id, resp.GetFile().GetId())
		assert.Equal(t, "Мой документ", resp.GetFile().GetName())
		assert.Equal(t, "document.txt", resp.GetFile().GetFilename())
	})

	t.Run("FileDelete - unauthenticated", func(t *testing.T) {
		req := api.IDRequest_builder{
			Id: &createdFile.Id,
		}.Build()
		resp, err := h.FileDelete(context.Background(), req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, st.Code())
		assert.Nil(t, resp.GetFile())
	})

	t.Run("FileDelete - empty id", func(t *testing.T) {
		req := api.IDRequest_builder{
			Id: stringPtr(""),
		}.Build()
		resp, err := h.FileDelete(authCtx, req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
		assert.Nil(t, resp.GetFile())
	})

	t.Run("FileDelete - not found", func(t *testing.T) {
		req := api.IDRequest_builder{
			Id: stringPtr("nonexistent-file-id"),
		}.Build()
		resp, err := h.FileDelete(authCtx, req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.NotFound, st.Code())
		assert.Nil(t, resp.GetFile())
	})

	t.Run("FileDelete - repository error", func(t *testing.T) {
		hErr, repErr, _, authCtxErr, _ := setupTest(t)
		repErr.fileDeleteErr = errors.New("db error")

		req := api.IDRequest_builder{
			Id: stringPtr("file-1"),
		}.Build()
		resp, err := hErr.FileDelete(authCtxErr, req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.Internal, st.Code())
		assert.Nil(t, resp.GetFile())
	})

	t.Run("FileDelete - authenticated success", func(t *testing.T) {
		req := api.IDRequest_builder{
			Id: &createdFile.Id,
		}.Build()
		resp, err := h.FileDelete(authCtx, req)
		require.NoError(t, err)
		require.NotNil(t, resp.GetFile())
		assert.Equal(t, createdFile.Id, resp.GetFile().GetId())
		assert.Equal(t, "Мой документ", resp.GetFile().GetName())
		assert.Equal(t, "document.txt", resp.GetFile().GetFilename())

		// Убеждаемся, что файл удален из репозитория
		showResp, err := h.FileShow(authCtx, req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.NotFound, st.Code())
		assert.Nil(t, showResp.GetFile())
	})
}

func TestHandler_FileStreamingIntegration(t *testing.T) {
	_ = os.RemoveAll("attachments")
	defer os.RemoveAll("attachments")

	secret := "secret-jwt-token"
	guard := auth.New(secret)
	rep := newMockRepository()
	cipherService, err := crypto.NewWithTempKey()
	require.NoError(t, err)
	val, err := validator.New()
	require.NoError(t, err)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	h := New(rep, guard, cipherService, val, log)

	tokenString, _, err := guard.NewUserToken("00000000-0000-0000-0000-000000000001")
	require.NoError(t, err)

	buffer := 1024 * 1024
	lis := bufconn.Listen(buffer)

	server := grpc.NewServer(
		grpc.UnaryInterceptor(guard.UnaryServerInterceptor()),
		grpc.StreamInterceptor(guard.StreamServerInterceptor()),
	)
	api.RegisterAPIServer(server, h)

	go func() {
		if err := server.Serve(lis); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			t.Logf("Server error: %v", err)
		}
	}()
	defer server.Stop()

	conn, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	defer conn.Close()

	client := api.NewAPIClient(conn)
	authCtx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("authorization", "Bearer "+tokenString))

	fileData := []byte("This is a test file content for chunk streaming verification in gkeep!")
	hasher := sha256.New()
	hasher.Write(fileData)
	expectedChecksum := hex.EncodeToString(hasher.Sum(nil))

	var uploadedFileID string
	var savedFilePath string

	t.Run("1. FileUpload - успешная загрузка по чанкам", func(t *testing.T) {
		stream, err := client.FileUpload(authCtx)
		require.NoError(t, err)

		recordName := "Пользовательское название"
		fileName := "sample.bin"
		fileSize := int64(len(fileData))

		// Первое сообщение - метаданные
		metaReq := &file.UploadFileRequest{}
		metaReq.SetMetadata(file.FileMetadata_builder{
			Name:     &recordName,
			Filename: &fileName,
			Size:     &fileSize,
			Checksum: &expectedChecksum,
		}.Build())

		err = stream.Send(metaReq)
		require.NoError(t, err)

		// Отправляем чанками по 10 байт
		chunkSize := 10
		for i := 0; i < len(fileData); i += chunkSize {
			end := i + chunkSize
			if end > len(fileData) {
				end = len(fileData)
			}
			chunkReq := &file.UploadFileRequest{}
			chunkReq.SetChunk(fileData[i:end])
			err = stream.Send(chunkReq)
			require.NoError(t, err)
		}

		resp, err := stream.CloseAndRecv()
		require.NoError(t, err)
		require.NotNil(t, resp.GetFile())
		assert.Equal(t, "Пользовательское название", resp.GetFile().GetName())
		assert.Equal(t, "sample.bin", resp.GetFile().GetFilename())
		assert.Equal(t, int64(len(fileData)), resp.GetFile().GetSize())
		assert.Equal(t, expectedChecksum, resp.GetFile().GetChecksum())
		assert.NotEmpty(t, resp.GetFile().GetId())

		uploadedFileID = resp.GetFile().GetId()

		// Проверяем, что файл сохранен в attachments/00000000-0000-0000-0000-000000000001/{date}
		dateStr := time.Now().Format("2006-01-02")
		expectedDir := filepath.Join("attachments", "00000000-0000-0000-0000-000000000001", dateStr)
		savedFile, err := rep.FileShow(context.Background(), "00000000-0000-0000-0000-000000000001", uploadedFileID)
		require.NoError(t, err)
		savedFilePath = savedFile.StoragePath
		assert.True(t, strings.HasPrefix(savedFile.StoragePath, expectedDir), "файл должен сохраняться в директории %s, получено: %s", expectedDir, savedFile.StoragePath)
		assert.FileExists(t, savedFile.StoragePath)

		// Проверяем, что файл на диске зашифрован (не совпадает с открытым текстом)
		diskData, err := os.ReadFile(savedFile.StoragePath)
		require.NoError(t, err)
		assert.NotEqual(t, fileData, diskData, "файл на сервере должен быть сохранен в зашифрованном виде")
		assert.Equal(t, len(fileData)+16, len(diskData), "размер файла на диске должен включать 16 байт IV вектора инициализации")
	})

	t.Run("2. FileList - проверка наличия загруженного файла", func(t *testing.T) {
		resp, err := client.FileList(authCtx, &api.ListRequest{})
		require.NoError(t, err)
		require.Len(t, resp.GetFiles(), 1)
		assert.Equal(t, uploadedFileID, resp.GetFiles()[0].GetId())
		assert.Equal(t, "Пользовательское название", resp.GetFiles()[0].GetName())
		assert.Equal(t, "sample.bin", resp.GetFiles()[0].GetFilename())
	})

	t.Run("3. FileShow - просмотр метаданных загруженного файла", func(t *testing.T) {
		req := api.IDRequest_builder{
			Id: &uploadedFileID,
		}.Build()
		resp, err := client.FileShow(authCtx, req)
		require.NoError(t, err)
		require.NotNil(t, resp.GetFile())
		assert.Equal(t, uploadedFileID, resp.GetFile().GetId())
		assert.Equal(t, "Пользовательское название", resp.GetFile().GetName())
		assert.Equal(t, "sample.bin", resp.GetFile().GetFilename())
		assert.Equal(t, int64(len(fileData)), resp.GetFile().GetSize())
	})

	t.Run("4. FileDownload - успешное скачивание по чанкам", func(t *testing.T) {
		req := file.DownloadFileRequest_builder{
			Id: &uploadedFileID,
		}.Build()

		downloadStream, err := client.FileDownload(authCtx, req)
		require.NoError(t, err)

		// Первое сообщение - метаданные
		firstMsg, err := downloadStream.Recv()
		require.NoError(t, err)
		require.NotNil(t, firstMsg.GetMetadata())
		assert.Equal(t, "Пользовательское название", firstMsg.GetMetadata().GetName())
		assert.Equal(t, "sample.bin", firstMsg.GetMetadata().GetFilename())
		assert.Equal(t, int64(len(fileData)), firstMsg.GetMetadata().GetSize())
		assert.Equal(t, expectedChecksum, firstMsg.GetMetadata().GetChecksum())

		var downloadedBytes []byte
		for {
			msg, err := downloadStream.Recv()
			if errors.Is(err, io.EOF) {
				break
			}
			require.NoError(t, err)
			chunk := msg.GetChunk()
			downloadedBytes = append(downloadedBytes, chunk...)
		}

		assert.Equal(t, fileData, downloadedBytes)
	})

	t.Run("5. FileUpload - несовпадение контрольной суммы", func(t *testing.T) {
		stream, err := client.FileUpload(authCtx)
		require.NoError(t, err)

		recordName := "Невалидный файл"
		fileName := "bad_checksum.bin"
		fileSize := int64(10)
		badChecksum := "invalid-checksum-value"

		metaReq := &file.UploadFileRequest{}
		metaReq.SetMetadata(file.FileMetadata_builder{
			Name:     &recordName,
			Filename: &fileName,
			Size:     &fileSize,
			Checksum: &badChecksum,
		}.Build())

		err = stream.Send(metaReq)
		require.NoError(t, err)

		chunkReq := &file.UploadFileRequest{}
		chunkReq.SetChunk([]byte("0123456789"))
		err = stream.Send(chunkReq)
		require.NoError(t, err)

		_, err = stream.CloseAndRecv()
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
	})

	t.Run("6. FileDelete - удаление файла", func(t *testing.T) {
		req := api.IDRequest_builder{
			Id: &uploadedFileID,
		}.Build()

		resp, err := client.FileDelete(authCtx, req)
		require.NoError(t, err)
		require.NotNil(t, resp.GetFile())
		assert.Equal(t, uploadedFileID, resp.GetFile().GetId())

		// Проверяем, что файл больше не отображается в списке
		listResp, err := client.FileList(authCtx, &api.ListRequest{})
		require.NoError(t, err)
		assert.Empty(t, listResp.GetFiles())

		// Проверяем, что физический файл был удален с диска
		assert.NoFileExists(t, savedFilePath)
	})
}
