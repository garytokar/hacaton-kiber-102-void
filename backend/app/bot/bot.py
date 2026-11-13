import asyncio
import logging
from typing import Optional
from aiogram import Bot, Dispatcher, types, F
from aiogram.filters import Command
from aiogram.fsm.context import FSMContext
from aiogram.fsm.state import State, StatesGroup
from aiogram.types import InlineKeyboardMarkup, InlineKeyboardButton
from sqlalchemy import select
from sqlalchemy.orm import selectinload

from app.core.config import settings
from app.core.database import AsyncSessionLocal
from app.models.models import User, Fragment, Translation, TranslationVote, UserStats
from app.services.gamification import (
    add_xp_and_currency, update_user_streak, get_level_info, update_quest_progress
)
from app.services.moderation import evaluate_translation_status

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

class TranslationFSM(StatesGroup):
    waiting_for_correction = State()
    current_fragment_id = State()
    current_translation_id = State()

dp = Dispatcher()

@dp.message(Command("start"))
async def cmd_start(message: types.Message):
    welcome_text = (
        "✨ *Һаумыһығыҙ! Добро пожаловать в бота платформы «ЙӘНЛӘНДЕР»!* ✨\n\n"
        "🏛 *Миссия:* Сохраняй язык. Создавай наследие.\n"
        "Здесь вы можете ежедневно изучать башкирский язык, проверять переводы и развивать цифровой корпус.\n\n"
        "Доступные команды:\n"
        "📝 /task — Получить ежедневное задание по переводу\n"
        "👤 /profile — Посмотреть ваш уровень, XP и серию дней\n"
        "🏆 /rating — Таблица лидеров Хранителей языка\n"
        "ℹ️ /help — Справка по проекту\n"
    )
    await message.answer(welcome_text, parse_mode="Markdown")

@dp.message(Command("help"))
async def cmd_help(message: types.Message):
    help_text = (
        "📖 *Как работает бот ЙӘНЛӘНДЕР:*\n\n"
        "1. Нажмите /task, чтобы получить башкирскую пословицу или строфу эпоса.\n"
        "2. Оцените предложенный перевод:\n"
        "   • Нажмите «✅ Подтвердить», если перевод точен (+5 XP).\n"
        "   • Нажмите «✏️ Предложить свой вариант», чтобы отправить уточнение (+8 XP).\n"
        "3. Ваш вклад сохраняется в единой базе знаний платформы.\n"
        "4. Поддерживайте ежедневную серию дней (streak)!\n"
    )
    await message.answer(help_text, parse_mode="Markdown")

@dp.message(Command("profile"))
async def cmd_profile(message: types.Message):
    async with AsyncSessionLocal() as db:
        # For demo in bot, query demo student or linked user
        res = await db.execute(select(User).options(selectinload(User.stats)).limit(1))
        user = res.scalar_one_or_none()
        if not user or not user.stats:
            await message.answer("Профиль пока не привязан. Воспользуйтесь платформой для создания профиля.")
            return

        lvl, title, next_xp = get_level_info(user.stats.xp)
        profile_text = (
            f"👤 *Профиль Хранителя: {user.username}*\n"
            f"🏅 Уровень: {lvl} — *{title}*\n"
            f"⚡ Опыт: {user.stats.xp} / {next_xp} XP\n"
            f"🔥 Серия активности: {user.stats.current_streak} дней (Лучшая: {user.stats.best_streak})\n"
            f"🪙 Валюта «Ҡурай»: {user.stats.currency}\n"
            f"✅ Подтверждённых переводов: {user.stats.total_translations_confirmed}\n"
        )
        await message.answer(profile_text, parse_mode="Markdown")

@dp.message(Command("rating"))
async def cmd_rating(message: types.Message):
    async with AsyncSessionLocal() as db:
        res = await db.execute(
            select(User).join(User.stats).options(selectinload(User.stats)).order_by(UserStats.xp.desc()).limit(5)
        )
        users = res.scalars().all()

        text = "🏆 *Топ-5 Хранителей башкирского языка:*\n\n"
        for idx, u in enumerate(users, start=1):
            lvl, title, _ = get_level_info(u.stats.xp)
            text += f"{idx}. *{u.username}* — {u.stats.xp} XP ({title}, серия: {u.stats.current_streak} дн.)\n"

        await message.answer(text, parse_mode="Markdown")

