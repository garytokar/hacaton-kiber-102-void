import random
import string
from typing import List, Optional
from fastapi import APIRouter, Depends, HTTPException, status
from pydantic import BaseModel
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy import select, func, desc
from sqlalchemy.orm import selectinload
from app.core.database import get_db
from app.models.models import User, School, SchoolClass, UserStats
from app.api.deps import get_current_user

router = APIRouter()

class CreateClassRequest(BaseModel):
    school_id: int
    name: str
    grade_number: Optional[int] = None
    letter: Optional[str] = None

class JoinClassRequest(BaseModel):
    invite_code: str

class AddStudentRequest(BaseModel):
    username_or_email: str

def generate_invite_code(prefix: str = "INV") -> str:
    rnd = ''.join(random.choices(string.ascii_uppercase + string.digits, k=6))
    return f"{prefix}-{rnd}"

@router.get("/schools")
async def get_schools_options(db: AsyncSession = Depends(get_db)):
    res = await db.execute(select(School).order_by(School.name))
    schools = res.scalars().all()
    return [{"id": s.id, "name": s.name, "city": s.city} for s in schools]

@router.post("/classes")
async def create_class(
    req: CreateClassRequest,
    user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    school = await db.get(School, req.school_id)
    if not school:
        raise HTTPException(status_code=404, detail="Школа не найдена")

    invite_code = generate_invite_code(f"INV-{req.grade_number or 'CL'}")
    new_class = SchoolClass(
        school_id=req.school_id,
        name=req.name,
        grade_number=req.grade_number,
        letter=req.letter,
        invite_code=invite_code
    )
    db.add(new_class)
    await db.commit()
    await db.refresh(new_class)

    return {
        "success": True,
        "id": new_class.id,
        "name": new_class.name,
        "school_id": new_class.school_id,
        "invite_code": new_class.invite_code,
        "message": f"Класс «{new_class.name}» успешно создан. Код приглашения: {new_class.invite_code}"
    }

@router.post("/classes/{class_id}/invite")
async def get_or_generate_invite(
    class_id: int,
    user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    school_class = await db.get(SchoolClass, class_id)
    if not school_class:
        raise HTTPException(status_code=404, detail="Класс не найден")

    if not school_class.invite_code:
        school_class.invite_code = generate_invite_code()
        await db.commit()

    return {
        "class_id": school_class.id,
        "invite_code": school_class.invite_code,
        "class_name": school_class.name
    }

@router.post("/classes/join")
async def join_class(
    req: JoinClassRequest,
    user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    code = req.invite_code.strip().upper()
    res = await db.execute(select(SchoolClass).where(SchoolClass.invite_code == code))
    school_class = res.scalar_one_or_none()
    if not school_class:
        raise HTTPException(status_code=404, detail="Класс с указанным кодом приглашения не найден")

    user.class_id = school_class.id
    user.school_id = school_class.school_id
    await db.commit()

    school = await db.get(School, school_class.school_id)
    school_name = school.name if school else "Школа"

    return {
        "success": True,
        "message": f"Вы успешно присоединились к классу {school_class.name} ({school_name})!",
        "class_id": school_class.id,
        "class_name": school_class.name,
        "school_name": school_name
    }

@router.post("/classes/{class_id}/students")
async def add_student_to_class(
    class_id: int,
    req: AddStudentRequest,
    user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    school_class = await db.get(SchoolClass, class_id)
    if not school_class:
        raise HTTPException(status_code=404, detail="Класс не найден")

    # Find user
    identifier = req.username_or_email.strip()
    res = await db.execute(
        select(User).where((User.username == identifier) | (User.email == identifier))
    )
    student = res.scalar_one_or_none()
    if not student:
        raise HTTPException(status_code=404, detail=f"Пользователь '{identifier}' не найден")

    student.class_id = school_class.id
    student.school_id = school_class.school_id
    await db.commit()

    return {
        "success": True,
        "message": f"Ученик {student.full_name or student.username} зачислен в класс {school_class.name}"
    }

@router.get("/classes/{class_id}")
async def get_class_stats(
    class_id: int,
    db: AsyncSession = Depends(get_db)
):
    school_class = await db.get(SchoolClass, class_id)
    if not school_class:
        raise HTTPException(status_code=404, detail="Класс не найден")

    # Get students
    res = await db.execute(
        select(User)
        .options(selectinload(User.stats))
        .where(User.class_id == class_id)
    )
    students = res.scalars().all()

    total_xp = sum(s.stats.xp for s in students if s.stats)
    total_tasks = sum(s.stats.total_tasks_completed for s in students if s.stats)

    students_list = []
    for s in sorted(students, key=lambda x: x.stats.xp if x.stats else 0, reverse=True):
        students_list.append({
            "id": s.id,
            "username": s.username,
            "full_name": s.full_name or s.username,
            "xp": s.stats.xp if s.stats else 0,
            "level": s.stats.level if s.stats else 1,
            "tasks_completed": s.stats.total_tasks_completed if s.stats else 0
        })

    school = await db.get(School, school_class.school_id)

    return {
        "class_id": school_class.id,
        "class_name": school_class.name,
        "school_name": school.name if school else "",
        "invite_code": school_class.invite_code,
        "total_students": len(students),
        "total_xp": total_xp,
        "total_tasks": total_tasks,
        "students": students_list
    }
