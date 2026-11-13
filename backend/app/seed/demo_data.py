from datetime import datetime, date, timedelta
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy import select
from app.core.security import get_password_hash
from app.models.models import (
    School, SchoolClass, Team, TeamMember, Language, LanguagePair,
    User, UserStats, Fragment, Translation, TranslationVote, TranslationHistory,
    Achievement, UserAchievement, DailyQuest, UserQuest, DictionaryEntry,
    WeeklyExpedition, Notification
)

async def seed_all(db: AsyncSession):
    # Check if already seeded
    res = await db.execute(select(User).limit(1))
    if res.scalar_one_or_none():
        return

    # 1. Languages & Pairs
    ba = Language(code="ba", name="Башкирский")
    ru = Language(code="ru", name="Русский")
    en = Language(code="en", name="Английский")
    db.add_all([ba, ru, en])
    await db.flush()

    pair_ba_ru = LanguagePair(source_language_id=ba.id, target_language_id=ru.id)
    pair_ba_en = LanguagePair(source_language_id=ba.id, target_language_id=en.id)
    db.add_all([pair_ba_ru, pair_ba_en])
    await db.flush()

    # 2. Schools and Classes
    schools_data = [
        ("БРГИ №1 им. Рами Гарипова", "Уфа", ["9-А", "10-Б", "11-А"]),
        ("РХГИ им. К.А. Давлеткильдеева", "Уфа", ["8-В", "9-Б", "10-А"]),
        ("Гимназия №39 им. А.Ш. Файзуллина", "Уфа", ["7-А", "8-А", "10-В"]),
        ("Башкирский лицей №136 им. М. Искужина", "Уфа", ["9-А", "11-Б"]),
        ("Гимназия №1 им. Ч.Т. Айтматова", "Стерлитамак", ["8-Б", "10-А"]),
    ]

    schools = []
    classes = []
    for s_name, city, cls_list in schools_data:
        school = School(name=s_name, city=city)
        db.add(school)
        await db.flush()
        schools.append(school)
        for c_name in cls_list:
            sc = SchoolClass(school_id=school.id, name=c_name)
            db.add(sc)
            await db.flush()
            classes.append(sc)

    # 3. Teams
    team1 = Team(name="Хранители Урала", description="Команда школы БРГИ №1 по сохранению фольклора", emblem="flame")
    team2 = Team(name="Алтын Ҡурай", description="Любители башкирской поэзии и эпоса", emblem="award")
    team3 = Team(name="Степные Беркуты", description="Команда знатоков родного языка", emblem="shield")
    db.add_all([team1, team2, team3])
    await db.flush()

    # 4. Users (10 users with realistic stats, including Admin and Demo Expert)
    hashed_pwd = get_password_hash("password123")
    
    users_info = [
        ("admin@yanlander.ru", "admin", "admin", "Администратор Системы", "avatar_admin", schools[0].id, classes[0].id, 4500, 8, 380, 24, 30, 150, 95, 25, 22),
        ("expert@yanlander.ru", "expert_fatima", "expert", "Фатима Ахметова (Филолог)", "avatar_expert", schools[0].id, classes[0].id, 6800, 9, 520, 45, 45, 210, 180, 40, 38),
        ("teacher_aygul@yanlander.ru", "aygul_ustaz", "teacher", "Айгуль Салаватовна", "avatar_teacher", schools[1].id, classes[3].id, 3200, 7, 290, 18, 20, 120, 85, 15, 14),
        ("student_timur@yanlander.ru", "timur_batyr", "student", "Тимур Ишмухаметов", "avatar_student1", schools[0].id, classes[0].id, 1650, 6, 140, 14, 14, 75, 42, 12, 10),
        ("student_gulnaz@yanlander.ru", "gulnaz_chechka", "student", "Гульназ Валиева", "avatar_student2", schools[0].id, classes[1].id, 980, 5, 110, 9, 12, 55, 34, 8, 6),
        ("student_ildar@yanlander.ru", "ildar_altin", "student", "Ильдар Хасанов", "avatar_student3", schools[1].id, classes[4].id, 560, 4, 75, 6, 8, 38, 20, 5, 4),
        ("student_zarina@yanlander.ru", "zarina_nur", "student", "Зарина Бикбаева", "avatar_student4", schools[2].id, classes[6].id, 310, 3, 50, 4, 5, 22, 15, 3, 2),
        ("student_artur@yanlander.ru", "artur_akbuzat", "student", "Артур Каримов", "avatar_student5", schools[2].id, classes[7].id, 120, 2, 40, 2, 3, 12, 8, 1, 1),
        ("student_aliya@yanlander.ru", "aliya_yoldoz", "student", "Алия Мусина", "avatar_student6", schools[3].id, classes[9].id, 85, 1, 25, 1, 2, 8, 5, 0, 0),
        ("student_rustam@yanlander.ru", "rustam_quray", "student", "Рустам Юсупов", "avatar_student7", schools[4].id, classes[11].id, 45, 1, 15, 1, 1, 5, 3, 0, 0),
    ]

    created_users = []
    for email, username, role, full_name, avatar, sch_id, cls_id, xp, lvl, curr, cur_str, best_str, tot_tsk, tot_cnf, tot_prop, tot_acc in users_info:
        user = User(
            email=email,
            username=username,
            password_hash=hashed_pwd,
            full_name=full_name,
            avatar=avatar,
            role=role,
            school_id=sch_id,
            class_id=cls_id,
        )
        db.add(user)
        await db.flush()
        created_users.append(user)

        stats = UserStats(
            user_id=user.id,
            xp=xp,
            level=lvl,
            currency=curr,
            current_streak=cur_str,
            best_streak=best_str,
            last_activity_date=date.today(),
            streak_protected=(cur_str > 10),
            total_tasks_completed=tot_tsk,
            total_translations_confirmed=tot_cnf,
            total_proposals_made=tot_prop,
            total_accepted_proposals=tot_acc
        )
        db.add(stats)

    # Assign team members
    db.add(TeamMember(team_id=team1.id, user_id=created_users[0].id))
    db.add(TeamMember(team_id=team1.id, user_id=created_users[1].id))
    db.add(TeamMember(team_id=team1.id, user_id=created_users[3].id))
    db.add(TeamMember(team_id=team2.id, user_id=created_users[2].id))
    db.add(TeamMember(team_id=team2.id, user_id=created_users[4].id))
    db.add(TeamMember(team_id=team3.id, user_id=created_users[5].id))
    db.add(TeamMember(team_id=team3.id, user_id=created_users[6].id))

    # 5. Achievements (Matching Section 7.4)
    achievements_data = [
        ("first_word", "Первое слово", "Выполните ваше первое задание на платформе", "sparkles", "tasks_count", 1, 20, 10),
        ("first_translation", "Первый перевод", "Предложите свое первое исправление или вариант перевода", "edit-3", "proposals_count", 1, 30, 15),
        ("streak_7", "Семидневный хранитель", "Поддерживайте серию ежедневных занятий 7 дней подряд", "flame", "streak_days", 7, 50, 25),
        ("streak_30", "Месяц знаний", "30 дней непрерывной помощи в сохранении языка", "calendar", "streak_days", 30, 150, 80),
        ("master_accuracy", "Мастер точности", "20 качественно подтверждённых вариантов перевода", "check-circle", "confirmed_count", 20, 60, 30),
        ("word_hunter", "Охотник за словами", "100 проверенных вариантов башкирского текста", "target", "tasks_count", 100, 120, 60),
        ("voice_of_people", "Голос народа", "50 полезных голосов за точность переводов", "thumbs-up", "confirmed_count", 50, 90, 45),
        ("collector", "Коллекционер", "50 сохранённых слов в персональном словаре", "bookmark", "tasks_count", 50, 70, 35),
        ("team_player", "Командный игрок", "Внесите вклад в 10 командных заданий школы", "users", "tasks_count", 10, 40, 20),
        ("community_legend", "Легенда сообщества", "Достигните 5000 очков опыта Хранителя языка", "crown", "xp_total", 5000, 250, 150),
    ]

    achievements = []
    for code, name, desc, icon, c_type, c_val, rxp, rcurr in achievements_data:
        ach = Achievement(
            code=code,
            name=name,
            description=desc,
            icon=icon,
            condition_type=c_type,
            condition_value=c_val,
            reward_xp=rxp,
            reward_currency=rcurr
        )
        db.add(ach)
        await db.flush()
        achievements.append(ach)

    # Unlock some achievements for top users
    for user_idx in [0, 1, 2, 3]:
        db.add(UserAchievement(user_id=created_users[user_idx].id, achievement_id=achievements[0].id))
        db.add(UserAchievement(user_id=created_users[user_idx].id, achievement_id=achievements[1].id))
        db.add(UserAchievement(user_id=created_users[user_idx].id, achievement_id=achievements[2].id))
        db.add(UserAchievement(user_id=created_users[user_idx].id, achievement_id=achievements[4].id))

    # 6. Daily Quests (Section 7.6)
    quests_data = [
        ("daily_confirm_2", "Первые шаги дня", "Подтвердите 2 проверенных перевода башкирского текста", 2, 15, 10, "confirm"),
        ("daily_verify_4", "Внимательный хранитель", "Проверьте и оцените 4 фрагмента из народных текстов", 4, 25, 15, "verify"),
        ("daily_propose_1", "Слово мастера", "Предложите 1 собственное исправление или вариант", 1, 30, 20, "propose"),
    ]

    for code, name, desc, target, rxp, rcurr, qtype in quests_data:
        dq = DailyQuest(
            code=code,
            name=name,
            description=desc,
            target=target,
            reward_xp=rxp,
            reward_currency=rcurr,
            quest_type=qtype
        )
        db.add(dq)

    # 7. Weekly Expeditions (Section 7.7)
    exp1 = WeeklyExpedition(
        title="Сокровища башкирских пословиц",
        description="Коллективная миссия школ Башкортостана: подтвердить 200 народных пословиц и поговорок за неделю",
        target_count=200,
        current_count=138,
        reward_xp=100,
        reward_currency=50,
        is_active=True
    )
    exp2 = WeeklyExpedition(
        title="Строки эпоса «Урал-батыр»",
        description="Сверим и перепроверим 150 строф древнего башкирского эпоса для цифрового корпуса",
        target_count=150,
        current_count=94,
        reward_xp=120,
        reward_currency=60,
        is_active=False
    )
    db.add_all([exp1, exp2])

    # 8. Authentic Bashkir Fragments & Translations (30+ items from folklore, literature, epics)
    fragments_catalog = [
        (
            "Тыуған тел — халыҡтың рухи байлығы.",
            "Башкирские народные пословицы",
            "Пословицы и поговорки",
            "easy",
            [
                ("Родной язык — духовное богатство народа.", "confirmed", 12, 0),
                ("Родная речь — это главное сокровище духа нации.", "pending", 2, 0)
            ]
        ),
        (
            "Берҙәмлектә — көс, айырымлыҡта — хәлһеҙлек.",
            "Башкирский фольклор",
            "Пословицы и поговорки",
            "easy",
            [
                ("В единстве — сила, в разобщении — бессилие.", "confirmed", 9, 0),
                ("Сила в согласии, в раздоре слабость.", "pending", 1, 0)
            ]
        ),
        (
            "Ир-егеттең ҡәҙере ил менән, илдең ҡәҙере ир менән.",
            "Эпос и фольклор",
            "Крылатые выражения",
            "medium",
            [
                ("Ценность джигита познаётся в служении родине, а величие родины — в её батырах.", "confirmed", 7, 0)
            ]
        ),
        (
            "Ат даны яуҙа беленер, ир даны дауҙа беленер.",
            "Башкирские поговорки",
            "Пословицы и поговорки",
            "medium",
            [
                ("Слава коня познаётся в бою, слава джигита — в споре за справедливость.", "confirmed", 8, 1),
                ("Слава коня познается в сражении, достоинство мужчины в испытании.", "pending", 2, 0)
            ]
        ),
        (
            "Аҡыл — баштан, байлыҡ — ҡулдан.",
            "Народная мудрость",
            "Пословицы и поговорки",
            "easy",
            [
                ("Ум исходит из головы, а богатство создается руками.", "confirmed", 11, 0)
            ]
        ),
        (
            "Башҡортостан — минең гүзәл төйәгем, ата-бабаларым төйәге.",
            "Мустай Карим",
            "Литература",
            "medium",
            [
                ("Башкортостан — мой прекрасный край, край моих отцов и дедов.", "confirmed", 15, 0),
                ("Башкирия — моя прекрасная родина, земля предков.", "pending", 3, 1)
            ]
        ),
        (
            "Яҡшы һүҙ — йән аҙығы, яман һүҙ — баш ҡаҙығы.",
            "Башкирские пословицы",
            "Пословицы и поговорки",
            "easy",
            [
                ("Доброе слово душу питает, худое слово ранит сердце.", "confirmed", 8, 0)
            ]
        ),
        (
            "Урал тауы бейек, уның башы күктәргә тейә.",
            "Эпос «Урал-батыр»",
            "Эпос",
            "medium",
            [
                ("Уральские горы высоки, их вершины касаются небес.", "confirmed", 10, 0)
            ]
        ),
        (
            "Әсә күңеле — балала, бала күңеле — далала.",
            "Башкирские пословицы",
            "Семейные ценности",
            "easy",
            [
                ("Сердце матери — с дитём, а помыслы дитя — в просторах степи.", "confirmed", 14, 0),
                ("Материнское сердце привязано к ребенку, а мысли ребенка гуляют в степи.", "pending", 2, 0)
            ]
        ),
        (
            "Кем эшләмәй, шул ашамай.",
            "Народные поговорки",
            "Труд и ремесло",
            "easy",
            [
                ("Кто не работает, тот не ест.", "confirmed", 16, 0)
            ]
        ),
        (
            "Ҡурай моңо — халыҡ күңеле.",
            "Башкирское музыкальное наследие",
            "Культура",
            "easy",
            [
                ("Мелодия курая — душа башкирского народа.", "confirmed", 18, 0)
            ]
        ),
        (
            "Яҙғы көн йылды туйҙыра.",
            "Аграрные пословицы",
            "Труд и природа",
            "easy",
            [
                ("Весенний день год кормит.", "confirmed", 7, 0)
            ]
        ),
        (
            "Китап — белем шишмәһе, аҡыл көҙгөһө.",
            "Педагогические наставления",
            "Образование",
            "easy",
            [
                ("Книга — родник знаний и зеркало разума.", "confirmed", 11, 0)
            ]
        ),
        (
            "Уҡыу — энә менән ҡойо ҡаҙыу кеүек.",
            "Башкирские пословицы",
            "Образование",
            "medium",
            [
                ("Учёба — всё равно что иголкой колодец копать.", "confirmed", 9, 0)
            ]
        ),
        (
            "Тырышҡан таш тишер, ялҡау баш ҡашыр.",
            "Башкирский фольклор",
            "Труд и упорство",
            "medium",
            [
                ("Старательный и камень просверлит, а ленивый лишь затылок почешет.", "confirmed", 13, 0)
            ]
        ),
        (
            "Ил төҙөгән ир арымай, ил боҙған оҙаҡ бармай.",
            "Эпос «Урал-батыр»",
            "Эпос",
            "hard",
            [
                ("Муж, созидающий отчизну, не устаёт, а разрушитель страны долго не протянет.", "confirmed", 6, 0),
                ("Тот, кто созидает родину, полон сил, а предатель скоро падет.", "expert_review", 2, 2)
            ]
        ),
        (
            "Һыу юлында көрәк табылған, һүҙ юлында йөрәк табылған.",
            "Народная лирика",
            "Поэзия",
            "hard",
            [
                ("На тропе к воде лопата найдётся, на стезе искреннего слова сердце откроется.", "expert_review", 3, 2),
                ("По пути к ручью орудие труда отыщется, в доброй беседе душа проявится.", "pending", 2, 1)
            ]
        ),
        (
            "Ағас үҙ тамыры менән, кеше үҙ дуҫтары менән көслө.",
            "Башкирские пословицы",
            "Дружба",
            "easy",
            [
                ("Дерево крепко своими корнями, а человек силен своими верными друзьями.", "confirmed", 15, 0)
            ]
        ),
        (
            "Ут менән уйнама, һыу менән шаярма.",
            "Народная мудрость",
            "Безопасность и жизнь",
            "easy",
            [
                ("С огнём не играй, со стихией воды не шути.", "confirmed", 9, 0)
            ]
        ),
        (
            "Ҡунаҡ ашы — ҡара-ҡаршы.",
            "Обычаи гостеприимства",
            "Традиции",
            "easy",
            [
                ("Угощение гостя требует взаимности и радушия.", "confirmed", 8, 0),
                ("Гостеприимство — долг взаимный.", "pending", 3, 0)
            ]
        ),
        (
            "Сәсән телен ҡылыс киҫмәҫ, батыр яуҙа һынмаҫ.",
            "Фольклор сэсэнов",
            "Культура",
            "medium",
            [
                ("Слово мудрого сэсэна и меч не разрубит, а батыр в бою не сломится.", "confirmed", 7, 0)
            ]
        ),
        (
            "Түҙгән — моратына еткән, түҙмәгән — ҡоро ҡалған.",
            "Башкирские поговорки",
            "Характер",
            "medium",
            [
                ("Терпеливый достигает желанной цели, а нетерпеливый остаётся ни с чем.", "confirmed", 12, 0)
            ]
        ),
        (
            "Ағым һыуға ҡарап аҡыл тупла, бейек тауға ҡарап рух ҡеүәтен ал.",
            "Народные наставления",
            "Философия",
            "hard",
            [
                ("Глядя на текущую воду, набирайся мудрости, взирая на высокую гору, укрепляй дух.", "confirmed", 5, 0)
            ]
        ),
        (
            "Бала сағың — алтын сағың, ҡәҙерен бел һин уның.",
            "Детский фольклор",
            "Детство",
            "easy",
            [
                ("Детские годы — золотая пора, цени и береги это светлое время.", "confirmed", 10, 0)
            ]
        ),
        (
            "Батыр бер үлер, ҡурҡаҡ мең үлер.",
            "Башкирские пословицы",
            "Мужество",
            "easy",
            [
                ("Смелый воин гибнет однажды, а трус умирает тысячу раз.", "confirmed", 14, 0)
            ]
        ),
        (
            "Тормош — ҙур диңгеҙ, белем — уның маяғы.",
            "Современная башкирская литература",
            "Образование",
            "medium",
            [
                ("Жизнь — это безбрежное море, а знание — путеводный маяк в нём.", "confirmed", 8, 0)
            ]
        ),
        (
            "Бал яратҡан айыуҙың мороно яныр.",
            "Башкирские сказки о животных",
            "Сказки",
            "easy",
            [
                ("У медведя, падкого до мёда, нос обожжён будет.", "confirmed", 6, 0)
            ]
        ),
        (
            "Үҙ ереңдә таш та алтынға тиң, сит илдә алтын да баҡыр кеүек.",
            "Патриотическая лирика",
            "Родина",
            "medium",
            [
                ("На родной земле и простой камень дороже золота, а на чужбине и золото покажется медью.", "confirmed", 11, 0)
            ]
        ),
        (
            "Һөнәрле егет ас ҡалмаҫ, ҡул ҡаушырып ятмаҫ.",
            "Народное ремесло",
            "Труд и ремесло",
            "easy",
            [
                ("Мастеровитый парень голодным не останется, сложа руки сидеть не станет.", "confirmed", 9, 0)
            ]
        ),
        (
            "Яҡты йондоҙ күктә яныр, яҡшы исем ерҙә ҡалыр.",
            "Рами Гарипов",
            "Поэзия",
            "medium",
            [
                ("Яркая звезда сияет на небесах, а доброе имя вечно живет на земле.", "confirmed", 13, 0),
                ("Ясная звезда в небе горит, светлое имя на земле остается.", "pending", 2, 0)
            ]
        ),
        (
            "Башҡорт теле — моңло тел, сәсәндәрҙең нәфис теле.",
            "Зайнаб Биишева",
            "Литература",
            "medium",
            [
                ("Башкирский язык — мелодичный и напевный, изящный язык древних сэсэнов.", "confirmed", 17, 0)
            ]
        ),
        (
            "Үткәнде онотҡандың киләсәге булмаҫ.",
            "Историческая мудрость",
            "История",
            "medium",
            [
                ("У того, кто забывает своё прошлое, нет достойного будущего.", "confirmed", 14, 0)
            ]
        ),
    ]

    for text_ba, src_title, cat, diff, trans_list in fragments_catalog:
        frag = Fragment(
            text_ba=text_ba,
            source_title=src_title,
            source_document="Цифровой корпус башкирского языка / Фольклорные сборники",
            category=cat,
            difficulty=diff
        )
        db.add(frag)
        await db.flush()

        for t_text, t_stat, v_count, r_count in trans_list:
            trans = Translation(
                fragment_id=frag.id,
                language_pair_id=pair_ba_ru.id,
                translation_text=t_text,
                status=t_stat,
                created_by=created_users[1].id if t_stat == "confirmed" else created_users[3].id,
                votes_count=v_count,
                reject_count=r_count
            )
            db.add(trans)
            await db.flush()

            # Add demo votes
            for u in created_users[:min(v_count, len(created_users))]:
                vote = TranslationVote(
                    translation_id=trans.id,
                    user_id=u.id,
                    vote_type="confirm"
                )
                db.add(vote)

    # 9. Initial dictionary entries for demo user
    demo_dict_words = [
        ("Тыуған тел", "Родной язык", "Тыуған тел — халыҡтың рухи байлығы.", True, "easy"),
        ("Берҙәмлек", "Единство, сплоченность", "Берҙәмлектә — көс.", True, "medium"),
        ("Ҡурай", "Курай (башкирский духовой инструмент)", "Ҡурай моңо — халыҡ күңеле.", True, "easy"),
        ("Сәсән", "Сэсэн, сказитель-импровизатор", "Сәсән телен ҡылыс киҫмәҫ.", False, "medium"),
        ("Батыр", "Богатырь, герой, храбрец", "Батыр бер үлер, ҡурҡаҡ мең үлер.", True, "easy"),
        ("Шишмә", "Родник, источник", "Китап — белем шишмәһе.", False, "easy"),
    ]

    for w, tr, ctx, fav, diff in demo_dict_words:
        db.add(DictionaryEntry(
            user_id=created_users[0].id,
            word=w,
            translation=tr,
            context=ctx,
            is_favorite=fav,
            difficulty=diff
        ))

    await db.commit()

    # 10. Import full external datasets (Hugging Face, BRSBS audiobooks, phrasebook)
    try:
        from app.seed.importer import import_all_datasets
        await import_all_datasets(db)
    except Exception as e:
        print(f"Warning: external dataset import encountered an error: {e}")