@dp.message(Command("task"))
async def cmd_task(message: types.Message, state: FSMContext):
    async with AsyncSessionLocal() as db:
        stmt = select(Fragment).options(
            selectinload(Fragment.translations)
        ).order_by(Fragment.id)
        res = await db.execute(stmt)
        fragments = res.scalars().all()
        if not fragments:
            await message.answer("Задания пока формируются. Попробуйте чуть позже!")
            return

        frag = fragments[0]
        translation = frag.translations[0] if frag.translations else None

        await state.update_data(fragment_id=frag.id, translation_id=translation.id if translation else None)

        msg_text = (
            f"📝 *Ежедневное задание:*\n\n"
            f"📜 *Текст на башкирском:*\n«{frag.text_ba}»\n\n"
            f"📚 *Источник:* {frag.source_title}\n"
            f"🏷 *Категория:* {frag.category}\n\n"
            f"🔍 *Предложенный перевод:*\n"
            f"«{translation.translation_text if translation else 'Перевод пока не добавлен'}»\n\n"
            f"Оцените корректность предложенного перевода:"
        )

        kb = InlineKeyboardMarkup(inline_keyboard=[
            [
                InlineKeyboardButton(text="✅ Перевод правильный (+5 XP)", callback_data=f"confirm_{translation.id if translation else 0}"),
            ],
            [
                InlineKeyboardButton(text="✏️ Предложить исправление (+8 XP)", callback_data=f"propose_{frag.id}")
            ]
        ])
        await message.answer(msg_text, reply_markup=kb, parse_mode="Markdown")

@dp.callback_query(F.data.startswith("confirm_"))
async def callback_confirm(callback: types.CallbackQuery):
    t_id = int(callback.data.split("_")[1])
    if t_id == 0:
        await callback.answer("Перевод отсутствует", show_alert=True)
        return

    async with AsyncSessionLocal() as db:
        res = await db.execute(select(Translation).where(Translation.id == t_id))
        trans = res.scalar_one_or_none()
        if not trans:
            await callback.answer("Перевод не найден", show_alert=True)
            return

        # Top user for demo in bot
        u_res = await db.execute(select(User).limit(1))
        user = u_res.scalar_one()

        await evaluate_translation_status(db, t_id, user.id)
        stats, level_up, new_level = await add_xp_and_currency(db, user.id, 5, 2)
        await db.commit()

        congrats = "🎉 *Голос учтён! +5 XP, +2 Ҡурай!*\nВы помогли подтвердить перевод для Сокровищницы языка."
        if level_up:
            congrats += f"\n\n🎊 *Поздравляем! Вы достигли {new_level} уровня!*"

        await callback.message.edit_text(congrats, parse_mode="Markdown")
        await callback.answer()

@dp.callback_query(F.data.startswith("propose_"))
async def callback_propose(callback: types.CallbackQuery, state: FSMContext):
    f_id = int(callback.data.split("_")[1])
    await state.set_state(TranslationFSM.waiting_for_correction)
    await state.update_data(current_fragment_id=f_id)
    await callback.message.answer(
        "✍️ Напишите ваш вариант перевода в ответном сообщении:\n(Или отправьте /cancel для отмены)"
    )
    await callback.answer()

@dp.message(TranslationFSM.waiting_for_correction)
async def process_correction(message: types.Message, state: FSMContext):
    if message.text and message.text.startswith("/"):
        await state.clear()
        await message.answer("Действие отменено.")
        return

    data = await state.get_data()
    f_id = data.get("current_fragment_id")

    async with AsyncSessionLocal() as db:
        u_res = await db.execute(select(User).limit(1))
        user = u_res.scalar_one()

        new_t = Translation(
            fragment_id=f_id,
            translation_text=message.text.strip(),
            status="pending",
            created_by=user.id,
            votes_count=1,
            reject_count=0
        )
        db.add(new_t)
        await db.flush()

        stats, level_up, new_level = await add_xp_and_currency(db, user.id, 8, 5)
        await db.commit()

        await state.clear()
        reply = (
            f"✨ *Ваш перевод сохранён и отправлен на голосование!*\n"
            f"Начислено: *+8 XP, +5 Ҡурай*\n"
            f"Спасибо за обогащение башкирского цифрового корпуса!"
        )
        await message.answer(reply, parse_mode="Markdown")

async def run_bot():
    if not settings.TELEGRAM_BOT_TOKEN:
        logger.warning("TELEGRAM_BOT_TOKEN не задан. Бот готов к запуску при наличии токена.")
        return
    bot = Bot(token=settings.TELEGRAM_BOT_TOKEN)
    logger.info("Запуск Telegram-бота ЙӘНЛӘНДЕР...")
    await dp.start_polling(bot)

if __name__ == "__main__":
    asyncio.run(run_bot())
