from typing import Dict, Any, List

def parse_user_cosmetics(avatar_str: str) -> Dict[str, Any]:
    default_res = {
        "avatar": "avatar_default",
        "equipped_frame": None,
        "equipped_title": None,
        "unlocked": []
    }
    if not avatar_str:
        return default_res

    parts = [p.strip() for p in avatar_str.split("|") if p.strip()]
    if not parts:
        return default_res

    res = {
        "avatar": parts[0] or "avatar_default",
        "equipped_frame": None,
        "equipped_title": None,
        "unlocked": []
    }

    for p in parts[1:]:
        if p.startswith("frame_"):
            if not res["equipped_frame"]:
                res["equipped_frame"] = p
            if p not in res["unlocked"]:
                res["unlocked"].append(p)
        elif p.startswith("title_"):
            if not res["equipped_title"]:
                res["equipped_title"] = p
            if p not in res["unlocked"]:
                res["unlocked"].append(p)
        elif p:
            if p not in res["unlocked"]:
                res["unlocked"].append(p)

    return res

def serialize_user_cosmetics(cosmetics: Dict[str, Any]) -> str:
    base_avatar = cosmetics.get("avatar") or "avatar_default"
    items: List[str] = [base_avatar]

    equipped_frame = cosmetics.get("equipped_frame")
    equipped_title = cosmetics.get("equipped_title")

    # Put equipped first
    if equipped_frame and equipped_frame not in items:
        items.append(equipped_frame)
    if equipped_title and equipped_title not in items:
        items.append(equipped_title)

    for u in cosmetics.get("unlocked", []):
        if u not in items:
            items.append(u)

    return "|".join(items)[:255]
