// Локальное кэширование данных клиента в зашифрованном виде
package client

import (
	"errors"
	"fmt"

	"github.com/alexsey-popov/gkeep/internal/cache"
	"github.com/alexsey-popov/gkeep/internal/models"
	"github.com/alexsey-popov/gkeep/proto"
	pbCard "github.com/alexsey-popov/gkeep/proto/card"
	pbFile "github.com/alexsey-popov/gkeep/proto/file"
	pbPassword "github.com/alexsey-popov/gkeep/proto/password"
	pbText "github.com/alexsey-popov/gkeep/proto/text"
)

// getCacheKey генерирует 32-байтный ключ шифрования для текущей сессии
func (a *App) getCacheKey() []byte {
	userID := ""
	login := ""
	if a.session != nil {
		userID = a.session.UserID
		login = a.session.Login
	}
	return cache.GetKey(userID, login)
}

// getCacheFilePath возвращает путь к файлу кэша
func (a *App) getCacheFilePath() string {
	if a.cacheFilePath != "" {
		return a.cacheFilePath
	}
	return ".gkeep_cache"
}

// SetCacheFilePath устанавливает путь к файлу кэша (используется для тестов или кастомной конфигурации)
func (a *App) SetCacheFilePath(path string) {
	a.cacheFilePath = path
}

// getCache загружает или инициализирует кэш для текущего пользователя
func (a *App) getCache() (*cache.Cache, error) {
	a.cacheMu.Lock()
	defer a.cacheMu.Unlock()

	c, err := cache.New(a.getCacheFilePath(), a.getCacheKey())
	if err != nil {
		return nil, fmt.Errorf("ошибка инициализации кэша: %w", err)
	}

	if err := c.Load(); err != nil {
		return nil, fmt.Errorf("ошибка загрузки кэша: %w", err)
	}

	return c, nil
}

// saveCachedPasswords сохраняет страницу паролей в кэш
func (a *App) saveCachedPasswords(page int, items []*pbPassword.Password) error {
	c, err := a.getCache()
	if err != nil {
		return err
	}

	list := make([]models.Password, 0, len(items))
	for _, item := range items {
		if item != nil {
			list = append(list, models.Password{
				Id:       item.GetId(),
				Name:     item.GetName(),
				Login:    item.GetLogin(),
				Password: item.GetPassword(),
			})
		}
	}

	c.SetPasswords(page, list)
	return c.Save()
}

// loadCachedPasswords загружает страницу паролей из кэша
func (a *App) loadCachedPasswords(page int) ([]*pbPassword.Password, error) {
	c, err := a.getCache()
	if err != nil {
		return nil, err
	}

	items, ok := c.GetPasswords(page)
	if !ok {
		return nil, errors.New("кэш не найден")
	}

	res := make([]*pbPassword.Password, 0, len(items))
	for _, item := range items {
		itemCopy := item
		res = append(res, proto.ConvertPassword(&itemCopy))
	}

	return res, nil
}

// saveCachedCards сохраняет страницу банковских карт в кэш
func (a *App) saveCachedCards(page int, items []*pbCard.Card) error {
	c, err := a.getCache()
	if err != nil {
		return err
	}

	list := make([]models.Card, 0, len(items))
	for _, item := range items {
		if item != nil {
			list = append(list, models.Card{
				Id:      item.GetId(),
				Name:    item.GetName(),
				Number:  item.GetNumber(),
				Expired: item.GetExpired(),
				Code:    item.GetCode(),
			})
		}
	}

	c.SetCards(page, list)
	return c.Save()
}

// loadCachedCards загружает страницу банковских карт из кэша
func (a *App) loadCachedCards(page int) ([]*pbCard.Card, error) {
	c, err := a.getCache()
	if err != nil {
		return nil, err
	}

	items, ok := c.GetCards(page)
	if !ok {
		return nil, errors.New("кэш не найден")
	}

	res := make([]*pbCard.Card, 0, len(items))
	for _, item := range items {
		itemCopy := item
		res = append(res, proto.ConvertCard(&itemCopy))
	}

	return res, nil
}

// saveCachedTexts сохраняет страницу текстов в кэш
func (a *App) saveCachedTexts(page int, items []*pbText.Text) error {
	c, err := a.getCache()
	if err != nil {
		return err
	}

	list := make([]models.Text, 0, len(items))
	for _, item := range items {
		if item != nil {
			list = append(list, models.Text{
				Id:   item.GetId(),
				Name: item.GetName(),
				Text: item.GetText(),
			})
		}
	}

	c.SetTexts(page, list)
	return c.Save()
}

// loadCachedTexts загружает страницу текстов из кэша
func (a *App) loadCachedTexts(page int) ([]*pbText.Text, error) {
	c, err := a.getCache()
	if err != nil {
		return nil, err
	}

	items, ok := c.GetTexts(page)
	if !ok {
		return nil, errors.New("кэш не найден")
	}

	res := make([]*pbText.Text, 0, len(items))
	for _, item := range items {
		itemCopy := item
		res = append(res, proto.ConvertText(&itemCopy))
	}

	return res, nil
}

// saveCachedFiles сохраняет страницу метаданных файлов в кэш
func (a *App) saveCachedFiles(page int, items []*pbFile.File) error {
	c, err := a.getCache()
	if err != nil {
		return err
	}

	list := make([]models.File, 0, len(items))
	for _, item := range items {
		if item != nil {
			list = append(list, models.File{
				Id:       item.GetId(),
				Name:     item.GetName(),
				Filename: item.GetFilename(),
				Size:     item.GetSize(),
				Checksum: item.GetChecksum(),
			})
		}
	}

	c.SetFiles(page, list)
	return c.Save()
}

// loadCachedFiles загружает страницу метаданных файлов из кэша
func (a *App) loadCachedFiles(page int) ([]*pbFile.File, error) {
	c, err := a.getCache()
	if err != nil {
		return nil, err
	}

	items, ok := c.GetFiles(page)
	if !ok {
		return nil, errors.New("кэш не найден")
	}

	res := make([]*pbFile.File, 0, len(items))
	for _, item := range items {
		itemCopy := item
		res = append(res, proto.ConvertFile(&itemCopy))
	}

	return res, nil
}
