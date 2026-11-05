package client

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	api "github.com/alexsey-popov/gkeep/proto"
	pbCard "github.com/alexsey-popov/gkeep/proto/card"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"google.golang.org/grpc/status"
)

// initCardsListView создает экран со списком банковских карт
func (a *App) initCardsListView() {
	a.cardsList = tview.NewList().
		ShowSecondaryText(true)

	openBtn := tview.NewButton("Открыть").
		SetSelectedFunc(func() {
			idx := a.cardsList.GetCurrentItem()
			if idx >= 0 && idx < len(a.currentCards) {
				a.openCardView(a.currentCards[idx])
			}
		})

	addBtn := tview.NewButton("Добавить").
		SetSelectedFunc(func() {
			a.clearCreateCardFields()
			a.pages.SwitchToPage("card_create")
		})

	a.prevCardPageBtn = tview.NewButton("< Пред.").
		SetSelectedFunc(a.prevCardPage)

	a.nextCardPageBtn = tview.NewButton("След. >").
		SetSelectedFunc(a.nextCardPage)

	refreshBtn := tview.NewButton("Обновить").
		SetSelectedFunc(func() {
			a.loadCards()
		})

	backBtn := tview.NewButton("Назад").
		SetSelectedFunc(func() {
			a.pages.SwitchToPage("data_types")
		})

	btnFlex := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(openBtn, 10, 1, true).
		AddItem(nil, 1, 0, false).
		AddItem(addBtn, 11, 1, false).
		AddItem(nil, 1, 0, false).
		AddItem(a.prevCardPageBtn, 10, 1, false).
		AddItem(nil, 1, 0, false).
		AddItem(a.nextCardPageBtn, 10, 1, false).
		AddItem(nil, 1, 0, false).
		AddItem(refreshBtn, 11, 1, false).
		AddItem(nil, 1, 0, false).
		AddItem(backBtn, 8, 1, false)

	a.cardsFlex = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(a.cardsList, 0, 1, true).
		AddItem(btnFlex, 3, 0, false)

	a.cardsFlex.SetBorder(true).
		SetTitle(a.getCardsListTitle()).
		SetTitleAlign(tview.AlignCenter).
		SetBorderColor(tcell.ColorCadetBlue)
}

// initCardDetailView создает экран просмотра отдельной банковской карты
func (a *App) initCardDetailView() {
	a.cardDetailsText = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft).
		SetWrap(true)

	editBtn := tview.NewButton("Изменить").
		SetSelectedFunc(func() {
			if a.currentCard != nil {
				a.openCardEdit(a.currentCard)
			}
		})

	deleteBtn := tview.NewButton("Удалить").
		SetSelectedFunc(func() {
			if a.currentCard != nil {
				a.showDeleteCardConfirmModal()
			}
		})

	backBtn := tview.NewButton("Назад").
		SetSelectedFunc(func() {
			a.pages.SwitchToPage("card_list")
		})

	btnFlex := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(nil, 0, 1, false).
		AddItem(editBtn, 14, 1, true).
		AddItem(nil, 2, 0, false).
		AddItem(deleteBtn, 14, 1, false).
		AddItem(nil, 2, 0, false).
		AddItem(backBtn, 12, 1, false).
		AddItem(nil, 0, 1, false)

	a.cardViewFlex = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(a.cardDetailsText, 0, 1, false).
		AddItem(btnFlex, 3, 0, true)

	a.cardViewFlex.SetBorder(true).
		SetTitle(" [::b]Просмотр карты[::-] ").
		SetTitleAlign(tview.AlignCenter).
		SetBorderColor(tcell.ColorDarkGoldenrod)
}

// initCardCreateForm создает форму добавления банковской карты
func (a *App) initCardCreateForm() {
	a.createCardNameField = tview.NewInputField().
		SetLabel("Название: ").
		SetFieldWidth(32)

	a.createCardNumberField = tview.NewInputField().
		SetLabel("Номер карты: ").
		SetFieldWidth(32)

	a.createCardExpiredField = tview.NewInputField().
		SetLabel("Срок действия: ").
		SetFieldWidth(32)

	a.createCardCodeField = tview.NewInputField().
		SetLabel("Код (CVV): ").
		SetFieldWidth(32)

	a.cardCreateForm = tview.NewForm().
		AddFormItem(a.createCardNameField).
		AddFormItem(a.createCardNumberField).
		AddFormItem(a.createCardExpiredField).
		AddFormItem(a.createCardCodeField).
		AddButton("Сохранить", a.handleCardCreate).
		AddButton("Назад", func() {
			a.clearCreateCardFields()
			a.pages.SwitchToPage("card_list")
		})

	a.cardCreateForm.SetBorder(true).
		SetTitle(" [::b]Добавление карты[::-] ").
		SetTitleAlign(tview.AlignCenter).
		SetBorderColor(tcell.ColorGreen)
}

