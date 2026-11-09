package client

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"

	api "github.com/alexsey-popov/gkeep/proto"
	pbFile "github.com/alexsey-popov/gkeep/proto/file"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestApp_Files(t *testing.T) {
	id1 := "f1"
	name1 := "Скан паспорта"
	filename1 := "passport.pdf"
	size1 := int64(1024)
	checksum1 := "sha256-hash-value"

	mockClient := &mockUsersClient{
		fileListFunc: func(ctx context.Context, in *api.ListRequest, opts ...grpc.CallOption) (*pbFile.ListResponse, error) {
			resp := &pbFile.ListResponse{}
			resp.SetFiles([]*pbFile.File{
				pbFile.File_builder{
					Id:       &id1,
					Name:     &name1,
					Filename: &filename1,
					Size:     &size1,
					Checksum: &checksum1,
				}.Build(),
			})
			return resp, nil
		},
		fileDeleteFunc: func(ctx context.Context, in *api.IDRequest, opts ...grpc.CallOption) (*pbFile.Response, error) {
			return &pbFile.Response{}, nil
		},
	}

	app := NewApp(mockClient)
	app.session = &Session{Token: "test-token", UserID: "uid-1", Login: "user1"}

	t.Run("positive - загрузка файла", func(t *testing.T) {
		app.openFileList()
		assert.Len(t, app.currentFiles, 1)
		assert.Equal(t, "Скан паспорта", app.currentFiles[0].GetName())
		assert.Equal(t, "passport.pdf", app.currentFiles[0].GetFilename())
		assert.Equal(t, int64(1024), app.currentFiles[0].GetSize())
	})

	t.Run("positive - просмотр страницы файла", func(t *testing.T) {
		app.openFileView(app.currentFiles[0])
		assert.Equal(t, "Скан паспорта", app.currentFile.GetName())
		assert.True(t, app.pages.HasPage("file_view"))
	})

	t.Run("positive - страница скачивания файла", func(t *testing.T) {
		app.openFileDownload(app.currentFiles[0])
		assert.Equal(t, "passport.pdf", app.downloadSavePathField.GetText())
		assert.True(t, app.pages.HasPage("file_download"))
	})

	t.Run("positive - удаление файла", func(t *testing.T) {
		app.currentFile = app.currentFiles[0]
		app.handleFileDelete()
		assert.Nil(t, app.currentFile)
	})

	t.Run("negative - запрос списка файлов без аутентификации", func(t *testing.T) {
		unauthApp := NewApp(mockClient)
		unauthApp.loadFiles()
		assert.True(t, unauthApp.pages.HasPage("error_modal"))
	})
}

