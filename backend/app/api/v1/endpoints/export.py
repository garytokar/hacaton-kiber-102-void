import io
import csv
from typing import Optional
from fastapi import APIRouter, Depends, Query, Response
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy import select
from sqlalchemy.orm import selectinload
from app.core.database import get_db
from app.models.models import Translation, Fragment

router = APIRouter()

@router.get("/translations")
async def export_translations(
    format: str = Query("json", pattern="^(json|csv)$"),
    status: str = Query("confirmed", pattern="^(confirmed|all)$"),
    db: AsyncSession = Depends(get_db)
):
    stmt = select(Translation).join(Translation.fragment).options(
        selectinload(Translation.fragment)
    )

    if status == "confirmed":
        stmt = stmt.where(Translation.status == "confirmed")

    res = await db.execute(stmt)
    translations = res.scalars().all()

    data = [
        {
            "fragment_id": t.fragment_id,
            "text_ba": t.fragment.text_ba,
            "translation_ru": t.translation_text,
            "category": t.fragment.category,
            "difficulty": t.fragment.difficulty,
            "source": t.fragment.source_title,
            "status": t.status,
            "confirmations": t.votes_count
        }
        for t in translations
    ]

    if format == "csv":
        output = io.StringIO()
        writer = csv.DictWriter(output, fieldnames=["fragment_id", "text_ba", "translation_ru", "category", "difficulty", "source", "status", "confirmations"])
        writer.writeheader()
        for row in data:
            writer.writerow(row)
        
        response = Response(content=output.getvalue(), media_type="text/csv")
        response.headers["Content-Disposition"] = "attachment; filename=yanlander_corpus_export.csv"
        return response

    return {
        "total": len(data),
        "corpus_name": "Корпус переводов ЙӘНЛӘНДЕР",
        "data": data
    }
