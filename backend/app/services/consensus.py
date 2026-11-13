"""
Fuzzy Consensus Engine for Bashkir Translations
Ported from Go backend (pkg/fuzzy/fuzzy.go & usecase/consensus.go).
Provides:
- Canonicalization of Bashkir and Russian text
- Levenshtein distance and normalized similarity
- Character Trigram similarity (Dice coefficient)
- Clustering and consensus matching against existing translations
"""

import re
import unicodedata
from typing import Tuple, List, Optional

DEFAULT_SIMILARITY_THRESHOLD = 0.82

def canonicalize(text: str) -> str:
    """
    Normalizes text for fuzzy comparison:
    - Preserves sentinel values like '[ERROR_REPORTED]'
    - Converts to lowercase
    - Normalizes dash variations to standard hyphen '-'
    - Normalizes quotation marks to standard double quote '"'
    - Collapses multiple whitespace characters to a single space
    - Strips leading and trailing punctuation and spaces
    """
    if not text:
        return ""
    
    trimmed = text.strip()
    if trimmed.upper() == "[ERROR_REPORTED]":
        return "[ERROR_REPORTED]"

    # Lowercase
    s = trimmed.lower()

    # Normalize dashes
    s = re.sub(r'[—–―−‒‐]', '-', s)

    # Normalize quotes
    s = re.sub(r'[«»“”„‘’‚`\']', '"', s)

    # Collapse multiple whitespace
    s = re.sub(r'\s+', ' ', s)

    # Strip punctuation and whitespace from edges
    s = re.sub(r'^[^\w\s]+|[^\w\s]+$', '', s)

    return s.strip()


def levenshtein_distance(s1: str, s2: str) -> int:
    """Calculates Levenshtein edit distance between two strings."""
    if s1 == s2:
        return 0
    if not s1:
        return len(s2)
    if not s2:
        return len(s1)

    r1, r2 = list(s1), list(s2)
    n, m = len(r1), len(r2)

    if n > m:
        r1, r2 = r2, r1
        n, m = m, n

    current = list(range(n + 1))
    for i in range(1, m + 1):
        previous, current = current, [i] + [0] * n
        for j in range(1, n + 1):
            add, delete = previous[j] + 1, current[j - 1] + 1
            change = previous[j - 1]
            if r1[j - 1] != r2[i - 1]:
                change += 1
            current[j] = min(add, delete, change)

    return current[n]


def levenshtein_similarity(s1: str, s2: str) -> float:
    """Calculates normalized Levenshtein similarity between 0.0 and 1.0."""
    if s1 == s2:
        return 1.0
    if not s1 or not s2:
        return 0.0

    max_len = max(len(s1), len(s2))
    if max_len == 0:
        return 1.0

    dist = levenshtein_distance(s1, s2)
    return max(0.0, 1.0 - (dist / max_len))


def extract_trigrams(text: str) -> List[str]:
    """Extracts character trigrams from text."""
    if len(text) < 3:
        return [text] if text else []
    return [text[i:i+3] for i in range(len(text) - 2)]


def trigram_similarity(s1: str, s2: str) -> float:
    """Calculates Dice coefficient over character trigrams."""
    if s1 == s2:
        return 1.0
    if not s1 or not s2:
        return 0.0

    t1 = extract_trigrams(s1)
    t2 = extract_trigrams(s2)

    if not t1 and not t2:
        return 1.0
    if not t1 or not t2:
        return 0.0

    counts1: dict[str, int] = {}
    for g in t1:
        counts1[g] = counts1.get(g, 0) + 1

    intersection = 0
    for g in t2:
        if counts1.get(g, 0) > 0:
            intersection += 1
            counts1[g] -= 1

    total = len(t1) + len(t2)
    if total == 0:
        return 0.0

    return (2.0 * intersection) / float(total)


def compute_combined_similarity(s1: str, s2: str) -> float:
    """
    Computes hybrid similarity score between two translation strings:
    uses canonicalized text, combining Levenshtein and Trigram similarity.
    """
    c1 = canonicalize(s1)
    c2 = canonicalize(s2)

    if c1 == c2:
        return 1.0

    lev_sim = levenshtein_similarity(c1, c2)
    tri_sim = trigram_similarity(c1, c2)

    # Weighted blend: Levenshtein captures word-level typo edits, Trigram captures phrasing similarity
    return max(lev_sim, 0.4 * lev_sim + 0.6 * tri_sim)


def find_matching_translation(
    candidate: str,
    existing_translations: list,
    threshold: float = DEFAULT_SIMILARITY_THRESHOLD
) -> Optional[Tuple[any, float]]:
    """
    Searches a list of existing Translation models for a fuzzy match with candidate text.
    Returns (translation, score) or None.
    """
    best_match = None
    best_score = 0.0

    for tr in existing_translations:
        score = compute_combined_similarity(candidate, tr.translation_text)
        if score > best_score:
            best_score = score
            best_match = tr

    if best_match and best_score >= threshold:
        return best_match, best_score

    return None
