(() => {
  const STORAGE_LANG = "idsai_site_lang";
  const LANGUAGES = ["ru", "en", "kk"];
  const LOCALES = {
    ru: "ru-RU",
    en: "en-US",
    kk: "kk-KZ",
  };
  const FLAG_ASSETS = {
    ru: "/dev/static/assets/lang-ru.svg",
    en: "/dev/static/assets/lang-en.svg",
    kk: "/dev/static/assets/lang-kk.svg",
  };
  const DOCK_MOUNT_SELECTORS = [
    ".header-actions",
    ".topbar-actions",
    ".head-actions",
    ".grading-hero__actions",
    ".prof-hero__panel",
    ".topbar-inner",
    ".topbar",
    ".admin-topbar",
    ".prof-header",
  ];
  const TRANSLATED_ATTRS = ["placeholder", "aria-label", "title", "alt"];
  const SKIP_SELECTOR = "[data-i18n-skip], .kb-markdown, script, style";
  const textSources = new WeakMap();
  const attrSources = new WeakMap();

  const UI_KEYS = {
    "prefs.title": {
      ru: "Языки интерфейса",
      en: "Interface languages",
      kk: "Интерфейс тілдері",
    },
    "prefs.language": {
      ru: "Язык",
      en: "Language",
      kk: "Тіл",
    },
    "lang.ru": {
      ru: "Русский",
      en: "Russian",
      kk: "Орысша",
    },
    "lang.en": {
      ru: "Английский",
      en: "English",
      kk: "Ағылшынша",
    },
    "lang.kk": {
      ru: "Казахский",
      en: "Kazakh",
      kk: "Қазақша",
    },
  };

  function normalizeBundle(source, entry = {}) {
    return {
      ru: entry.ru || source,
      en: entry.en || source,
      kk: entry.kk || source,
    };
  }

  function buildTranslationIndex() {
    const index = new Map();

    const addEntry = (source, entry) => {
      const bundle = normalizeBundle(source, entry);
      [source, bundle.ru, bundle.en, bundle.kk].forEach((variant) => {
        const normalized = normalizeInlineText(variant);
        if (!normalized) return;
        index.set(normalized, bundle);
      });
    };

    Object.values(UI_KEYS).forEach((entry) => {
      if (!entry || typeof entry !== "object") return;
      addEntry(entry.ru || entry.en || entry.kk || "", entry);
    });

    [EXACT_TRANSLATIONS, SUPPLEMENTAL_TRANSLATIONS].forEach((table) => {
      Object.entries(table).forEach(([source, entry]) => {
        addEntry(source, entry);
      });
    });

    return index;
  }

  const EXACT_TRANSLATIONS = {
    "Айболат · IDSAI Corp.": {
      en: "Aibolat · IDSAI Corp.",
      kk: "Айболат · IDSAI Corp.",
    },
    "Логотип IDSAI Corp.": {
      en: "IDSAI Corp. logo",
      kk: "IDSAI Corp. логотипі",
    },
    "IDSAI Corp. Settings": {
      ru: "IDSAI Corp. Настройки",
      en: "IDSAI Corp. Settings",
      kk: "IDSAI Corp. Баптаулар",
    },
    "IDSAI Corp. Projects": {
      ru: "IDSAI Corp. Проекты",
      en: "IDSAI Corp. Projects",
      kk: "IDSAI Corp. Жобалар",
    },
    "IDSAI Corp. Project": {
      ru: "IDSAI Corp. Проект",
      en: "IDSAI Corp. Project",
      kk: "IDSAI Corp. Жоба",
    },
    "IDSAI Corp. Profile": {
      ru: "IDSAI Corp. Профиль",
      en: "IDSAI Corp. Profile",
      kk: "IDSAI Corp. Профиль",
    },
    "IDSAI Corp. Professor Dashboard": {
      ru: "IDSAI Corp. Панель преподавателя",
      en: "IDSAI Corp. Professor Dashboard",
      kk: "IDSAI Corp. Оқытушы панелі",
    },
    "IDSAI Corp. Review Requests": {
      ru: "IDSAI Corp. Заявки на ревью",
      en: "IDSAI Corp. Review Requests",
      kk: "IDSAI Corp. Ревью өтініштері",
    },
    "IDSAI Corp. Project Grading": {
      ru: "IDSAI Corp. Оценивание проекта",
      en: "IDSAI Corp. Project Grading",
      kk: "IDSAI Corp. Жобаны бағалау",
    },
    "IDSAI Corp. Criteria Setup": {
      ru: "IDSAI Corp. Настройка критериев",
      en: "IDSAI Corp. Criteria Setup",
      kk: "IDSAI Corp. Критерий баптауы",
    },
    "IDSAI Corp. Login": {
      ru: "IDSAI Corp. Вход",
      en: "IDSAI Corp. Login",
      kk: "IDSAI Corp. Кіру",
    },
    "IDSAI Corp. Invites": {
      ru: "IDSAI Corp. Приглашения",
      en: "IDSAI Corp. Invites",
      kk: "IDSAI Corp. Шақырулар",
    },
    "IDSAI Corp. Admin": {
      ru: "IDSAI Corp. Админка",
      en: "IDSAI Corp. Admin",
      kk: "IDSAI Corp. Әкімші панелі",
    },
    "IDSAI Corp. Groups": {
      ru: "IDSAI Corp. Группы",
      en: "IDSAI Corp. Groups",
      kk: "IDSAI Corp. Топтар",
    },
    "IDSAI Corp. 404": {
      ru: "IDSAI Corp. 404",
      en: "IDSAI Corp. 404",
      kk: "IDSAI Corp. 404",
    },
    "Аватар Айболата": {
      en: "Aibolat avatar",
      kk: "Айболаттың аватары",
    },
    "Главная навигация": {
      en: "Main navigation",
      kk: "Негізгі навигация",
    },
    "Навигация страницы автора": {
      en: "Author page navigation",
      kk: "Автор бетінің навигациясы",
    },
    "Возможности": {
      en: "Features",
      kk: "Мүмкіндіктер",
    },
    "О проекте": {
      en: "About",
      kk: "Жоба туралы",
    },
    "Сообщество": {
      en: "Community",
      kk: "Қауымдастық",
    },
    "Контакты": {
      en: "Contacts",
      kk: "Байланыс",
    },
    "Вход": {
      en: "Sign in",
      kk: "Кіру",
    },
    "Присоединиться": {
      en: "Join",
      kk: "Қосылу",
    },
    "v2.4.0 стабильный релиз": {
      en: "v2.4.0 stable release",
      kk: "v2.4.0 тұрақты релиз",
    },
    "Управление проектами для студентов нового поколения": {
      en: "Project management for the next generation of students",
      kk: "Жаңа буын студенттеріне арналған жобаларды басқару",
    },
    "IDSAI Corp. Инициализируйте проект, распределяйте задачи и ведите командную разработку в едином пространстве.": {
      en: "IDSAI Corp. helps you kick off a project, assign tasks, and run team development in one shared workspace.",
      kk: "IDSAI Corp. жобаны бастап, міндеттерді бөліп, командалық әзірлеуді бір ортада жүргізуге көмектеседі.",
    },
    "Попробовать бесплатно": {
      en: "Try for free",
      kk: "Тегін көру",
    },
    "Связаться с нами": {
      en: "Contact us",
      kk: "Бізбен байланысу",
    },
    "План спринта #4": {
      en: "Sprint plan #4",
      kk: "Спринт жоспары №4",
    },
    "Задач: 14/20": {
      en: "Tasks: 14/20",
      kk: "Тапсырмалар: 14/20",
    },
    "Ваш рабочий процесс": {
      en: "Your workflow",
      kk: "Сіздің жұмыс ағымыңыз",
    },
    "Процессы": {
      en: "Processes",
      kk: "Процестер",
    },
    "Задачи": {
      en: "Tasks",
      kk: "Тапсырмалар",
    },
    "Тайм-трекинг": {
      en: "Time tracking",
      kk: "Уақыт есебі",
    },
    "Аналитика": {
      en: "Analytics",
      kk: "Аналитика",
    },
    "Оценка": {
      en: "Assessment",
      kk: "Бағалау",
    },
    "Канбан-доски, приоритеты и дедлайны в одном потоке работы.": {
      en: "Kanban boards, priorities, and deadlines in one workflow.",
      kk: "Kanban-тақталар, басымдықтар және дедлайндар бір ағында.",
    },
    "Отслеживание вклада участников и прозрачная история активности.": {
      en: "Track contribution and keep a transparent activity history.",
      kk: "Қатысушылар үлесін қадағалап, белсенділік тарихын ашық ұстаңыз.",
    },
    "Панель метрик прогресса, скорости спринтов и выполненных целей.": {
      en: "A metrics panel for progress, sprint velocity, and completed goals.",
      kk: "Прогресс, спринт жылдамдығы және орындалған мақсаттар метрикасы.",
    },
    "Система оценки качества работы и обратной связи между участниками.": {
      en: "A quality assessment and feedback system for participants.",
      kk: "Қатысушылар арасындағы жұмыс сапасын бағалау және кері байланыс жүйесі.",
    },
    "Подробнее →": {
      en: "Learn more →",
      kk: "Толығырақ →",
    },
    "создано для разработчиков": {
      en: "built for developers",
      kk: "әзірлеушілер үшін жасалған",
    },
    "Платформа для разработчиков, построенная на базе открытого кода.": {
      en: "A developer platform built on top of open source.",
      kk: "Ашық кодқа негізделген әзірлеушілер платформасы.",
    },
    "Документация": {
      en: "Documentation",
      kk: "Құжаттама",
    },
    "Совместная работа студентов и преподавателей": {
      en: "Collaboration between students and professors",
      kk: "Студенттер мен оқытушылардың бірлескен жұмысы",
    },
    "Прозрачная модерация и управление ролями": {
      en: "Transparent moderation and role management",
      kk: "Ашық модерация және рөлдерді басқару",
    },
    "Начать работу →": {
      en: "Start now →",
      kk: "Бастау →",
    },
    "отзывы": {
      en: "reviews",
      kk: "пікірлер",
    },
    "Что говорят наши пользователи": {
      en: "What our users say",
      kk: "Пайдаланушыларымыз не дейді",
    },
    "«Работа с IDSAI Corp. дала порядок в спринтах.»": {
      en: "\"Working with IDSAI Corp. brought structure to our sprints.\"",
      kk: "\"IDSAI Corp. спринттерімізге тәртіп әкелді.\"",
    },
    "Jim Halpert · Руководитель разработки": {
      en: "Jim Halpert · Engineering Manager",
      kk: "Jim Halpert · Әзірлеу жетекшісі",
    },
    "«Интерфейс интуитивный и помогает сфокусироваться.»": {
      en: "\"The interface is intuitive and helps you stay focused.\"",
      kk: "\"Интерфейс интуитивті және назарды ұстап тұруға көмектеседі.\"",
    },
    "Alexey Petrov · Студент": {
      en: "Alexey Petrov · Student",
      kk: "Alexey Petrov · Студент",
    },
    "«Лучший инструмент для учебных командных проектов.»": {
      en: "\"The best tool for academic team projects.\"",
      kk: "\"Оқу командалық жобаларына арналған ең жақсы құрал.\"",
    },
    "Ivan Ivanov · Преподаватель": {
      en: "Ivan Ivanov · Professor",
      kk: "Ivan Ivanov · Оқытушы",
    },
    "«Преподавателю удобно отслеживать реальный вклад.»": {
      en: "\"It is easy for professors to track real contribution.\"",
      kk: "\"Оқытушыға нақты үлесті бақылау ыңғайлы.\"",
    },
    "Sarah J. · Куратор": {
      en: "Sarah J. · Curator",
      kk: "Sarah J. · Куратор",
    },
    "Продукт": {
      en: "Product",
      kk: "Өнім",
    },
    "Отзывы": {
      en: "Reviews",
      kk: "Пікірлер",
    },
    "Войти": {
      en: "Sign in",
      kk: "Кіру",
    },
    "Привет, я Айболат": {
      en: "Hi, I'm Aibolat",
      kk: "Сәлем, мен Айболатпын",
    },
    "Бэкенд разработчик. Развиваю IDSAI Corp. как практическую платформу для управления студенческими проектами, совместной работы и прозрачных процессов разработки.": {
      en: "Backend developer. I'm building IDSAI Corp. as a practical platform for student project management, collaboration, and transparent engineering workflows.",
      kk: "Backend әзірлеушісі. Мен IDSAI Corp.-ты студенттік жобаларды басқаруға, бірлескен жұмысқа және ашық әзірлеу процестеріне арналған практикалық платформа ретінде дамытып жатырмын.",
    },
    "Резюме": {
      en: "Resume",
      kk: "Түйіндеме",
    },
    "Обо мне": {
      en: "About me",
      kk: "Мен туралы",
    },
    "Я сосредоточен на решении реальных рабочих задач в учебных командах: жизненный цикл проекта, модерация, уведомления и прозрачное отслеживание вклада. Мой приоритет - надёжная бэкенд-архитектура с понятным пользовательским опытом поверх неё.": {
      en: "I focus on solving real operational problems in academic teams: the project lifecycle, moderation, notifications, and transparent contribution tracking. My priority is reliable backend architecture with a clear user experience on top.",
      kk: "Мен оқу командаларындағы нақты жұмыс міндеттерін шешуге назар аударамын: жоба өмірлік циклі, модерация, хабарламалар және үлесті ашық бақылау. Мен үшін басымдық - түсінікті пайдаланушы тәжірибесі бар сенімді backend архитектурасы.",
    },
    "Последние проекты": {
      en: "Recent projects",
      kk: "Соңғы жобалар",
    },
    "Превью проекта IDSAI Core API": {
      en: "IDSAI Core API preview",
      kk: "IDSAI Core API жобасының көрінісі",
    },
    "Мультитенантный бэкенд с JWT, RBAC, рабочими процессами и уведомлениями.": {
      en: "A multi-tenant backend with JWT, RBAC, workflows, and notifications.",
      kk: "JWT, RBAC, жұмыс процестері мен хабарламалары бар көптенантты backend.",
    },
    "Превью проекта Padel App": {
      en: "Padel App preview",
      kk: "Padel App жобасының көрінісі",
    },
    "Корты, система бронирования и управление игроками.": {
      en: "Courts, booking flows, and player management.",
      kk: "Корттар, брондау жүйесі және ойыншыларды басқару.",
    },
    "Технологии": {
      en: "Technologies",
      kk: "Технологиялар",
    },
    "Связаться": {
      en: "Get in touch",
      kk: "Байланысу",
    },
    "Если есть вопросы, предложения или просто хотите пообщаться - пишите! Я открыт к сотрудничеству, новым идеям и интересным проектам. Буду рад обсудить, как я могу помочь вам или вашей команде достичь целей.": {
      en: "If you have questions, ideas, or just want to talk, write to me. I'm open to collaboration, new ideas, and interesting projects, and I'd be glad to discuss how I can help you or your team reach your goals.",
      kk: "Сұрақтарыңыз, ұсыныстарыңыз болса немесе жай сөйлескіңіз келсе, жазыңыз. Мен ынтымақтастыққа, жаңа идеяларға және қызықты жобаларға ашықпын, сізге не командаңызға мақсатқа жетуге қалай көмектесе алатынымды қуана талқылаймын.",
    },
    "Ваши данные": {
      en: "Your contact details",
      kk: "Сіздің байланыс деректеріңіз",
    },
    "Имя, телефон, email или Telegram": {
      en: "Name, phone, email, or Telegram",
      kk: "Аты-жөніңіз, телефон, email немесе Telegram",
    },
    "Чем я могу помочь?": {
      en: "How can I help?",
      kk: "Мен қалай көмектесе аламын?",
    },
    "Опишите задачу, проект или вопрос": {
      en: "Describe the task, project, or question",
      kk: "Міндетті, жобаны немесе сұрақты сипаттаңыз",
    },
    "Отправить": {
      en: "Send",
      kk: "Жіберу",
    },
    "Профессиональная платформа для учебных проектных команд.": {
      en: "A professional platform for academic project teams.",
      kk: "Оқу жобалық командаларына арналған кәсіби платформа.",
    },
    "Page Missing": {
      ru: "Страница отсутствует",
      en: "Page Missing",
      kk: "Бет табылмады",
    },
    "Страница не найдена": {
      en: "Page not found",
      kk: "Бет табылмады",
    },
    "Адрес изменился, ссылка устарела или запрошенный объект больше не существует.": {
      en: "The address changed, the link is outdated, or the requested item no longer exists.",
      kk: "Мекенжай өзгерген, сілтеме ескірген немесе сұралған нысан енді жоқ.",
    },
    "Запрошенный путь": {
      en: "Requested path",
      kk: "Сұралған жол",
    },
    "В кабинет": {
      en: "Open workspace",
      kk: "Кеңістікке өту",
    },
    "На главную": {
      en: "Back home",
      kk: "Басты бетке",
    },
    "Информация о платформе": {
      en: "Platform information",
      kk: "Платформа туралы ақпарат",
    },
    "Совместная образовательная экосистема ENU × IDSAI Corp.": {
      en: "A collaborative educational ecosystem by ENU × IDSAI Corp.",
      kk: "ENU × IDSAI Corp. бірлескен білім беру экожүйесі.",
    },
    "Платформа для студентов, преподавателей и командных проектных команд.": {
      en: "A platform for students, professors, and project teams.",
      kk: "Студенттерге, оқытушыларға және жобалық командаларға арналған платформа.",
    },
    "Auth tabs": {
      ru: "Вкладки авторизации",
      en: "Auth tabs",
      kk: "Авторизация қойындылары",
    },
    "Регистрация": {
      en: "Sign up",
      kk: "Тіркелу",
    },
    "ID студента / Email": {
      en: "Student ID / Email",
      kk: "Студент ID / Email",
    },
    "Пароль": {
      en: "Password",
      kk: "Құпиясөз",
    },
    "Запомнить меня": {
      en: "Remember me",
      kk: "Мені есте сақтау",
    },
    "Забыли пароль?": {
      en: "Forgot password?",
      kk: "Құпиясөзді ұмыттыңыз ба?",
    },
    "Войти →": {
      en: "Sign in →",
      kk: "Кіру →",
    },
    "ФИО": {
      en: "Full name",
      kk: "Аты-жөні",
    },
    "Кафедра": {
      en: "Department",
      kk: "Кафедра",
    },
    "Группа": {
      en: "Group",
      kk: "Топ",
    },
    "Выберите кафедру": {
      en: "Choose a department",
      kk: "Кафедраны таңдаңыз",
    },
    "Сначала выберите кафедру": {
      en: "Choose a department first",
      kk: "Алдымен кафедраны таңдаңыз",
    },
    "Код группы:": {
      en: "Group code:",
      kk: "Топ коды:",
    },
    "Email": {
      en: "Email",
      kk: "Email",
    },
    "минимум 10 символов": {
      en: "minimum 10 characters",
      kk: "кемі 10 таңба",
    },
    "Повторите пароль": {
      en: "Repeat password",
      kk: "Құпиясөзді қайталаңыз",
    },
    "Создать аккаунт ☐": {
      en: "Create account ☐",
      kk: "Аккаунт ашу ☐",
    },
    "© 2026 IDSAI Corp. Все системы работают штатно.": {
      en: "© 2026 IDSAI Corp. All systems are operating normally.",
      kk: "© 2026 IDSAI Corp. Барлық жүйелер қалыпты жұмыс істеп тұр.",
    },
    "Конфиденциальность": {
      en: "Privacy",
      kk: "Құпиялық",
    },
    "Условия": {
      en: "Terms",
      kk: "Шарттар",
    },
    "Статус": {
      en: "Status",
      kk: "Күйі",
    },
    "Project": {
      ru: "Проект",
      en: "Project",
      kk: "Жоба",
    },
    "dashboard": {
      ru: "дашборд",
      en: "dashboard",
      kk: "дашборд",
    },
    "projects": {
      ru: "проекты",
      en: "projects",
      kk: "жобалар",
    },
    "profile": {
      ru: "профиль",
      en: "profile",
      kk: "профиль",
    },
    "Refresh": {
      ru: "Обновить",
      en: "Refresh",
      kk: "Жаңарту",
    },
    "Search projects (Cmd+K)": {
      ru: "Поиск проектов (Cmd+K)",
      en: "Search projects (Cmd+K)",
      kk: "Жобаларды іздеу (Cmd+K)",
    },
    "Student workspace": {
      ru: "Студенческое рабочее пространство",
      en: "Student workspace",
      kk: "Студенттің жұмыс кеңістігі",
    },
    "Student Workspace": {
      ru: "Студенческое рабочее пространство",
      en: "Student Workspace",
      kk: "Студенттің жұмыс кеңістігі",
    },
    "Admin Console": {
      ru: "Консоль администратора",
      en: "Admin Console",
      kk: "Әкімші консолі",
    },
    "Professor Workspace": {
      ru: "Пространство преподавателя",
      en: "Professor Workspace",
      kk: "Оқытушының жұмыс кеңістігі",
    },
    "Мои проекты": {
      en: "My projects",
      kk: "Менің жобаларым",
    },
    "Управляйте репозиториями и отслеживайте прогресс команды.": {
      en: "Manage repositories and keep track of team progress.",
      kk: "Репозиторийлерді басқарып, команда прогресін бақылаңыз.",
    },
    "Создать проект": {
      en: "Create project",
      kk: "Жоба құру",
    },
    "Все": {
      en: "All",
      kk: "Барлығы",
    },
    "Проверка": {
      en: "Review",
      kk: "Тексеру",
    },
    "Готово": {
      en: "Done",
      kk: "Дайын",
    },
    "View mode": {
      ru: "Режим отображения",
      en: "View mode",
      kk: "Көрініс режимі",
    },
    "Режим сетки": {
      en: "Grid view",
      kk: "Тор көрінісі",
    },
    "Режим списка": {
      en: "List view",
      kk: "Тізім көрінісі",
    },
    "Поиск проектов сообщества": {
      en: "Search community projects",
      kk: "Қауымдастық жобаларын іздеу",
    },
    "Найти проект по названию, стеку...": {
      en: "Find a project by title or stack...",
      kk: "Жобаны атауы не стекі бойынша табыңыз...",
    },
    "Фильтры:": {
      en: "Filters:",
      kk: "Сүзгілер:",
    },
    "Все технологии": {
      en: "All technologies",
      kk: "Барлық технологиялар",
    },
    "Любая сложность": {
      en: "Any difficulty",
      kk: "Кез келген күрделілік",
    },
    "status console": {
      ru: "консоль статуса",
      en: "status console",
      kk: "күй консолі",
    },
    "console.js": {
      ru: "console.js",
      en: "console.js",
      kk: "console.js",
    },
    "waiting for events...": {
      ru: "ожидание событий...",
      en: "waiting for events...",
      kk: "оқиғаларды күту...",
    },
    "Создание нового проекта": {
      en: "Create a new project",
      kk: "Жаңа жоба құру",
    },
    "Инициализируйте рабочее пространство и настройте публичный или приватный доступ.": {
      en: "Initialize the workspace and configure public or private access.",
      kk: "Жұмыс кеңістігін бастап, жария немесе жеке қолжетімділікті баптаңыз.",
    },
    "Название проекта": {
      en: "Project title",
      kk: "Жоба атауы",
    },
    "Краткое описание": {
      en: "Short description",
      kk: "Қысқаша сипаттама",
    },
    "README / Полное описание": {
      en: "README / Full description",
      kk: "README / Толық сипаттама",
    },
    "Введите подробное описание проекта...": {
      en: "Enter a detailed project description...",
      kk: "Жобаның толық сипаттамасын енгізіңіз...",
    },
    "Ссылка на репозиторий GitHub": {
      en: "GitHub repository link",
      kk: "GitHub репозиторий сілтемесі",
    },
    "Технологический стек": {
      en: "Technology stack",
      kk: "Технологиялық стек",
    },
    "Например: Go, React, PostgreSQL": {
      en: "For example: Go, React, PostgreSQL",
      kk: "Мысалы: Go, React, PostgreSQL",
    },
    "Приватность проекта": {
      en: "Project visibility",
      kk: "Жоба құпиялығы",
    },
    "Публичный": {
      en: "Public",
      kk: "Жария",
    },
    "Доступен всем участникам системы": {
      en: "Available to all users in the system",
      kk: "Жүйедегі барлық қатысушыларға қолжетімді",
    },
    "Приватный": {
      en: "Private",
      kk: "Жеке",
    },
    "Только для вашей группы": {
      en: "Visible only to your group",
      kk: "Тек сіздің тобыңызға ғана",
    },
    "Номер группы": {
      en: "Group number",
      kk: "Топ нөмірі",
    },
    "Отмена": {
      en: "Cancel",
      kk: "Бас тарту",
    },
    "Аккаунт и безопасность": {
      en: "Account and security",
      kk: "Аккаунт және қауіпсіздік",
    },
    "Настройки учетной записи": {
      en: "Account settings",
      kk: "Аккаунт баптаулары",
    },
    "Управляйте личной информацией, email и безопасностью аккаунта.": {
      en: "Manage your personal information, email, and account security.",
      kk: "Жеке ақпаратты, email-ды және аккаунт қауіпсіздігін басқарыңыз.",
    },
    "Личные данные": {
      en: "Personal details",
      kk: "Жеке деректер",
    },
    "Публичный профиль и контактный email.": {
      en: "Your public profile and contact email.",
      kk: "Жария профиліңіз және байланыс email-ы.",
    },
    "Рекомендуется квадратное изображение, минимум 400x400 px.": {
      en: "A square image is recommended, at least 400x400 px.",
      kk: "Кемі 400x400 px болатын шаршы сурет ұсынылады.",
    },
    "Загрузить новое": {
      en: "Upload new",
      kk: "Жаңасын жүктеу",
    },
    "Удалить": {
      en: "Delete",
      kk: "Жою",
    },
    "Ваше полное имя": {
      en: "Your full name",
      kk: "Толық атыңыз",
    },
    "Имя будет показано в карточках проектов и комментариях.": {
      en: "Your name will be shown in project cards and comments.",
      kk: "Атыңыз жоба карточкалары мен пікірлерде көрсетіледі.",
    },
    "Сохранить профиль": {
      en: "Save profile",
      kk: "Профильді сақтау",
    },
    "Учебная группа": {
      en: "Academic group",
      kk: "Оқу тобы",
    },
    "Смена группы выполняется только после подтверждения администратора.": {
      en: "A group change is applied only after an administrator approves it.",
      kk: "Топты ауыстыру әкімші растағаннан кейін ғана орындалады.",
    },
    "Текущая кафедра": {
      en: "Current department",
      kk: "Ағымдағы кафедра",
    },
    "Текущая группа": {
      en: "Current group",
      kk: "Ағымдағы топ",
    },
    "Новая кафедра": {
      en: "New department",
      kk: "Жаңа кафедра",
    },
    "Новая группа": {
      en: "New group",
      kk: "Жаңа топ",
    },
    "Отправить заявку на смену группы": {
      en: "Request a group change",
      kk: "Топты ауыстыруға өтініш жіберу",
    },
    "История заявок": {
      en: "Request history",
      kk: "Өтініш тарихы",
    },
    "Текущий email": {
      en: "Current email",
      kk: "Ағымдағы email",
    },
    "Текущий email остается активным до подтверждения нового адреса.": {
      en: "Your current email stays active until the new address is confirmed.",
      kk: "Жаңа мекенжай расталғанша ағымдағы email белсенді болып қалады.",
    },
    "Новый email": {
      en: "New email",
      kk: "Жаңа email",
    },
    "Отправить подтверждение": {
      en: "Send confirmation",
      kk: "Растауды жіберу",
    },
    "Отправить повторно": {
      en: "Send again",
      kk: "Қайта жіберу",
    },
    "Код подтверждения": {
      en: "Confirmation code",
      kk: "Растау коды",
    },
    "Введите 6-значный код из письма": {
      en: "Enter the 6-digit code from the email",
      kk: "Хаттағы 6 таңбалы кодты енгізіңіз",
    },
    "Подтвердить email": {
      en: "Confirm email",
      kk: "Email растау",
    },
    "Безопасность": {
      en: "Security",
      kk: "Қауіпсіздік",
    },
    "Смена пароля с проверкой текущего значения.": {
      en: "Change your password after verifying the current one.",
      kk: "Ағымдағы құпиясөзді тексеріп, оны ауыстырыңыз.",
    },
    "Текущий пароль": {
      en: "Current password",
      kk: "Ағымдағы құпиясөз",
    },
    "Новый пароль": {
      en: "New password",
      kk: "Жаңа құпиясөз",
    },
    "Подтверждение пароля": {
      en: "Password confirmation",
      kk: "Құпиясөзді растау",
    },
    "Минимум 8 символов": {
      en: "At least 8 characters",
      kk: "Кемі 8 таңба",
    },
    "Хотя бы одна буква": {
      en: "At least one letter",
      kk: "Кемі бір әріп",
    },
    "Хотя бы одна цифра": {
      en: "At least one digit",
      kk: "Кемі бір сан",
    },
    "Обновить пароль": {
      en: "Update password",
      kk: "Құпиясөзді жаңарту",
    },
    "Контакты": {
      en: "Contacts",
      kk: "Байланыстар",
    },
    "Редактировать профиль": {
      en: "Edit profile",
      kk: "Профильді өңдеу",
    },
    "Обновить аватар": {
      en: "Update avatar",
      kk: "Аватарды жаңарту",
    },
    "Удалить аватар": {
      en: "Remove avatar",
      kk: "Аватарды жою",
    },
    "Добавьте специализацию": {
      en: "Add a specialization",
      kk: "Мамандандыруды қосыңыз",
    },
    "Профиль 0%": {
      en: "Profile 0%",
      kk: "Профиль 0%",
    },
    "Кафедра: —": {
      en: "Department: —",
      kk: "Кафедра: —",
    },
    "Группа: —": {
      en: "Group: —",
      kk: "Топ: —",
    },
    "Сводка профиля": {
      en: "Profile summary",
      kk: "Профиль қорытындысы",
    },
    "Направления": {
      en: "Focus areas",
      kk: "Бағыттар",
    },
    "в фокусе": {
      en: "in focus",
      kk: "фокуста",
    },
    "готовы": {
      en: "ready",
      kk: "дайын",
    },
    "Не указан": {
      en: "Not set",
      kk: "Көрсетілмеген",
    },
    "загрузка": {
      en: "loading",
      kk: "жүктелуде",
    },
    "Основная информация": {
      en: "Main information",
      kk: "Негізгі ақпарат",
    },
    "Полное имя": {
      en: "Full name",
      kk: "Толық аты-жөні",
    },
    "Фокус": {
      en: "Focus",
      kk: "Фокус",
    },
    "Например: Backend Developer": {
      en: "For example: Backend Developer",
      kk: "Мысалы: Backend Developer",
    },
    "О себе": {
      en: "About you",
      kk: "Өзіңіз туралы",
    },
    "Расскажите о своем опыте и сильных сторонах.": {
      en: "Tell us about your experience and strengths.",
      kk: "Тәжірибеңіз бен мықты жақтарыңыз туралы жазыңыз.",
    },
    "Предпочтительная роль": {
      en: "Preferred role",
      kk: "Қалаулы рөл",
    },
    "Выберите роль": {
      en: "Choose a role",
      kk: "Рөлді таңдаңыз",
    },
    "Семестр / курс": {
      en: "Semester / year",
      kk: "Семестр / курс",
    },
    "Контакты": {
      en: "Contacts",
      kk: "Байланыстар",
    },
    "Портфолио / CV": {
      en: "Portfolio / CV",
      kk: "Портфолио / CV",
    },
    "Сводка структуры": {
      en: "Structure summary",
      kk: "Құрылым қорытындысы",
    },
    "Пользователи": {
      en: "Users",
      kk: "Пайдаланушылар",
    },
    "Проекты": {
      en: "Projects",
      kk: "Жобалар",
    },
    "Группы": {
      en: "Groups",
      kk: "Топтар",
    },
    "База знаний": {
      en: "Knowledge base",
      kk: "Білім қоры",
    },
    "Дашборд": {
      en: "Dashboard",
      kk: "Басқару панелі",
    },
    "Заявки": {
      en: "Requests",
      kk: "Өтініштер",
    },
    "Заявки на ревью": {
      en: "Review requests",
      kk: "Ревью өтініштері",
    },
    "Критерии": {
      en: "Criteria",
      kk: "Критерийлер",
    },
    "Оценивание": {
      en: "Grading",
      kk: "Бағалау",
    },
    "Поиск по базе знаний…": {
      en: "Search the knowledge base…",
      kk: "Білім қорынан іздеу…",
    },
    "К базе знаний": {
      en: "Back to knowledge base",
      kk: "Білім қорына оралу",
    },
    "IDSAI Corp. — База знаний": {
      en: "IDSAI Corp. — Knowledge base",
      kk: "IDSAI Corp. — Білім қоры",
    },
    "IDSAI Corp. — Статья": {
      en: "IDSAI Corp. — Article",
      kk: "IDSAI Corp. — Мақала",
    },
    "Пока нет уведомлений": {
      en: "No notifications yet",
      kk: "Әзірге хабарландырулар жоқ",
    },
    "Уведомление": {
      en: "Notification",
      kk: "Хабарландыру",
    },
    "Новое событие в системе.": {
      en: "A new event in the system.",
      kk: "Жүйеде жаңа оқиға болды.",
    },
    "Успех": {
      en: "Success",
      kk: "Сәтті",
    },
    "Внимание": {
      en: "Attention",
      kk: "Назар аударыңыз",
    },
    "Ошибка": {
      en: "Error",
      kk: "Қате",
    },
    "Инфо": {
      en: "Info",
      kk: "Ақпарат",
    },
    "только что": {
      en: "just now",
      kk: "жаңа ғана",
    },
    "Подготовка": {
      en: "Preparation",
      kk: "Дайындық",
    },
    "Набор": {
      en: "Recruiting",
      kk: "Іріктеу",
    },
    "В работе": {
      en: "In progress",
      kk: "Жұмыста",
    },
    "Завершен": {
      en: "Completed",
      kk: "Аяқталды",
    },
    "Закрыт": {
      en: "Closed",
      kk: "Жабық",
    },
    "Ожидает": {
      en: "Pending",
      kk: "Күтуде",
    },
    "Одобрено": {
      en: "Approved",
      kk: "Мақұлданды",
    },
    "Отклонено": {
      en: "Rejected",
      kk: "Қабылданбады",
    },
    "Новичок": {
      en: "Beginner",
      kk: "Бастаушы",
    },
    "Средний": {
      en: "Intermediate",
      kk: "Орташа",
    },
    "Продвинутый": {
      en: "Advanced",
      kk: "Жоғары деңгей",
    },
    "Нет проектов под текущий фильтр.": {
      en: "No projects match the current filter.",
      kk: "Ағымдағы сүзгіге сай жоба жоқ.",
    },
    "Нет проектов под текущий фильтр. Нажми, чтобы создать новый.": {
      en: "No projects match the current filter. Click to create a new one.",
      kk: "Ағымдағы сүзгіге сай жоба жоқ. Жаңасын құру үшін басыңыз.",
    },
    "Новый проект": {
      en: "New project",
      kk: "Жаңа жоба",
    },
    "Без описания": {
      en: "No description",
      kk: "Сипаттамасыз",
    },
    "Публичные проекты не найдены.": {
      en: "No public projects found.",
      kk: "Жария жобалар табылмады.",
    },
    "Открыть проект": {
      en: "Open project",
      kk: "Жобаны ашу",
    },
    "Набор закрыт": {
      en: "Recruiting closed",
      kk: "Іріктеу жабық",
    },
    "Обзор проектов": {
      en: "Project overview",
      kk: "Жобалар шолуы",
    },
    "Введите название проекта для подтверждения.": {
      en: "Enter the project title to confirm.",
      kk: "Растау үшін жоба атауын енгізіңіз.",
    },
    "Название проекта обязательно": {
      en: "Project title is required",
      kk: "Жоба атауы міндетті",
    },
    "Сессия истекла. Войди снова.": {
      en: "Your session expired. Sign in again.",
      kk: "Сессияңыз аяқталды. Қайта кіріңіз.",
    },
    "Нет выбранного проекта": {
      en: "No project selected",
      kk: "Жоба таңдалмаған",
    },
    "Administrator": {
      ru: "Администратор",
      en: "Administrator",
      kk: "Әкімші",
    },
    "Workspace: CORE": {
      ru: "Контур: CORE",
      en: "Workspace: CORE",
      kk: "Контур: CORE",
    },
    "Live RBAC": {
      ru: "RBAC онлайн",
      en: "Live RBAC",
      kk: "RBAC онлайн",
    },
    "Control center": {
      ru: "Центр управления",
      en: "Control center",
      kk: "Басқару орталығы",
    },
    "Quick action 01": {
      ru: "Быстрое действие 01",
      en: "Quick action 01",
      kk: "Жылдам әрекет 01",
    },
    "Quick action 02": {
      ru: "Быстрое действие 02",
      en: "Quick action 02",
      kk: "Жылдам әрекет 02",
    },
    "Quick action 03": {
      ru: "Быстрое действие 03",
      en: "Quick action 03",
      kk: "Жылдам әрекет 03",
    },
    "Quick action 04": {
      ru: "Быстрое действие 04",
      en: "Quick action 04",
      kk: "Жылдам әрекет 04",
    },
    "Identity / Access": {
      ru: "Идентификация / доступ",
      en: "Identity / Access",
      kk: "Идентификация / қолжетімділік",
    },
    "Projects / Flow": {
      ru: "Проекты / поток",
      en: "Projects / Flow",
      kk: "Жобалар / ағын",
    },
    "Operations / Focus": {
      ru: "Операции / фокус",
      en: "Operations / Focus",
      kk: "Операциялар / фокус",
    },
    "Project cover": {
      ru: "Обложка проекта",
      en: "Project cover",
      kk: "Жоба мұқабасы",
    },
    "Project tabs": {
      ru: "Вкладки проекта",
      en: "Project tabs",
      kk: "Жоба қойындылары",
    },
    "Backlog": {
      ru: "Бэклог",
      en: "Backlog",
      kk: "Бэклог",
    },
    "Review cockpit": {
      ru: "Панель ревью",
      en: "Review cockpit",
      kk: "Ревью панелі",
    },
    "SUPER_ADMIN": {
      ru: "СУПЕР_АДМИН",
      en: "SUPER_ADMIN",
      kk: "СУПЕР_ӘКІМШІ",
    },
    "PUBLIC": {
      ru: "ПУБЛИЧНЫЙ",
      en: "PUBLIC",
      kk: "ЖАРИЯ",
    },
    "PRIVATE": {
      ru: "ПРИВАТНЫЙ",
      en: "PRIVATE",
      kk: "ЖЕКЕ",
    },
    "DRAFT": {
      ru: "ЧЕРНОВИК",
      en: "DRAFT",
      kk: "БАСТАПҚЫ НҰСҚА",
    },
    "RECRUITMENT": {
      ru: "НАБОР",
      en: "RECRUITMENT",
      kk: "ІРІКТЕУ",
    },
    "ACTIVE": {
      ru: "АКТИВНЫЙ",
      en: "ACTIVE",
      kk: "БЕЛСЕНДІ",
    },
    "GRADING": {
      ru: "ОЦЕНИВАНИЕ",
      en: "GRADING",
      kk: "БАҒАЛАУ",
    },
    "COMPLETED": {
      ru: "ЗАВЕРШЕН",
      en: "COMPLETED",
      kk: "АЯҚТАЛДЫ",
    },
    "Created from projects dashboard": {
      ru: "Создано из панели проектов",
      en: "Created from projects dashboard",
      kk: "Жобалар панелінен жасалды",
    },
    "Последнее обновление: {date}": {
      en: "Last updated: {date}",
      kk: "Соңғы жаңарту: {date}",
    },
    "Public": {
      ru: "Публичный",
      en: "Public",
      kk: "Жария",
    },
    "Private": {
      ru: "Приватный",
      en: "Private",
      kk: "Жеке",
    },
    "＋ Создать проект": {
      en: "＋ Create project",
      kk: "＋ Жоба құру",
    },
    "★ В избранное": {
      en: "★ Add to favorites",
      kk: "★ Таңдаулыға қосу",
    },
    "★ В избранном": {
      en: "★ In favorites",
      kk: "★ Таңдаулыда",
    },
    "✎ Редактировать": {
      en: "✎ Edit",
      kk: "✎ Өңдеу",
    },
    "Жизненный цикл проекта": {
      en: "Project lifecycle",
      kk: "Жоба өмірлік циклі",
    },
    "Путь проекта от черновика до публикации итоговой оценки.": {
      en: "The project path from draft to the final published grade.",
      kk: "Жобаның бастапқы нұсқадан қорытынды баға жарияланғанға дейінгі жолы.",
    },
    "Подсказка по этапу": {
      en: "Stage hint",
      kk: "Кезең бойынша кеңес",
    },
    "Что нужно для следующего этапа": {
      en: "What is needed for the next stage",
      kk: "Келесі кезеңге не қажет",
    },
    "Покажем только ближайшие требования, чтобы команде было проще двигаться по этапам.": {
      en: "We show only the nearest requirements so the team can move through stages more easily.",
      kk: "Команда кезеңдерден оңай өтуі үшін тек жақын талаптарды көрсетеміз.",
    },
    "Требования следующего этапа": {
      en: "Next stage requirements",
      kk: "Келесі кезең талаптары",
    },
    "Что нужно, чтобы перейти в работу": {
      en: "What you need to move into active work",
      kk: "Жұмыс кезеңіне өту үшін не қажет",
    },
    "Видны только требования ближайшего перехода. Выполненные пункты сразу зачеркиваются.": {
      en: "Only the requirements for the nearest transition are shown. Completed items are crossed out immediately.",
      kk: "Тек келесі ауысуға қажет талаптар көрінеді. Орындалған тармақтар бірден сызылып көрсетіледі.",
    },
    "Набрать команду по всем ролям": {
      en: "Fill the team for all roles",
      kk: "Барлық рөлдер бойынша команданы толықтыру",
    },
    "Подтвердить преподавателя": {
      en: "Confirm the professor",
      kk: "Оқытушыны растау",
    },
    "Преподаватель должен добавить критерии оценки": {
      en: "The professor must add grading criteria",
      kk: "Оқытушы бағалау критерийлерін қосуы керек",
    },
    "Назначьте преподавателя на проект.": {
      en: "Assign a professor to the project.",
      kk: "Жобаға оқытушы тағайындаңыз.",
    },
    "Что нужно, чтобы отправить проект на оценивание": {
      en: "What is needed to submit the project for grading",
      kk: "Жобаны бағалауға жіберу үшін не қажет",
    },
    "Команда видит только то, что блокирует ближайшую сдачу проекта.": {
      en: "The team sees only what blocks the nearest project submission.",
      kk: "Команда тек жобаны жақын тапсыруға кедергі болып тұрған нәрселерді көреді.",
    },
    "Подтверждение преподавателя сохранено": {
      en: "The professor confirmation is in place",
      kk: "Оқытушының растауы сақталған",
    },
    "Сначала назначьте преподавателя.": {
      en: "Assign a professor first.",
      kk: "Алдымен оқытушыны тағайындаңыз.",
    },
    "Создать хотя бы одну задачу": {
      en: "Create at least one task",
      kk: "Кемінде бір тапсырма құру",
    },
    "Первую задачу обычно создает тимлид или task manager, иначе проект нельзя отправить на оценивание.": {
      en: "The first task is usually created by the team lead or task manager, otherwise the project cannot be submitted for grading.",
      kk: "Алғашқы тапсырманы әдетте тимлид немесе task manager жасайды, әйтпесе жобаны бағалауға жіберу мүмкін емес.",
    },
    "Закрыть все задачи": {
      en: "Close all tasks",
      kk: "Барлық тапсырманы жабу",
    },
    "Сначала добавьте задачи в канбан. Это делает тимлид или task manager.": {
      en: "First add tasks to the kanban. This is usually done by the team lead or task manager.",
      kk: "Алдымен kanban-ға тапсырмалар қосыңыз. Мұны әдетте тимлид немесе task manager жасайды.",
    },
    "Что нужно, чтобы завершить проект": {
      en: "What is needed to complete the project",
      kk: "Жобаны аяқтау үшін не қажет",
    },
    "На этом этапе команда видит только ближайшие требования до публикации итоговой оценки.": {
      en: "At this stage the team sees only the nearest requirements before the final grade is published.",
      kk: "Бұл кезеңде команда қорытынды баға жарияланғанға дейінгі ең жақын талаптарды ғана көреді.",
    },
    "Критерии оценки настроены": {
      en: "The grading criteria are configured",
      kk: "Бағалау критерийлері бапталған",
    },
    "Оценки выставлены по всем критериям": {
      en: "Grades are set for all criteria",
      kk: "Барлық критерий бойынша бағалар қойылған",
    },
    "Финал достигнут": {
      en: "Final stage reached",
      kk: "Финалға жетті",
    },
    "Проект завершен": {
      en: "Project completed",
      kk: "Жоба аяқталды",
    },
    "Все обязательные шаги выполнены. Здесь остается только итоговый статус проекта.": {
      en: "All required steps are completed. Only the final project status remains here.",
      kk: "Барлық міндетті қадам орындалды. Мұнда тек жобаның қорытынды күйі ғана қалады.",
    },
    "Итоговая оценка опубликована": {
      en: "The final grade is published",
      kk: "Қорытынды баға жарияланды",
    },
    "Карточка проекта зафиксирована как завершенный кейс.": {
      en: "The project card is locked as a completed case.",
      kk: "Жоба картасы аяқталған кейс ретінде бекітілді.",
    },
    "Что подготовить перед открытием набора": {
      en: "What to prepare before opening recruitment",
      kk: "Іріктеуді ашар алдында не дайындау керек",
    },
    "Блок показывает только ближайший шаг, чтобы команде было понятнее, что делать прямо сейчас.": {
      en: "The block shows only the nearest step so the team can better understand what to do right now.",
      kk: "Блок командаға дәл қазір не істеу керегін түсініктірек ету үшін тек келесі қадамды көрсетеді.",
    },
    "Заполнить описание или README проекта": {
      en: "Fill in the project description or README",
      kk: "Жоба сипаттамасын немесе README-ді толтыру",
    },
    "Базовое описание проекта уже есть.": {
      en: "The project already has a basic description.",
      kk: "Жобада негізгі сипаттама қазірдің өзінде бар.",
    },
    "Кратко опишите идею, стек и ожидаемый результат.": {
      en: "Briefly describe the idea, stack, and expected result.",
      kk: "Идеяны, стекті және күтілетін нәтижені қысқаша сипаттаңыз.",
    },
    "Добавить хотя бы одну роль в команду": {
      en: "Add at least one role to the team",
      kk: "Командаға кемінде бір рөл қосу",
    },
    "Создайте роли, чтобы открыть набор осознанно.": {
      en: "Create roles first so recruitment can be opened deliberately.",
      kk: "Іріктеуді саналы түрде ашу үшін алдымен рөлдерді жасаңыз.",
    },
    "Ближайшие требования пока не определены.": {
      en: "The nearest requirements are not defined yet.",
      kk: "Ең жақын талаптар әлі анықталмаған.",
    },
    "Сейчас главное закрыть все роли": {
      en: "The main goal now is to fill every role",
      kk: "Қазір ең бастысы барлық рөлді жабу",
    },
    "Команда уже в рабочем составе": {
      en: "The team is already in working shape",
      kk: "Команда қазірдің өзінде жұмыс құрамында",
    },
    "Состав команды должен быть понятен для оценки": {
      en: "The team composition should be clear for grading",
      kk: "Бағалау үшін команда құрамы түсінікті болуы керек",
    },
    "Это финальный состав команды": {
      en: "This is the final team composition",
      kk: "Бұл команданың финалдық құрамы",
    },
    "Сначала подготовьте каркас команды": {
      en: "Prepare the team structure first",
      kk: "Алдымен команда қаңқасын дайындаңыз",
    },
    "Создайте первую задачу": {
      en: "Create the first task",
      kk: "Алғашқы тапсырманы құрыңыз",
    },
    "Доведите канбан до конца": {
      en: "Finish the kanban board",
      kk: "Kanban-ды соңына дейін жеткізіңіз",
    },
    "Все задачи уже закрыты": {
      en: "All tasks are already closed",
      kk: "Барлық тапсырма қазірдің өзінде жабылған",
    },
    "Задачи здесь работают как история выполнения": {
      en: "Tasks work here as an execution history",
      kk: "Мұнда тапсырмалар орындалу тарихы ретінде қызмет етеді",
    },
    "Канбан зафиксирован как история проекта": {
      en: "The kanban board is locked as project history",
      kk: "Kanban жоба тарихы ретінде бекітілген",
    },
    "До запуска задачи можно не расписывать подробно": {
      en: "Before launch you do not need to describe tasks in detail",
      kk: "Іске қоспай тұрып тапсырмаларды егжей-тегжейлі жазудың қажеті жоқ",
    },
    "По этим критериям идет финальная оценка": {
      en: "The final grade is based on these criteria",
      kk: "Қорытынды баға осы критерийлер бойынша қойылады",
    },
    "Команда должна сверяться с критериями": {
      en: "The team should align with the criteria",
      kk: "Команда критерийлермен салыстырып отыруы керек",
    },
    "Критерии еще не готовы": {
      en: "The criteria are not ready yet",
      kk: "Критерийлер әлі дайын емес",
    },
    "Критерии уже отработали свою роль": {
      en: "The criteria have already served their purpose",
      kk: "Критерийлер өз рөлін орындап болды",
    },
    "Критерии уже подготовлены": {
      en: "The criteria are already prepared",
      kk: "Критерийлер қазірдің өзінде дайын",
    },
    "Подготовьте критерии заранее": {
      en: "Prepare the criteria in advance",
      kk: "Критерийлерді алдын ала дайындаңыз",
    },
    "Ждем ревьюера": {
      en: "Waiting for reviewer",
      kk: "Ревьюерді күтіп отырмыз",
    },
    "Ревьюер подтвержден": {
      en: "Reviewer confirmed",
      kk: "Ревьюер расталды",
    },
    "Финальная стадия": {
      en: "Final stage",
      kk: "Қорытынды кезең",
    },
    "Без данных": {
      en: "No data",
      kk: "Дерек жоқ",
    },
    "Стек пока не заполнен.": {
      en: "The stack has not been filled in yet.",
      kk: "Стек әлі толтырылмаған.",
    },
    "Активных участников пока нет.": {
      en: "There are no active members yet.",
      kk: "Белсенді қатысушылар әзірге жоқ.",
    },
    "Текущий статус: {status}.": {
      en: "Current status: {status}.",
      kk: "Ағымдағы күйі: {status}.",
    },
    "О проекте": {
      en: "About the project",
      kk: "Жоба туралы",
    },
    "Заявка в команду": {
      en: "Team application",
      kk: "Командаға өтінім",
    },
    "Оставьте короткий комментарий и отправьте заявку в проект.": {
      en: "Leave a short comment and send your application to the project.",
      kk: "Қысқа пікір қалдырып, жобаға өтінім жіберіңіз.",
    },
    "Комментарий к заявке": {
      en: "Application comment",
      kk: "Өтінімге пікір",
    },
    "Например: хочу взять backend-роль, есть опыт с Go/PostgreSQL": {
      en: "For example: I want a backend role, I have experience with Go/PostgreSQL",
      kk: "Мысалы: backend рөлін алғым келеді, Go/PostgreSQL бойынша тәжірибем бар",
    },
    "Подать заявку": {
      en: "Send application",
      kk: "Өтінім беру",
    },
    "Активность": {
      en: "Activity",
      kk: "Белсенділік",
    },
    "Управление": {
      en: "Manage",
      kk: "Басқару",
    },
    "Пайплайн запуска": {
      en: "Launch pipeline",
      kk: "Іске қосу пайплайны",
    },
    "Поиск преподавателя": {
      en: "Professor search",
      kk: "Оқытушыны іздеу",
    },
    "Поиск преподавателя по имени или email": {
      en: "Search for a professor by name or email",
      kk: "Оқытушыны аты немесе email бойынша іздеу",
    },
    "Пригласить": {
      en: "Invite",
      kk: "Шақыру",
    },
    "Назначение преподавателя и состав команды управляются здесь, а главные действия по этапу вынесены в верхний цветной блок.": {
      en: "Professor assignment and team composition are managed here, while the main stage actions are moved to the colored block above.",
      kk: "Оқытушыны тағайындау мен команда құрамын осында басқарасыз, ал кезең бойынша негізгі әрекеттер жоғарыдағы түсті блокқа шығарылған.",
    },
    "Команда проекта": {
      en: "Project team",
      kk: "Жоба командасы",
    },
    "Управляйте участниками и ролями в проекте.": {
      en: "Manage project members and roles.",
      kk: "Жобадағы қатысушылар мен рөлдерді басқарыңыз.",
    },
    "Название роли": {
      en: "Role name",
      kk: "Рөл атауы",
    },
    "Код роли": {
      en: "Role code",
      kk: "Рөл коды",
    },
    "Количество мест": {
      en: "Number of slots",
      kk: "Орын саны",
    },
    "Добавить роль": {
      en: "Add role",
      kk: "Рөл қосу",
    },
    "Участник": {
      en: "Member",
      kk: "Қатысушы",
    },
    "Действия": {
      en: "Actions",
      kk: "Әрекеттер",
    },
    "Нужно больше рук?": {
      en: "Need more hands?",
      kk: "Қосымша көмек керек пе?",
    },
    "Пригласите студентов из других групп присоединиться к проекту.": {
      en: "Invite students from other groups to join the project.",
      kk: "Басқа топтардағы студенттерді жобаға қосылуға шақырыңыз.",
    },
    "Найти участников": {
      en: "Find members",
      kk: "Қатысушыларды табу",
    },
    "Поиск участников": {
      en: "Find members",
      kk: "Қатысушыларды іздеу",
    },
    "Найдите студентов и отправьте приглашение в команду проекта.": {
      en: "Find students and send them an invite to the project team.",
      kk: "Студенттерді тауып, оларды жоба командасына шақырыңыз.",
    },
    "Поиск по имени или email": {
      en: "Search by name or email",
      kk: "Аты немесе email бойынша іздеу",
    },
    "Назад к команде": {
      en: "Back to team",
      kk: "Командаға оралу",
    },
    "Прогресс закрыт": {
      en: "Progress is closed",
      kk: "Прогресс жабық",
    },
    "+ Новая задача": {
      en: "+ New task",
      kk: "+ Жаңа тапсырма",
    },
    "К выполнению": {
      en: "To do",
      kk: "Орындауға",
    },
    "Прогресс проекта": {
      en: "Project progress",
      kk: "Жоба прогресі",
    },
    "Чек-лист требований": {
      en: "Requirements checklist",
      kk: "Талаптар чек-листі",
    },
    "Результат ревью": {
      en: "Review result",
      kk: "Ревью нәтижесі",
    },
    "Оценивание выполняет преподаватель после завершения проекта на своей странице.": {
      en: "The professor performs grading after the project is completed on their own page.",
      kk: "Бағалауды оқытушы жоба аяқталғаннан кейін өз парақшасында жүргізеді.",
    },
    "Пока нет опубликованной оценки.": {
      en: "There is no published grade yet.",
      kk: "Жарияланған баға әзірге жоқ.",
    },
    "Чек-лист ревью": {
      en: "Review checklist",
      kk: "Ревью чек-листі",
    },
    "проходной": {
      en: "passing",
      kk: "өту шегі",
    },
    "Итоговая оценка": {
      en: "Final grade",
      kk: "Қорытынды баға",
    },
    "Выполнено критериев": {
      en: "Criteria met",
      kk: "Орындалған критерий",
    },
    "Есть замечания": {
      en: "Issues found",
      kk: "Ескертулер бар",
    },
    "Дата проверки": {
      en: "Review date",
      kk: "Тексеру күні",
    },
    "Проверяющий": {
      en: "Reviewer",
      kk: "Тексеруші",
    },
    "Общий комментарий": {
      en: "Overall comment",
      kk: "Жалпы пікір",
    },
    "Комментарий преподавателя появится после ревью.": {
      en: "The professor comment will appear after the review.",
      kk: "Оқытушы пікірі ревьюден кейін пайда болады.",
    },
    "Редактирование проекта": {
      en: "Project editing",
      kk: "Жобаны өңдеу",
    },
    "Закрыть режим редактирования": {
      en: "Close editing mode",
      kk: "Өңдеу режимін жабу",
    },
    "Загрузить": {
      en: "Upload",
      kk: "Жүктеу",
    },
    "Рекомендуем формат 16:9, JPG/PNG/WEBP.": {
      en: "Recommended format: 16:9, JPG/PNG/WEBP.",
      kk: "Ұсынылатын формат: 16:9, JPG/PNG/WEBP.",
    },
    "GitHub репозиторий": {
      en: "GitHub repository",
      kk: "GitHub репозиторийі",
    },
    "Стек технологий": {
      en: "Technology stack",
      kk: "Технология стегі",
    },
    "React, Node.js, PostgreSQL": {
      en: "React, Node.js, PostgreSQL",
      kk: "React, Node.js, PostgreSQL",
    },
    "Сохранить изменения": {
      en: "Save changes",
      kk: "Өзгерістерді сақтау",
    },
    "Добавление новой задачи": {
      en: "Add a new task",
      kk: "Жаңа тапсырма қосу",
    },
    "Закрыть окно задачи": {
      en: "Close task window",
      kk: "Тапсырма терезесін жабу",
    },
    "Название задачи": {
      en: "Task title",
      kk: "Тапсырма атауы",
    },
    "Например: Разработать дизайн-систему": {
      en: "For example: Build the design system",
      kk: "Мысалы: Дизайн жүйесін әзірлеу",
    },
    "Сложность": {
      en: "Complexity",
      kk: "Күрделілік",
    },
    "Роль": {
      en: "Role",
      kk: "Рөл",
    },
    "Исполнитель": {
      en: "Assignee",
      kk: "Орындаушы",
    },
    "Теги": {
      en: "Tags",
      kk: "Тегтер",
    },
    "Design, Frontend": {
      en: "Design, Frontend",
      kk: "Design, Frontend",
    },
    "Описание": {
      en: "Description",
      kk: "Сипаттама",
    },
    "Добавьте подробности...": {
      en: "Add more details...",
      kk: "Толығырақ жазыңыз...",
    },
    "Срок": {
      en: "Deadline",
      kk: "Мерзім",
    },
    "Создать задачу": {
      en: "Create task",
      kk: "Тапсырма құру",
    },
    "Фиксация результата задачи": {
      en: "Submit task result",
      kk: "Тапсырма нәтижесін тіркеу",
    },
    "Закрыть окно результата задачи": {
      en: "Close task result window",
      kk: "Тапсырма нәтижесі терезесін жабу",
    },
    "Комментарий к результату": {
      en: "Result comment",
      kk: "Нәтижеге пікір",
    },
    "Что сделано, какие проверки пройдены, что важно знать команде...": {
      en: "What was done, which checks passed, and what the team should know...",
      kk: "Не істелді, қандай тексерулерден өтті, команда нені білуі керек...",
    },
    "Вложения (фото / видео / ссылка)": {
      en: "Attachments (photo / video / link)",
      kk: "Тіркемелер (фото / видео / сілтеме)",
    },
    "(каждая ссылка с новой строки)": {
      en: "(each link on a new line)",
      kk: "(әр сілтеме жаңа жолдан)",
    },
    "Настройка прав доступа:": {
      en: "Access permissions setup:",
      kk: "Қолжетімділік құқықтарын баптау:",
    },
    "Закрыть окно прав доступа": {
      en: "Close permissions window",
      kk: "Құқықтар терезесін жабу",
    },
    "Загрузка данных…": {
      en: "Loading data…",
      kk: "Деректер жүктелуде…",
    },
    "Системные роли": {
      en: "System roles",
      kk: "Жүйелік рөлдер",
    },
    "Делегируемые роли": {
      en: "Delegated roles",
      kk: "Делегирленетін рөлдер",
    },
    "Выберите роли, которые вы хотите назначить участнику.": {
      en: "Choose the roles you want to assign to the member.",
      kk: "Қатысушыға тағайындағыңыз келетін рөлдерді таңдаңыз.",
    },
    "Итоговые разрешения": {
      en: "Effective permissions",
      kk: "Қорытынды рұқсаттар",
    },
    "Сохранить права": {
      en: "Save permissions",
      kk: "Құқықтарды сақтау",
    },
    "Панель управления платформой": {
      en: "Platform control panel",
      kk: "Платформаны басқару панелі",
    },
    "Операционный контур платформы: пользователи, проекты, структура групп и точка контроля доступа.": {
      en: "The platform operations layer: users, projects, group structure, and the access control checkpoint.",
      kk: "Платформаның операциялық контуры: пайдаланушылар, жобалар, топ құрылымы және қолжетімділікті басқару нүктесі.",
    },
    "Пользовательский поток": {
      en: "User flow",
      kk: "Пайдаланушы ағыны",
    },
    "активных аккаунтов": {
      en: "active accounts",
      kk: "белсенді аккаунт",
    },
    "Проектный контур": {
      en: "Project flow",
      kk: "Жоба контуры",
    },
    "проектов под наблюдением": {
      en: "projects under watch",
      kk: "бақылаудағы жоба",
    },
    "Требует внимания": {
      en: "Needs attention",
      kk: "Назар аударуды қажет етеді",
    },
    "рисков и блокеров": {
      en: "risks and blockers",
      kk: "тәуекел мен бөгет",
    },
    "Сводка администратора": {
      en: "Administrator summary",
      kk: "Әкімші қорытындысы",
    },
    "Аккаунты, роли, статусы и быстрый переход в реестр пользователей.": {
      en: "Accounts, roles, statuses, and a quick jump into the user registry.",
      kk: "Аккаунттар, рөлдер, күйлер және пайдаланушылар тізіліміне жылдам өту.",
    },
    "Запуск, статусы, удаление и наблюдение за проектным потоком.": {
      en: "Launch, statuses, deletion, and monitoring of the project flow.",
      kk: "Іске қосу, күйлер, жою және жоба ағынын бақылау.",
    },
    "Структура потоков, кафедры и административные переводы в одном контуре.": {
      en: "Cohort structure, departments, and administrative transfers in one place.",
      kk: "Ағын құрылымы, кафедралар және әкімшілік ауыстырулар бір контурда.",
    },
    "Внутренние гайды, редактор статей и документация для команды.": {
      en: "Internal guides, the article editor, and team documentation.",
      kk: "Ішкі гайдтар, мақала редакторы және команда құжаттамасы.",
    },
    "Пользовательский контур": {
      en: "User domain",
      kk: "Пайдаланушы контуры",
    },
    "Всего пользователей": {
      en: "Total users",
      kk: "Барлық пайдаланушы",
    },
    "Все роли системы": {
      en: "All system roles",
      kk: "Жүйедегі барлық рөл",
    },
    "Активный учебный поток": {
      en: "Active academic cohort",
      kk: "Белсенді оқу ағыны",
    },
    "Преподаватели": {
      en: "Professors",
      kk: "Оқытушылар",
    },
    "Назначенные кураторы": {
      en: "Assigned supervisors",
      kk: "Тағайындалған кураторлар",
    },
    "Заблокированные": {
      en: "Blocked",
      kk: "Бұғатталғандар",
    },
    "Требуют внимания": {
      en: "Need attention",
      kk: "Назар аударуды қажет етеді",
    },
    "Сегодня в фокусе": {
      en: "In focus today",
      kk: "Бүгін назарда",
    },
    "Собираю пульс системы...": {
      en: "Collecting the system pulse...",
      kk: "Жүйенің пульсін жинап жатырмын...",
    },
    "Сейчас подтяну пользователей, проекты и список точек внимания.": {
      en: "Loading users, projects, and the current attention points now.",
      kk: "Қазір пайдаланушыларды, жобаларды және назар аударатын нүктелер тізімін жүктеймін.",
    },
    "Recent signal": {
      ru: "Последний сигнал",
      en: "Recent signal",
      kk: "Соңғы сигнал",
    },
    "Последняя активность": {
      en: "Recent activity",
      kk: "Соңғы белсенділік",
    },
    "Пользователи и проекты": {
      en: "Users and projects",
      kk: "Пайдаланушылар мен жобалар",
    },
    "Админка / Пользователи": {
      en: "Admin / Users",
      kk: "Әкімші / Пайдаланушылар",
    },
    "Управление пользователями": {
      en: "User management",
      kk: "Пайдаланушыларды басқару",
    },
    "Фильтр по ролям": {
      en: "Role filter",
      kk: "Рөлдер сүзгісі",
    },
    "Админы": {
      en: "Admins",
      kk: "Әкімшілер",
    },
    "Поиск по имени или email...": {
      en: "Search by name or email...",
      kk: "Аты немесе email бойынша іздеу...",
    },
    "Имя": {
      en: "Name",
      kk: "Аты",
    },
    "Загрузка...": {
      en: "Loading...",
      kk: "Жүктелуде...",
    },
    "Админка / Проекты": {
      en: "Admin / Projects",
      kk: "Әкімші / Жобалар",
    },
    "Управление проектами": {
      en: "Project management",
      kk: "Жобаларды басқару",
    },
    "COMPLETED и исторические записи": {
      en: "COMPLETED and historical entries",
      kk: "COMPLETED және тарихи жазбалар",
    },
    "Фильтр по статусу проектов": {
      en: "Project status filter",
      kk: "Жоба күйі сүзгісі",
    },
    "Активные": {
      en: "Active",
      kk: "Белсенділер",
    },
    "Завершенные": {
      en: "Completed",
      kk: "Аяқталғандар",
    },
    "Поиск проекта или автора...": {
      en: "Search by project or author...",
      kk: "Жоба немесе автор бойынша іздеу...",
    },
    "ID проекта": {
      en: "Project ID",
      kk: "Жоба ID-і",
    },
    "Название": {
      en: "Title",
      kk: "Атауы",
    },
    "Автор": {
      en: "Author",
      kk: "Автор",
    },
    "Дата": {
      en: "Date",
      kk: "Күні",
    },
    "Укажите ваши контактные данные.": {
      en: "Please provide your contact details.",
      kk: "Байланыс деректеріңізді көрсетіңіз.",
    },
    "Опишите, чем я могу помочь.": {
      en: "Describe how I can help.",
      kk: "Қалай көмектесе алатынымды сипаттаңыз.",
    },
    "Сообщение пока не отправилось. Попробуйте ещё раз чуть позже.": {
      en: "The message has not been delivered yet. Please try again a bit later.",
      kk: "Хабарлама әзірге жіберілмеді. Сәл кейінірек қайта көріңіз.",
    },
    "Не удалось отправить форму.": {
      en: "Failed to send the form.",
      kk: "Форманы жіберу мүмкін болмады.",
    },
    "Заполните поле с контактами и опишите ваш запрос.": {
      en: "Fill in your contact details and describe your request.",
      kk: "Байланыс деректерін толтырып, сұрағыңызды сипаттаңыз.",
    },
    "Отправляем...": {
      en: "Sending...",
      kk: "Жіберілуде...",
    },
    "Сообщение отправлено. Я свяжусь с вами в ближайшее время.": {
      en: "Your message has been sent. I will get in touch with you soon.",
      kk: "Хабарламаңыз жіберілді. Жақын уақытта сізбен байланысамын.",
    },
    "Публичные проекты на стадии набора и активной разработки.": {
      en: "Public projects in recruitment and active development stages.",
      kk: "Іріктеу және белсенді әзірлеу кезеңіндегі жария жобалар.",
    },
    "Выбери кафедру": {
      en: "Choose a department",
      kk: "Кафедраны таңда",
    },
    "Выбери номер группы": {
      en: "Choose a group number",
      kk: "Топ нөмірін таңда",
    },
    "Выбери кафедру и номер группы для PRIVATE visibility": {
      en: "Choose a department and group number for PRIVATE visibility",
      kk: "PRIVATE visibility үшін кафедра мен топ нөмірін таңдаңыз",
    },
    "Поиск пользователей...": {
      en: "Search users...",
      kk: "Пайдаланушыларды іздеу...",
    },
    "Поиск проектов...": {
      en: "Search projects...",
      kk: "Жобаларды іздеу...",
    },
    "Поиск по активности...": {
      en: "Search activity...",
      kk: "Белсенділік бойынша іздеу...",
    },
    "Добавить преподавателя": {
      en: "Add professor",
      kk: "Оқытушы қосу",
    },
    "Добавить студента": {
      en: "Add student",
      kk: "Студент қосу",
    },
    "Пользователь будет создан в статусе ACTIVE с привязкой к кафедре.": {
      en: "The user will be created in ACTIVE status and linked to a department.",
      kk: "Пайдаланушы ACTIVE күйінде құрылып, кафедраға байланыстырылып жасалады.",
    },
    "Заполните все поля.": {
      en: "Fill in all fields.",
      kk: "Барлық өрісті толтырыңыз.",
    },
    "Сохраняем...": {
      en: "Saving...",
      kk: "Сақталуда...",
    },
    "Наблюдение проекта": {
      en: "Project observation",
      kk: "Жобаны бақылау",
    },
    "Нет участников": {
      en: "No members",
      kk: "Қатысушылар жоқ",
    },
    "Нет задач": {
      en: "No tasks",
      kk: "Тапсырмалар жоқ",
    },
    "Нет критериев": {
      en: "No criteria",
      kk: "Критерий жоқ",
    },
    "Выполнено": {
      en: "Completed",
      kk: "Орындалды",
    },
    "Не выполнено": {
      en: "Not completed",
      kk: "Орындалмаған",
    },
    "Не оценено": {
      en: "Not graded",
      kk: "Бағаланбаған",
    },
    "Подтверждение действия": {
      en: "Confirm action",
      kk: "Әрекетті растау",
    },
    "Подтвердите смену статуса проекта.": {
      en: "Confirm the project status change.",
      kk: "Жоба күйін өзгертуді растаңыз.",
    },
    "Название проекта не совпадает.": {
      en: "The project title does not match.",
      kk: "Жоба атауы сәйкес келмейді.",
    },
    "Обновляем статус...": {
      en: "Updating status...",
      kk: "Күй жаңартылуда...",
    },
    "Сброс пароля": {
      en: "Reset password",
      kk: "Құпиясөзді қалпына келтіру",
    },
    "Новый пароль": {
      en: "New password",
      kk: "Жаңа құпиясөз",
    },
    "Пароль должен быть не короче 10 символов.": {
      en: "The password must be at least 10 characters long.",
      kk: "Құпиясөз кемінде 10 таңбадан тұруы керек.",
    },
    "Пароль обновлен.": {
      en: "Password updated.",
      kk: "Құпиясөз жаңартылды.",
    },
    "Удалить пользователя": {
      en: "Delete user",
      kk: "Пайдаланушыны жою",
    },
    "Ошибка наблюдения проекта": {
      en: "Project observation error",
      kk: "Жобаны бақылау қатесі",
    },
    "Ожидаем ответа": {
      en: "Awaiting response",
      kk: "Жауап күтілуде",
    },
    "Подтвержден": {
      en: "Confirmed",
      kk: "Расталды",
    },
    "Назначен": {
      en: "Assigned",
      kk: "Тағайындалды",
    },
    "Не назначен": {
      en: "Not assigned",
      kk: "Тағайындалмаған",
    },
    "Пройдено": {
      en: "Passed",
      kk: "Өтілді",
    },
    "Опционально": {
      en: "Optional",
      kk: "Қосымша",
    },
    "Впереди": {
      en: "Ahead",
      kk: "Алда",
    },
    "Описание проекта пока не заполнено.": {
      en: "The project description has not been filled in yet.",
      kk: "Жоба сипаттамасы әлі толтырылмаған.",
    },
    "Идея проекта, описание, README, стек и базовые роли команды.": {
      en: "The project idea, description, README, stack, and the team's base roles.",
      kk: "Жоба идеясы, сипаттама, README, стек және команданың негізгі рөлдері.",
    },
    "Формируются роли, команда, преподаватель и критерии готовности проекта.": {
      en: "Roles, the team, the professor, and project readiness criteria are being formed.",
      kk: "Рөлдер, команда, оқытушы және жоба дайындығының критерийлері қалыптасуда.",
    },
    "Команда выполняет задачи, двигает канбан и готовит проект к сдаче.": {
      en: "The team completes tasks, moves the kanban, and prepares the project for submission.",
      kk: "Команда тапсырмаларды орындайды, kanban-ды жүргізеді және жобаны тапсыруға дайындайды.",
    },
    "Преподаватель проверяет проект по критериям и выставляет итог.": {
      en: "The professor reviews the project against the criteria and assigns the final result.",
      kk: "Оқытушы жобаны критерийлер бойынша тексеріп, қорытындыны қояды.",
    },
    "Итоговая оценка опубликована, проект завершен и доступен как завершенный кейс.": {
      en: "The final grade is published, the project is completed, and it is available as a finished case.",
      kk: "Қорытынды баға жарияланды, жоба аяқталды және аяқталған кейс ретінде қолжетімді.",
    },
    "Открыть набор": {
      en: "Open recruitment",
      kk: "Іріктеуді ашу",
    },
    "Дать разрешение на запуск": {
      en: "Grant launch approval",
      kk: "Іске қосуға рұқсат беру",
    },
    "Перевести проект в ACTIVE": {
      en: "Move the project to ACTIVE",
      kk: "Жобаны ACTIVE күйіне ауыстыру",
    },
    "Для запуска должны быть готовы команда, подтвержден преподаватель и настроены критерии.": {
      en: "The team must be ready, the professor confirmed, and the criteria configured before launch.",
      kk: "Іске қосу үшін команда дайын болып, оқытушы расталып, критерийлер бапталуы керек.",
    },
    "Нельзя отправить на оценивание:": {
      en: "Cannot submit for grading:",
      kk: "Бағалауға жіберу мүмкін емес:",
    },
    "преподаватель еще не подтвердил участие": {
      en: "the professor has not confirmed participation yet",
      kk: "оқытушы әлі қатысуын растаған жоқ",
    },
    "нет задач для проверки": {
      en: "there are no tasks for review",
      kk: "тексеруге арналған тапсырмалар жоқ",
    },
    "Когда базовая структура готова, откройте набор и начните собирать команду.": {
      en: "When the base structure is ready, open recruitment and start assembling the team.",
      kk: "Негізгі құрылым дайын болғанда, іріктеуді ашып, команданы жинай бастаңыз.",
    },
    "Открыть набор может тимлид проекта. Подготовьте описание и роли, чтобы следующий шаг был очевиден для команды.": {
      en: "The project team lead can open recruitment. Prepare the description and roles so the next step is obvious for the team.",
      kk: "Іріктеуді жоба тимлиді аша алады. Келесі қадам командаға түсінікті болуы үшін сипаттама мен рөлдерді дайындаңыз.",
    },
    "Все условия собраны. Преподаватель может дать разрешение и перевести проект в рабочую фазу.": {
      en: "All conditions are met. The professor can grant approval and move the project into the working phase.",
      kk: "Барлық шарт жиналды. Оқытушы рұқсат беріп, жобаны жұмыс фазасына ауыстыра алады.",
    },
    "Для запуска еще не хватает обязательных условий. Проверьте чеклист слева и доберите недостающие пункты.": {
      en: "Mandatory launch conditions are still missing. Check the checklist on the left and complete the missing items.",
      kk: "Іске қосуға әлі міндетті шарттар жетіспейді. Сол жақтағы чек-листті тексеріп, жетіспейтін тармақтарды толықтырыңыз.",
    },
    "Команда готова. Следующий шаг за преподавателем: дать разрешение на запуск.": {
      en: "The team is ready. The next step is up to the professor: grant launch approval.",
      kk: "Команда дайын. Келесі қадам оқытушыға тиесілі: іске қосуға рұқсат беру.",
    },
    "Доведите набор до готовности, и после этого преподаватель сможет запустить проект.": {
      en: "Bring recruitment to readiness and after that the professor will be able to launch the project.",
      kk: "Іріктеуді дайын күйге жеткізіңіз, содан кейін оқытушы жобаны іске қоса алады.",
    },
    "Сначала создайте задачи в канбане, иначе проект нельзя будет отправить на оценивание.": {
      en: "Create tasks in the kanban first, otherwise the project cannot be submitted for grading.",
      kk: "Алдымен kanban-да тапсырмалар жасаңыз, әйтпесе жобаны бағалауға жіберу мүмкін болмайды.",
    },
    "Все задачи готовы, но нужно дождаться подтверждения преподавателя.": {
      en: "All tasks are ready, but you still need to wait for the professor's confirmation.",
      kk: "Барлық тапсырма дайын, бірақ оқытушының растауын күту керек.",
    },
    "Проект готов к передаче на оценивание. Кнопка отправки вынесена рядом.": {
      en: "The project is ready to be sent for grading. The submit button is placed nearby.",
      kk: "Жоба бағалауға жіберуге дайын. Жіберу батырмасы қасында орналасқан.",
    },
    "Преподавателю нужно сначала добавить критерии, иначе финальная оценка не будет опубликована.": {
      en: "The professor needs to add criteria first, otherwise the final grade will not be published.",
      kk: "Оқытушы алдымен критерийлерді қосуы керек, әйтпесе қорытынды баға жарияланбайды.",
    },
    "Все критерии заполнены. Осталось опубликовать итоговую оценку.": {
      en: "All criteria are filled in. The final grade just needs to be published.",
      kk: "Барлық критерий толтырылған. Енді қорытынды бағаны жариялау ғана қалды.",
    },
    "Финальная оценка уже опубликована. Проект завершен.": {
      en: "The final grade has already been published. The project is completed.",
      kk: "Қорытынды баға қазірдің өзінде жарияланған. Жоба аяқталды.",
    },
    "Комментарий к приглашению": {
      en: "Invite comment",
      kk: "Шақыруға пікір",
    },
    "не задан": {
      en: "not set",
      kk: "көрсетілмеген",
    },
    "Описание отсутствует": {
      en: "No description available",
      kk: "Сипаттама жоқ",
    },
    "не назначен": {
      en: "not assigned",
      kk: "тағайындалмаған",
    },
    "Срок истек, пока задача не завершена.": {
      en: "The deadline has passed and the task is still not complete.",
      kk: "Мерзім өтіп кетті, бірақ тапсырма әлі аяқталмаған.",
    },
    "Проект возвращен на пересдачу. После доработки команда сможет снова отправить его преподавателю.": {
      en: "The project has been returned for retake. After rework, the team will be able to send it to the professor again.",
      kk: "Жоба қайта тапсыруға қайтарылды. Пысықтаудан кейін команда оны оқытушыға қайта жібере алады.",
    },
    "Проект ожидает преподавательскую проверку. Итоговая оценка появится после завершения оценки.": {
      en: "The project is waiting for the professor review. The final grade will appear after grading is completed.",
      kk: "Жоба оқытушы тексеруін күтіп отыр. Қорытынды баға бағалау аяқталғаннан кейін пайда болады.",
    },
    "Критерии настраивает преподаватель, а итоговое оценивание появится после завершения проекта и запуска проверки.": {
      en: "The professor configures the criteria, and the final grade will appear after the project is completed and the review starts.",
      kk: "Критерийлерді оқытушы баптайды, ал қорытынды бағалау жоба аяқталып, тексеру басталғаннан кейін пайда болады.",
    },
    "Проект отправлен преподавателю на оценивание. Результаты появятся после проверки.": {
      en: "The project has been sent to the professor for grading. Results will appear after the review.",
      kk: "Жоба оқытушыға бағалауға жіберілді. Нәтижелер тексеруден кейін пайда болады.",
    },
    "Преподаватель вернул проект на доработку. После повторной сдачи итоговая оценка будет немного снижена.": {
      en: "The professor returned the project for rework. After resubmission the final grade will be slightly reduced.",
      kk: "Оқытушы жобаны пысықтауға қайтарды. Қайта тапсырғаннан кейін қорытынды баға сәл төмендетіледі.",
    },
    "После завершения проекта преподаватель выставляет оценки по критериям. Здесь отображаются результаты ревью.": {
      en: "After the project is completed, the professor assigns grades by criteria. The review results are shown here.",
      kk: "Жоба аяқталғаннан кейін оқытушы критерийлер бойынша баға қояды. Мұнда ревью нәтижелері көрсетіледі.",
    },
    "Оставьте короткий комментарий и отправьте заявку в команду проекта.": {
      en: "Leave a short comment and send your application to the project team.",
      kk: "Қысқа пікір қалдырып, жоба командасына өтінім жіберіңіз.",
    },
    "У вас нет права создавать задачи в этом проекте.": {
      en: "You do not have permission to create tasks in this project.",
      kk: "Сізде бұл жобада тапсырма құру құқығы жоқ.",
    },
    "Создание задач доступно только после перевода проекта в ACTIVE.": {
      en: "Task creation becomes available only after the project is moved to ACTIVE.",
      kk: "Тапсырма құру жоба ACTIVE күйіне ауыстырылғаннан кейін ғана қолжетімді болады.",
    },
    "Задача не найдена.": {
      en: "Task not found.",
      kk: "Тапсырма табылмады.",
    },
    "Фиксировать результат может только назначенный исполнитель.": {
      en: "Only the assigned performer can submit the result.",
      kk: "Нәтижені тек тағайындалған орындаушы ғана тіркей алады.",
    },
    "Задача должна быть в статусе IN_PROGRESS.": {
      en: "The task must be in IN_PROGRESS status.",
      kk: "Тапсырма IN_PROGRESS күйінде болуы керек.",
    },
    "Не выбрана задача для фиксации результата.": {
      en: "No task selected for result submission.",
      kk: "Нәтижені тіркеу үшін тапсырма таңдалмаған.",
    },
    "Результат зафиксирован. Задача переведена в DONE.": {
      en: "Result saved. The task was moved to DONE.",
      kk: "Нәтиже тіркелді. Тапсырма DONE күйіне ауыстырылды.",
    },
    "Заполните название задачи и роль.": {
      en: "Fill in the task title and role.",
      kk: "Тапсырма атауы мен рөлін толтырыңыз.",
    },
    "Переводить новую задачу сразу в IN_PROGRESS может только участник с правом изменения статуса.": {
      en: "Only a member with status-change permission can create a new task directly in IN_PROGRESS.",
      kk: "Жаңа тапсырманы бірден IN_PROGRESS күйінде тек мәртебені өзгерту құқығы бар қатысушы ғана жасай алады.",
    },
    "Для статуса IN_PROGRESS нужен исполнитель этой роли.": {
      en: "An assignee for this role is required for IN_PROGRESS status.",
      kk: "IN_PROGRESS күйі үшін осы рөлдің орындаушысы қажет.",
    },
    "Задача создана.": {
      en: "Task created.",
      kk: "Тапсырма құрылды.",
    },
    "Поддерживаются JPG/PNG/WEBP.": {
      en: "Supported formats: JPG/PNG/WEBP.",
      kk: "Қолдау көрсетілетін форматтар: JPG/PNG/WEBP.",
    },
    "Файл слишком большой (макс. 12MB).": {
      en: "The file is too large (max 12MB).",
      kk: "Файл тым үлкен (макс. 12MB).",
    },
    "Обложка проекта обновлена.": {
      en: "Project cover updated.",
      kk: "Жоба мұқабасы жаңартылды.",
    },
    "Обложка проекта удалена. Используется вариант по умолчанию.": {
      en: "Project cover removed. The default variant is being used.",
      kk: "Жоба мұқабасы жойылды. Әдепкі нұсқа қолданылып жатыр.",
    },
    "Название проекта обязательно.": {
      en: "Project title is required.",
      kk: "Жоба атауы міндетті.",
    },
    "Изменение visibility в этом UI пока локальное (API update visibility не подключен).": {
      en: "Changing visibility in this UI is still local only (the API update for visibility is not wired yet).",
      kk: "Бұл UI-дегі visibility өзгерісі әзірге тек жергілікті түрде жұмыс істейді (visibility update API әлі қосылмаған).",
    },
    "Изменения проекта сохранены.": {
      en: "Project changes saved.",
      kk: "Жоба өзгерістері сақталды.",
    },
    "Удалять проект может только его создатель.": {
      en: "Only the project creator can delete the project.",
      kk: "Жобаны тек оның авторы ғана жоя алады.",
    },
    "Проект будет удален без возможности восстановления. Это действие затронет команду, задачи и материалы проекта.": {
      en: "The project will be deleted with no recovery option. This will affect the team, tasks, and project materials.",
      kk: "Жоба қалпына келтіру мүмкіндігінсіз жойылады. Бұл әрекет командаға, тапсырмаларға және жоба материалдарына әсер етеді.",
    },
    "Название роли обязательно.": {
      en: "Role name is required.",
      kk: "Рөл атауы міндетті.",
    },
    "Роль добавлена.": {
      en: "Role added.",
      kk: "Рөл қосылды.",
    },
    "Права можно настраивать только для ACTIVE участников.": {
      en: "Permissions can be configured only for ACTIVE members.",
      kk: "Құқықтарды тек ACTIVE қатысушылары үшін баптауға болады.",
    },
    "Приглашение принято.": {
      en: "Invite accepted.",
      kk: "Шақыру қабылданды.",
    },
    "Приглашение отклонено.": {
      en: "Invite declined.",
      kk: "Шақыру қабылданбады.",
    },
    "Участник будет исключен из проекта. Если за ним были закреплены задачи, назначения будут сняты.": {
      en: "The member will be removed from the project. If tasks were assigned to them, those assignments will be cleared.",
      kk: "Қатысушы жобадан шығарылады. Егер оған тапсырмалар бекітілсе, тағайындаулар алынады.",
    },
    "Выберите роль участника.": {
      en: "Choose a member role.",
      kk: "Қатысушы рөлін таңдаңыз.",
    },
    "Выберите статус.": {
      en: "Choose a status.",
      kk: "Күйді таңдаңыз.",
    },
    "Выберите исполнителя.": {
      en: "Choose an assignee.",
      kk: "Орындаушыны таңдаңыз.",
    },
    "Задача будет удалена вместе с ее историей выполнения.": {
      en: "The task will be deleted together with its execution history.",
      kk: "Тапсырма оның орындалу тарихымен бірге жойылады.",
    },
    "Набор в проект открыт.": {
      en: "Recruitment for the project is open.",
      kk: "Жобаға іріктеу ашылды.",
    },
    "Приглашение отправлено.": {
      en: "Invite sent.",
      kk: "Шақыру жіберілді.",
    },
    "Выберите преподавателя из подсказок.": {
      en: "Choose a professor from the suggestions.",
      kk: "Ұсыныстардан оқытушыны таңдаңыз.",
    },
    "Приглашение преподавателю отправлено.": {
      en: "Professor invite sent.",
      kk: "Оқытушыға шақыру жіберілді.",
    },
    "Проект переведен в ACTIVE.": {
      en: "The project has been moved to ACTIVE.",
      kk: "Жоба ACTIVE күйіне ауыстырылды.",
    },
    "После подтверждения проект перейдет в статус GRADING и будет ждать финального решения преподавателя.": {
      en: "After confirmation the project will move to GRADING and wait for the professor's final decision.",
      kk: "Расталғаннан кейін жоба GRADING күйіне өтеді және оқытушының соңғы шешімін күтеді.",
    },
    "Проект отправлен на оценивание преподавателю.": {
      en: "The project has been sent to the professor for grading.",
      kk: "Жоба оқытушыға бағалауға жіберілді.",
    },
    "Подать заявку можно только на этапе набора (RECRUITMENT).": {
      en: "You can apply only during the recruitment stage (RECRUITMENT).",
      kk: "Өтінімді тек іріктеу кезеңінде (RECRUITMENT) беруге болады.",
    },
    "Вы уже связаны с этим проектом.": {
      en: "You are already linked to this project.",
      kk: "Сіз бұл жобамен әлдеқашан байланыстасыз.",
    },
    "Заявка отправлена. Ожидайте решения тимлида.": {
      en: "Application sent. Wait for the team lead's decision.",
      kk: "Өтінім жіберілді. Тимлидтің шешімін күтіңіз.",
    },
    "Заявка отправлена.": {
      en: "Application sent.",
      kk: "Өтінім жіберілді.",
    },
    "Права участника обновлены.": {
      en: "Member permissions updated.",
      kk: "Қатысушы құқықтары жаңартылды.",
    },
    "Данные обновлены.": {
      en: "Data updated.",
      kk: "Деректер жаңартылды.",
    },
    "Удаление...": {
      en: "Deleting...",
      kk: "Жойылуда...",
    },
    "Обложка удалена.": {
      en: "Cover removed.",
      kk: "Мұқаба жойылды.",
    },
    "Открыть": {
      en: "Open",
      kk: "Ашу",
    },
    "Проект прошел подготовку и готов к запуску.": {
      en: "The project has passed preparation and is ready to launch.",
      kk: "Жоба дайындық кезеңінен өтіп, іске қосуға дайын.",
    },
    "Проект находится на финальной подготовке перед запуском команды.": {
      en: "The project is in final preparation before the team launch.",
      kk: "Жоба команданы іске қосар алдындағы соңғы дайындықта тұр.",
    },
    "Команда и критерии готовы: проект можно переводить в ACTIVE.": {
      en: "The team and criteria are ready: the project can move to ACTIVE.",
      kk: "Команда мен критерийлер дайын: жобаны ACTIVE күйіне ауыстыруға болады.",
    },
    "Проект запущен. Следующий шаг: создать и выполнить задачи перед отправкой на оценивание.": {
      en: "The project is launched. The next step is to create and complete tasks before submitting for grading.",
      kk: "Жоба іске қосылды. Келесі қадам: бағалауға жібермес бұрын тапсырмалар құрып, орындау.",
    },
    "Все задачи закрыты: проект можно отправлять преподавателю на оценивание.": {
      en: "All tasks are closed: the project can be sent to the professor for grading.",
      kk: "Барлық тапсырма жабылды: жобаны оқытушыға бағалауға жіберуге болады.",
    },
    "Проект находится на оценивании. Для публикации итогов нужно добавить критерии.": {
      en: "The project is being graded. Criteria must be added before the final result can be published.",
      kk: "Жоба бағалануда. Қорытындыны жариялау үшін критерийлерді қосу қажет.",
    },
    "Все критерии оценены: можно публиковать итоговую оценку и завершать проект.": {
      en: "All criteria are graded: the final grade can be published and the project can be completed.",
      kk: "Барлық критерий бағаланды: қорытынды бағаны жариялап, жобаны аяқтауға болады.",
    },
    "Проект завершен: итоговая оценка опубликована и доступна в карточке проекта.": {
      en: "The project is completed: the final grade is published and available on the project card.",
      kk: "Жоба аяқталды: қорытынды баға жарияланып, жоба картасында қолжетімді.",
    },
    "Стартовая стадия: заполните описание, роли и стек, затем откройте набор проекта.": {
      en: "Initial stage: fill in the description, roles, and stack, then open project recruitment.",
      kk: "Бастапқы кезең: сипаттаманы, рөлдерді және стекті толтырып, содан кейін жобаға іріктеуді ашыңыз.",
    },
    "Создание задач доступно только участникам проекта с правом управления": {
      en: "Task creation is available only to project members with management rights",
      kk: "Тапсырма құру тек басқару құқығы бар жоба қатысушыларына қолжетімді",
    },
    "Создание задач доступно только после ACTIVE": {
      en: "Task creation is available only after ACTIVE",
      kk: "Тапсырма құру тек ACTIVE кезеңінен кейін қолжетімді",
    },
    "Команда не набрана.": {
      en: "The team is not assembled yet.",
      kk: "Команда әлі жиналмаған.",
    },
    "Пока пусто": {
      en: "Empty for now",
      kk: "Әзірге бос",
    },
    "Проверено преподавателем": {
      en: "Reviewed by professor",
      kk: "Оқытушы тексерді",
    },
    "На пересдаче": {
      en: "On retake",
      kk: "Қайта тапсыруда",
    },
    "Идет оценивание": {
      en: "Grading in progress",
      kk: "Бағалау жүріп жатыр",
    },
    "На оценивании": {
      en: "Awaiting grading",
      kk: "Бағалауда",
    },
    "Оценка не опубликована": {
      en: "The grade is not published",
      kk: "Баға жарияланбаған",
    },
    "Комментарий не оставлен.": {
      en: "No comment left.",
      kk: "Пікір қалдырылмаған.",
    },
    "Тимлид": {
      en: "Team lead",
      kk: "Тимлид",
    },
    "Без роли": {
      en: "No role",
      kk: "Рөлсіз",
    },
    "Нет ролей": {
      en: "No roles",
      kk: "Рөлдер жоқ",
    },
    "Одобрить": {
      en: "Approve",
      kk: "Мақұлдау",
    },
    "Сменить роль": {
      en: "Change role",
      kk: "Рөлді ауыстыру",
    },
    "Принять": {
      en: "Accept",
      kk: "Қабылдау",
    },
    "Права": {
      en: "Permissions",
      kk: "Құқықтар",
    },
    "Удалить проект": {
      en: "Delete project",
      kk: "Жобаны жою",
    },
    "Avatar": {
      ru: "Аватар",
      en: "Avatar",
      kk: "Аватар",
    },
  };

  const SUPPLEMENTAL_TRANSLATIONS = {
    "Закрыть": {
      en: "Close",
      kk: "Жабу",
    },
    "Сохранить": {
      en: "Save",
      kk: "Сақтау",
    },
    "Понятно": {
      en: "Got it",
      kk: "Түсінікті",
    },
    "Сообщение": {
      en: "Message",
      kk: "Хабарлама",
    },
    "Подтвердите действие": {
      en: "Confirm the action",
      kk: "Әрекетті растаңыз",
    },
    "Нет доступа": {
      en: "Access denied",
      kk: "Қолжетімділік жоқ",
    },
    "Нет доступа. Попробуйте сменить контекст.": {
      en: "Access denied. Try changing the context.",
      kk: "Қолжетімділік жоқ. Контексті өзгертіп көріңіз.",
    },
    "Закрыть окно наблюдения": {
      en: "Close observation window",
      kk: "Бақылау терезесін жабу",
    },
    "Закрыть окно подтверждения": {
      en: "Close confirmation window",
      kk: "Растау терезесін жабу",
    },
    "Закрыть окно создания пользователя": {
      en: "Close user creation window",
      kk: "Пайдаланушы құру терезесін жабу",
    },
    "Поиск пользователей или проектов...": {
      en: "Search users or projects...",
      kk: "Пайдаланушыларды немесе жобаларды іздеу...",
    },
    "+ Студент": {
      en: "+ Student",
      kk: "+ Студент",
    },
    "+ Преподаватель": {
      en: "+ Professor",
      kk: "+ Оқытушы",
    },
    "Заполните данные нового пользователя.": {
      en: "Fill in the new user's details.",
      kk: "Жаңа пайдаланушының деректерін толтырыңыз.",
    },
    "Создать пользователя": {
      en: "Create user",
      kk: "Пайдаланушы құру",
    },
    "Создать": {
      en: "Create",
      kk: "Құру",
    },
    "Код кафедры": {
      en: "Department code",
      kk: "Кафедра коды",
    },
    "Задача": {
      en: "Task",
      kk: "Тапсырма",
    },
    "Позиция": {
      en: "Position",
      kk: "Позиция",
    },
    "Вес": {
      en: "Weight",
      kk: "Салмақ",
    },
    "Итог опубликован": {
      en: "Final result published",
      kk: "Қорытынды жарияланды",
    },
    "Черновики и предзапуск": {
      en: "Drafts and pre-launch",
      kk: "Бастапқы нұсқалар мен іске қосуға дейінгі кезең",
    },
    "Набор, работа и оценивание": {
      en: "Recruiting, active work, and grading",
      kk: "Іріктеу, жұмыс және бағалау",
    },
    "Студенты": {
      en: "Students",
      kk: "Студенттер",
    },
    "Участники": {
      en: "Members",
      kk: "Қатысушылар",
    },
    "Админка": {
      en: "Admin",
      kk: "Әкімші панелі",
    },
    "Административный контур переводов между группами и комментариев по заявкам.": {
      en: "The administrative flow for transfers between groups and request comments.",
      kk: "Топтар арасындағы ауысулар мен өтінім пікірлерінің әкімшілік контуры.",
    },
    "Единая структура по кафедрам, группам и студентам. Здесь легче контролировать академический контур и переходы между группами.": {
      en: "A unified structure of departments, groups, and students. It is easier to control the academic flow and transfers between groups here.",
      kk: "Кафедралар, топтар және студенттер бойынша бірыңғай құрылым. Мұнда академиялық контур мен топтар арасындағы ауысуларды бақылау оңай.",
    },
    "Кафедры и студенческие группы": {
      en: "Departments and student groups",
      kk: "Кафедралар мен студенттік топтар",
    },
    "Кафедра → группа → студенты. Открывайте только нужный уровень и быстро находите точку управления.": {
      en: "Department → group → students. Open only the level you need and quickly find the control point.",
      kk: "Кафедра → топ → студенттер. Тек керек деңгейді ашып, басқару нүктесін жылдам табыңыз.",
    },
    "Заявки на смену группы": {
      en: "Group change requests",
      kk: "Топ ауыстыру өтінімдері",
    },
    "Кафедры": {
      en: "Departments",
      kk: "Кафедралар",
    },
    "Структура": {
      en: "Structure",
      kk: "Құрылым",
    },
    "Показано": {
      en: "Shown",
      kk: "Көрсетілгені",
    },
    "академическая структура": {
      en: "academic structure",
      kk: "академиялық құрылым",
    },
    "в дереве ниже": {
      en: "in the tree below",
      kk: "төмендегі ағашта",
    },
    "в текущем срезе": {
      en: "in the current slice",
      kk: "ағымдағы көріністе",
    },
    "после фильтров": {
      en: "after filters",
      kk: "сүзгілерден кейін",
    },
    "Фильтры групп": {
      en: "Group filters",
      kk: "Топ сүзгілері",
    },
    "Поиск студента или группы": {
      en: "Search student or group",
      kk: "Студентті немесе топты іздеу",
    },
    "Например: IS-45 или Иванов": {
      en: "For example: IS-45 or Ivanov",
      kk: "Мысалы: IS-45 немесе Иванов",
    },
    "Обновить заявки": {
      en: "Refresh requests",
      kk: "Өтінімдерді жаңарту",
    },
    "Academic directory": {
      ru: "Академический справочник",
      en: "Academic directory",
      kk: "Академиялық анықтама",
    },
    "Admin only": {
      ru: "Только для админа",
      en: "Admin only",
      kk: "Тек әкімшіге",
    },
    "Все кафедры": {
      en: "All departments",
      kk: "Барлық кафедра",
    },
    "Группы не найдены": {
      en: "Groups not found",
      kk: "Топтар табылмады",
    },
    "Для этой кафедры пока нет привязанных академических групп.": {
      en: "There are no linked academic groups for this department yet.",
      kk: "Бұл кафедраға әлі байланыстырылған академиялық топтар жоқ.",
    },
    "Список студентов доступен ниже": {
      en: "The student list is available below",
      kk: "Студенттер тізімі төменде қолжетімді",
    },
    "В этой группе пока нет студентов": {
      en: "There are no students in this group yet",
      kk: "Бұл топта әзірге студенттер жоқ",
    },
    "Пока пусто": {
      en: "Nothing here yet",
      kk: "Әзірге бос",
    },
    "В этой группе еще нет студентов.": {
      en: "There are no students in this group yet.",
      kk: "Бұл топта әлі студенттер жоқ.",
    },
    "Новых заявок нет": {
      en: "There are no new requests",
      kk: "Жаңа өтінімдер жоқ",
    },
    "Сейчас нет переводов, которые требуют решения администратора.": {
      en: "There are currently no transfers that require an administrator's decision.",
      kk: "Қазіргі кезде әкімшінің шешімін қажет ететін ауысулар жоқ.",
    },
    "Комментарий администратора": {
      en: "Administrator comment",
      kk: "Әкімші түсіндірмесі",
    },
    "Загрузка структуры групп...": {
      en: "Loading group structure...",
      kk: "Топ құрылымы жүктелуде...",
    },
    "Заявка обновлена.": {
      en: "Request updated.",
      kk: "Өтінім жаңартылды.",
    },
    "Входящие": {
      en: "Incoming",
      kk: "Кіріс",
    },
    "Исходящие": {
      en: "Outgoing",
      kk: "Шығыс",
    },
    "Заявки и приглашения": {
      en: "Requests and invites",
      kk: "Өтінімдер мен шақырулар",
    },
    "Одна страница с входящими приглашениями в проекты и вашими исходящими заявками.": {
      en: "One page with incoming project invites and your outgoing applications.",
      kk: "Жобаларға кіріс шақырулар мен сіздің шығыс өтінімдеріңізге арналған бір бет.",
    },
    "Открыть проекты": {
      en: "Open projects",
      kk: "Жобаларды ашу",
    },
    "Invites tabs": {
      ru: "Вкладки заявок",
      en: "Invites tabs",
      kk: "Өтінім қойындылары",
    },
    "Входящих заявок и приглашений пока нет.": {
      en: "There are no incoming requests or invites yet.",
      kk: "Кіріс өтінімдер мен шақырулар әзірге жоқ.",
    },
    "Исходящих заявок пока нет.": {
      en: "There are no outgoing applications yet.",
      kk: "Шығыс өтінімдер әзірге жоқ.",
    },
    "Открыть проект": {
      en: "Open project",
      kk: "Жобаны ашу",
    },
    "Принять": {
      en: "Accept",
      kk: "Қабылдау",
    },
    "Отклонить": {
      en: "Decline",
      kk: "Қабылдамау",
    },
    "От:": {
      en: "From:",
      kk: "Кімнен:",
    },
    "Создано:": {
      en: "Created:",
      kk: "Құрылған:",
    },
    "Ответ:": {
      en: "Response:",
      kk: "Жауап:",
    },
    "Команда проекта": {
      en: "Project team",
      kk: "Жоба командасы",
    },
    "ЗАЯВКА": {
      en: "APPLICATION",
      kk: "ӨТІНІМ",
    },
    "ПРИГЛАШЕНИЕ": {
      en: "INVITE",
      kk: "ШАҚЫРУ",
    },
    "НА РАССМОТРЕНИИ": {
      en: "UNDER REVIEW",
      kk: "ҚАРАЛУДА",
    },
    "ПРИНЯТО": {
      en: "ACCEPTED",
      kk: "ҚАБЫЛДАНДЫ",
    },
    "ОТКЛОНЕНО": {
      en: "DECLINED",
      kk: "ҚАБЫЛДАНБАДЫ",
    },
    "ОТОЗВАНО": {
      en: "WITHDRAWN",
      kk: "ҚАЙТАРЫЛДЫ",
    },
    "СТАТУС": {
      en: "STATUS",
      kk: "КҮЙІ",
    },
    "ПОДГОТОВКА": {
      en: "PREPARATION",
      kk: "ДАЙЫНДЫҚ",
    },
    "НАБОР": {
      en: "RECRUITMENT",
      kk: "ІРІКТЕУ",
    },
    "В РАБОТЕ": {
      en: "IN PROGRESS",
      kk: "ЖҰМЫСТА",
    },
    "ОЦЕНИВАНИЕ": {
      en: "GRADING",
      kk: "БАҒАЛАУ",
    },
    "ЗАВЕРШЕН": {
      en: "COMPLETED",
      kk: "АЯҚТАЛДЫ",
    },
    "Решение: заявка принята.": {
      en: "Decision: application accepted.",
      kk: "Шешім: өтінім қабылданды.",
    },
    "Решение: заявка отклонена.": {
      en: "Decision: application declined.",
      kk: "Шешім: өтінім қабылданбады.",
    },
    "Решение: ожидает ответа тимлида.": {
      en: "Decision: waiting for the team lead's response.",
      kk: "Шешім: тимлидтің жауабын күтуде.",
    },
    "Решение: заявка отозвана/снята.": {
      en: "Decision: application withdrawn/removed.",
      kk: "Шешім: өтінім қайтарылды/алынды.",
    },
    "Загрузка заявок...": {
      en: "Loading requests...",
      kk: "Өтінімдер жүктелуде...",
    },
    "Заявка отклонена.": {
      en: "Request declined.",
      kk: "Өтінім қабылданбады.",
    },
    "Заявка принята. Участник добавлен в команду.": {
      en: "Request accepted. The member has been added to the team.",
      kk: "Өтінім қабылданды. Қатысушы командаға қосылды.",
    },
    "Заявки загружены.": {
      en: "Requests loaded.",
      kk: "Өтінімдер жүктелді.",
    },
    "Не удалось определить пользователя заявки.": {
      en: "Failed to determine the request user.",
      kk: "Өтінім иесін анықтау мүмкін болмады.",
    },
    "Отклонить заявку": {
      en: "Decline application",
      kk: "Өтінімді қабылдамау",
    },
    "Отклонить приглашение": {
      en: "Decline invite",
      kk: "Шақыруды қабылдамау",
    },
    "Подтвердите действие: {...} заявку участника в проект.": {
      en: "Confirm the action: {...} the participant's application to the project.",
      kk: "Әрекетті растаңыз: қатысушының жобаға өтінімін {...}.",
    },
    "Подтвердите действие: {...} приглашение в проект.": {
      en: "Confirm the action: {...} the project invite.",
      kk: "Әрекетті растаңыз: жобаға шақыруды {...}.",
    },
    "Принять заявку": {
      en: "Accept application",
      kk: "Өтінімді қабылдау",
    },
    "Принять приглашение": {
      en: "Accept invite",
      kk: "Шақыруды қабылдау",
    },
    "принять": {
      en: "accept",
      kk: "қабылдау",
    },
    "отклонить": {
      en: "decline",
      kk: "қабылдамау",
    },
    "Настройки": {
      en: "Settings",
      kk: "Баптаулар",
    },
    "Выйти": {
      en: "Log out",
      kk: "Шығу",
    },
    "Администратор": {
      en: "Administrator",
      kk: "Әкімші",
    },
    "Преподаватель": {
      en: "Professor",
      kk: "Оқытушы",
    },
    "Профиль": {
      en: "Profile",
      kk: "Профиль",
    },
    "Студент": {
      en: "Student",
      kk: "Студент",
    },
    "Обновить": {
      en: "Refresh",
      kk: "Жаңарту",
    },
    "Подтвердить": {
      en: "Confirm",
      kk: "Растау",
    },
    "Всего проектов": {
      en: "Total projects",
      kk: "Барлық жоба",
    },
    "Все доступные проекты": {
      en: "All available projects",
      kk: "Барлық қолжетімді жоба",
    },
    "Завершены": {
      en: "Completed",
      kk: "Аяқталды",
    },
    "Комментарий": {
      en: "Comment",
      kk: "Пікір",
    },
    "Критерий": {
      en: "Criterion",
      kk: "Критерий",
    },
    "Критерии и оценка": {
      en: "Criteria and grading",
      kk: "Критерийлер мен бағалау",
    },
    "Пользователь": {
      en: "User",
      kk: "Пайдаланушы",
    },
    "Привет, admin!": {
      en: "Hi, admin!",
      kk: "Сәлем, admin!",
    },
    "Документация, заметки и разборы в Markdown.": {
      en: "Documentation, notes, and walkthroughs in Markdown.",
      kk: "Markdown форматындағы құжаттама, жазбалар және талдаулар.",
    },
    "Заголовок": {
      en: "Title",
      kk: "Тақырып",
    },
    "Загрузка…": {
      en: "Loading…",
      kk: "Жүктелуде…",
    },
    "Материал": {
      en: "Article",
      kk: "Материал",
    },
    "На этой странице": {
      en: "On this page",
      kk: "Осы бетте",
    },
    "Опубликовать": {
      en: "Publish",
      kk: "Жариялау",
    },
    "Предпросмотр": {
      en: "Preview",
      kk: "Алдын ала қарау",
    },
    "Редактировать": {
      en: "Edit",
      kk: "Өңдеу",
    },
    "Редактор": {
      en: "Editor",
      kk: "Редактор",
    },
    "Статья базы знаний IDSAI": {
      en: "IDSAI knowledge base article",
      kk: "IDSAI білім қоры мақаласы",
    },
    "Теги (через запятую)": {
      en: "Tags (comma separated)",
      kk: "Тегтер (үтір арқылы)",
    },
    "Черновик": {
      en: "Draft",
      kk: "Бастапқы нұсқа",
    },
    "Чтение 1 мин": {
      en: "1 min read",
      kk: "1 мин оқу",
    },
    "База знаний IDSAI — инструкции, гайды и документация": {
      en: "IDSAI knowledge base — instructions, guides, and documentation",
      kk: "IDSAI білім қоры — нұсқаулықтар, гайдтар және құжаттама",
    },
    "Вперёд →": {
      en: "Forward →",
      kk: "Алға →",
    },
    "Все статьи": {
      en: "All articles",
      kk: "Барлық мақалалар",
    },
    "Гайды по Git": {
      en: "Git guides",
      kk: "Git гайдтары",
    },
    "Загрузить .md": {
      en: "Upload .md",
      kk: ".md жүктеу",
    },
    "Категории": {
      en: "Categories",
      kk: "Санаттар",
    },
    "Категория": {
      en: "Category",
      kk: "Санат",
    },
    "Новая категория": {
      en: "New category",
      kk: "Жаңа санат",
    },
    "Новая статья": {
      en: "New article",
      kk: "Жаңа мақала",
    },
    "Родительская категория": {
      en: "Parent category",
      kk: "Ата-санат",
    },
    "Статьи пока не добавлены": {
      en: "No articles have been added yet",
      kk: "Мақалалар әзірге қосылмаған",
    },
    "— Корень —": {
      en: "— Root —",
      kk: "— Түбір —",
    },
    "← Назад": {
      en: "← Back",
      kk: "← Артқа",
    },
    "Email, пароль и смена группы настраиваются в настройках аккаунта.": {
      en: "Email, password, and group changes are configured in account settings.",
      kk: "Email, құпиясөз және топты ауыстыру аккаунт баптауларында реттеледі.",
    },
    "Аккаунт": {
      en: "Account",
      kk: "Аккаунт",
    },
    "Добавить": {
      en: "Add",
      kk: "Қосу",
    },
    "Есть время на 1 активный проект": {
      en: "I have time for 1 active project",
      kk: "1 белсенді жобаға уақытым бар",
    },
    "Загрузка": {
      en: "Loading",
      kk: "Жүктеу",
    },
    "Интересы и направления": {
      en: "Interests and directions",
      kk: "Қызығушылықтар мен бағыттар",
    },
    "Квадратное изображение от 400×400 px. JPG, PNG, WEBP.": {
      en: "Square image from 400×400 px. JPG, PNG, WEBP.",
      kk: "400×400 px бастап квадрат сурет. JPG, PNG, WEBP.",
    },
    "Не указано": {
      en: "Not specified",
      kk: "Көрсетілмеген",
    },
    "Работаю точечно по задачам": {
      en: "I work selectively on tasks",
      kk: "Тапсырмалармен нүктелі жұмыс істеймін",
    },
    "Свободен для новых проектов": {
      en: "Available for new projects",
      kk: "Жаңа жобаларға ашықпын",
    },
    "Сейчас высокая учебная нагрузка": {
      en: "My current study load is high",
      kk: "Қазір оқу жүктемем жоғары",
    },
    "Цель на ближайший проект": {
      en: "Goal for the next project",
      kk: "Келесі жобаға мақсат",
    },
    "в стеке": {
      en: "in stack",
      kk: "стекте",
    },
    "Баллы": {
      en: "Score",
      kk: "Ұпай",
    },
    "Выберите проект и продолжайте ревью по критериям.": {
      en: "Choose a project and continue the review by criteria.",
      kk: "Жобаны таңдап, критерийлер бойынша ревьюді жалғастырыңыз.",
    },
    "Короткая визуальная сводка по участникам на основе задач и событий проекта.": {
      en: "A short visual summary of participants based on project tasks and events.",
      kk: "Жоба тапсырмалары мен оқиғалары негізіндегі қатысушылардың қысқа визуалды қорытындысы.",
    },
    "Лист проверки": {
      en: "Checklist",
      kk: "Тексеру парағы",
    },
    "Отмечайте каждый критерий и оставляйте комментарий только там, где он действительно помогает команде.": {
      en: "Mark every criterion and leave comments only where they really help the team.",
      kk: "Әр критерийді белгілеңіз және пікірді тек командаға шын мәнінде көмектесетін жерде қалдырыңыз.",
    },
    "Покрытие": {
      en: "Coverage",
      kk: "Қамту",
    },
    "Покрытие, итоговый балл и статус завершения по текущему проекту.": {
      en: "Coverage, final score, and completion status for the current project.",
      kk: "Ағымдағы жоба бойынша қамту, қорытынды балл және аяқталу күйі.",
    },
    "После выбора проекта здесь появятся подсказки, что преподавателю делать на этом этапе.": {
      en: "After choosing a project, hints about what the professor should do at this stage will appear here.",
      kk: "Жобаны таңдағаннан кейін мұнда оқытушының осы кезеңде не істеуі керегі туралы кеңестер шығады.",
    },
    "Прогресс проверки": {
      en: "Review progress",
      kk: "Тексеру прогресі",
    },
    "Сводка ревью": {
      en: "Review summary",
      kk: "Ревью қорытындысы",
    },
    "Сохранить оценку": {
      en: "Save grading",
      kk: "Бағаны сақтау",
    },
    "Статус проверки": {
      en: "Review status",
      kk: "Тексеру күйі",
    },
    "Активность команды": {
      en: "Team activity",
      kk: "Команда белсенділігі",
    },
    "Входящие проекты": {
      en: "Incoming projects",
      kk: "Кіріс жобалар",
    },
    "Назад к дашборду": {
      en: "Back to dashboard",
      kk: "Дашбордқа оралу",
    },
    "Ожидают решения": {
      en: "Awaiting decision",
      kk: "Шешімді күтуде",
    },
    "Проект": {
      en: "Project",
      kk: "Жоба",
    },
    "1. Ответить на приглашение": {
      en: "1. Respond to the invite",
      kk: "1. Шақыруға жауап беру",
    },
    "2. Настроить критерии": {
      en: "2. Set up criteria",
      kk: "2. Критерийлерді баптау",
    },
    "3. Сопровождать команду": {
      en: "3. Support the team",
      kk: "3. Команданы сүйемелдеу",
    },
    "4. Завершить ревью": {
      en: "4. Complete the review",
      kk: "4. Ревьюді аяқтау",
    },
    "Ближайшие действия": {
      en: "Next actions",
      kk: "Келесі әрекеттер",
    },
    "Во время ACTIVE преподаватель следит за проектом, но не оценивает преждевременно.": {
      en: "During ACTIVE the professor watches the project, but does not grade it prematurely.",
      kk: "ACTIVE кезінде оқытушы жобаны бақылайды, бірақ ерте бағаламайды.",
    },
    "Все проекты": {
      en: "All projects",
      kk: "Барлық жобалар",
    },
    "Все проекты в вашем faculty-контуре.": {
      en: "All projects in your faculty scope.",
      kk: "Сіздің faculty аймағыңыздағы барлық жобалар.",
    },
    "Все проекты факультета": {
      en: "All faculty projects",
      kk: "Факультеттің барлық жобалары",
    },
    "До старта проекта преподаватель задает чек-лист и подтверждает, что проект можно запускать.": {
      en: "Before the project starts, the professor sets the checklist and confirms that the project can be launched.",
      kk: "Жоба басталмай тұрып, оқытушы чек-лист орнатып, жобаны іске қосуға болатынын растайды.",
    },
    "Загружаю приглашения, назначенные проекты и этапы ревью.": {
      en: "Loading invites, assigned projects, and review stages.",
      kk: "Шақыруларды, тағайындалған жобаларды және ревью кезеңдерін жүктеп жатырмын.",
    },
    "Загрузка проектов...": {
      en: "Loading projects...",
      kk: "Жобалар жүктелуде...",
    },
    "Здесь собраны все проекты вашего faculty-контура. Приглашения и ревью-поток подсвечиваются отдельно, но открыть и посмотреть можно любой проект.": {
      en: "All projects in your faculty scope are collected here. Invites and the review flow are highlighted separately, but any project can be opened and viewed.",
      kk: "Мұнда сіздің faculty аймағыңыздағы барлық жобалар жиналған. Шақырулар мен ревью ағыны бөлек ерекшеленеді, бірақ кез келген жобаны ашып көруге болады.",
    },
    "Как теперь работает преподаватель": {
      en: "How the professor works now",
      kk: "Енді оқытушы қалай жұмыс істейді",
    },
    "Команды в работе": {
      en: "Teams in progress",
      kk: "Жұмыстағы командалар",
    },
    "Контур преподавателя": {
      en: "Professor workspace",
      kk: "Оқытушы контуры",
    },
    "Можно открыть любой проект и посмотреть детали. Приоритетные преподавательские действия все равно подсвечиваются отдельно.": {
      en: "You can open any project and view details. Priority professor actions are still highlighted separately.",
      kk: "Кез келген жобаны ашып, егжей-тегжейін көруге болады. Басым профессорлық әрекеттер бәрібір бөлек белгіленеді.",
    },
    "На оценке": {
      en: "Under grading",
      kk: "Бағалауда",
    },
    "На этапе GRADING выставляется итог и фиксируются комментарии по критериям.": {
      en: "At the GRADING stage the final result is assigned and comments on criteria are recorded.",
      kk: "GRADING кезеңінде қорытынды қойылып, критерийлер бойынша пікірлер бекітіледі.",
    },
    "Новые приглашения и запросы на ревью.": {
      en: "New invites and review requests.",
      kk: "Жаңа шақырулар мен ревью сұраулары.",
    },
    "Нужно ответить": {
      en: "Need to respond",
      kk: "Жауап беру керек",
    },
    "Преподаватель видит весь каталог проектов своей faculty-зоны, а write-действия остаются только там, где у него есть назначение или проектная роль.": {
      en: "The professor sees the entire project catalog of their faculty zone, while write actions remain only where they have an assignment or project role.",
      kk: "Оқытушы өз faculty аймағындағы жобалардың толық каталогын көреді, ал write әрекеттері тек тағайындауы немесе жобалық рөлі бар жерлерде ғана қалады.",
    },
    "Преподаватель принимает или отклоняет ревью, а не назначает себя вручную.": {
      en: "The professor accepts or declines the review instead of assigning themselves manually.",
      kk: "Оқытушы өзін қолмен тағайындамай, ревьюді қабылдайды немесе қабылдамайды.",
    },
    "Привет!": {
      en: "Hi!",
      kk: "Сәлем!",
    },
    "Проекты уже запущены и идут к сдаче.": {
      en: "Projects are already launched and moving toward submission.",
      kk: "Жобалар қазірдің өзінде іске қосылып, тапсыруға қарай жылжып жатыр.",
    },
    "Проекты, где можно открыть оценивание.": {
      en: "Projects where grading can be opened.",
      kk: "Бағалауды ашуға болатын жобалар.",
    },
    "Сначала то, что двигает проект вперед уже сегодня.": {
      en: "First, the things that move the project forward today.",
      kk: "Алдымен жобаны бүгін алға жылжытатын нәрселер.",
    },
    "Собираю контекст...": {
      en: "Collecting context...",
      kk: "Контекст жиналуда...",
    },
    "Собираю полный faculty-каталог, приглашения и доступные преподавателю действия.": {
      en: "Collecting the full faculty catalog, invites, and actions available to the professor.",
      kk: "Толық faculty каталогын, шақыруларды және оқытушыға қолжетімді әрекеттерді жинап жатырмын.",
    },
    "Четкая последовательность вместо набора несвязанных кнопок.": {
      en: "A clear sequence instead of a set of unrelated buttons.",
      kk: "Бір-бірімен байланыссыз батырмалар жиынының орнына айқын рет.",
    },
    "0 активных": {
      en: "0 active",
      kk: "0 белсенді",
    },
    "Поиск в проекте (Cmd+K)": {
      en: "Search in project (Cmd+K)",
      kk: "Жоба ішінде іздеу (Cmd+K)",
    },
    "Обновить проект": {
      en: "Refresh project",
      kk: "Жобаны жаңарту",
    },
    "Средняя": {
      en: "Medium",
      kk: "Орташа",
    },
    "Сложная": {
      en: "Hard",
      kk: "Күрделі",
    },
    "Легкая": {
      en: "Easy",
      kk: "Жеңіл",
    },
    "Сейчас: подготовка": {
      en: "Current: preparation",
      kk: "Қазір: дайындық",
    },
    "Дальше: набор": {
      en: "Next: recruitment",
      kk: "Келесі: іріктеу",
    },
    "0 критериев": {
      en: "0 criteria",
      kk: "0 критерий",
    },
    "Ревью": {
      en: "Review",
      kk: "Ревью",
    },
    "Например: Покрытие тестами > 80%": {
      en: "For example: Test coverage > 80%",
      kk: "Мысалы: Тестпен қамту > 80%",
    },
    "Что именно проверяет критерий": {
      en: "What exactly the criterion checks",
      kk: "Критерий нақты нені тексереді",
    },
    "Вес (1-100)": {
      en: "Weight (1-100)",
      kk: "Салмақ (1-100)",
    },
    "Вес 0 / 100": {
      en: "Weight 0 / 100",
      kk: "Салмақ 0 / 100",
    },
    "Добавить критерий": {
      en: "Add criterion",
      kk: "Критерий қосу",
    },
    "Загрузить шаблон": {
      en: "Upload template",
      kk: "Үлгіні жүктеу",
    },
    "Изменения применяются сразу для команды проекта. Участники увидят обновлённый список после сохранения.": {
      en: "Changes apply to the project team immediately. Participants will see the updated list after saving.",
      kk: "Өзгерістер жоба командасына бірден қолданылады. Қатысушылар жаңартылған тізімді сақтағаннан кейін көреді.",
    },
    "Критерии оценки": {
      en: "Grading criteria",
      kk: "Бағалау критерийлері",
    },
    "Настройка критериев оценки": {
      en: "Grading criteria setup",
      kk: "Бағалау критерийлерін баптау",
    },
    "Режим редактирования": {
      en: "Editing mode",
      kk: "Өңдеу режимі",
    },
    "Сохранить как шаблон": {
      en: "Save as template",
      kk: "Үлгі ретінде сақтау",
    },
    "Управление чек-листами для проверки студенческих проектов и лабораторных работ.": {
      en: "Checklist management for reviewing student projects and lab work.",
      kk: "Студенттік жобалар мен зертханалық жұмыстарды тексеруге арналған чек-листтерді басқару.",
    },
    "Чему хотите научиться?": {
      en: "What do you want to learn?",
      kk: "Нені үйренгіңіз келеді?",
    },
    "4 курс, 2 семестр": {
      en: "4th year, 2nd semester",
      kk: "4 курс, 2 семестр",
    },
    "Иванов Иван Иванович": {
      en: "Ivanov Ivan Ivanovich",
      kk: "Иванов Иван Иванович",
    },
    "напр. s1234567@university.edu": {
      en: "e.g. s1234567@university.edu",
      kk: "мыс. s1234567@university.edu",
    },
    "Команда": {
      en: "Team",
      kk: "Команда",
    },
    "Отменить": {
      en: "Cancel",
      kk: "Болдырмау",
    },
    "Опубликовано": {
      en: "Published",
      kk: "Жарияланды",
    },
    "Уведомления": {
      en: "Notifications",
      kk: "Хабарландырулар",
    },
    "Прочитать все": {
      en: "Mark all as read",
      kk: "Барлығын оқылған деп белгілеу",
    },
    "Очистить": {
      en: "Clear",
      kk: "Тазалау",
    },
    "Редактировать критерии": {
      en: "Edit criteria",
      kk: "Критерийлерді өңдеу",
    },
    "Отправить на пересдачу": {
      en: "Return for resubmission",
      kk: "Қайта тапсыруға қайтару",
    },
    "Завершить оценивание": {
      en: "Complete grading",
      kk: "Бағалауды аяқтау",
    },
    "Оценивание проекта": {
      en: "Project grading",
      kk: "Жобаны бағалау",
    },
    "Без названия": {
      en: "Untitled",
      kk: "Атаусыз",
    },
    "Без описания": {
      en: "No description",
      kk: "Сипаттама жоқ",
    },
    "Описание проекта пока не заполнено.": {
      en: "The project description has not been filled in yet.",
      kk: "Жоба сипаттамасы әлі толтырылмаған.",
    },
    "Нет проектов": {
      en: "No projects",
      kk: "Жобалар жоқ",
    },
    "Нет проектов для настройки": {
      en: "No projects for setup",
      kk: "Баптауға арналған жобалар жоқ",
    },
    "Нет категорий": {
      en: "No categories",
      kk: "Санаттар жоқ",
    },
    "Выберите категорию для загрузки": {
      en: "Choose a category for upload",
      kk: "Жүктеу үшін санатты таңдаңыз",
    },
    "Заполните заголовок и выберите категорию": {
      en: "Fill in the title and choose a category",
      kk: "Тақырыпты толтырып, санатты таңдаңыз",
    },
    "Редактировать категорию": {
      en: "Edit category",
      kk: "Санатты өңдеу",
    },
    "Создайте категорию перед загрузкой": {
      en: "Create a category before uploading",
      kk: "Жүктеу алдында санат құрыңыз",
    },
    "Важно": {
      en: "Important",
      kk: "Маңызды",
    },
    "Осторожно": {
      en: "Caution",
      kk: "Абайлаңыз",
    },
    "Примечание": {
      en: "Note",
      kk: "Ескертпе",
    },
    "Совет": {
      en: "Tip",
      kk: "Кеңес",
    },
    "Заголовок обязателен": {
      en: "Title is required",
      kk: "Тақырып міндетті",
    },
    "Короткая статья из базы знаний IDSAI с заметками, примерами и быстрым входом в тему.": {
      en: "A short IDSAI knowledge base article with notes, examples, and a quick entry into the topic.",
      kk: "IDSAI білім қорындағы жазбалармен, мысалдармен және тақырыпқа жылдам кіріспемен қысқа мақала.",
    },
    "Похоже, материал был удален или ссылка на него устарела.": {
      en: "It looks like the article was deleted or the link is outdated.",
      kk: "Материал жойылған немесе оған сілтеме ескірген сияқты.",
    },
    "Статья не найдена": {
      en: "Article not found",
      kk: "Мақала табылмады",
    },
    "Удалить статью": {
      en: "Delete article",
      kk: "Мақаланы жою",
    },
    "Это действие необратимо. Статья будет удалена без возможности восстановления.": {
      en: "This action cannot be undone. The article will be deleted permanently.",
      kk: "Бұл әрекетті қайтару мүмкін емес. Мақала қалпына келтірусіз жойылады.",
    },
    "Email еще не подтвержден": {
      en: "Email has not been confirmed yet",
      kk: "Email әлі расталмаған",
    },
    "Email обязателен для сброса пароля.": {
      en: "Email is required to reset the password.",
      kk: "Құпиясөзді қалпына келтіру үшін email міндетті.",
    },
    "Email подтвержден. Теперь можно войти.": {
      en: "Email confirmed. You can sign in now.",
      kk: "Email расталды. Енді кіруге болады.",
    },
    "Email успешно подтвержден и обновлен. Войдите с новым адресом.": {
      en: "Email was confirmed and updated successfully. Sign in with the new address.",
      kk: "Email сәтті расталып, жаңартылды. Жаңа мекенжаймен кіріңіз.",
    },
    "Аккаунт с таким email не найден или для него недоступен сброс пароля.": {
      en: "No account with this email was found, or password reset is unavailable for it.",
      kk: "Мұндай email бар аккаунт табылмады немесе ол үшін құпиясөзді қалпына келтіру қолжетімсіз.",
    },
    "Аккаунт создан. Подтвердите email по ссылке из письма, затем войдите.": {
      en: "Account created. Confirm your email via the link in the message, then sign in.",
      kk: "Аккаунт құрылды. Хаттағы сілтеме арқылы email-ді растаңыз, содан кейін кіріңіз.",
    },
    "Аккаунт создан. Теперь можно войти.": {
      en: "Account created. You can sign in now.",
      kk: "Аккаунт құрылды. Енді кіруге болады.",
    },
    "Введите email, на который нужно отправить код для смены пароля.": {
      en: "Enter the email where the password reset code should be sent.",
      kk: "Құпиясөзді ауыстыру коды жіберілетін email-ді енгізіңіз.",
    },
    "Введите новый пароль для аккаунта. После подтверждения можно будет сразу войти с ним.": {
      en: "Enter a new password for the account. After confirmation, you will be able to sign in with it immediately.",
      kk: "Аккаунт үшін жаңа құпиясөзді енгізіңіз. Растағаннан кейін онымен бірден кіруге болады.",
    },
    "Вход выполнен. Переход в кабинет...": {
      en: "Signed in. Redirecting to the workspace...",
      kk: "Кіру орындалды. Кабинетке өту...",
    },
    "Если такой аккаунт существует, код для сброса уже отправлен на email.": {
      en: "If such an account exists, a reset code has already been sent to the email.",
      kk: "Егер мұндай аккаунт бар болса, қалпына келтіру коды email-ге жіберілді.",
    },
    "Заполните обязательные поля и укажите номер группы.": {
      en: "Fill in the required fields and specify the group number.",
      kk: "Міндетті өрістерді толтырып, топ нөмірін көрсетіңіз.",
    },
    "Код должен содержать ровно 6 цифр.": {
      en: "The code must contain exactly 6 digits.",
      kk: "Код дәл 6 цифрдан тұруы керек.",
    },
    "Код из письма": {
      en: "Code from the email",
      kk: "Хаттағы код",
    },
    "Мы можем повторно отправить письмо подтверждения на этот адрес прямо сейчас.": {
      en: "We can resend the confirmation email to this address right now.",
      kk: "Растау хатын осы мекенжайға дәл қазір қайта жібере аламыз.",
    },
    "Не удалось загрузить кафедры": {
      en: "Failed to load departments",
      kk: "Кафедраларды жүктеу мүмкін болмады",
    },
    "Не удалось загрузить кафедры для регистрации": {
      en: "Failed to load departments for registration",
      kk: "Тіркелу үшін кафедраларды жүктеу мүмкін болмады",
    },
    "Не удалось обновить пароль.": {
      en: "Failed to update the password.",
      kk: "Құпиясөзді жаңарту мүмкін болмады.",
    },
    "Не удалось отправить письмо для сброса пароля.": {
      en: "Failed to send the password reset email.",
      kk: "Құпиясөзді қалпына келтіру хатын жіберу мүмкін болмады.",
    },
    "Не удалось повторно отправить письмо подтверждения": {
      en: "Failed to resend the confirmation email",
      kk: "Растау хатын қайта жіберу мүмкін болмады",
    },
    "Новый пароль не был задан.": {
      en: "The new password was not set.",
      kk: "Жаңа құпиясөз орнатылмады.",
    },
    "Номер группы должен содержать от 1 до 4 цифр.": {
      en: "The group number must contain 1 to 4 digits.",
      kk: "Топ нөмірі 1-ден 4-ке дейінгі цифрдан тұруы керек.",
    },
    "Нужно указать код и новый пароль.": {
      en: "You need to specify the code and a new password.",
      kk: "Код пен жаңа құпиясөзді көрсету керек.",
    },
    "Отправить код": {
      en: "Send code",
      kk: "Код жіберу",
    },
    "Отправить письмо": {
      en: "Send email",
      kk: "Хат жіберу",
    },
    "Ошибка входа:": {
      en: "Sign-in error:",
      kk: "Кіру қатесі:",
    },
    "Ошибка регистрации:": {
      en: "Registration error:",
      kk: "Тіркелу қатесі:",
    },
    "Пароли не совпадают": {
      en: "Passwords do not match",
      kk: "Құпиясөздер сәйкес келмейді",
    },
    "Пароли не совпадают.": {
      en: "Passwords do not match.",
      kk: "Құпиясөздер сәйкес келмейді.",
    },
    "Пароль должен содержать буквы и цифры.": {
      en: "The password must contain letters and digits.",
      kk: "Құпиясөзде әріптер мен сандар болуы керек.",
    },
    "Пароль обновлен. Войдите снова с новым паролем.": {
      en: "Password updated. Sign in again with the new password.",
      kk: "Құпиясөз жаңартылды. Жаңа құпиясөзбен қайта кіріңіз.",
    },
    "Пароль обновлен. Теперь можно войти с новым паролем.": {
      en: "Password updated. You can sign in with the new password now.",
      kk: "Құпиясөз жаңартылды. Енді жаңа құпиясөзбен кіруге болады.",
    },
    "Письмо подтверждения отправлено повторно.": {
      en: "Confirmation email sent again.",
      kk: "Растау хаты қайта жіберілді.",
    },
    "Повторите новый пароль": {
      en: "Repeat the new password",
      kk: "Жаңа құпиясөзді қайталаңыз",
    },
    "Подтвердите email перед входом.": {
      en: "Confirm your email before signing in.",
      kk: "Кірер алдында email-ді растаңыз.",
    },
    "Подтвердите код сброса": {
      en: "Confirm the reset code",
      kk: "Қалпына келтіру кодын растаңыз",
    },
    "Придумайте новый пароль": {
      en: "Create a new password",
      kk: "Жаңа құпиясөз ойлап табыңыз",
    },
    "Сбой запроса входа": {
      en: "Sign-in request failed",
      kk: "Кіру сұрауы сәтсіз аяқталды",
    },
    "Сбой запроса регистрации": {
      en: "Registration request failed",
      kk: "Тіркелу сұрауы сәтсіз аяқталды",
    },
    "Сбой запроса сброса пароля": {
      en: "Password reset request failed",
      kk: "Құпиясөзді қалпына келтіру сұрауы сәтсіз аяқталды",
    },
    "Сбой сброса пароля": {
      en: "Password reset failed",
      kk: "Құпиясөзді қалпына келтіру сәтсіз аяқталды",
    },
    "Слишком много попыток. Попробуйте немного позже.": {
      en: "Too many attempts. Please try again a little later.",
      kk: "Тым көп әрекет жасалды. Сәл кейінірек қайталап көріңіз.",
    },
    "Ссылка подтверждения недействительна или истекла.": {
      en: "The confirmation link is invalid or has expired.",
      kk: "Растау сілтемесі жарамсыз немесе мерзімі өтіп кеткен.",
    },
    "Ссылка сброса пароля недействительна или уже истекла.": {
      en: "The password reset link is invalid or has already expired.",
      kk: "Құпиясөзді қалпына келтіру сілтемесі жарамсыз немесе мерзімі өтіп кеткен.",
    },
    "Ссылка сброса подтверждена. Задайте новый пароль.": {
      en: "The reset link has been confirmed. Set a new password.",
      kk: "Қалпына келтіру сілтемесі расталды. Жаңа құпиясөз орнатыңыз.",
    },
    "Система выглядит стабильно": {
      en: "The system looks stable",
      kk: "Жүйе тұрақты көрінеді",
    },
    "Срочных сигналов сейчас нет. Можно перейти к пользователям, проектам или структуре групп по плану.": {
      en: "There are no urgent signals right now. You can move on to users, projects, or the group structure as planned.",
      kk: "Қазір шұғыл сигналдар жоқ. Енді жоспар бойынша пайдаланушыларға, жобаларға немесе топ құрылымына өте аласыз.",
    },
    "Наблюдать": {
      en: "Observe",
      kk: "Бақылау",
    },
    "Нет данных для отображения": {
      en: "No data to display",
      kk: "Көрсетуге дерек жоқ",
    },
    "Активен": {
      en: "Enabled",
      kk: "Белсенді",
    },
    "Активировать": {
      en: "Activate",
      kk: "Белсендіру",
    },
    "Активный": {
      en: "Active",
      kk: "Белсенді",
    },
    "Блокировать": {
      en: "Block",
      kk: "Бұғаттау",
    },
    "Выполнена": {
      en: "Completed",
      kk: "Орындалды",
    },
    "Неактивен": {
      en: "Inactive",
      kk: "Белсенді емес",
    },
    "Открыта": {
      en: "Open",
      kk: "Ашық",
    },
    "Есть проекты в промежуточной фазе": {
      en: "There are projects in an intermediate phase",
      kk: "Аралық кезеңдегі жобалар бар",
    },
    "Подготовка проектов еще идет": {
      en: "Project preparation is still in progress",
      kk: "Жобаларды дайындау әлі жүріп жатыр",
    },
    "Открыть пользователей": {
      en: "Open users",
      kk: "Пайдаланушыларды ашу",
    },
    "Пользователи требуют проверки": {
      en: "Users require review",
      kk: "Пайдаланушылар тексеруді қажет етеді",
    },
    "Сделать преподом": {
      en: "Make professor",
      kk: "Оқытушы ету",
    },
    "Сделать студентом": {
      en: "Make student",
      kk: "Студент ету",
    },
    "Смотреть поток": {
      en: "View flow",
      kk: "Ағынды қарау",
    },
    "По выбранным фильтрам данных нет": {
      en: "No data for the selected filters",
      kk: "Таңдалған сүзгілер бойынша дерек жоқ",
    },
    "Попробуйте снять часть ограничений или обновить структуру кафедр и групп.": {
      en: "Try removing some restrictions or refreshing the structure of departments and groups.",
      kk: "Шектеулердің бір бөлігін алып тастап көріңіз немесе кафедралар мен топтар құрылымын жаңартыңыз.",
    },
    "Ожидают решения": {
      en: "Awaiting decision",
      kk: "Шешімді күтуде",
    },
    "Пока нет уведомлений": {
      en: "No notifications yet",
      kk: "Хабарландырулар әзірге жоқ",
    },
    "Готов к ревью": {
      en: "Ready for review",
      kk: "Ревьюге дайын",
    },
    "Ждет ваш ответ": {
      en: "Waiting for your response",
      kk: "Жауабыңызды күтуде",
    },
    "Набор команды": {
      en: "Team recruitment",
      kk: "Команда жинау",
    },
    "Назначен другой преподаватель": {
      en: "Another professor is assigned",
      kk: "Басқа оқытушы тағайындалған",
    },
    "Настроить критерии": {
      en: "Set up criteria",
      kk: "Критерийлерді баптау",
    },
    "Нужно ответить на приглашение преподавателя-ревьюера.": {
      en: "You need to respond to the reviewer professor's invite.",
      kk: "Ревьюер-оқытушының шақыруына жауап беру керек.",
    },
    "Обновлен": {
      en: "Updated",
      kk: "Жаңартылды",
    },
    "Обновляю контур преподавателя...": {
      en: "Refreshing professor workspace...",
      kk: "Оқытушы контурын жаңартып жатырмын...",
    },
    "Открыть критерии": {
      en: "Open criteria",
      kk: "Критерийлерді ашу",
    },
    "Открыть оценивание": {
      en: "Open grading",
      kk: "Бағалауды ашу",
    },
    "Оценить сейчас": {
      en: "Grade now",
      kk: "Қазір бағалау",
    },
    "Преподаватель пока не закреплен": {
      en: "The professor is not assigned yet",
      kk: "Оқытушы әлі бекітілмеген",
    },
    "Приглашение на ревью отклонено.": {
      en: "The review invite was declined.",
      kk: "Ревьюге шақыру қабылданбады.",
    },
    "Приглашение на ревью принято.": {
      en: "The review invite was accepted.",
      kk: "Ревьюге шақыру қабылданды.",
    },
    "Приглашение отклонено": {
      en: "Invite declined",
      kk: "Шақыру қабылданбады",
    },
    "Принять ревью": {
      en: "Accept review",
      kk: "Ревьюді қабылдау",
    },
    "Проекты пока не найдены": {
      en: "No projects found yet",
      kk: "Жобалар әзірге табылмады",
    },
    "Срочных действий нет": {
      en: "No urgent actions",
      kk: "Шұғыл әрекеттер жоқ",
    },
    "Сейчас можно спокойно посмотреть каталог проектов или вернуться к заявкам на ревью.": {
      en: "You can calmly browse the project catalog or return to review requests now.",
      kk: "Қазір жобалар каталогын асықпай қарап шығуға немесе ревью өтінімдеріне оралуға болады.",
    },
    "Введите название критерия.": {
      en: "Enter the criterion title.",
      kk: "Критерий атауын енгізіңіз.",
    },
    "Вес должен быть числом от 1 до 100.": {
      en: "Weight must be a number from 1 to 100.",
      kk: "Салмақ 1-ден 100-ге дейінгі сан болуы керек.",
    },
    "Готово к редактированию критериев.": {
      en: "Ready to edit criteria.",
      kk: "Критерийлерді өңдеуге дайын.",
    },
    "Готово. Список критериев сохранён.": {
      en: "Done. The criteria list has been saved.",
      kk: "Дайын. Критерийлер тізімі сақталды.",
    },
    "Для этого проекта у вас нет доступа к просмотру критериев.": {
      en: "You do not have access to view criteria for this project.",
      kk: "Бұл жоба үшін критерийлерді көруге қолжетімділігіңіз жоқ.",
    },
    "Добавьте первый пункт проверки, чтобы сформировать чек-лист проекта.": {
      en: "Add the first checkpoint to form the project's checklist.",
      kk: "Жобаның чек-листін қалыптастыру үшін бірінші тексеру тармағын қосыңыз.",
    },
    "Добавьте следующий критерий": {
      en: "Add the next criterion",
      kk: "Келесі критерийді қосыңыз",
    },
    "Достигнут лимит суммарного веса (100).": {
      en: "The total weight limit (100) has been reached.",
      kk: "Жалпы салмақ лимитіне (100) жетті.",
    },
    "Достигнут лимит суммарного веса 100. Для добавления нового критерия уменьшите существующие веса.": {
      en: "The total weight limit of 100 has been reached. Reduce existing weights to add a new criterion.",
      kk: "Жалпы салмақтың 100 лимитіне жетті. Жаңа критерий қосу үшін бар салмақтарды азайтыңыз.",
    },
    "Загрузка данных...": {
      en: "Loading data...",
      kk: "Деректер жүктелуде...",
    },
    "Загрузка шаблонов будет добавлена в следующем обновлении.": {
      en: "Template upload will be added in the next update.",
      kk: "Үлгілерді жүктеу келесі жаңартуда қосылады.",
    },
    "Критерий добавлен.": {
      en: "Criterion added.",
      kk: "Критерий қосылды.",
    },
    "Начните вводить новый критерий...": {
      en: "Start entering a new criterion...",
      kk: "Жаңа критерийді енгізуді бастаңыз...",
    },
    "Нет прав на редактирование.": {
      en: "No editing permissions.",
      kk: "Өңдеуге құқық жоқ.",
    },
    "Нет проектов, где вы подтверждены как преподаватель-ревьюер.": {
      en: "There are no projects where you are confirmed as the review professor.",
      kk: "Сіз ревьюер-оқытушы ретінде расталған жобалар жоқ.",
    },
    "Проект выбран в режиме read-only: примите приглашение на ревью, чтобы редактировать критерии.": {
      en: "The project is opened in read-only mode: accept the review invite to edit criteria.",
      kk: "Жоба тек оқу режимінде ашылды: критерийлерді өңдеу үшін ревью шақыруын қабылдаңыз.",
    },
    "Редактирование недоступно: нет прав на этот проект.": {
      en: "Editing is unavailable: no permissions for this project.",
      kk: "Өңдеу қолжетімсіз: бұл жобаға құқық жоқ.",
    },
    "Редактирование недоступно: преподаватель должен быть подтверждён в проекте.": {
      en: "Editing is unavailable: the professor must be confirmed in the project.",
      kk: "Өңдеу қолжетімсіз: оқытушы жобаға расталуы керек.",
    },
    "Сначала выберите проект для настройки критериев.": {
      en: "Choose a project first to configure criteria.",
      kk: "Критерийлерді баптау үшін алдымен жобаны таңдаңыз.",
    },
    "Без штрафа": {
      en: "No penalty",
      kk: "Айыппұл жоқ",
    },
    "Вернуть на пересдачу": {
      en: "Return for resubmission",
      kk: "Қайта тапсыруға қайтару",
    },
    "Вернуть на пересдачу можно только на этапе финального оценивания.": {
      en: "You can return for resubmission only at the final grading stage.",
      kk: "Қайта тапсыруға тек қорытынды бағалау кезеңінде қайтаруға болады.",
    },
    "Выберите проект из списка, чтобы открыть рабочее место ревьюера.": {
      en: "Choose a project from the list to open the review workspace.",
      kk: "Ревьюердің жұмыс орнын ашу үшін тізімнен жобаны таңдаңыз.",
    },
    "Готово к завершению": {
      en: "Ready to complete",
      kk: "Аяқтауға дайын",
    },
    "Да": {
      en: "Yes",
      kk: "Иә",
    },
    "Да, выполнено": {
      en: "Yes, completed",
      kk: "Иә, орындалды",
    },
    "Нет": {
      en: "No",
      kk: "Жоқ",
    },
    "Нет, не выполнено": {
      en: "No, not completed",
      kk: "Жоқ, орындалмады",
    },
    "Для этого статуса проекта редактирование оценки недоступно.": {
      en: "Editing the grading is unavailable for this project status.",
      kk: "Бұл жоба күйінде бағаны өңдеу қолжетімсіз.",
    },
    "Добавьте короткий комментарий только там, где он поможет команде понять решение.": {
      en: "Add a short comment only where it helps the team understand the decision.",
      kk: "Қысқа пікірді тек командаға шешімді түсінуге көмектесетін жерде қосыңыз.",
    },
    "Завершение доступно только на этапе финального оценивания.": {
      en: "Completion is available only at the final grading stage.",
      kk: "Аяқтау тек қорытынды бағалау кезеңінде қолжетімді.",
    },
    "Команда в активной работе": {
      en: "The team is actively working",
      kk: "Команда белсенді жұмыста",
    },
    "Команда еще работает над проектом. Форма оценки откроется после отправки на ревью.": {
      en: "The team is still working on the project. The grading form will open after submission for review.",
      kk: "Команда жобамен әлі жұмыс істеп жатыр. Бағалау формасы ревьюге жібергеннен кейін ашылады.",
    },
    "Команда исправляет замечания после пересдачи. Когда проект снова отправят на финальное ревью, итоговая оценка будет учитывать небольшой штраф.": {
      en: "The team is fixing remarks after resubmission. When the project is sent again for final review, the final grade will include a small penalty.",
      kk: "Команда қайта тапсырудан кейін ескертулерді түзетіп жатыр. Жоба қайтадан қорытынды ревьюге жіберілгенде, қорытынды баға аздаған айыппұлды ескереді.",
    },
    "Комментарий преподавателя": {
      en: "Professor comment",
      kk: "Оқытушы пікірі",
    },
    "Контекст оценивания": {
      en: "Grading context",
      kk: "Бағалау контексті",
    },
    "Критерии пока не настроены": {
      en: "Criteria are not configured yet",
      kk: "Критерийлер әлі бапталмаған",
    },
    "Критерий не выполнен": {
      en: "Criterion not met",
      kk: "Критерий орындалмады",
    },
    "Критерий подтвержден": {
      en: "Criterion confirmed",
      kk: "Критерий расталды",
    },
    "Лидер проекта": {
      en: "Project lead",
      kk: "Жоба жетекшісі",
    },
    "Лидирует": {
      en: "Leading",
      kk: "Бастап тұр",
    },
    "Нельзя завершить ревью, пока не отмечены все критерии.": {
      en: "You cannot complete the review until all criteria are marked.",
      kk: "Барлық критерийлер белгіленбейінше ревьюді аяқтау мүмкін емес.",
    },
    "Нет доступных проектов для оценивания.": {
      en: "No available projects for grading.",
      kk: "Бағалауға қолжетімді жобалар жоқ.",
    },
    "Ответ еще не выбран": {
      en: "No answer selected yet",
      kk: "Жауап әлі таңдалмаған",
    },
    "Откройте проект из списка сверху, чтобы увидеть условия ревью и прогресс оценивания.": {
      en: "Open a project from the list above to see the review conditions and grading progress.",
      kk: "Ревью шарттары мен бағалау прогресін көру үшін жоғарыдағы тізімнен жобаны ашыңыз.",
    },
    "Откройте проект из списка справа, чтобы увидеть контекст ревью и критерии.": {
      en: "Open a project from the list on the right to see the review context and criteria.",
      kk: "Ревью контексті мен критерийлерді көру үшін оң жақтағы тізімнен жобаны ашыңыз.",
    },
    "Отметьте каждый критерий, добавьте точечные комментарии и завершите оценивание только после полного покрытия чек-листа.": {
      en: "Mark every criterion, add focused comments, and complete the grading only after the checklist is fully covered.",
      kk: "Әр критерийді белгілеңіз, нақты пікірлер қосыңыз және чек-лист толық жабылғаннан кейін ғана бағалауды аяқтаңыз.",
    },
    "Оценивание еще не открыто": {
      en: "Grading is not open yet",
      kk: "Бағалау әлі ашылған жоқ",
    },
    "Оценивание завершено. Итог закреплен в проекте.": {
      en: "Grading is complete. The final result has been fixed in the project.",
      kk: "Бағалау аяқталды. Қорытынды жобаға бекітілді.",
    },
    "Оценивание завершено. Страница работает как итоговый отчет по проекту.": {
      en: "Grading is complete. The page now serves as the final report for the project.",
      kk: "Бағалау аяқталды. Бұл бет енді жоба бойынша қорытынды есеп ретінде жұмыс істейді.",
    },
    "Оценивание уже завершено": {
      en: "Grading is already complete",
      kk: "Бағалау әлдеқашан аяқталған",
    },
    "Оценка сохранена. Можно продолжать ревью или завершить его позже.": {
      en: "Grading saved. You can continue the review or finish it later.",
      kk: "Баға сақталды. Ревьюді жалғастыруға немесе кейін аяқтауға болады.",
    },
    "Переключаю проект...": {
      en: "Switching project...",
      kk: "Жоба ауыстырылып жатыр...",
    },
    "Пересдачи": {
      en: "Resubmissions",
      kk: "Қайта тапсырулар",
    },
    "Повторное финальное ревью": {
      en: "Repeated final review",
      kk: "Қайталама қорытынды ревью",
    },
    "Пока без зафиксированной активности": {
      en: "No recorded activity yet",
      kk: "Әзірге тіркелген белсенділік жоқ",
    },
    "Пока нет данных для сводки": {
      en: "No summary data yet",
      kk: "Қорытынды үшін дерек әзірге жоқ",
    },
    "Проект возвращен на доработку": {
      en: "Project returned for revision",
      kk: "Жоба пысықтауға қайтарылды",
    },
    "Проект еще готовится к запуску": {
      en: "The project is still being prepared for launch",
      kk: "Жоба әлі іске қосуға дайындалып жатыр",
    },
    "Проект открыт для финального ревью": {
      en: "The project is open for final review",
      kk: "Жоба қорытынды ревьюге ашық",
    },
    "Рабочее место ревьюера готово. Можно сохранять оценку.": {
      en: "The reviewer workspace is ready. You can save the grading.",
      kk: "Ревьюердің жұмыс орны дайын. Бағаны сақтауға болады.",
    },
    "Ревью-форма пока закрыта. Следующий преподавательский шаг появится после отправки проекта на финальную оценку.": {
      en: "The review form is closed for now. The next professor action will appear after the project is sent for final grading.",
      kk: "Ревью формасы әзірге жабық. Келесі оқытушы әрекеті жоба қорытынды бағалауға жіберілгеннен кейін пайда болады.",
    },
    "Редактирование доступно": {
      en: "Editing available",
      kk: "Өңдеу қолжетімді",
    },
    "Режим страницы": {
      en: "Page mode",
      kk: "Бет режимі",
    },
    "Собираю контекст ревьюера...": {
      en: "Collecting reviewer context...",
      kk: "Ревьюер контексін жинап жатырмын...",
    },
    "Состояние проекта определяет, что преподаватель может сделать прямо сейчас.": {
      en: "The project state determines what the professor can do right now.",
      kk: "Жоба күйі оқытушының дәл қазір не істей алатынын анықтайды.",
    },
    "Страница остается полезной как итоговый отчет: можно просмотреть критерии, покрытие и комментарии без редактирования.": {
      en: "The page remains useful as a final report: criteria, coverage, and comments can be viewed without editing.",
      kk: "Бұл бет қорытынды есеп ретінде пайдалы болып қалады: критерийлерді, қамтуды және пікірлерді өңдеусіз көруге болады.",
    },
    "Только просмотр": {
      en: "View only",
      kk: "Тек көру",
    },
    "Чтобы завершить оценивание, отметьте каждый критерий как выполненный или невыполненный.": {
      en: "To complete the grading, mark each criterion as completed or not completed.",
      kk: "Бағалауды аяқтау үшін әр критерийді орындалды немесе орындалмады деп белгілеңіз.",
    },
    "Этап подготовки к активной фазе": {
      en: "Preparation stage before the active phase",
      kk: "Белсенді кезеңге дайындық сатысы",
    },
    "Это повторная попытка после пересдачи. Можно завершить оценивание или вернуть проект на еще одну доработку.": {
      en: "This is a repeated attempt after resubmission. You can complete the grading or return the project for another revision.",
      kk: "Бұл қайта тапсырудан кейінгі қайталама әрекет. Бағалауды аяқтауға немесе жобаны тағы бір пысықтауға қайтаруға болады.",
    },
    "README / описание": {
      en: "README / description",
      kk: "README / сипаттама",
    },
    "Без исполнителя": {
      en: "No assignee",
      kk: "Орындаушы жоқ",
    },
    "Взять": {
      en: "Claim",
      kk: "Алу",
    },
    "Добавьте хотя бы один стек.": {
      en: "Add at least one stack item.",
      kk: "Кемінде бір стек қосыңыз.",
    },
    "Завершение": {
      en: "Completion",
      kk: "Аяқтау",
    },
    "Задача взята в работу": {
      en: "Task claimed",
      kk: "Тапсырма жұмысқа алынды",
    },
    "Задача завершена": {
      en: "Task completed",
      kk: "Тапсырма аяқталды",
    },
    "Задача закрыта": {
      en: "Task closed",
      kk: "Тапсырма жабылды",
    },
    "Изменение задачи": {
      en: "Task update",
      kk: "Тапсырманы өзгерту",
    },
    "Итоговая оценка опубликована. Детали по критериям видны только участникам команды.": {
      en: "The final grading has been published. Details by criteria are visible only to team members.",
      kk: "Қорытынды баға жарияланды. Критерийлер бойынша егжей-тегжей тек команда қатысушыларына көрінеді.",
    },
    "Итоговая оценка опубликована. Детализация критериев доступна только участникам команды.": {
      en: "The final grading has been published. Criteria details are available only to team members.",
      kk: "Қорытынды баға жарияланды. Критерийлердің егжей-тегжейі тек команда қатысушыларына қолжетімді.",
    },
    "Критерии добавляет преподаватель на вкладке «Критерии». Команда здесь только сверяется с ними.": {
      en: "The professor adds criteria on the “Criteria” tab. The team only checks against them here.",
      kk: "Критерийлерді оқытушы «Критерийлер» қойындысында қосады. Команда мұнда тек солармен салыстырады.",
    },
    "Критерии пока не настроены преподавателем.": {
      en: "Criteria have not been configured by the professor yet.",
      kk: "Критерийлерді оқытушы әлі баптамаған.",
    },
    "Лента задачи": {
      en: "Task timeline",
      kk: "Тапсырма лентасы",
    },
    "Лог активности появится после первых действий по задаче.": {
      en: "The activity log will appear after the first actions on the task.",
      kk: "Белсенділік журналы тапсырма бойынша алғашқы әрекеттерден кейін пайда болады.",
    },
    "Назначен исполнитель": {
      en: "Assignee selected",
      kk: "Орындаушы тағайындалды",
    },
    "Назначить": {
      en: "Assign",
      kk: "Тағайындау",
    },
    "Нет базовых ролей": {
      en: "No base roles",
      kk: "Негізгі рөлдер жоқ",
    },
    "Нет разрешений": {
      en: "No permissions",
      kk: "Рұқсаттар жоқ",
    },
    "Ожидайте назначения или обновлений.": {
      en: "Wait for assignment or updates.",
      kk: "Тағайындауды немесе жаңартуларды күтіңіз.",
    },
    "Отправить на оценивание": {
      en: "Submit for grading",
      kk: "Бағалауға жіберу",
    },
    "Отправить проект на оценивание": {
      en: "Submit project for grading",
      kk: "Жобаны бағалауға жіберу",
    },
    "Подсказка": {
      en: "Hint",
      kk: "Кеңес",
    },
    "Подходящие студенты не найдены.": {
      en: "No matching students found.",
      kk: "Сәйкес студенттер табылмады.",
    },
    "Преподаватель еще не добавил критерии. Студенты здесь только видят этот список и сверяются с ним.": {
      en: "The professor has not added criteria yet. Students only see this list here and use it as a reference.",
      kk: "Оқытушы критерийлерді әлі қоспады. Студенттер мұнда тек осы тізімді көріп, онымен салыстырады.",
    },
    "Преподаватель отклонил приглашение. Выберите другого преподавателя.": {
      en: "The professor declined the invite. Choose another professor.",
      kk: "Оқытушы шақыруды қабылдамады. Басқа оқытушыны таңдаңыз.",
    },
    "Преподаватель подтвердил участие в ревью.": {
      en: "The professor confirmed participation in the review.",
      kk: "Оқытушы ревьюге қатысуын растады.",
    },
    "Преподаватель пока не приглашён.": {
      en: "The professor has not been invited yet.",
      kk: "Оқытушы әлі шақырылмаған.",
    },
    "Приглашён": {
      en: "Invited",
      kk: "Шақырылған",
    },
    "Просрочено": {
      en: "Overdue",
      kk: "Мерзімі өткен",
    },
    "Работа": {
      en: "Work",
      kk: "Жұмыс",
    },
    "Сейчас": {
      en: "Current",
      kk: "Қазір",
    },
    "Система": {
      en: "System",
      kk: "Жүйе",
    },
    "Создана задача": {
      en: "Task created",
      kk: "Тапсырма құрылды",
    },
    "Текущий студент": {
      en: "Current student",
      kk: "Ағымдағы студент",
    },
    "Нажми, чтобы создать новый.": {
      en: "Click to create a new one.",
      kk: "Жаңасын құру үшін басыңыз.",
    },
    "Открыть результат": {
      en: "Open result",
      kk: "Нәтижені ашу",
    },
    "Открыть": {
      en: "Open",
      kk: "Ашу",
    },
    "Отмена": {
      en: "Cancel",
      kk: "Болдырмау",
    },
    "# Заголовок Напишите статью в формате Markdown…": {
      en: "# Title Write the article in Markdown…",
      kk: "# Тақырып Мақаланы Markdown форматында жазыңыз…",
    },
    "git, ветвление, basics": {
      en: "git, branching, basics",
      kk: "git, тармақталу, basics",
    },
    "Управление проектами для": {
      en: "Project management for",
      kk: "Жобаларды басқару",
    },
    "студентов": {
      en: "students",
      kk: "студенттер",
    },
    "нового поколения": {
      en: "the next generation",
      kk: "жаңа буынға арналған",
    },
    "Платформа для разработчиков, построенная на базе": {
      en: "A developer platform built on top of",
      kk: "Негізінде құрылған әзірлеушілер платформасы",
    },
    "открытого кода.": {
      en: "open source.",
      kk: "ашық код.",
    },
    "https://... https://... (каждая ссылка с новой строки)": {
      en: "https://... https://... (each link on a new line)",
      kk: "https://... https://... (әр сілтеме жаңа жолдан)",
    },
    "Админ": {
      en: "Admin",
      kk: "Әкімші",
    },
    "Активных:": {
      en: "Active:",
      kk: "Белсенді:",
    },
    "Минимум 10 символов": {
      en: "Minimum 10 characters",
      kk: "Кемі 10 таңба",
    },
    "Пользователи не найдены": {
      en: "Users not found",
      kk: "Пайдаланушылар табылмады",
    },
    "Проекты не найдены": {
      en: "Projects not found",
      kk: "Жобалар табылмады",
    },
    "друг": {
      en: "friend",
      kk: "дос",
    },
    "недавно": {
      en: "recently",
      kk: "жуырда",
    },
    "пользователь": {
      en: "user",
      kk: "пайдаланушы",
    },
    "админ": {
      en: "admin",
      kk: "әкімші",
    },
    "Ошибка загрузки:": {
      en: "Loading error:",
      kk: "Жүктеу қатесі:",
    },
    "Ошибка:": {
      en: "Error:",
      kk: "Қате:",
    },
    "Например 101": {
      en: "For example 101",
      kk: "Мысалы 101",
    },
    "заверш": {
      en: "complete",
      kk: "аяқтал",
    },
    "опублик": {
      en: "publish",
      kk: "жарияла",
    },
    "отклон": {
      en: "decline",
      kk: "қабылдама",
    },
    "пересдач": {
      en: "resubm",
      kk: "қайта тапсыр",
    },
    "пересдача": {
      en: "resubmission",
      kk: "қайта тапсыру",
    },
    "пересдачи": {
      en: "resubmissions",
      kk: "қайта тапсырулар",
    },
    "принят": {
      en: "accepted",
      kk: "қабылданды",
    },
    "создан": {
      en: "created",
      kk: "құрылды",
    },
    "убрали": {
      en: "removed",
      kk: "алып тасталды",
    },
    "Дать разрешение на запуск и запустить": {
      en: "Approve launch and start",
      kk: "Іске қосуға рұқсат беріп, бастау",
    },
    "До запуска не хватает критериев. Здесь главное действие преподавателя именно настройка чек-листа.": {
      en: "Criteria are still missing before launch. The professor's main action here is setting up the checklist.",
      kk: "Іске қосуға дейін критерийлер жетіспейді. Мұндағы оқытушының басты әрекеті - чек-листті баптау.",
    },
    "Когда в faculty-контуре появятся проекты, они отобразятся здесь автоматически.": {
      en: "When projects appear in the faculty scope, they will show up here automatically.",
      kk: "Faculty контурында жобалар пайда болғанда, олар мұнда автоматты түрде көрінеді.",
    },
    "Команда готова к старту. Вы можете дать разрешение на запуск и перевести проект в ACTIVE.": {
      en: "The team is ready to start. You can approve the launch and move the project to ACTIVE.",
      kk: "Команда бастауға дайын. Іске қосуға рұқсат беріп, жобаны ACTIVE күйіне ауыстыра аласыз.",
    },
    "Команда ждет вашего решения. После принятия вы сможете вести критерии и финальное ревью.": {
      en: "The team is waiting for your decision. After accepting, you will be able to manage criteria and the final review.",
      kk: "Команда сіздің шешіміңізді күтіп отыр. Қабылдағаннан кейін критерийлер мен қорытынды ревьюді жүргізе аласыз.",
    },
    "Команда завершила работу и отправила проект на итоговое оценивание.": {
      en: "The team finished the work and sent the project for final grading.",
      kk: "Команда жұмысты аяқтап, жобаны қорытынды бағалауға жіберді.",
    },
    "Команда сейчас в работе. На этом этапе преподаватель сопровождает проект и ждет отправки на оценивание.": {
      en: "The team is currently working. At this stage the professor supports the project and waits for grading submission.",
      kk: "Команда қазір жұмыс істеп жатыр. Бұл кезеңде оқытушы жобаны сүйемелдеп, бағалауға жіберуді күтеді.",
    },
    "Команда, преподавательское ревью и критерии готовы. Вы можете дать разрешение на запуск и перевести проект в ACTIVE.": {
      en: "The team, professor review, and criteria are ready. You can approve the launch and move the project to ACTIVE.",
      kk: "Команда, оқытушы ревьюі және критерийлер дайын. Іске қосуға рұқсат беріп, жобаны ACTIVE күйіне ауыстыра аласыз.",
    },
    "Проект в активной фазе. Здесь полезнее открыть карточку проекта и посмотреть динамику команды.": {
      en: "The project is in the active phase. It is more useful to open the project card and view the team dynamics here.",
      kk: "Жоба белсенді кезеңде. Мұнда жоба картасын ашып, команда динамикасын қарау пайдалырақ.",
    },
    "Проект доступен для просмотра, но преподавательские действия здесь пока не требуются.": {
      en: "The project is available for viewing, but no professor actions are required here yet.",
      kk: "Жобаны көруге болады, бірақ әзірге мұнда оқытушы әрекеттері қажет емес.",
    },
    "Проект еще готовится к активной фазе. Проверьте критерии и дождитесь готовности команды.": {
      en: "The project is still preparing for the active phase. Check the criteria and wait for the team to be ready.",
      kk: "Жоба әлі белсенді кезеңге дайындалып жатыр. Критерийлерді тексеріп, команданың дайын болуын күтіңіз.",
    },
    "Проект принадлежит вам. Управление набором, запуском и составом команды остается внутри карточки самого проекта.": {
      en: "The project belongs to you. Recruitment, launch, and team composition management remain inside the project card itself.",
      kk: "Жоба сізге тиесілі. Іріктеуді, іске қосуды және команда құрамын басқару жобаның өз картасының ішінде қалады.",
    },
    "Проект уже передан на финальную проверку. Откройте оценивание и завершите ревью по критериям.": {
      en: "The project has already been sent for final review. Open grading and complete the review by criteria.",
      kk: "Жоба қорытынды тексеруге жіберілді. Бағалауды ашып, критерийлер бойынша ревьюді аяқтаңыз.",
    },
    "Разрешение на запуск выдано. Проект переведен в ACTIVE.": {
      en: "Launch approval granted. The project has been moved to ACTIVE.",
      kk: "Іске қосуға рұқсат берілді. Жоба ACTIVE күйіне ауыстырылды.",
    },
    "Ревью завершено. Можно открыть оценивание, чтобы посмотреть итоговую картину и комментарии.": {
      en: "The review is complete. You can open grading to see the final picture and comments.",
      kk: "Ревью аяқталды. Қорытынды көрініс пен пікірлерді көру үшін бағалауды ашуға болады.",
    },
    "Ревью закреплено за вами": {
      en: "The review is assigned to you",
      kk: "Ревью сізге бекітілген",
    },
    "У проекта еще нет критериев. Без них команда не выйдет в понятный ревью-поток.": {
      en: "The project does not have criteria yet. Without them the team will not enter a clear review flow.",
      kk: "Жобада әлі критерийлер жоқ. Оларсыз команда түсінікті ревью ағынына кіре алмайды.",
    },
    "коллега": {
      en: "colleague",
      kk: "әріптес",
    },
    "Нажмите кнопку ниже, чтобы открыть форму добавления.": {
      en: "Click the button below to open the add form.",
      kk: "Қосу формасын ашу үшін төмендегі батырманы басыңыз.",
    },
    "Редактирование станет доступно после принятия приглашения на ревью.": {
      en: "Editing will become available after you accept the review invite.",
      kk: "Өңдеу ревью шақыруын қабылдағаннан кейін қолжетімді болады.",
    },
    "Сначала выберите проект.": {
      en: "Choose a project first.",
      kk: "Алдымен жобаны таңдаңыз.",
    },
    "Сохранение шаблона пока недоступно в API.": {
      en: "Saving a template is not available in the API yet.",
      kk: "Үлгіні сақтау API-де әзірге қолжетімді емес.",
    },
    "Описание критерия отсутствует.": {
      en: "Criterion description is missing.",
      kk: "Критерий сипаттамасы жоқ.",
    },
    "Когда в проекте появятся задачи и события, здесь отобразится вклад участников.": {
      en: "When tasks and events appear in the project, participants' contribution will be shown here.",
      kk: "Жобада тапсырмалар мен оқиғалар пайда болғанда, мұнда қатысушылардың үлесі көрінеді.",
    },
    "Оценивание уже можно черновиком заполнять, но основной акцент здесь на полноте критериев и готовности команды.": {
      en: "Grading can already be filled in as a draft, but the main focus here is the completeness of criteria and the team's readiness.",
      kk: "Бағалауды қазірдің өзінде черновик ретінде толтыруға болады, бірақ мұнда негізгі назар критерийлердің толықтығы мен команданың дайындығына аударылады.",
    },
    "Сейчас ценнее проверить критерии и убедиться, что преподавательское ревью подтверждено, чем пытаться выставлять оценку заранее.": {
      en: "It is more valuable now to check the criteria and make sure the professor review is confirmed than to try to grade too early.",
      kk: "Қазір ерте баға қоюға тырысқаннан гөрі критерийлерді тексеріп, оқытушы ревьюінің расталғанына көз жеткізу маңыздырақ.",
    },
    "Сначала откройте страницу критериев и соберите чек-лист оценки. После этого форма ревью автоматически станет осмысленной.": {
      en: "Open the criteria page first and assemble the grading checklist. After that the review form will automatically become meaningful.",
      kk: "Алдымен критерийлер бетін ашып, бағалау чек-листін жинаңыз. Осыдан кейін ревью формасы автоматты түрде мағыналы болады.",
    },
    "Эта страница останется рабочим местом ревьюера, но сами отметки станут доступны после перехода проекта в REVIEW или GRADING.": {
      en: "This page will remain the reviewer's workspace, but the marks themselves will become available after the project moves to REVIEW or GRADING.",
      kk: "Бұл бет ревьюердің жұмыс орны болып қалады, бірақ белгілердің өзі жоба REVIEW немесе GRADING күйіне өткеннен кейін қолжетімді болады.",
    },
    "Аватар обновлен.": {
      en: "Avatar updated.",
      kk: "Аватар жаңартылды.",
    },
    "Аватар удален.": {
      en: "Avatar removed.",
      kk: "Аватар жойылды.",
    },
    "Введите корректное имя (минимум 2 символа).": {
      en: "Enter a valid name (at least 2 characters).",
      kk: "Дұрыс атты енгізіңіз (кемі 2 таңба).",
    },
    "Загружаем...": {
      en: "Loading...",
      kk: "Жүктелуде...",
    },
    "Мой профиль": {
      en: "My profile",
      kk: "Менің профилім",
    },
    "Поддерживаются JPG, PNG и WEBP.": {
      en: "JPG, PNG, and WEBP are supported.",
      kk: "JPG, PNG және WEBP қолданылады.",
    },
    "Профиль обновлен.": {
      en: "Profile updated.",
      kk: "Профиль жаңартылды.",
    },
    "Удаляем...": {
      en: "Removing...",
      kk: "Жойылып жатыр...",
    },
    "Файл слишком большой (макс. 8MB).": {
      en: "The file is too large (max 8MB).",
      kk: "Файл тым үлкен (макс. 8MB).",
    },
    "Введите корректный email.": {
      en: "Enter a valid email.",
      kk: "Дұрыс email енгізіңіз.",
    },
    "Выберите кафедру и введите номер группы.": {
      en: "Choose a department and enter the group number.",
      kk: "Кафедраны таңдап, топ нөмірін енгізіңіз.",
    },
    "Заполните все поля пароля.": {
      en: "Fill in all password fields.",
      kk: "Құпиясөздің барлық өрістерін толтырыңыз.",
    },
    "Заявка отправлена администратору.": {
      en: "The request has been sent to the administrator.",
      kk: "Өтінім әкімшіге жіберілді.",
    },
    "Заявок пока нет.": {
      en: "No requests yet.",
      kk: "Өтінімдер әзірге жоқ.",
    },
    "Код должен содержать 6 цифр.": {
      en: "The code must contain 6 digits.",
      kk: "Код 6 цифрдан тұруы керек.",
    },
    "Новый email ожидает подтверждения.": {
      en: "The new email is awaiting confirmation.",
      kk: "Жаңа email растауды күтіп тұр.",
    },
    "Новый email ожидает подтверждения. Письмо отправлено.": {
      en: "The new email is awaiting confirmation. The message has been sent.",
      kk: "Жаңа email растауды күтіп тұр. Хат жіберілді.",
    },
    "Новый пароль и подтверждение не совпадают.": {
      en: "The new password and confirmation do not match.",
      kk: "Жаңа құпиясөз бен растау сәйкес келмейді.",
    },
    "Пароль должен быть не короче 8 символов и содержать буквы и цифры.": {
      en: "The password must be at least 8 characters long and contain letters and digits.",
      kk: "Құпиясөз кемінде 8 таңбадан тұрып, әріптер мен сандарды қамтуы керек.",
    },
    "Пароль обновлен. Выполняется выход из системы.": {
      en: "Password updated. Signing out...",
      kk: "Құпиясөз жаңартылды. Жүйеден шығу орындалып жатыр.",
    },
    "Письмо отправлено повторно.": {
      en: "The message has been sent again.",
      kk: "Хат қайта жіберілді.",
    },
    "Письмо подтверждения отправлено на новый email.": {
      en: "A confirmation email has been sent to the new address.",
      kk: "Растау хаты жаңа email-ге жіберілді.",
    },
    "Подтверждаем...": {
      en: "Confirming...",
      kk: "Расталып жатыр...",
    },
    "Ссылка подтверждения устарела. Отправьте письмо повторно.": {
      en: "The confirmation link is outdated. Send the email again.",
      kk: "Растау сілтемесі ескірген. Хатты қайта жіберіңіз.",
    },
    "Укажите код подтверждения.": {
      en: "Enter the confirmation code.",
      kk: "Растау кодын көрсетіңіз.",
    },
    "Активных участников пока нет.": {
      en: "There are no active members yet.",
      kk: "Белсенді қатысушылар әзірге жоқ.",
    },
    "Ближайшие требования пока не определены.": {
      en: "The nearest requirements have not been defined yet.",
      kk: "Ең жақын талаптар әлі анықталмаған.",
    },
    "Команда не набрана.": {
      en: "The team is not assembled.",
      kk: "Команда жиналмаған.",
    },
    "Стек пока не заполнен.": {
      en: "The stack is not filled in yet.",
      kk: "Стек әзірге толтырылмаған.",
    },
    "Активных участников {...} из {...}. Составом обычно управляют тимлид, ко-лид и рекрутер. Когда каждое место будет занято и преподаватель подтвердит участие, проект можно будет запускать.": {
      en: "There are {...} active members out of {...}. The composition is usually managed by the team lead, co-lead, and recruiter. When every seat is filled and the professor confirms participation, the project can be launched.",
      kk: "Қазір {...} белсенді қатысушының {...}-і бар. Құрамды әдетте тимлид, ко-лид және рекрутер басқарады. Әр орын толып, оқытушы қатысуын растаған кезде, жобаны іске қосуға болады.",
    },
    "Без задач проект нельзя отправить на оценивание. Первые карточки обычно создает тимлид или task manager, а затем распределяет их по ролям.": {
      en: "Without tasks, the project cannot be submitted for grading. The first cards are usually created by the team lead or task manager and then distributed by roles.",
      kk: "Тапсырмаларсыз жобаны бағалауға жіберу мүмкін емес. Алғашқы карточкаларды әдетте тимлид немесе task manager жасап, содан кейін оларды рөлдер бойынша бөледі.",
    },
    "Без критериев проект не сможет перейти дальше. Их заранее добавляет преподаватель, а студенты здесь только видят готовый список.": {
      en: "Without criteria, the project will not be able to move forward. The professor adds them in advance, and students only see the finished list here.",
      kk: "Критерийлерсіз жоба әрі қарай өте алмайды. Оларды оқытушы алдын ала қосады, ал студенттер мұнда тек дайын тізімді көреді.",
    },
    "Без критериев проект упрется в блокер. Их должен добавить преподаватель до момента отправки проекта на проверку.": {
      en: "Without criteria, the project will hit a blocker. The professor must add them before the project is sent for review.",
      kk: "Критерийлерсіз жоба блокерге тіреледі. Оларды жоба тексеруге жіберілгенге дейін оқытушы қосуы керек.",
    },
    "Во время оценивания обычно уже не меняют участников. Здесь важно, чтобы роли и вклад команды были отражены корректно.": {
      en: "Participants are usually no longer changed during grading. It is important here that the roles and the team's contribution are reflected correctly.",
      kk: "Бағалау кезінде қатысушыларды әдетте енді өзгертпейді. Мұнда рөлдер мен команданың үлесі дұрыс көрсетілуі маңызды.",
    },
    "Все проверенные критерии отмечены как выполненные.": {
      en: "All reviewed criteria are marked as completed.",
      kk: "Барлық тексерілген критерийлер орындалды деп белгіленген.",
    },
    "Добавьте роли и количество мест заранее. Так при открытии набора студентам будет сразу понятно, кого именно вы ищете.": {
      en: "Add roles and seat counts in advance. This way, when recruitment opens, students will immediately understand whom exactly you are looking for.",
      kk: "Рөлдер мен орын санын алдын ала қосыңыз. Сонда іріктеу ашылғанда студенттерге кімді іздеп отырғаныңыз бірден түсінікті болады.",
    },
    "Здесь можно посмотреть, как команда прошла путь от первых задач до завершения кейса.": {
      en: "Here you can see how the team went from the first tasks to the completion of the case.",
      kk: "Мұнда команданың алғашқы тапсырмалардан кейсті аяқтауға дейінгі жолын көруге болады.",
    },
    "Здесь остается финальный набор требований, по которым преподаватель оценивал проект.": {
      en: "The final set of requirements by which the professor graded the project remains here.",
      kk: "Оқытушы жобаны бағалаған қорытынды талаптар жиыны осы жерде қалады.",
    },
    "Именно этот состав будет виден в завершенном кейсе проекта и в истории работы.": {
      en: "This exact team composition will be visible in the completed project case and in the work history.",
      kk: "Дәл осы құрам аяқталған жоба кейсінде және жұмыс тарихында көрінеді.",
    },
    "Канбан готов. Если преподаватель уже подтвержден, проект можно отправлять на оценивание.": {
      en: "The Kanban board is ready. If the professor is already confirmed, the project can be submitted for grading.",
      kk: "Kanban дайын. Егер оқытушы әлдеқашан расталған болса, жобаны бағалауға жіберуге болады.",
    },
    "Комментарий преподавателя пока не добавлен.": {
      en: "The professor comment has not been added yet.",
      kk: "Оқытушының пікірі әлі қосылмаған.",
    },
    "На этапе оценивания канбан нужен как подтверждение сделанной работы. Новые задачи обычно уже не добавляют.": {
      en: "At the grading stage, the Kanban board is needed as confirmation of the work done. New tasks are usually no longer added.",
      kk: "Бағалау кезеңінде Kanban жасалған жұмыстың растауы ретінде қажет. Жаңа тапсырмалар әдетте енді қосылмайды.",
    },
    "Нет участников под текущий фильтр.": {
      en: "No members match the current filter.",
      kk: "Ағымдағы сүзгіге сәйкес қатысушылар жоқ.",
    },
    "Ожидаем подтверждения преподавателя в его кабинете ревью.": {
      en: "Waiting for the professor's confirmation in their review workspace.",
      kk: "Оқытушының өз ревью кабинетінде растауын күтіп отырмыз.",
    },
    "Проверьте, что у каждого участника есть своя роль и доступы. Если нужно, здесь же можно управлять составом и приглашениями.": {
      en: "Make sure every participant has a role and permissions. If needed, you can manage the team composition and invites here as well.",
      kk: "Әр қатысушының өз рөлі мен рұқсаттары бар екенін тексеріңіз. Қажет болса, мұнда құрам мен шақыруларды да басқаруға болады.",
    },
    "Проект возвращен на пересдачу. Закройте замечания преподавателя и отправьте его на оценивание повторно.": {
      en: "The project has been returned for resubmission. Close the professor's remarks and send it for grading again.",
      kk: "Жоба қайта тапсыруға қайтарылды. Оқытушының ескертулерін жауып, оны бағалауға қайта жіберіңіз.",
    },
    "Сначала соберите команду, преподавателя и критерии. Полноценный канбан удобнее вести, когда проект уже перешел в рабочую фазу.": {
      en: "First assemble the team, professor, and criteria. A full Kanban board is easier to maintain once the project has already moved into the work phase.",
      kk: "Алдымен команданы, оқытушыны және критерийлерді жинаңыз. Толыққанды Kanban-ды жоба жұмыс кезеңіне өткенде жүргізу ыңғайлырақ.",
    },
    "У вас есть приглашение на ревью. Откройте страницу": {
      en: "You have an invite for review. Open the page",
      kk: "Сізде ревьюге шақыру бар. Бетті ашыңыз",
    },
    "и примите его.": {
      en: "and accept it.",
      kk: "және оны қабылдаңыз.",
    },
    "добавить критерии": {
      en: "add criteria",
      kk: "критерийлер қосу",
    },
    "назначить преподавателя": {
      en: "assign professor",
      kk: "оқытушы тағайындау",
    },
    "получить подтверждение преподавателя": {
      en: "get professor confirmation",
      kk: "оқытушының растауын алу",
    },
    "создать роли": {
      en: "create roles",
      kk: "рөлдер құру",
    },
    "команда": {
      en: "team",
      kk: "команда",
    },
    "текущее состояние": {
      en: "current state",
      kk: "ағымдағы күй",
    },
    "Удалить задачу": {
      en: "Delete task",
      kk: "Тапсырманы жою",
    },
    "Удалить участника": {
      en: "Remove member",
      kk: "Қатысушыны жою",
    },
    "Набор закрыт": {
      en: "Recruitment closed",
      kk: "Іріктеу жабық",
    },
    "Задач:": {
      en: "Tasks:",
      kk: "Тапсырмалар:",
    },
    "Критериев:": {
      en: "Criteria:",
      kk: "Критерийлер:",
    },
    "Участников:": {
      en: "Members:",
      kk: "Қатысушылар:",
    },
    "Запустить": {
      en: "Launch",
      kk: "Іске қосу",
    },
    "Показано 0 пользователей": {
      en: "0 users shown",
      kk: "0 пайдаланушы көрсетілді",
    },
    "Показано 0 проектов": {
      en: "0 projects shown",
      kk: "0 жоба көрсетілді",
    },
    "Проект будет переведен в завершенный статус, а итоговые оценки зафиксируются в карточке проекта.": {
      en: "The project will be moved to completed status, and the final grades will be fixed in the project card.",
      kk: "Жоба аяқталған күйге ауыстырылады, ал қорытынды бағалар жоба картасында бекітіледі.",
    },
    "Прогресс закрыт до ACTIVE": {
      en: "Progress is closed until ACTIVE",
      kk: "Прогресс ACTIVE кезеңіне дейін жабық",
    },
    "Email подтвержден. Войдите заново для обновления сессии.": {
      en: "Email confirmed. Sign in again to refresh the session.",
      kk: "Email расталды. Сессияны жаңарту үшін қайта кіріңіз.",
    },
    "Email подтвержден. Войдите заново.": {
      en: "Email confirmed. Sign in again.",
      kk: "Email расталды. Қайта кіріңіз.",
    },
  };

  const TRANSLATION_INDEX = buildTranslationIndex();

  const state = {
    lang: normalizeLanguage(readStorage(STORAGE_LANG)) || "ru",
  };

  let observer = null;
  let dock = null;
  let dockHost = null;

  function readStorage(key) {
    try {
      return window.localStorage.getItem(key);
    } catch (_) {
      return null;
    }
  }

  function writeStorage(key, value) {
    try {
      window.localStorage.setItem(key, value);
    } catch (_) {}
  }

  function normalizeLanguage(lang) {
    const value = String(lang || "").trim().toLowerCase();
    return LANGUAGES.includes(value) ? value : null;
  }

  function locale() {
    return LOCALES[state.lang] || LOCALES.ru;
  }

  function key(name, params) {
    const entry = UI_KEYS[name];
    const template = entry ? entry[state.lang] || entry.ru || name : name;
    return applyParams(template, params);
  }

  function t(source, params) {
    const normalized = normalizeInlineText(source);
    const translated = translateText(normalized);
    return applyParams(translated, params);
  }

  function applyParams(template, params) {
    return String(template || "").replace(/\{(\w+)\}/g, (_, name) => {
      if (!params || params[name] === undefined || params[name] === null) return "";
      return String(params[name]);
    });
  }

  function normalizeInlineText(value) {
    return String(value || "")
      .replace(/\s+/g, " ")
      .trim();
  }

  function normalizeNodeText(value, preformatted) {
    if (preformatted) return String(value || "").trim();
    return normalizeInlineText(value);
  }

  function isPreformattedNode(node) {
    const parent = node && node.parentElement;
    return Boolean(parent && parent.closest("pre, code, textarea"));
  }

  function translateExact(source) {
    const entry = TRANSLATION_INDEX.get(normalizeInlineText(source));
    if (!entry) return "";
    return entry[state.lang] || entry.ru || source;
  }

  function translatePattern(source) {
    let match = source.match(/^Привет,\s*(.+)!$/);
    if (match) {
      if (state.lang === "en") return `Hi, ${match[1]}!`;
      if (state.lang === "kk") return `Сәлем, ${match[1]}!`;
      return source;
    }

    match = source.match(/^(\d+)\s+активных$/);
    if (match) {
      const count = Number(match[1] || 0);
      if (state.lang === "en") return `${count} active`;
      if (state.lang === "kk") return `${count} белсенді`;
      return source;
    }

    match = source.match(/^Создано:\s*(.+)$/);
    if (match) {
      if (state.lang === "en") return `Created: ${match[1]}`;
      if (state.lang === "kk") return `Құрылған: ${match[1]}`;
      return source;
    }

    match = source.match(/^(?:Проверено|Тексерілген|Reviewed):\s*(\d+)\/(\d+)\.\s+(?:Итоговый балл|Қорытынды балл|Final score):\s*([0-9.]+)\/5\.0\s+\((\d+)%\)\.(.*)$/);
    if (match) {
      const suffixRaw = String(match[5] || "").trim();
      const suffix = suffixRaw ? ` ${translateText(suffixRaw)}` : "";
      if (state.lang === "en") return `Reviewed: ${match[1]}/${match[2]}. Final score: ${match[3]}/5.0 (${match[4]}%).${suffix}`;
      if (state.lang === "kk") return `Тексерілген: ${match[1]}/${match[2]}. Қорытынды балл: ${match[3]}/5.0 (${match[4]}%).${suffix}`;
      return `Проверено: ${match[1]}/${match[2]}. Итоговый балл: ${match[3]}/5.0 (${match[4]}%).${suffix}`;
    }

    match = source.match(/^Проверено:\s*(.+)$/);
    if (match) {
      if (state.lang === "en") return `Reviewed: ${match[1]}`;
      if (state.lang === "kk") return `Тексерілген: ${match[1]}`;
      return source;
    }

    match = source.match(/^Комментарий:\s*(.+)$/);
    if (match) {
      if (state.lang === "en") return `Comment: ${match[1]}`;
      if (state.lang === "kk") return `Пікір: ${match[1]}`;
      return source;
    }

    match = source.match(/^Кафедра:\s*(.+)$/);
    if (match) {
      if (state.lang === "en") return `Department: ${match[1]}`;
      if (state.lang === "kk") return `Кафедра: ${match[1]}`;
      return source;
    }

    match = source.match(/^Группа:\s*(.+)$/);
    if (match) {
      if (state.lang === "en") return `Group: ${match[1]}`;
      if (state.lang === "kk") return `Топ: ${match[1]}`;
      return source;
    }

    match = source.match(/^Код группы:\s*(.+)$/);
    if (match) {
      if (state.lang === "en") return `Group code: ${match[1]}`;
      if (state.lang === "kk") return `Топ коды: ${match[1]}`;
      return source;
    }

    match = source.match(/^Профиль\s+(\d+)%$/);
    if (match) {
      if (state.lang === "en") return `Profile ${match[1]}%`;
      if (state.lang === "kk") return `Профиль ${match[1]}%`;
      return source;
    }

    match = source.match(/^(\d+)\s+мин\.\s+назад$/);
    if (match) {
      const count = Number(match[1] || 0);
      if (state.lang === "en") return `${count} min ago`;
      if (state.lang === "kk") return `${count} мин бұрын`;
      return source;
    }

    match = source.match(/^(\d+)\s+ч\.\s+назад$/);
    if (match) {
      const count = Number(match[1] || 0);
      if (state.lang === "en") return `${count} hr ago`;
      if (state.lang === "kk") return `${count} сағ бұрын`;
      return source;
    }

    match = source.match(/^(\d+)\s+дн\.\s+назад$/);
    if (match) {
      const count = Number(match[1] || 0);
      if (state.lang === "en") return `${count} d ago`;
      if (state.lang === "kk") return `${count} күн бұрын`;
      return source;
    }

    match = source.match(/^(\d+)\s+мин\s+назад$/);
    if (match) {
      const count = Number(match[1] || 0);
      if (state.lang === "en") return `${count} min ago`;
      if (state.lang === "kk") return `${count} мин бұрын`;
      return `${count} мин назад`;
    }

    match = source.match(/^(\d+)\s+ч\s+назад$/);
    if (match) {
      const count = Number(match[1] || 0);
      if (state.lang === "en") return `${count} hr ago`;
      if (state.lang === "kk") return `${count} сағ бұрын`;
      return `${count} ч назад`;
    }

    match = source.match(/^(\d+)\s+дн\s+назад$/);
    if (match) {
      const count = Number(match[1] || 0);
      if (state.lang === "en") return `${count} d ago`;
      if (state.lang === "kk") return `${count} күн бұрын`;
      return `${count} дн назад`;
    }

    match = source.match(/^👥\s*(\d+)\/(\d+)\s+участников$/);
    if (match) {
      if (state.lang === "en") return `👥 ${match[1]}/${match[2]} members`;
      if (state.lang === "kk") return `👥 ${match[1]}/${match[2]} қатысушы`;
      return source;
    }

    match = source.match(/^Не удалось загрузить настройки:\s*(.+)$/);
    if (match) {
      if (state.lang === "en") return `Failed to load settings: ${match[1]}`;
      if (state.lang === "kk") return `Баптауларды жүктеу мүмкін болмады: ${match[1]}`;
      return source;
    }

    match = source.match(/^Вижу проектов в faculty-контуре:\s*(\d+)\.$/);
    if (match) {
      if (state.lang === "en") return `Projects visible in the faculty scope: ${match[1]}.`;
      if (state.lang === "kk") return `Факультет аясында көрінетін жобалар: ${match[1]}.`;
      return source;
    }

    match = source.match(/^Последнее обновление:\s*(.+)$/);
    if (match) {
      if (state.lang === "en") return `Last updated: ${match[1]}`;
      if (state.lang === "kk") return `Соңғы жаңарту: ${match[1]}`;
      return source;
    }

    match = source.match(/^Last updated:\s*(.+)$/);
    if (match) {
      if (state.lang === "ru") return `Последнее обновление: ${match[1]}`;
      if (state.lang === "kk") return `Соңғы жаңарту: ${match[1]}`;
      return source;
    }

    match = source.match(/^Статус проекта:\s*(.+)$/);
    if (match) {
      if (state.lang === "en") return `Project status: ${match[1]}`;
      if (state.lang === "kk") return `Жоба күйі: ${match[1]}`;
      return source;
    }

    match = source.match(/^Project status:\s*(.+)$/);
    if (match) {
      if (state.lang === "ru") return `Статус проекта: ${match[1]}`;
      if (state.lang === "kk") return `Жоба күйі: ${match[1]}`;
      return source;
    }

    match = source.match(/^Обновлено\s+(.+)$/);
    if (match) {
      if (state.lang === "en") return `Updated ${match[1]}`;
      if (state.lang === "kk") return `${match[1]} жаңартылды`;
      return source;
    }

    match = source.match(/^Updated\s+(.+)$/);
    if (match) {
      if (state.lang === "ru") return `Обновлено ${match[1]}`;
      if (state.lang === "kk") return `${match[1]} жаңартылды`;
      return source;
    }

    match = source.match(/^Чтение\s+(\d+)\s+мин$/);
    if (match) {
      if (state.lang === "en") return `${match[1]} min read`;
      if (state.lang === "kk") return `${match[1]} мин оқу`;
      return source;
    }

    match = source.match(/^(\d+)\s+мин\s+read$/);
    if (match) {
      if (state.lang === "ru") return `Чтение ${match[1]} мин`;
      if (state.lang === "kk") return `${match[1]} мин оқу`;
      return source;
    }

    match = source.match(/^(.+)\s+—\s+IDSAI База знаний$/);
    if (match) {
      if (state.lang === "en") return `${match[1]} — IDSAI Knowledge Base`;
      if (state.lang === "kk") return `${match[1]} — IDSAI Білім қоры`;
      return source;
    }

    match = source.match(/^(.+)\s+—\s+IDSAI Knowledge Base$/);
    if (match) {
      if (state.lang === "ru") return `${match[1]} — IDSAI База знаний`;
      if (state.lang === "kk") return `${match[1]} — IDSAI Білім қоры`;
      return source;
    }

    match = source.match(/^Сейчас:\s*(.+)$/);
    if (match) {
      const label = translateText(match[1]);
      if (state.lang === "en") return `Current: ${label}`;
      if (state.lang === "kk") return `Қазір: ${label}`;
      return source;
    }

    match = source.match(/^Current:\s*(.+)$/);
    if (match) {
      const label = translateText(match[1]);
      if (state.lang === "ru") return `Сейчас: ${label}`;
      if (state.lang === "kk") return `Қазір: ${label}`;
      return source;
    }

    match = source.match(/^Қазір:\s*(.+)$/);
    if (match) {
      const label = translateText(match[1]);
      if (state.lang === "ru") return `Сейчас: ${label}`;
      if (state.lang === "en") return `Current: ${label}`;
      return source;
    }

    match = source.match(/^Дальше:\s*(.+)$/);
    if (match) {
      const label = translateText(match[1]);
      if (state.lang === "en") return `Next: ${label}`;
      if (state.lang === "kk") return `Келесі: ${label}`;
      return source;
    }

    match = source.match(/^Next:\s*(.+)$/);
    if (match) {
      const label = translateText(match[1]);
      if (state.lang === "ru") return `Дальше: ${label}`;
      if (state.lang === "kk") return `Келесі: ${label}`;
      return source;
    }

    match = source.match(/^Келесі:\s*(.+)$/);
    if (match) {
      const label = translateText(match[1]);
      if (state.lang === "ru") return `Дальше: ${label}`;
      if (state.lang === "en") return `Next: ${label}`;
      return source;
    }

    match = source.match(/^До запуска осталось:\s*(.+)\.$/);
    if (match) {
      const blockers = String(match[1] || "")
        .split(/\s*,\s*/)
        .filter(Boolean)
        .map((item) => translateText(item))
        .join(", ");
      if (state.lang === "en") return `Remaining before launch: ${blockers}.`;
      if (state.lang === "kk") return `Іске қосуға дейін қалғаны: ${blockers}.`;
      return source;
    }

    match = source.match(/^добрать команду\s+(\d+)\/(\d+)$/);
    if (match) {
      if (state.lang === "en") return `complete the team ${match[1]}/${match[2]}`;
      if (state.lang === "kk") return `команданы толықтыру ${match[1]}/${match[2]}`;
      return source;
    }

    match = source.match(/^Сейчас занято\s+(\d+)\s+из\s+(\d+)\s+мест\.$/);
    if (match) {
      if (state.lang === "en") return `${match[1]} of ${match[2]} seats are filled now.`;
      if (state.lang === "kk") return `Қазір ${match[1]} орынның ${match[2]}-і толды.`;
      return source;
    }

    match = source.match(/^Преподаватель уже настроил\s+(\d+)\s+критериев\.$/);
    if (match) {
      if (state.lang === "en") return `The professor has already configured ${match[1]} criteria.`;
      if (state.lang === "kk") return `Оқытушы ${match[1]} критерийді баптап қойды.`;
      return source;
    }

    match = source.match(/^В проекте уже\s+(\d+)\s+задач\.$/);
    if (match) {
      if (state.lang === "en") return `The project already has ${match[1]} tasks.`;
      if (state.lang === "kk") return `Жобада қазірдің өзінде ${match[1]} тапсырма бар.`;
      return source;
    }

    match = source.match(/^Готово\s+(\d+)\s+из\s+(\d+)\.$/);
    if (match) {
      if (state.lang === "en") return `${match[1]} of ${match[2]} completed.`;
      if (state.lang === "kk") return `${match[2]} тапсырманың ${match[1]}-і дайын.`;
      return source;
    }

    match = source.match(/^Всего критериев:\s*(\d+)\.$/);
    if (match) {
      if (state.lang === "en") return `Total criteria: ${match[1]}.`;
      if (state.lang === "kk") return `Барлық критерий саны: ${match[1]}.`;
      return source;
    }

    match = source.match(/^Сейчас проверено\s+(\d+)\s+из\s+(\d+)\.$/);
    if (match) {
      if (state.lang === "en") return `${match[1]} of ${match[2]} checked so far.`;
      if (state.lang === "kk") return `Қазір ${match[2]} критерийдің ${match[1]}-і тексерілді.`;
      return source;
    }

    match = source.match(/^Сейчас в проекте\s+(\d+)\s+ролей\.$/);
    if (match) {
      if (state.lang === "en") return `The project currently has ${match[1]} roles.`;
      if (state.lang === "kk") return `Жобада қазір ${match[1]} рөл бар.`;
      return source;
    }

    match = source.match(/^Сейчас в проекте\s+(\d+)\s+критериев\.$/);
    if (match) {
      if (state.lang === "en") return `The project currently has ${match[1]} criteria.`;
      if (state.lang === "kk") return `Жобада қазір ${match[1]} критерий бар.`;
      return source;
    }

    match = source.match(/^Пока проверено\s+(\d+)\s+из\s+(\d+)\s+критериев\.$/);
    if (match) {
      if (state.lang === "en") return `${match[1]} of ${match[2]} criteria have been checked so far.`;
      if (state.lang === "kk") return `Қазірге дейін ${match[2]} критерийдің ${match[1]}-і тексерілді.`;
      return source;
    }

    match = source.match(/^До сдачи осталось закрыть\s+(\d+)\s+задач\.$/);
    if (match) {
      if (state.lang === "en") return `${match[1]} tasks still need to be closed before submission.`;
      if (state.lang === "kk") return `Тапсыруға дейін ${match[1]} тапсырманы жабу керек.`;
      return source;
    }

    match = source.match(/^Есть замечания по\s+(\d+)\s+критериям\.$/);
    if (match) {
      if (state.lang === "en") return `There are remarks on ${match[1]} criteria.`;
      if (state.lang === "kk") return `${match[1]} критерий бойынша ескертулер бар.`;
      return source;
    }

    match = source.match(/^Статус приглашения:\s*(.+)\.$/);
    if (match) {
      const label = translateText(match[1]);
      if (state.lang === "en") return `Invite status: ${label}.`;
      if (state.lang === "kk") return `Шақыру күйі: ${label}.`;
      return source;
    }

    match = source.match(/^Прогресс открыт · (\d+)%(?: · просрочено (\d+))?$/);
    if (match) {
      if (state.lang === "en") return match[2] ? `Progress open · ${match[1]}% · overdue ${match[2]}` : `Progress open · ${match[1]}%`;
      if (state.lang === "kk") return match[2] ? `Прогресс ашық · ${match[1]}% · мерзімі өткен ${match[2]}` : `Прогресс ашық · ${match[1]}%`;
      return source;
    }

    match = source.match(/^Задачи:\s*(\d+)\s+todo\s*\/\s*(\d+)\s+in progress\s*\/\s*(\d+)\s+done$/);
    if (match) {
      if (state.lang === "en") return `Tasks: ${match[1]} todo / ${match[2]} in progress / ${match[3]} done`;
      if (state.lang === "kk") return `Тапсырмалар: ${match[1]} todo / ${match[2]} орындалуда / ${match[3]} дайын`;
      return `Задачи: ${match[1]} к выполнению / ${match[2]} в работе / ${match[3]} готово`;
    }

    match = source.match(/^Участники:\s*(\d+)\s+активных$/);
    if (match) {
      if (state.lang === "en") return `Members: ${match[1]} active`;
      if (state.lang === "kk") return `Қатысушылар: ${match[1]} белсенді`;
      return source;
    }

    match = source.match(/^Прогресс закрыт до ACTIVE$/);
    if (match) {
      if (state.lang === "en") return "Progress is closed until ACTIVE";
      if (state.lang === "kk") return "Прогресс ACTIVE кезеңіне дейін жабық";
      return source;
    }

    match = source.match(/^Показано\s+(\d+)\s+пользователей$/);
    if (match) {
      if (state.lang === "en") return `${match[1]} users shown`;
      if (state.lang === "kk") return `${match[1]} пайдаланушы көрсетілді`;
      return source;
    }

    match = source.match(/^Показано\s+(\d+)\s+проектов$/);
    if (match) {
      if (state.lang === "en") return `${match[1]} projects shown`;
      if (state.lang === "kk") return `${match[1]} жоба көрсетілді`;
      return source;
    }

    match = source.match(/^Кафедр:\s*(\d+)$/);
    if (match) {
      if (state.lang === "en") return `Departments: ${match[1]}`;
      if (state.lang === "kk") return `Кафедралар: ${match[1]}`;
      return source;
    }

    match = source.match(/^(\d+)\s+групп$/);
    if (match) {
      if (state.lang === "en") return `${match[1]} groups`;
      if (state.lang === "kk") return `${match[1]} топ`;
      return source;
    }

    match = source.match(/^(\d+)\s+студентов$/);
    if (match) {
      if (state.lang === "en") return `${match[1]} students`;
      if (state.lang === "kk") return `${match[1]} студент`;
      return source;
    }

    match = source.match(/^От:\s*(.+)\s+·\s+(.+)$/);
    if (match) {
      if (state.lang === "en") return `From: ${match[1]} · ${match[2]}`;
      if (state.lang === "kk") return `Кімнен: ${match[1]} · ${match[2]}`;
      return source;
    }

    match = source.match(/^·\s*Ответ:\s*(.+)$/);
    if (match) {
      if (state.lang === "en") return `· Response: ${match[1]}`;
      if (state.lang === "kk") return `· Жауап: ${match[1]}`;
      return source;
    }

    match = source.match(/^Мы отправили 6-значный код на\s+(.+)\.\s+Введите его и задайте новый пароль\.$/);
    if (match) {
      if (state.lang === "en") return `We sent a 6-digit code to ${match[1]}. Enter it and set a new password.`;
      if (state.lang === "kk") return `Біз ${match[1]} адресіне 6 таңбалы код жібердік. Оны енгізіп, жаңа құпиясөз орнатыңыз.`;
      return source;
    }

    match = source.match(/^Пароль должен быть не короче\s+(\d+)\s+символов\.$/);
    if (match) {
      if (state.lang === "en") return `The password must be at least ${match[1]} characters long.`;
      if (state.lang === "kk") return `Құпиясөз кемінде ${match[1]} таңбадан тұруы керек.`;
      return source;
    }

    match = source.match(/^Не удалось загрузить статью\.\s*(.+)$/);
    if (match) {
      if (state.lang === "en") return `Failed to load article. ${match[1]}`;
      if (state.lang === "kk") return `Мақаланы жүктеу мүмкін болмады. ${match[1]}`;
      return source;
    }

    match = source.match(/^Введите новый пароль для пользователя\s+"(.+)"\.$/);
    if (match) {
      if (state.lang === "en") return `Enter a new password for user "${match[1]}".`;
      if (state.lang === "kk") return `"${match[1]}" пайдаланушысы үшін жаңа құпиясөз енгізіңіз.`;
      return source;
    }

    match = source.match(/^Пользователь\s+"(.+)"\s+будет удален без возможности восстановления\.$/);
    if (match) {
      if (state.lang === "en") return `User "${match[1]}" will be deleted permanently.`;
      if (state.lang === "kk") return `"${match[1]}" пайдаланушысы қалпына келтірусіз жойылады.`;
      return source;
    }

    match = source.match(/^Проект\s+"(.+)"\s+будет удален без возможности восстановления\.$/);
    if (match) {
      if (state.lang === "en") return `Project "${match[1]}" will be deleted permanently.`;
      if (state.lang === "kk") return `"${match[1]}" жобасы қалпына келтірусіз жойылады.`;
      return source;
    }

    match = source.match(/^Ошибка загрузки:\s*(.+)$/);
    if (match) {
      if (state.lang === "en") return `Loading error: ${match[1]}`;
      if (state.lang === "kk") return `Жүктеу қатесі: ${match[1]}`;
      return source;
    }

    match = source.match(/^Ошибка наблюдения проекта:\s*(.+)$/);
    if (match) {
      if (state.lang === "en") return `Project observation error: ${match[1]}`;
      if (state.lang === "kk") return `Жобаны бақылау қатесі: ${match[1]}`;
      return source;
    }

    match = source.match(/^Ошибка сброса пароля:\s*(.+)$/);
    if (match) {
      if (state.lang === "en") return `Password reset error: ${match[1]}`;
      if (state.lang === "kk") return `Құпиясөзді қалпына келтіру қатесі: ${match[1]}`;
      return source;
    }

    match = source.match(/^Ошибка смены роли:\s*(.+)$/);
    if (match) {
      if (state.lang === "en") return `Role change error: ${match[1]}`;
      if (state.lang === "kk") return `Рөлді ауыстыру қатесі: ${match[1]}`;
      return source;
    }

    match = source.match(/^Ошибка смены статуса:\s*(.+)$/);
    if (match) {
      if (state.lang === "en") return `Status change error: ${match[1]}`;
      if (state.lang === "kk") return `Күйді өзгерту қатесі: ${match[1]}`;
      return source;
    }

    match = source.match(/^Ошибка удаления пользователя:\s*(.+)$/);
    if (match) {
      if (state.lang === "en") return `User deletion error: ${match[1]}`;
      if (state.lang === "kk") return `Пайдаланушыны жою қатесі: ${match[1]}`;
      return source;
    }

    match = source.match(/^Ошибка удаления проекта:\s*(.+)$/);
    if (match) {
      if (state.lang === "en") return `Project deletion error: ${match[1]}`;
      if (state.lang === "kk") return `Жобаны жою қатесі: ${match[1]}`;
      return source;
    }

    match = source.match(/^Ошибка сохранения прав:\s*(.+)$/);
    if (match) {
      if (state.lang === "en") return `Permissions save error: ${match[1]}`;
      if (state.lang === "kk") return `Рұқсаттарды сақтау қатесі: ${match[1]}`;
      return source;
    }

    match = source.match(/^Не удалось загрузить профиль:\s*(.+)$/);
    if (match) {
      if (state.lang === "en") return `Failed to load profile: ${match[1]}`;
      if (state.lang === "kk") return `Профильді жүктеу мүмкін болмады: ${match[1]}`;
      return source;
    }

    match = source.match(/^Не удалось загрузить проект:\s*(.+)$/);
    if (match) {
      if (state.lang === "en") return `Failed to load project: ${match[1]}`;
      if (state.lang === "kk") return `Жобаны жүктеу мүмкін болмады: ${match[1]}`;
      return source;
    }

    match = source.match(/^Недостаточно условий:\s*роли\s+(\d+)\/(\d+),\s*критерии\s+(\d+)\.$/);
    if (match) {
      if (state.lang === "en") return `Not enough conditions: roles ${match[1]}/${match[2]}, criteria ${match[3]}.`;
      if (state.lang === "kk") return `Шарттар жеткіліксіз: рөлдер ${match[1]}/${match[2]}, критерийлер ${match[3]}.`;
      return source;
    }

    match = source.match(/^Активных:\s*(.+)$/);
    if (match) {
      if (state.lang === "en") return `Active: ${match[1]}`;
      if (state.lang === "kk") return `Белсенді: ${match[1]}`;
      return source;
    }

    match = source.match(/^Участников:\s*(.+)$/);
    if (match) {
      if (state.lang === "en") return `Members: ${match[1]}`;
      if (state.lang === "kk") return `Қатысушылар: ${match[1]}`;
      return source;
    }

    match = source.match(/^Задач:\s*(.+)$/);
    if (match) {
      if (state.lang === "en") return `Tasks: ${match[1]}`;
      if (state.lang === "kk") return `Тапсырмалар: ${match[1]}`;
      return source;
    }

    match = source.match(/^Критериев:\s*(.+)$/);
    if (match) {
      if (state.lang === "en") return `Criteria: ${match[1]}`;
      if (state.lang === "kk") return `Критерийлер: ${match[1]}`;
      return source;
    }

    match = source.match(/^Ожидается:\s*(.+)$/);
    if (match) {
      const label = translateText(match[1]);
      if (state.lang === "en") return `Expected: ${label}`;
      if (state.lang === "kk") return `Күтілуде: ${label}`;
      return source;
    }

    match = source.match(/^Нельзя отправить на оценивание:\s*(.+)$/);
    if (match) {
      const reasons = String(match[1] || "")
        .split(/\s*,\s*/)
        .filter(Boolean)
        .map((item) => translateText(item))
        .join(", ");
      if (state.lang === "en") return `Cannot submit for grading: ${reasons}`;
      if (state.lang === "kk") return `Бағалауға жіберу мүмкін емес: ${reasons}`;
      return source;
    }

    match = source.match(/^выполнено задач\s+(\d+)\/(\d+)$/);
    if (match) {
      if (state.lang === "en") return `completed tasks ${match[1]}/${match[2]}`;
      if (state.lang === "kk") return `орындалған тапсырма ${match[1]}/${match[2]}`;
      return source;
    }

    match = source.match(/^(\d+)\s+критерий$/);
    if (match) {
      if (state.lang === "en") return `${match[1]} criterion`;
      if (state.lang === "kk") return `${match[1]} критерий`;
      return source;
    }

    match = source.match(/^(\d+)\s+критерия$/);
    if (match) {
      if (state.lang === "en") return `${match[1]} criteria`;
      if (state.lang === "kk") return `${match[1]} критерий`;
      return source;
    }

    match = source.match(/^(\d+)\s+критериев$/);
    if (match) {
      if (state.lang === "en") return `${match[1]} criteria`;
      if (state.lang === "kk") return `${match[1]} критерий`;
      return source;
    }

    match = source.match(/^Критерий\s+(\d+)$/);
    if (match) {
      if (state.lang === "en") return `Criterion ${match[1]}`;
      if (state.lang === "kk") return `${match[1]}-критерий`;
      return source;
    }

    match = source.match(/^Пересдачи:\s*(\d+)$/);
    if (match) {
      if (state.lang === "en") return `Resubmissions: ${match[1]}`;
      if (state.lang === "kk") return `Қайта тапсырулар: ${match[1]}`;
      return source;
    }

    match = source.match(/^Уровень\s+(\d+)%$/);
    if (match) {
      if (state.lang === "en") return `Level ${match[1]}%`;
      if (state.lang === "kk") return `Деңгей ${match[1]}%`;
      return source;
    }

    match = source.match(/^Команда:\s*(.+)$/);
    if (match) {
      if (state.lang === "en") return `Team: ${match[1]}`;
      if (state.lang === "kk") return `Команда: ${match[1]}`;
      return source;
    }

    match = source.match(/^Статус:\s*(.+)$/);
    if (match) {
      const label = translateText(match[1]);
      if (state.lang === "en") return `Status: ${label}`;
      if (state.lang === "kk") return `Күйі: ${label}`;
      return source;
    }

    match = source.match(/^Вес\s+(.+)$/);
    if (match) {
      if (state.lang === "en") return `Weight ${match[1]}`;
      if (state.lang === "kk") return `Салмақ ${match[1]}`;
      return source;
    }

    match = source.match(/^Срок:\s*(.+)$/);
    if (match) {
      if (state.lang === "en") return `Due: ${match[1]}`;
      if (state.lang === "kk") return `Мерзім: ${match[1]}`;
      return source;
    }

    match = source.match(/^Исполнитель:\s*(.+)$/);
    if (match) {
      if (state.lang === "en") return `Assignee: ${match[1]}`;
      if (state.lang === "kk") return `Орындаушы: ${match[1]}`;
      return source;
    }

    match = source.match(/^Роль:\s*(.+)$/);
    if (match) {
      if (state.lang === "en") return `Role: ${match[1]}`;
      if (state.lang === "kk") return `Рөл: ${match[1]}`;
      return source;
    }

    match = source.match(/^Текущий статус:\s*(.+)\.?$/);
    if (match) {
      const label = translateText(match[1]);
      if (state.lang === "en") return `Current status: ${label}`;
      if (state.lang === "kk") return `Ағымдағы күйі: ${label}`;
      return source;
    }

    match = source.match(/^Удалить\s+(.+)$/);
    if (match) {
      if (state.lang === "en") return `Delete ${match[1]}`;
      if (state.lang === "kk") return `${match[1]} жою`;
      return source;
    }

    match = source.match(/^(\d+)\s+аккаунтов сейчас выключены или ограничены\. Это первое место, где админ может быстро убрать блокер\.$/);
    if (match) {
      if (state.lang === "en") return `${match[1]} accounts are currently disabled or limited. This is the first place where an admin can quickly remove a blocker.`;
      if (state.lang === "kk") return `Қазір ${match[1]} аккаунт өшірілген немесе шектелген. Әкімші блокерді тез алып тастай алатын бірінші орын осы.`;
      return source;
    }

    match = source.match(/^(\d+)\s+проектов находятся в REVIEW или GRADING\. Здесь чаще всего нужны точечные решения по статусам\.$/);
    if (match) {
      if (state.lang === "en") return `${match[1]} projects are in REVIEW or GRADING. Precise status decisions are usually needed here.`;
      if (state.lang === "kk") return `${match[1]} жоба REVIEW немесе GRADING күйінде. Мұнда көбіне күй бойынша нақты шешімдер қажет.`;
      return source;
    }

    match = source.match(/^(\d+)\s+проектов пока не вышли в стабильную рабочую фазу\. Проверьте, не застрял ли запуск или набор\.$/);
    if (match) {
      if (state.lang === "en") return `${match[1]} projects have not yet reached a stable work phase. Check whether launch or recruitment is stuck.`;
      if (state.lang === "kk") return `${match[1]} жоба әлі тұрақты жұмыс кезеңіне шыққан жоқ. Іске қосу немесе іріктеу тұрып қалмағанын тексеріңіз.`;
      return source;
    }

    match = source.match(/^(.+)\s+·\s+штраф\s+(\d+)%$/);
    if (match) {
      if (state.lang === "en") return `${match[1]} · penalty ${match[2]}%`;
      if (state.lang === "kk") return `${match[1]} · айыппұл ${match[2]}%`;
      return source;
    }

    match = source.match(/^·\s*Пересдачи:\s*(\d+)$/);
    if (match) {
      if (state.lang === "en") return `· Resubmissions: ${match[1]}`;
      if (state.lang === "kk") return `· Қайта тапсырулар: ${match[1]}`;
      return source;
    }

    match = source.match(/^Автор:\s*(.+)\s+·\s+Обновлен:\s*(.+)\s+·\s+Ревью:\s*(.+)$/);
    if (match) {
      if (state.lang === "en") return `Author: ${match[1]} · Updated: ${match[2]} · Review: ${match[3]}`;
      if (state.lang === "kk") return `Автор: ${match[1]} · Жаңартылды: ${match[2]} · Ревью: ${match[3]}`;
      return source;
    }

    match = source.match(/^Оценивание завершено\. Итог закреплен в проекте с учетом штрафа\s+(\d+)%\.$/);
    if (match) {
      if (state.lang === "en") return `Grading is complete. The final result is fixed in the project with a ${match[1]}% penalty.`;
      if (state.lang === "kk") return `Бағалау аяқталды. Қорытынды жобаға ${match[1]}% айыппұлмен бекітілді.`;
      return source;
    }

    match = source.match(/^Проект будет переведен в завершенный статус, а итоговая оценка зафиксируется с учетом штрафа\s+(\d+)%\s+за пересдачу\.$/);
    if (match) {
      if (state.lang === "en") return `The project will move to completed status, and the final grade will be fixed with a ${match[1]}% resubmission penalty.`;
      if (state.lang === "kk") return `Жоба аяқталған күйге ауысады, ал қорытынды баға қайта тапсыру үшін ${match[1]}% айыппұлмен бекітіледі.`;
      return source;
    }

    match = source.match(/^Проект будет переведен в завершенный статус, а итоговые оценки зафиксируются в карточке проекта\.$/);
    if (match) {
      if (state.lang === "en") return "The project will move to completed status, and the final grades will be fixed in the project card.";
      if (state.lang === "kk") return "Жоба аяқталған күйге ауысады, ал қорытынды бағалар жоба картасында бекітіледі.";
      return source;
    }

    match = source.match(/^Проект вернется в ACTIVE\. После повторной сдачи итоговая оценка будет включать штраф\s+(\d+)%\s+\(всего пересдач:\s*(\d+)\)\.$/);
    if (match) {
      if (state.lang === "en") return `The project will return to ACTIVE. After the repeated submission, the final grade will include a ${match[1]}% penalty (total resubmissions: ${match[2]}).`;
      if (state.lang === "kk") return `Жоба ACTIVE күйіне оралады. Қайта тапсырғаннан кейін қорытынды баға ${match[1]}% айыппұлды қамтиды (барлық қайта тапсыру саны: ${match[2]}).`;
      return source;
    }

    match = source.match(/^Проект возвращен на пересдачу\. При следующем финале будет учтен штраф\s+(\d+)%\.$/);
    if (match) {
      if (state.lang === "en") return `The project has been returned for resubmission. A ${match[1]}% penalty will be applied at the next final review.`;
      if (state.lang === "kk") return `Жоба қайта тапсыруға қайтарылды. Келесі финалда ${match[1]}% айыппұл ескеріледі.`;
      return source;
    }

    match = source.match(/^Проект отправлен на пересдачу\. Текущий накопленный штраф:\s*(\d+)%\.$/);
    if (match) {
      if (state.lang === "en") return `The project has been sent for resubmission. Current accumulated penalty: ${match[1]}%.`;
      if (state.lang === "kk") return `Жоба қайта тапсыруға жіберілді. Ағымдағы жинақталған айыппұл: ${match[1]}%.`;
      return source;
    }

    match = source.match(/^Сейчас проект находится в статусе\s+(.+)\.\s+Редактирование оценок на этом этапе ограничено\.$/);
    if (match) {
      const label = translateText(match[1]);
      if (state.lang === "en") return `The project is currently in status ${label}. Editing grades is limited at this stage.`;
      if (state.lang === "kk") return `Жоба қазір ${label} күйінде. Бұл кезеңде бағаларды өңдеу шектеулі.`;
      return source;
    }

    match = source.match(/^Активных участников\s+(\d+)\s+из\s+(\d+)\./);
    if (match) {
      if (state.lang === "en") return source.replace(/^Активных участников\s+\d+\s+из\s+\d+\./, `${match[1]} active members out of ${match[2]}.`);
      if (state.lang === "kk") return source.replace(/^Активных участников\s+\d+\s+из\s+\d+\./, `${match[2]} қатысушының ${match[1]}-і белсенді.`);
      return source;
    }

    match = source.match(/^В проекте уже есть\s+(\d+)\s+критериев\./);
    if (match) {
      if (state.lang === "en") return source.replace(/^В проекте уже есть\s+\d+\s+критериев\./, `The project already has ${match[1]} criteria.`);
      if (state.lang === "kk") return source.replace(/^В проекте уже есть\s+\d+\s+критериев\./, `Жобада қазірдің өзінде ${match[1]} критерий бар.`);
      return source;
    }

    match = source.match(/^Задача\s+(.+)\s+взята в работу\.$/);
    if (match) {
      if (state.lang === "en") return `Task ${match[1]} has been claimed.`;
      if (state.lang === "kk") return `${match[1]} тапсырмасы жұмысқа алынды.`;
      return source;
    }

    match = source.match(/^Задача\s+(.+)\s+удалена\.$/);
    if (match) {
      if (state.lang === "en") return `Task ${match[1]} has been removed.`;
      if (state.lang === "kk") return `${match[1]} тапсырмасы жойылды.`;
      return source;
    }

    match = source.match(/^Задача «(.+)» будет удалена вместе с ее историей выполнения\.$/);
    if (match) {
      if (state.lang === "en") return `Task “${match[1]}” will be deleted together with its execution history.`;
      if (state.lang === "kk") return `«${match[1]}» тапсырмасы орындалу тарихымен бірге жойылады.`;
      return source;
    }

    match = source.match(/^Задача «(.+)» удалена\.$/);
    if (match) {
      if (state.lang === "en") return `Task “${match[1]}” has been deleted.`;
      if (state.lang === "kk") return `«${match[1]}» тапсырмасы жойылды.`;
      return source;
    }

    match = source.match(/^Заявка участника\s+(.+)\s+отклонена\.$/);
    if (match) {
      if (state.lang === "en") return `Participant application ${match[1]} was declined.`;
      if (state.lang === "kk") return `${match[1]} қатысушысының өтінімі қабылданбады.`;
      return source;
    }

    match = source.match(/^Исполнитель задачи\s+(.+)\s+обновлен\.$/);
    if (match) {
      if (state.lang === "en") return `Task assignee ${match[1]} updated.`;
      if (state.lang === "kk") return `${match[1]} тапсырмасының орындаушысы жаңартылды.`;
      return source;
    }

    match = source.match(/^Роль участника\s+(.+)\s+обновлена\.$/);
    if (match) {
      if (state.lang === "en") return `Member role ${match[1]} updated.`;
      if (state.lang === "kk") return `${match[1]} қатысушысының рөлі жаңартылды.`;
      return source;
    }

    match = source.match(/^Участник\s+(.+)\s+принят в команду\.$/);
    if (match) {
      if (state.lang === "en") return `Member ${match[1]} has been accepted to the team.`;
      if (state.lang === "kk") return `${match[1]} қатысушысы командаға қабылданды.`;
      return source;
    }

    match = source.match(/^Участник\s+(.+)\s+удален из проекта\.$/);
    if (match) {
      if (state.lang === "en") return `Member ${match[1]} has been removed from the project.`;
      if (state.lang === "kk") return `${match[1]} қатысушысы жобадан жойылды.`;
      return source;
    }

    match = source.match(/^Статус задачи\s+(.+)\s+обновлен\.$/);
    if (match) {
      if (state.lang === "en") return `Task status ${match[1]} updated.`;
      if (state.lang === "kk") return `${match[1]} тапсырмасының күйі жаңартылды.`;
      return source;
    }

    match = source.match(/^(?:Проверено|Тексерілген|Reviewed):\s*(\d+)\/(\d+)\.\s+(?:Итоговый балл|Қорытынды балл|Final score):\s*([0-9.]+)\/5\.0\s+\((\d+)%\)\.(.*)$/);
    if (match) {
      const suffixRaw = String(match[5] || "").trim();
      const suffix = suffixRaw ? ` ${translateText(suffixRaw)}` : "";
      if (state.lang === "en") return `Reviewed: ${match[1]}/${match[2]}. Final score: ${match[3]}/5.0 (${match[4]}%).${suffix}`;
      if (state.lang === "kk") return `Тексерілген: ${match[1]}/${match[2]}. Қорытынды балл: ${match[3]}/5.0 (${match[4]}%).${suffix}`;
      return `Проверено: ${match[1]}/${match[2]}. Итоговый балл: ${match[3]}/5.0 (${match[4]}%).${suffix}`;
    }

    match = source.match(/^Ревью по критериям сохранено\. Выполнено:\s*(\d+)\/(\d+)\.(.*)$/);
    if (match) {
      const suffix = match[3] ? ` ${match[3].trim()}` : "";
      if (state.lang === "en") return `Criteria review saved. Completed: ${match[1]}/${match[2]}.${suffix}`;
      if (state.lang === "kk") return `Критерийлер бойынша ревью сақталды. Орындалды: ${match[1]}/${match[2]}.${suffix}`;
      return source;
    }

    match = source.match(/^С учетом пересдачи:\s*-(\d+)%\.$/);
    if (match) {
      if (state.lang === "en") return `Including resubmission: -${match[1]}%.`;
      if (state.lang === "kk") return `Қайта тапсыруды ескере отырып: -${match[1]}%.`;
      return source;
    }

    match = source.match(/^Сейчас оценено\s+(\d+)\s+из\s+(\d+)\./);
    if (match) {
      if (state.lang === "en") return source.replace(/^Сейчас оценено\s+\d+\s+из\s+\d+\./, `${match[1]} of ${match[2]} graded so far.`);
      if (state.lang === "kk") return source.replace(/^Сейчас оценено\s+\d+\s+из\s+\d+\./, `Қазір ${match[2]} критерийдің ${match[1]}-і бағаланды.`);
      return source;
    }

    match = source.match(/^Сейчас осталось закрыть\s+(\d+)\s+задач\./);
    if (match) {
      if (state.lang === "en") return source.replace(/^Сейчас осталось закрыть\s+\d+\s+задач\./, `${match[1]} tasks still need to be closed now.`);
      if (state.lang === "kk") return source.replace(/^Сейчас осталось закрыть\s+\d+\s+задач\./, `Қазір ${match[1]} тапсырманы жабу қалды.`);
      return source;
    }

    match = source.match(/^Проект на проверке:\s*оценки выставлены по\s+(\d+)\/(\d+)\s+критериям\.$/);
    if (match) {
      if (state.lang === "en") return `The project is under review: grades have been given for ${match[1]}/${match[2]} criteria.`;
      if (state.lang === "kk") return `Жоба тексеруде: бағалар ${match[1]}/${match[2]} критерий бойынша қойылды.`;
      return source;
    }

    match = source.match(/^Роли\s+(.+)$/);
    if (match) {
      if (state.lang === "en") return `Roles ${match[1]}`;
      if (state.lang === "kk") return `Рөлдер ${match[1]}`;
      return source;
    }

    match = source.match(/^Roles\s+(.+)$/);
    if (match) {
      if (state.lang === "ru") return `Роли ${match[1]}`;
      if (state.lang === "kk") return `Рөлдер ${match[1]}`;
      return source;
    }

    match = source.match(/^Стек\s+(.+)$/);
    if (match) {
      if (state.lang === "en") return `Stack ${match[1]}`;
      if (state.lang === "kk") return `Стек ${match[1]}`;
      return source;
    }

    match = source.match(/^Stack\s+(.+)$/);
    if (match) {
      if (state.lang === "ru") return `Стек ${match[1]}`;
      if (state.lang === "kk") return `Стек ${match[1]}`;
      return source;
    }

    match = source.match(/^Задачи\s+(.+)$/);
    if (match) {
      if (state.lang === "en") return `Tasks ${match[1]}`;
      if (state.lang === "kk") return `Тапсырмалар ${match[1]}`;
      return source;
    }

    match = source.match(/^Tasks\s+(.+)$/);
    if (match) {
      if (state.lang === "ru") return `Задачи ${match[1]}`;
      if (state.lang === "kk") return `Тапсырмалар ${match[1]}`;
      return source;
    }

    match = source.match(/^Критерии\s+(.+)$/);
    if (match) {
      if (state.lang === "en") return `Criteria ${match[1]}`;
      if (state.lang === "kk") return `Критерийлер ${match[1]}`;
      return source;
    }

    match = source.match(/^Criteria\s+(.+)$/);
    if (match) {
      if (state.lang === "ru") return `Критерии ${match[1]}`;
      if (state.lang === "kk") return `Критерийлер ${match[1]}`;
      return source;
    }

    match = source.match(/^Оценено\s+(.+)$/);
    if (match) {
      if (state.lang === "en") return `Graded ${match[1]}`;
      if (state.lang === "kk") return `Бағаланғаны ${match[1]}`;
      return source;
    }

    match = source.match(/^Graded\s+(.+)$/);
    if (match) {
      if (state.lang === "ru") return `Оценено ${match[1]}`;
      if (state.lang === "kk") return `Бағаланғаны ${match[1]}`;
      return source;
    }

    match = source.match(/^Преподаватель\s+(.+)$/);
    if (match) {
      if (state.lang === "en") return `Professor ${match[1]}`;
      if (state.lang === "kk") return `Оқытушы ${match[1]}`;
      return source;
    }

    match = source.match(/^Профессор\s+(.+)$/);
    if (match) {
      if (state.lang === "en") return `Professor ${match[1]}`;
      if (state.lang === "kk") return `Профессор ${match[1]}`;
      return source;
    }

    match = source.match(/^Professor\s+(.+)$/);
    if (match) {
      if (state.lang === "ru") return `Преподаватель ${match[1]}`;
      if (state.lang === "kk") return `Оқытушы ${match[1]}`;
      return source;
    }

    return "";
  }

  function translateText(source) {
    if (!source) return source;
    return translateExact(source) || translatePattern(source) || source;
  }

  function compareStrings(a, b) {
    return String(a || "").localeCompare(String(b || ""), locale(), {
      sensitivity: "base",
      numeric: true,
    });
  }

  function formatDate(value, options = {}) {
    if (!value) return "";
    const date = new Date(value);
    if (Number.isNaN(date.getTime())) return String(value);
    return date.toLocaleDateString(locale(), options);
  }

  function formatDateTime(value, options = {}) {
    if (!value) return "";
    const date = new Date(value);
    if (Number.isNaN(date.getTime())) return String(value);
    return date.toLocaleString(locale(), options);
  }

  function formatTime(value, options = {}) {
    const date = value instanceof Date ? value : new Date(value || Date.now());
    if (Number.isNaN(date.getTime())) return String(value || "");
    return date.toLocaleTimeString(locale(), options);
  }

  function relativeTime(value) {
    if (!value) return translateText("только что");
    const ts = typeof value === "number" ? value : Date.parse(value);
    if (!Number.isFinite(ts)) return translateText("только что");
    const diffSec = Math.max(0, Math.floor((Date.now() - ts) / 1000));
    if (diffSec < 60) return translateText("только что");
    if (diffSec < 3600) {
      const count = Math.floor(diffSec / 60);
      if (state.lang === "en") return `${count} min ago`;
      if (state.lang === "kk") return `${count} мин бұрын`;
      return `${count} мин. назад`;
    }
    if (diffSec < 86400) {
      const count = Math.floor(diffSec / 3600);
      if (state.lang === "en") return `${count} hr ago`;
      if (state.lang === "kk") return `${count} сағ бұрын`;
      return `${count} ч. назад`;
    }
    if (diffSec < 604800) {
      const count = Math.floor(diffSec / 86400);
      if (state.lang === "en") return `${count} d ago`;
      if (state.lang === "kk") return `${count} күн бұрын`;
      return `${count} дн. назад`;
    }
    return formatDate(value, { day: "numeric", month: "short", year: "numeric" });
  }

  function shouldSkipNode(node) {
    if (!node) return false;
    if (node instanceof HTMLElement) {
      return node.matches(SKIP_SELECTOR) || Boolean(node.closest(SKIP_SELECTOR));
    }
    const parent = node.parentElement;
    return Boolean(parent && parent.closest(SKIP_SELECTOR));
  }

  function matchesStoredVariant(source, value) {
    const normalizedValue = normalizeInlineText(value);
    if (!normalizedValue || !source) return false;
    if (normalizedValue === normalizeInlineText(source)) return true;
    const bundle = TRANSLATION_INDEX.get(normalizeInlineText(source));
    if (!bundle) return false;
    return [bundle.ru, bundle.en, bundle.kk].some((variant) => normalizeInlineText(variant) === normalizedValue);
  }

  function rememberTextSource(node) {
    if (!(node instanceof Text) || shouldSkipNode(node)) return null;
    const raw = node.nodeValue || "";
    const preformatted = isPreformattedNode(node);
    const leading = raw.match(/^\s*/)?.[0] || "";
    const trailing = raw.match(/\s*$/)?.[0] || "";
    const source = normalizeNodeText(raw, preformatted);
    if (!source) return null;
    const current = textSources.get(node);
    if (current) {
      const normalizedRaw = normalizeNodeText(raw, preformatted);
      const rendered = current.preformatted ? current.rendered : normalizeNodeText(current.rendered, current.preformatted);
      if (current.raw === raw || normalizedRaw === rendered || matchesStoredVariant(current.source, normalizedRaw)) {
        current.raw = raw;
        current.leading = leading;
        current.trailing = trailing;
        current.preformatted = preformatted;
        return current;
      }
    }
    const stored = { source, leading, trailing, preformatted };
    stored.raw = raw;
    stored.rendered = raw;
    textSources.set(node, stored);
    return stored;
  }

  function applyTextNode(node) {
    const stored = rememberTextSource(node);
    if (!stored) return;
    const next = translateText(stored.source);
    const value = stored.preformatted ? next : `${stored.leading}${next}${stored.trailing}`;
    if (node.nodeValue !== value) {
      node.nodeValue = value;
    }
    stored.rendered = value;
  }

  function rememberAttrSource(el, attrName) {
    if (!(el instanceof HTMLElement) || shouldSkipNode(el)) return "";
    const current = el.getAttribute(attrName);
    if (!current) return "";
    let stored = attrSources.get(el);
    if (!stored) {
      stored = {};
      attrSources.set(el, stored);
    }
    const source = normalizeInlineText(current);
    const previous = stored[attrName];
    if (previous) {
      if (previous.raw === current || source === normalizeInlineText(previous.rendered) || matchesStoredVariant(previous.source, source)) {
        previous.raw = current;
        return previous.source;
      }
    }
    stored[attrName] = {
      source,
      raw: current,
      rendered: current,
    };
    return source;
  }

  function applyElementAttrs(el) {
    if (!(el instanceof HTMLElement) || shouldSkipNode(el)) return;
    TRANSLATED_ATTRS.forEach((attrName) => {
      if (!el.hasAttribute(attrName)) return;
      const source = rememberAttrSource(el, attrName);
      if (!source) return;
      const translated = translateText(source);
      if (el.getAttribute(attrName) !== translated) {
        el.setAttribute(attrName, translated);
      }
      const stored = attrSources.get(el)?.[attrName];
      if (stored) stored.rendered = translated;
    });
  }

  function apply(root = document.body) {
    if (!root) return;
    if (root instanceof Text) {
      applyTextNode(root);
      return;
    }

    if (root instanceof HTMLElement) {
      applyElementAttrs(root);
    }

    const walker = document.createTreeWalker(root, NodeFilter.SHOW_TEXT);
    let current = walker.nextNode();
    while (current) {
      applyTextNode(current);
      current = walker.nextNode();
    }

    if (root.querySelectorAll) {
      root.querySelectorAll("[placeholder],[aria-label],[title],[alt]").forEach((el) => {
        applyElementAttrs(el);
      });
    }
  }

  function ensureObserver() {
    if (observer || !document.body) return;
    observer = new MutationObserver((mutations) => {
      mutations.forEach((mutation) => {
        if (mutation.type === "characterData" && mutation.target instanceof Text) {
          applyTextNode(mutation.target);
          return;
        }

        if (mutation.type === "attributes" && mutation.target instanceof HTMLElement) {
          applyElementAttrs(mutation.target);
          return;
        }

        mutation.addedNodes.forEach((node) => {
          if (node instanceof Text) {
            applyTextNode(node);
            return;
          }
          if (node instanceof HTMLElement) {
            apply(node);
          }
        });
      });
    });

    observer.observe(document.body, {
      subtree: true,
      childList: true,
      characterData: true,
      attributes: true,
      attributeFilter: TRANSLATED_ATTRS,
    });
  }

  function getLanguage() {
    return state.lang;
  }

  function setLanguage(lang) {
    const next = normalizeLanguage(lang);
    if (!next || next === state.lang) return;
    state.lang = next;
    writeStorage(STORAGE_LANG, next);
    document.documentElement.lang = next;
    apply(document.documentElement);
    updateDock();
    window.dispatchEvent(new CustomEvent("idsai:languagechange", { detail: { lang: next } }));
  }

  function buildDock() {
    if (dock || !document.body) return;

    dock = document.createElement("aside");
    dock.className = "idsai-site-prefs idsai-site-prefs--fixed";
    dock.setAttribute("aria-live", "polite");
    dock.setAttribute("role", "group");
    dock.dataset.i18nSkip = "true";
    dock.dataset.idsaiPrefs = "dock";

    const langGroup = document.createElement("div");
    langGroup.className = "idsai-site-prefs__lang-group";
    langGroup.setAttribute("role", "group");
    langGroup.dataset.idsaiPrefs = "languageGroup";
    dock.appendChild(langGroup);

    LANGUAGES.forEach((lang) => {
      const button = document.createElement("button");
      button.type = "button";
      button.className = "idsai-site-prefs__flag";
      button.dataset.idsaiPrefsLang = lang;
      button.addEventListener("click", () => {
        setLanguage(lang);
      });

      const image = document.createElement("img");
      image.className = "idsai-site-prefs__flag-image";
      image.src = FLAG_ASSETS[lang];
      image.alt = "";
      image.width = 24;
      image.height = 24;
      image.decoding = "async";
      button.appendChild(image);

      const label = document.createElement("span");
      label.className = "idsai-site-prefs__sr";
      label.dataset.idsaiPrefsLangLabel = lang;
      button.appendChild(label);

      langGroup.appendChild(button);
    });

    mountDock();
    updateDock();
  }

  function resolveDockMount() {
    for (const selector of DOCK_MOUNT_SELECTORS) {
      const candidate = document.querySelector(selector);
      if (candidate instanceof HTMLElement) return candidate;
    }
    return null;
  }

  function mountDock() {
    if (!dock || !document.body) return;
    const nextHost = resolveDockMount();

    if (dockHost && dockHost !== nextHost) {
      dockHost.classList.remove("idsai-site-prefs-host");
    }

    dockHost = nextHost;

    if (dockHost) {
      dockHost.classList.add("idsai-site-prefs-host");
    }

    dock.classList.add("idsai-site-prefs--fixed");
    dock.classList.remove("idsai-site-prefs--inline");
    if (dock.parentElement !== document.body) {
      document.body.appendChild(dock);
    }
  }

  function updateDock() {
    if (!dock) return;
    mountDock();
    const languageGroupEl = dock.querySelector("[data-idsai-prefs='languageGroup']");

    dock.setAttribute("aria-label", key("prefs.title"));

    if (languageGroupEl) {
      languageGroupEl.setAttribute("aria-label", key("prefs.language"));
    }

    dock.querySelectorAll("[data-idsai-prefs-lang]").forEach((button) => {
      const lang = button.dataset.idsaiPrefsLang;
      const label = key(`lang.${lang}`);
      const active = lang === state.lang;
      button.setAttribute("aria-label", label);
      button.setAttribute("title", label);
      button.setAttribute("aria-pressed", String(active));
      button.classList.toggle("is-active", active);

      const srLabel = button.querySelector("[data-idsai-prefs-lang-label]");
      if (srLabel) srLabel.textContent = label;
    });
  }

  function init() {
    document.documentElement.lang = state.lang;
    buildDock();
    ensureObserver();
    apply(document.documentElement);
    window.addEventListener("resize", mountDock, { passive: true });
  }

  init();

  window.IDSAI18n = {
    key,
    t,
    apply,
    locale,
    compareStrings,
    formatDate,
    formatDateTime,
    formatTime,
    relativeTime,
    getLanguage,
    setLanguage,
  };
})();
