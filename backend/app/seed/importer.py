import os
import json
import logging
from typing import List, Dict, Any, Optional
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy import select
from app.models.models import Fragment, Translation, LanguagePair, User

logger = logging.getLogger("importer")

POSSIBLE_DATA_DIRS = [
    "data",
    "backend/data",
    "../data",
    os.path.join(os.path.dirname(__file__), "..", "..", "data"),
]

def find_data_file(filename: str) -> Optional[str]:
    for base in POSSIBLE_DATA_DIRS:
        p = os.path.normpath(os.path.join(base, filename))
        if os.path.isfile(p):
            return p
    return None

async def import_all_datasets(db: AsyncSession) -> Dict[str, int]:
    """
    Imports all datasets from git repository (data-ba-ru-kk.jsonl, hf_phrasebook.json, audiobooks_brsbs.jsonl, texts_samples.jsonl)
    into the database, avoiding duplicates.
    """
    stats = {
        "phrasebook_imported": 0,
        "audiobooks_imported": 0,
        "hf_imported": 0,
        "samples_imported": 0,
    }

    # Ensure LanguagePair 1 exists
    res_pair = await db.execute(select(LanguagePair).limit(1))
    pair = res_pair.scalar_one_or_none()
    pair_id = pair.id if pair else 1

    # Find demo user for attribution
    res_user = await db.execute(select(User).limit(1))
    first_user = res_user.scalar_one_or_none()
    user_id = first_user.id if first_user else None

    # Cache existing fragment texts to prevent duplication
    res_frags = await db.execute(select(Fragment.text_ba))
    existing_texts = set(res_frags.scalars().all())

    # 1. Import Audiobooks (Mustai Karim from Hugging Face)
    audiobooks_path = find_data_file("audiobooks_brsbs.jsonl")
    if audiobooks_path:
        try:
            with open(audiobooks_path, "r", encoding="utf-8") as f:
                for line in f:
                    line = line.strip()
                    if not line:
                        continue
                    item = json.loads(line)
                    text_ba = item.get("text_ba", "").strip()
                    if not text_ba or text_ba in existing_texts:
                        continue

                    frag = Fragment(
                        text_ba=text_ba,
                        source_title=f"{item.get('author', 'Мустай Карим')} «{item.get('book_title', 'Оҙон-оҙаҡ бала саҡ')}»",
                        source_document=item.get("chapter", "Глава 1"),
                        chapter=item.get("chapter"),
                        audio_url=item.get("audio_url"),
                        category="Художественная литература",
                        difficulty="medium"
                    )
                    db.add(frag)
                    await db.flush()
                    existing_texts.add(text_ba)

                    trans_ru = item.get("translation_ru", "").strip()
                    if trans_ru:
                        trans = Translation(
                            fragment_id=frag.id,
                            language_pair_id=pair_id,
                            translation_text=trans_ru,
                            status="confirmed",
                            created_by=user_id,
                            votes_count=3,
                            reject_count=0
                        )
                        db.add(trans)
                    stats["audiobooks_imported"] += 1
        except Exception as e:
            logger.error(f"Error importing audiobooks: {e}")

    # 2. Import Parallel phrasebook dataset (1,397 lines)
    phrasebook_path = find_data_file("data-ba-ru-kk.jsonl")
    if phrasebook_path:
        try:
            with open(phrasebook_path, "r", encoding="utf-8") as f:
                idx = 0
                for line in f:
                    line = line.strip()
                    if not line:
                        continue
                    item = json.loads(line)
                    text_ba = item.get("text_ba", "").strip()
                    if not text_ba or text_ba in existing_texts:
                        continue

                    idx += 1
                    diff = "easy" if len(text_ba) < 35 else "hard" if len(text_ba) > 80 else "medium"
                    frag = Fragment(
                        text_ba=text_ba,
                        source_title="БГПУ им. М. Акмуллы: трехъязычный разговорник",
                        source_document=f"Разговорная речь (параллельный корпус) #{idx}",
                        category="Разговорный язык",
                        difficulty=diff
                    )
                    db.add(frag)
                    await db.flush()
                    existing_texts.add(text_ba)

                    text_ru = item.get("text_ru", "").strip()
                    if text_ru:
                        # 70% confirmed, 30% pending for active user review
                        is_confirmed = (idx % 3 != 0)
                        trans = Translation(
                            fragment_id=frag.id,
                            language_pair_id=pair_id,
                            translation_text=text_ru,
                            status="confirmed" if is_confirmed else "pending",
                            created_by=user_id,
                            votes_count=3 if is_confirmed else 1,
                            reject_count=0
                        )
                        db.add(trans)
                    stats["phrasebook_imported"] += 1
        except Exception as e:
            logger.error(f"Error importing phrasebook: {e}")

    # 3. Import Hugging Face Phrasebook
    hf_path = find_data_file("hf_phrasebook.json")
    if hf_path:
        try:
            with open(hf_path, "r", encoding="utf-8-sig") as f:
                data = json.load(f)
                for item in data:
                    text_ba = item.get("text_ba", "").strip()
                    if not text_ba or text_ba in existing_texts:
                        continue
                    frag = Fragment(
                        text_ba=text_ba,
                        source_title="Hugging Face / bashkorttele",
                        source_document="БГПУ: Trilingual phrasebook",
                        category="Диалоги и фразеология",
                        difficulty="easy"
                    )
                    db.add(frag)
                    await db.flush()
                    existing_texts.add(text_ba)

                    text_ru = item.get("text_ru", "").strip()
                    if text_ru:
                        trans = Translation(
                            fragment_id=frag.id,
                            language_pair_id=pair_id,
                            translation_text=text_ru,
                            status="confirmed",
                            created_by=user_id,
                            votes_count=3,
                            reject_count=0
                        )
                        db.add(trans)
                    stats["hf_imported"] += 1
        except Exception as e:
            logger.error(f"Error importing HF phrasebook: {e}")

    # 4. Import Text Samples (Folklore)
    texts_path = find_data_file("texts_samples.jsonl")
    if texts_path:
        try:
            with open(texts_path, "r", encoding="utf-8") as f:
                for line in f:
                    line = line.strip()
                    if not line:
                        continue
                    item = json.loads(line)
                    text_ba = item.get("text_ba", "").strip()
                    if not text_ba or text_ba in existing_texts:
                        continue
                    frag = Fragment(
                        text_ba=text_ba,
                        source_title=item.get("source_title", "Башкирский эпос и фольклор"),
                        source_document="Архив Академии наук РБ",
                        category=item.get("category", "Эпос и фольклор"),
                        difficulty="hard"
                    )
                    db.add(frag)
                    await db.flush()
                    existing_texts.add(text_ba)

                    text_ru = item.get("translation_ru", "").strip()
                    if text_ru:
                        trans = Translation(
                            fragment_id=frag.id,
                            language_pair_id=pair_id,
                            translation_text=text_ru,
                            status="confirmed",
                            created_by=user_id,
                            votes_count=3,
                            reject_count=0
                        )
                        db.add(trans)
                    stats["samples_imported"] += 1
        except Exception as e:
            logger.error(f"Error importing text samples: {e}")

    await db.commit()
    return stats

if __name__ == "__main__":
    import asyncio
    from app.core.database import AsyncSessionLocal

    async def run():
        print("=== Импорт датасетов из репозитория в базу данных ===")
        async with AsyncSessionLocal() as session:
            result = await import_all_datasets(session)
            print("Результат импорта:", result)

    asyncio.run(run())