// initCardEditForm создает форму редактирования банковской карты
func (a *App) initCardEditForm() {
	a.editCardIDField = tview.NewInputField().
		SetLabel("ID: ").
		SetFieldWidth(32)
	a.editCardIDField.SetDisabled(true)

	a.editCardNameField = tview.NewInputField().
		SetLabel("Название: ").
		SetFieldWidth(32)

	a.editCardNumberField = tview.NewInputField().
		SetLabel("Номер карты: ").
		SetFieldWidth(32)

	a.editCardExpiredField = tview.NewInputField().
		SetLabel("Срок действия: ").
		SetFieldWidth(32)

	a.editCardCodeField = tview.NewInputField().
		SetLabel("Код (CVV): ").
		SetFieldWidth(32)

	a.cardEditForm = tview.NewForm().
		AddFormItem(a.editCardIDField).
		AddFormItem(a.editCardNameField).
		AddFormItem(a.editCardNumberField).
		AddFormItem(a.editCardExpiredField).
		AddFormItem(a.editCardCodeField).
		AddButton("Сохранить", a.handleCardUpdate).
		AddButton("Назад", func() {
			a.pages.SwitchToPage("card_view")
		})

	a.cardEditForm.SetBorder(true).
		SetTitle(" [::b]Редактирование карты[::-] ").
		SetTitleAlign(tview.AlignCenter).
		SetBorderColor(tcell.ColorOrange)
}

// openCardList загружает список банковских карт и переключает на экран списка
func (a *App) openCardList() {
	a.currentCardPage = 1
	a.loadCards()
	a.pages.SwitchToPage("card_list")
}

// loadCards выполняет запрос к серверу для получения списка банковских карт пользователя
func (a *App) loadCards() {
	if a.session == nil || a.session.Token == "" {
		a.showErrorModal("Сессия не авторизована. Пожалуйста, выполните вход.", "login")
		return
	}

	if a.currentCardPage < 1 {
		a.currentCardPage = 1
	}

	ctx, cancel := a.authCtx(context.Background())
	defer cancel()

	req := &api.ListRequest{}
	req.SetPage(strconv.Itoa(a.currentCardPage))

	resp, err := a.usersClient.CardList(ctx, req)
	if err != nil {
		st := status.Convert(err)
		a.showErrorModal(fmt.Sprintf("Ошибка получения списка карт:\n%s", st.Message()), "data_types")
		return
	}

	a.currentCards = resp.GetCard()
	a.cardsList.Clear()
	a.cardsFlex.SetTitle(a.getCardsListTitle())

	if len(a.currentCards) == 0 {
		if a.currentCardPage > 1 {
			a.cardsList.AddItem("[gray]На этой странице нет записей[-]", "Нажмите кнопку '< Пред.' для возврата", 0, nil)
		} else {
			a.cardsList.AddItem("[gray]Нет сохраненных карт[-]", "Нажмите кнопку 'Добавить' для создания записи", 0, nil)
		}
		return
	}

	for i, item := range a.currentCards {
		itemCopy := item
		shortcut := rune(0)
		if i < 9 {
			shortcut = rune('1' + i)
		} else if i == 9 {
			shortcut = '0'
		}
		a.cardsList.AddItem(
			itemCopy.GetName(),
			"",
			shortcut,
			func() {
				a.openCardView(itemCopy)
			},
		)
	}
}

// openCardView открывает экран подробного просмотра банковской карты
func (a *App) openCardView(c *pbCard.Card) {
	if c == nil {
		return
	}
	a.currentCard = c

	info := fmt.Sprintf(
		"\n  [yellow]Название:[white]      %s\n\n"+
			"  [yellow]Номер карты:[white]   %s\n\n"+
			"  [yellow]Срок действия:[white] %s\n\n"+
			"  [yellow]Код (CVV):[white]     %s\n",
		c.GetName(),
		c.GetNumber(),
		c.GetExpired(),
		c.GetCode(),
	)

	a.cardDetailsText.SetText(info)
	a.pages.SwitchToPage("card_view")
}

// openCardEdit открывает экран редактирования банковской карты
func (a *App) openCardEdit(c *pbCard.Card) {
	if c == nil {
		return
	}
	a.currentCard = c
	a.editCardIDField.SetText(c.GetId())
	a.editCardNameField.SetText(c.GetName())
	a.editCardNumberField.SetText(c.GetNumber())
	a.editCardExpiredField.SetText(c.GetExpired())
	a.editCardCodeField.SetText(c.GetCode())
	a.pages.SwitchToPage("card_edit")
}

// handleCardCreate создает новую запись банковской карты
func (a *App) handleCardCreate() {
	name := strings.TrimSpace(a.createCardNameField.GetText())
	number := strings.TrimSpace(a.createCardNumberField.GetText())
	expired := strings.TrimSpace(a.createCardExpiredField.GetText())
	code := strings.TrimSpace(a.createCardCodeField.GetText())

	if name == "" {
		a.showErrorModal("Название записи не может быть пустым", "card_create")
		return
	}

	if a.session == nil || a.session.Token == "" {
		a.showErrorModal("Сессия не авторизована. Пожалуйста, выполните вход.", "login")
		return
	}

	ctx, cancel := a.authCtx(context.Background())
	defer cancel()

	req := &pbCard.CreateRequest{}
	req.SetName(name)
	req.SetNumber(number)
	req.SetExpired(expired)
	req.SetCode(code)

	_, err := a.usersClient.CardCreate(ctx, req)
	if err != nil {
		st := status.Convert(err)
		a.showErrorModal(fmt.Sprintf("Ошибка создания записи:\n%s", st.Message()), "card_create")
		return
	}

	a.clearCreateCardFields()
	a.currentCardPage = 1
	a.loadCards()
	a.pages.SwitchToPage("card_list")
}

