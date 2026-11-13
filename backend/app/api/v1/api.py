from fastapi import APIRouter
from app.api.v1.endpoints import (
    auth, users, fragments, translations, quests,
    leaderboard, achievements, dictionary, admin, export, stats, lms
)

api_router = APIRouter()

api_router.include_router(auth.router, prefix="/auth", tags=["auth"])
api_router.include_router(users.router, prefix="/users", tags=["users"])
api_router.include_router(fragments.router, prefix="/fragments", tags=["fragments"])
api_router.include_router(translations.router, prefix="/translations", tags=["translations"])
api_router.include_router(quests.router, prefix="/quests", tags=["quests"])
api_router.include_router(leaderboard.router, prefix="/leaderboard", tags=["leaderboard"])
api_router.include_router(achievements.router, prefix="/achievements", tags=["achievements"])
api_router.include_router(dictionary.router, prefix="/dictionary", tags=["dictionary"])
api_router.include_router(admin.router, prefix="/admin", tags=["admin"])
api_router.include_router(export.router, prefix="/export", tags=["export"])
api_router.include_router(stats.router, prefix="/stats", tags=["stats"])
api_router.include_router(lms.router, prefix="/lms", tags=["lms"])
api_router.include_router(lms.router, prefix="/teacher", tags=["teacher"])