func TestApp_FileUploadStreaming(t *testing.T) {
	tmpDir := t.TempDir()
	testFilePath := filepath.Join(tmpDir, "upload_sample.txt")
	testData := []byte("hello world this is a test upload file content stream")
	err := os.WriteFile(testFilePath, testData, 0644)
	require.NoError(t, err)

	hasher := sha256.New()
	hasher.Write(testData)
	expectedChecksum := hex.EncodeToString(hasher.Sum(nil))

	var streamClient *mockUploadStreamClient

	mockClient := &mockUsersClient{
		fileUploadFunc: func(ctx context.Context, opts ...grpc.CallOption) (grpc.ClientStreamingClient[pbFile.UploadFileRequest, pbFile.UploadFileResponse], error) {
			fId := "new-file-id"
			fName := "Тестовый файл"
			fFilename := "upload_sample.txt"
			fSize := int64(len(testData))

			streamClient = &mockUploadStreamClient{
				closeResp: &pbFile.UploadFileResponse{},
			}
			streamClient.closeResp.SetFile(pbFile.File_builder{
				Id:       &fId,
				Name:     &fName,
				Filename: &fFilename,
				Size:     &fSize,
				Checksum: &expectedChecksum,
			}.Build())
			return streamClient, nil
		},
		fileListFunc: func(ctx context.Context, in *api.ListRequest, opts ...grpc.CallOption) (*pbFile.ListResponse, error) {
			return &pbFile.ListResponse{}, nil
		},
	}

	app := NewApp(mockClient)
	app.session = &Session{Token: "test-token", UserID: "uid-1", Login: "user1"}

	t.Run("negative - валидация пустой формы", func(t *testing.T) {
		app.uploadFileNameField.SetText("")
		app.uploadFilePathField.SetText("")
		app.handleFileUpload()
		assert.True(t, app.pages.HasPage("error_modal"))
		app.pages.RemovePage("error_modal")
	})

	t.Run("negative - несуществующий файл", func(t *testing.T) {
		app.uploadFileNameField.SetText("Test")
		app.uploadFilePathField.SetText("/nonexistent/file/path.dat")
		app.handleFileUpload()
		assert.True(t, app.pages.HasPage("error_modal"))
		app.pages.RemovePage("error_modal")
	})

	t.Run("negative - папка вместо файла", func(t *testing.T) {
		app.uploadFileNameField.SetText("DirectoryTest")
		app.uploadFilePathField.SetText(tmpDir)
		app.handleFileUpload()
		assert.True(t, app.pages.HasPage("error_modal"))
		app.pages.RemovePage("error_modal")
	})

	t.Run("negative - слишком большой файл", func(t *testing.T) {
		largeFilePath := filepath.Join(tmpDir, "large_file.dat")
		f, createErr := os.Create(largeFilePath)
		require.NoError(t, createErr)
		// Truncate to 101 MB without allocating real disk space
		err = f.Truncate(101 * 1024 * 1024)
		require.NoError(t, err)
		f.Close()

		app.uploadFileNameField.SetText("Large File")
		app.uploadFilePathField.SetText(largeFilePath)
		app.handleFileUpload()
		assert.True(t, app.pages.HasPage("error_modal"))
		app.pages.RemovePage("error_modal")
	})

	t.Run("positive - успешная загрузка файла", func(t *testing.T) {
		app.uploadFileNameField.SetText("Тестовый файл")
		app.uploadFilePathField.SetText(testFilePath)
		app.handleFileUpload()

		assert.False(t, app.pages.HasPage("error_modal"))
		require.NotNil(t, streamClient)
		require.GreaterOrEqual(t, len(streamClient.sentReqs), 2) // 1 metadata + >=1 chunk

		// Проверяем метаданные в первом сообщении
		firstReq := streamClient.sentReqs[0]
		require.NotNil(t, firstReq.GetMetadata())
		assert.Equal(t, "Тестовый файл", firstReq.GetMetadata().GetName())
		assert.Equal(t, "upload_sample.txt", firstReq.GetMetadata().GetFilename())
		assert.Equal(t, int64(len(testData)), firstReq.GetMetadata().GetSize())
		assert.Equal(t, expectedChecksum, firstReq.GetMetadata().GetChecksum())

		// Проверяем чанк во втором сообщении
		secondReq := streamClient.sentReqs[1]
		assert.Equal(t, testData, secondReq.GetChunk())
	})
}

func TestApp_FileDownloadStreaming(t *testing.T) {
	tmpDir := t.TempDir()
	downloadSavePath := filepath.Join(tmpDir, "nested", "saved_sample.txt")
	testData := []byte("downloaded binary stream chunk payload")

	hasher := sha256.New()
	hasher.Write(testData)
	expectedChecksum := hex.EncodeToString(hasher.Sum(nil))

	fName := "Скачиваемый файл"
	fFilename := "saved_sample.txt"
	fSize := int64(len(testData))

	metaMsg := &pbFile.DownloadFileResponse{}
	metaMsg.SetMetadata(pbFile.FileMetadata_builder{
		Name:     &fName,
		Filename: &fFilename,
		Size:     &fSize,
		Checksum: &expectedChecksum,
	}.Build())

	chunkMsg := &pbFile.DownloadFileResponse{}
	chunkMsg.SetChunk(testData)

	mockClient := &mockUsersClient{
		fileDownloadFunc: func(ctx context.Context, in *pbFile.DownloadFileRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[pbFile.DownloadFileResponse], error) {
			assert.Equal(t, "f123", in.GetId())
			return &mockDownloadStreamClient{
				msgs: []*pbFile.DownloadFileResponse{metaMsg, chunkMsg},
			}, nil
		},
	}

	app := NewApp(mockClient)
	app.session = &Session{Token: "test-token", UserID: "uid-1", Login: "user1"}
	fId := "f123"
	app.currentFile = pbFile.File_builder{
		Id:       &fId,
		Name:     &fName,
		Filename: &fFilename,
		Size:     &fSize,
		Checksum: &expectedChecksum,
	}.Build()

	t.Run("negative - скачивание несуществующего файла", func(t *testing.T) {
		app.downloadSavePathField.SetText("")
		app.handleFileDownload()
		assert.True(t, app.pages.HasPage("error_modal"))
		app.pages.RemovePage("error_modal")
	})

	t.Run("positive - скачивание файла", func(t *testing.T) {
		app.downloadSavePathField.SetText(downloadSavePath)
		app.handleFileDownload()

		assert.False(t, app.pages.HasPage("error_modal"))
		assert.True(t, app.pages.HasPage("info_modal"))

		// Проверяем, что файл действительно записан на диск и его содержимое совпадает
		savedContent, err := os.ReadFile(downloadSavePath)
		require.NoError(t, err)
		assert.Equal(t, testData, savedContent)
	})
}