// handleCardUpdate обновляет данные выбранной банковской карты
func (a *App) handleCardUpdate() {
	id := strings.TrimSpace(a.editCardIDField.GetText())
	name := strings.TrimSpace(a.editCardNameField.GetText())
	number := strings.TrimSpace(a.editCardNumberField.GetText())
	expired := strings.TrimSpace(a.editCardExpiredField.GetText())
	code := strings.TrimSpace(a.editCardCodeField.GetText())

	if name == "" {
		a.showErrorModal("Название не может быть пустым", "card_edit")
		return
	}

	if a.session == nil || a.session.Token == "" {
		a.showErrorModal("Сессия не авторизована. Пожалуйста, выполните вход.", "login")
		return
	}

	ctx, cancel := a.authCtx(context.Background())
	defer cancel()

	req := &pbCard.UpdateRequest{}
	req.SetId(id)
	req.SetName(name)
	req.SetNumber(number)
	req.SetExpired(expired)
	req.SetCode(code)

	resp, err := a.usersClient.CardUpdate(ctx, req)
	if err != nil {
		st := status.Convert(err)
		a.showErrorModal(fmt.Sprintf("Ошибка обновления записи:\n%s", st.Message()), "card_edit")
		return
	}

	if resp.GetCard() != nil {
		a.currentCard = resp.GetCard()
	} else if a.currentCard != nil {
		a.currentCard.SetName(name)
		a.currentCard.SetNumber(number)
		a.currentCard.SetExpired(expired)
		a.currentCard.SetCode(code)
	}

	a.openCardView(a.currentCard)
}

// showDeleteCardConfirmModal показывает окно подтверждения удаления банковской карты
func (a *App) showDeleteCardConfirmModal() {
	if a.currentCard == nil {
		return
	}

	name := a.currentCard.GetName()
	modal := tview.NewModal().
		SetText(fmt.Sprintf("Вы уверены, что хотите удалить карту '%s'?", name)).
		AddButtons([]string{"Удалить", "Отмена"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			a.pages.RemovePage("confirm_delete_card_modal")
			if buttonLabel == "Удалить" {
				a.handleCardDelete()
			} else {
				a.pages.SwitchToPage("card_view")
			}
		})

	a.pages.AddPage("confirm_delete_card_modal", modal, true, true)
}

// handleCardDelete отправляет запрос на удаление банковской карты
func (a *App) handleCardDelete() {
	if a.currentCard == nil {
		return
	}

	if a.session == nil || a.session.Token == "" {
		a.showErrorModal("Сессия не авторизована. Пожалуйста, выполните вход.", "login")
		return
	}

	ctx, cancel := a.authCtx(context.Background())
	defer cancel()

	req := &api.IDRequest{}
	req.SetId(a.currentCard.GetId())

	_, err := a.usersClient.CardDelete(ctx, req)
	if err != nil {
		st := status.Convert(err)
		a.showErrorModal(fmt.Sprintf("Ошибка удаления записи:\n%s", st.Message()), "card_view")
		return
	}

	a.currentCard = nil
	a.loadCards()
	if len(a.currentCards) == 0 && a.currentCardPage > 1 {
		a.currentCardPage--
		a.loadCards()
	}
	a.pages.SwitchToPage("card_list")
}

// prevCardPage переключает на предыдущую страницу списка карт
func (a *App) prevCardPage() {
	if a.currentCardPage > 1 {
		a.currentCardPage--
		a.loadCards()
	}
}

// nextCardPage переключает на следующую страницу списка карт
func (a *App) nextCardPage() {
	if len(a.currentCards) > 0 {
		a.currentCardPage++
		a.loadCards()
	}
}

// GetCurrentCardPage возвращает номер текущей страницы списка карт
func (a *App) GetCurrentCardPage() int {
	return a.currentCardPage
}

// SetCurrentCardPage устанавливает номер текущей страницы списка карт
func (a *App) SetCurrentCardPage(p int) {
	if p < 1 {
		p = 1
	}
	a.currentCardPage = p
}

func (a *App) getCardsListTitle() string {
	if a.currentCardPage <= 1 {
		return " [::b]Список банковских карт (Стр. 1)[::-] "
	}
	return fmt.Sprintf(" [::b]Список банковских карт (Стр. %d)[::-] ", a.currentCardPage)
}

// clearCreateCardFields очищает поля формы создания карты
func (a *App) clearCreateCardFields() {
	a.createCardNameField.SetText("")
	a.createCardNumberField.SetText("")
	a.createCardExpiredField.SetText("")
	a.createCardCodeField.SetText("")
}
