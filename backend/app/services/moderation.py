from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy import select, func
from app.models.models import Translation, TranslationVote, TranslationHistory, WeeklyExpedition

CONFIRMATION_THRESHOLD = 3

async def evaluate_translation_status(
    db: AsyncSession,
    translation_id: int,
    changed_by_user_id: int
) -> str:
    """Evaluates status of translation based on votes count and ratio."""
    res = await db.execute(select(Translation).where(Translation.id == translation_id))
    translation = res.scalar_one_or_none()
    if not translation:
        return "not_found"

    # Count confirms vs rejects
    confirms_res = await db.execute(
        select(func.count(TranslationVote.id)).where(
            TranslationVote.translation_id == translation_id,
            TranslationVote.vote_type == "confirm"
        )
    )
    confirms = confirms_res.scalar() or 0

    rejects_res = await db.execute(
        select(func.count(TranslationVote.id)).where(
            TranslationVote.translation_id == translation_id,
            TranslationVote.vote_type == "reject"
        )
    )
    rejects = rejects_res.scalar() or 0

    translation.votes_count = confirms
    translation.reject_count = rejects

    old_status = translation.status
    new_status = old_status

    if rejects > 1 and rejects >= confirms:
        new_status = "expert_review"
    elif confirms >= CONFIRMATION_THRESHOLD and (confirms / (confirms + rejects)) >= 0.75:
        new_status = "confirmed"
        # Increment active weekly expedition if any
        exp_res = await db.execute(
            select(WeeklyExpedition).where(WeeklyExpedition.is_active == True).limit(1)
        )
        expedition = exp_res.scalar_one_or_none()
        if expedition:
            expedition.current_count += 1
    elif confirms > 0 or rejects > 0:
        new_status = "pending"

    if new_status != old_status:
        translation.status = new_status
        history = TranslationHistory(
            translation_id=translation_id,
            previous_text=f"Статус: {old_status}",
            new_text=f"Статус: {new_status} (Голосов: +{confirms}/-{rejects})",
            changed_by=changed_by_user_id
        )
        db.add(history)

    return new_status