func TestApp_FilePagination(t *testing.T) {
	requestedPages := make([]string, 0)
	mockClient := &mockUsersClient{
		fileListFunc: func(ctx context.Context, in *api.ListRequest, opts ...grpc.CallOption) (*pbFile.ListResponse, error) {
			page := in.GetPage()
			requestedPages = append(requestedPages, page)

			resp := &pbFile.ListResponse{}
			if page == "1" || page == "" {
				fName := "File_Page1"
				fFilename := "file1.txt"
				resp.SetFiles([]*pbFile.File{
					pbFile.File_builder{Name: &fName, Filename: &fFilename}.Build(),
				})
			} else if page == "2" {
				fName := "File_Page2"
				fFilename := "file2.txt"
				resp.SetFiles([]*pbFile.File{
					pbFile.File_builder{Name: &fName, Filename: &fFilename}.Build(),
				})
			} else {
				resp.SetFiles([]*pbFile.File{})
			}
			return resp, nil
		},
	}

	app := NewApp(mockClient)
	app.session = &Session{Token: "test-token", UserID: "uid-1", Login: "user1"}

	t.Run("positive - проверка пагинации. Загрузка списка файлов", func(t *testing.T) {
		requestedPages = nil
		app.loadFiles()
		assert.Equal(t, 1, app.GetCurrentFilePage())
		require.Len(t, requestedPages, 1)
		assert.Equal(t, "1", requestedPages[0])
		assert.Len(t, app.currentFiles, 1)
		assert.Equal(t, "File_Page1", app.currentFiles[0].GetName())
	})

	t.Run("negative - проверка пагинации. Кнопка предыдущей страницы должна быть заблокирована", func(t *testing.T) {
		requestedPages = nil
		app.prevFilePage()
		assert.Equal(t, 1, app.GetCurrentFilePage())
		assert.Empty(t, requestedPages)
	})

	t.Run("positive - проверка пагинации. Переход к следующей странице", func(t *testing.T) {
		requestedPages = nil
		app.nextFilePage()
		assert.Equal(t, 2, app.GetCurrentFilePage())
		require.Len(t, requestedPages, 1)
		assert.Equal(t, "2", requestedPages[0])
		assert.Len(t, app.currentFiles, 1)
		assert.Equal(t, "File_Page2", app.currentFiles[0].GetName())
	})

	t.Run("positive - проверка пагинации. Переход к предыдущей странице", func(t *testing.T) {
		requestedPages = nil
		app.prevFilePage()
		assert.Equal(t, 1, app.GetCurrentFilePage())
		require.Len(t, requestedPages, 1)
		assert.Equal(t, "1", requestedPages[0])
		assert.Len(t, app.currentFiles, 1)
		assert.Equal(t, "File_Page1", app.currentFiles[0].GetName())
	})

	t.Run("positive - проверка пагинации. Сброс номера страницы при переходе на форму отображения списка", func(t *testing.T) {
		app.SetCurrentFilePage(5)
		requestedPages = nil
		app.openFileList()
		assert.Equal(t, 1, app.GetCurrentFilePage())
		require.Len(t, requestedPages, 1)
		assert.Equal(t, "1", requestedPages[0])
	})

	t.Run("positive - загрузка списка файлов", func(t *testing.T) {
		errMockClient := &mockUsersClient{
			fileListFunc: func(ctx context.Context, in *api.ListRequest, opts ...grpc.CallOption) (*pbFile.ListResponse, error) {
				return nil, status.Error(codes.Internal, "storage error")
			},
		}
		errApp := NewApp(errMockClient)
		errApp.session = &Session{Token: "test-token", UserID: "uid-1", Login: "user1"}
		errApp.loadFiles()
		assert.True(t, errApp.pages.HasPage("error_modal"))
	})
}
