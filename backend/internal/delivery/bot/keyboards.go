package bot

import (
	tele "gopkg.in/telebot.v3"
)

type Keyboards struct {
	MainMenu     *tele.ReplyMarkup
	TaskControls *tele.ReplyMarkup
	BtnConfirm   tele.Btn
	BtnEdit      tele.Btn
	BtnReport    tele.Btn
	BtnCancel    tele.Btn
}

func NewKeyboards(webAppURL string) *Keyboards {
	k := &Keyboards{}

	// Главное меню с кнопкой запуска WebApp
	k.MainMenu = &tele.ReplyMarkup{ResizeKeyboard: true}
	btnWebApp := k.MainMenu.WebApp("🚀 Открыть приложение (Duolingo)", &tele.WebApp{URL: webAppURL})
	btnDaily := k.MainMenu.Text("📝 Тренировка в чате")
	btnStats := k.MainMenu.Text("📊 Моя статистика")
	btnRating := k.MainMenu.Text("🏆 Рейтинг школ")

	k.MainMenu.Reply(
		k.MainMenu.Row(btnWebApp),
		k.MainMenu.Row(btnDaily, btnStats),
		k.MainMenu.Row(btnRating),
	)

	// Управление карточкой задания (Inline кнопки)
	k.TaskControls = &tele.ReplyMarkup{}
	k.BtnConfirm = k.TaskControls.Data("✅ Перевод верен", "confirm_task")
	k.BtnEdit = k.TaskControls.Data("✏️ Предложить исправление", "edit_task")
	k.BtnReport = k.TaskControls.Data("⚠️ Сообщить об ошибке", "report_task")
	k.BtnCancel = k.TaskControls.Data("❌ Завершить", "cancel_task")

	k.TaskControls.Inline(
		k.TaskControls.Row(k.BtnConfirm),
		k.TaskControls.Row(k.BtnEdit),
		k.TaskControls.Row(k.BtnReport, k.BtnCancel),
	)

	return k
}
