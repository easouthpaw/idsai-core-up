(() => {
  const auth = window.IDSAIAuth;

  function confirmAction(options) {
    if (auth && typeof auth.showConfirmDialog === "function") {
      return auth.showConfirmDialog(options);
    }
    return Promise.resolve(window.confirm(String((options && options.message) || "")));
  }

  const EDITABLE_STATUSES = new Set(["REVIEW", "GRADING"]);
  const FINALIZABLE_STATUSES = new Set(["GRADING"]);
  const RETAKE_PENALTY_PER_ATTEMPT = 5;
  const RETAKE_PENALTY_CAP = 25;

  const ui = {
    projectTitle: document.getElementById("projectTitle"),
    projectMeta: document.getElementById("projectMeta"),
    projectStatusBadge: document.getElementById("projectStatusBadge"),
    projectSelect: document.getElementById("projectSelect"),
    openCriteriaBtn: document.getElementById("openCriteriaBtn"),
    gradingStageTitle: document.getElementById("gradingStageTitle"),
    gradingStageText: document.getElementById("gradingStageText"),
    gradingSignalChips: document.getElementById("gradingSignalChips"),
    gradingChecklistIntro: document.getElementById("gradingChecklistIntro"),
    gradingList: document.getElementById("gradingList"),
    summaryCoverage: document.getElementById("summaryCoverage"),
    summaryMet: document.getElementById("summaryMet"),
    summaryScore: document.getElementById("summaryScore"),
    summaryPenalty: document.getElementById("summaryPenalty"),
    gradingProgressFill: document.getElementById("gradingProgressFill"),
    gradingProgressText: document.getElementById("gradingProgressText"),
    returnForRetakeBtn: document.getElementById("returnForRetakeBtn"),
    publishGradingBtn: document.getElementById("publishGradingBtn"),
    saveGradingBtn: document.getElementById("saveGradingBtn"),
    teamActivityList: document.getElementById("teamActivityList"),
    pageStatus: document.getElementById("pageStatus"),
  };

  const state = {
    profile: null,
    projectID: "",
    projects: [],
    project: null,
    readiness: null,
    criteria: [],
    members: [],
    tasks: [],
    taskActivities: [],
    grading: new Map(),
    canEdit: false,
    canPublish: false,
    isComplete: false,
    gradingRestricted: false,
  };

  function escapeHTML(value) {
    return String(value || "")
      .replaceAll("&", "&amp;")
      .replaceAll("<", "&lt;")
      .replaceAll(">", "&gt;")
      .replaceAll('"', "&quot;")
      .replaceAll("'", "&#039;");
  }

  function extractItems(payload) {
    if (Array.isArray(payload)) return payload;
    if (payload && Array.isArray(payload.items)) return payload.items;
    return [];
  }

  function retakeCount(project = state.project) {
    const count = Number(project && project.retake_count ? project.retake_count : 0);
    return Number.isFinite(count) && count > 0 ? Math.round(count) : 0;
  }

  function retakePenaltyPercent(project = state.project) {
    const explicit = Number(project && project.retake_penalty_percent ? project.retake_penalty_percent : 0);
    if (Number.isFinite(explicit) && explicit >= 0) {
      return Math.round(explicit);
    }
    return Math.min(RETAKE_PENALTY_CAP, retakeCount(project) * RETAKE_PENALTY_PER_ATTEMPT);
  }

  function retakeWord(count) {
    const mod10 = count % 10;
    const mod100 = count % 100;
    if (mod10 === 1 && mod100 !== 11) return "пересдача";
    if (mod10 >= 2 && mod10 <= 4 && (mod100 < 10 || mod100 >= 20)) return "пересдачи";
    return "пересдач";
  }

  function displayName(userID, fallbackEmail) {
    const normalizedUserID = String(userID || "").trim();
    const normalizedEmail = String(fallbackEmail || "").trim();
    const member = extractItems(state.members).find((item) => String(item && item.user_id || "") === normalizedUserID);
    if (member) {
      const fullName = String(member.full_name || "").trim();
      if (fullName) return fullName;
      const email = String(member.email || "").trim();
      if (email) return email;
    }
    if (normalizedUserID && state.project && normalizedUserID === String(state.project.created_by || "")) {
      return state.project.created_by_name || state.project.created_by_email || normalizedEmail || "Лидер проекта";
    }
    return normalizedEmail || "Участник";
  }

  function setStatus(message, isError) {
    if (!ui.pageStatus) return;
    ui.pageStatus.textContent = message || "";
    ui.pageStatus.classList.toggle("err", Boolean(isError));
  }

  async function request(method, url, body, extra = {}) {
    const options = { method, ...extra };
    if (body !== undefined) {
      options.body = body;
    }
    const { resp, data } = await auth.requestJSON(url, options);
    if (!resp.ok) {
      const err = new Error(data && data.error ? data.error : `${method} ${url} failed (${resp.status})`);
      err.status = resp.status;
      err.data = data;
      throw err;
    }
    return data;
  }

  function projectIDFromQuery() {
    return String(new URLSearchParams(window.location.search).get("project_id") || "").trim();
  }

  function setProjectInURL(projectID) {
    const url = new URL(window.location.href);
    if (projectID) {
      url.searchParams.set("project_id", projectID);
    } else {
      url.searchParams.delete("project_id");
    }
    window.history.replaceState({}, "", url.toString());
  }

  function statusCode(project) {
    return String(project?.status || "DRAFT").toUpperCase();
  }

  function reviewCode(project) {
    return String(project?.professor_review_status || "NONE").toUpperCase();
  }

  function isAssignedToMe(project) {
    return Boolean(project?.professor_id) && String(project.professor_id) === String(state.profile?.sub || "");
  }

  function isCreatedByMe(project) {
    return String(project?.created_by || "") === String(state.profile?.sub || "");
  }

  function statusMeta(status) {
    const code = String(status || "DRAFT").toUpperCase();
    if (code === "REVIEW") return { label: "Готов к ревью", tone: "review" };
    if (code === "RECRUITMENT") return { label: "Набор команды", tone: "recruitment" };
    if (code === "ACTIVE") return { label: "В работе", tone: "active" };
    if (code === "GRADING") return { label: "На оценке", tone: "grading" };
    if (code === "COMPLETED") return { label: "Завершен", tone: "done" };
    if (code === "ARCHIVE") return { label: "Закрыт", tone: "default" };
    return { label: "Подготовка", tone: "default" };
  }

  function formatDate(value) {
    if (!value) return "—";
    try {
      return new Date(value).toLocaleString("ru-RU");
    } catch (_) {
      return String(value);
    }
  }

  function sortProjects(items) {
    return [...items].sort((a, b) => {
      const score = (project) => {
        let value = 0;
        if (isAssignedToMe(project)) value += 30;
        if (statusCode(project) === "GRADING") value += 20;
        if (statusCode(project) === "REVIEW") value += 12;
        if (isCreatedByMe(project)) value += 8;
        return value;
      };
      const diff = score(b) - score(a);
      if (diff !== 0) return diff;
      const ad = new Date(a.updated_at || a.created_at || 0).getTime();
      const bd = new Date(b.updated_at || b.created_at || 0).getTime();
      return bd - ad;
    });
  }

  function normalizeProjects(items) {
    const list = Array.isArray(items) ? items : [];
    const filtered = list.filter((item) => statusCode(item) !== "ARCHIVE");
    return sortProjects(filtered);
  }

  async function loadProjectList() {
    const [mine, pub] = await Promise.all([
      request("GET", "/v2/projects/my"),
      request("GET", "/v2/projects/public"),
    ]);

    const merged = new Map();
    [mine, pub].forEach((list) => {
      (Array.isArray(list) ? list : []).forEach((item) => {
        if (!item || !item.id) return;
        merged.set(item.id, item);
      });
    });

    const preferred = Array.from(merged.values()).filter((project) => (
      isAssignedToMe(project) || isCreatedByMe(project) || statusCode(project) === "GRADING" || statusCode(project) === "REVIEW"
    ));

    state.projects = normalizeProjects(preferred.length ? preferred : Array.from(merged.values()));

    const queryProjectID = projectIDFromQuery();
    if (queryProjectID) {
      state.projectID = queryProjectID;
      return;
    }

    if (!state.projectID && state.projects.length > 0) {
      const pick = state.projects.find((project) => {
        const status = statusCode(project);
        return status === "GRADING" || status === "REVIEW";
      }) || state.projects[0];
      state.projectID = String(pick.id || "");
    }
  }

  function renderProjectPicker() {
    if (!ui.projectSelect) return;
    const options = [...state.projects];
    if (state.project && state.project.id && !options.some((item) => String(item.id) === String(state.project.id))) {
      options.unshift(state.project);
    }

    if (!options.length) {
      ui.projectSelect.innerHTML = `<option value="">Нет проектов</option>`;
      ui.projectSelect.disabled = true;
      return;
    }

    ui.projectSelect.disabled = false;
    ui.projectSelect.innerHTML = options.map((project) => {
      const status = statusMeta(project.status);
      const owner = project.created_by_name || project.created_by_email || "Команда";
      const selected = String(project.id) === String(state.projectID) ? "selected" : "";
      return `<option value="${escapeHTML(project.id)}" ${selected}>${escapeHTML(project.title || "Без названия")} · ${escapeHTML(status.label)} · ${escapeHTML(owner)}</option>`;
    }).join("");
  }

  function ensureGradeEntry(criterionID) {
    const id = String(criterionID || "");
    if (!id) return { criterion_id: "", is_met: null, comment: "" };
    if (!state.grading.has(id)) {
      state.grading.set(id, {
        criterion_id: id,
        is_met: null,
        comment: "",
      });
    }
    return state.grading.get(id);
  }

  function applyGradingPayload(items) {
    state.grading.clear();
    (Array.isArray(items) ? items : []).forEach((item) => {
      const id = String(item.criterion_id || "").trim();
      if (!id) return;
      state.grading.set(id, {
        criterion_id: id,
        is_met: item.is_met === true ? true : item.is_met === false ? false : null,
        comment: String(item.comment || ""),
      });
    });
  }

  function renderProjectHeader() {
    const project = state.project;
    if (!project) {
      ui.projectTitle.textContent = "Оценивание проекта";
      ui.projectMeta.textContent = "Выберите проект из списка, чтобы открыть рабочее место ревьюера.";
      ui.projectStatusBadge.textContent = "—";
      ui.projectStatusBadge.className = "status-pill default";
      return;
    }

    const status = statusMeta(project.status);
    const reviewerState = reviewCode(project);
    const owner = project.created_by_name || project.created_by_email || "Команда";
    const retakes = retakeCount(project);
    const retakeMeta = retakes > 0 ? ` · Пересдачи: ${retakes}` : "";

    ui.projectTitle.textContent = project.title || "Без названия";
    ui.projectMeta.textContent = `Автор: ${owner} · Обновлен: ${formatDate(project.updated_at || project.created_at)} · Ревью: ${reviewerState}${retakeMeta}`;
    ui.projectStatusBadge.textContent = status.label;
    ui.projectStatusBadge.className = `status-pill ${status.tone}`;

    ui.openCriteriaBtn.href = state.projectID
      ? `/dev/professor/criteria?project_id=${encodeURIComponent(state.projectID)}`
      : "/dev/professor/criteria";
  }

  function guideChip(label, tone) {
    return `<span class="prof-check prof-check--${escapeHTML(tone)}">${escapeHTML(label)}</span>`;
  }

  function renderStagePanel() {
    const project = state.project;
    if (!project) {
      if (ui.gradingStageTitle) ui.gradingStageTitle.textContent = "Нет выбранного проекта";
      if (ui.gradingStageText) ui.gradingStageText.textContent = "Откройте проект из списка справа, чтобы увидеть контекст ревью и критерии.";
      if (ui.gradingSignalChips) ui.gradingSignalChips.innerHTML = "";
      return;
    }

    const status = statusCode(project);
    const readiness = state.readiness || {};
    const professorState = reviewCode(project);
    const criteriaCount = Number(readiness.criteria_count || state.criteria.length || 0);
    const membersReady = `${Number(readiness.active_members || 0)}/${Number(readiness.required_members || 0)}`;
    const retakes = retakeCount(project);

    let title = "Контекст оценивания";
    let text = "Состояние проекта определяет, что преподаватель может сделать прямо сейчас.";

    if (status === "DRAFT" || status === "RECRUITMENT") {
      title = "Проект еще готовится к запуску";
      text = "Сейчас ценнее проверить критерии и убедиться, что преподавательское ревью подтверждено, чем пытаться выставлять оценку заранее.";
    } else if (status === "REVIEW") {
      title = "Этап подготовки к активной фазе";
      text = "Оценивание уже можно черновиком заполнять, но основной акцент здесь на полноте критериев и готовности команды.";
    } else if (status === "ACTIVE") {
      if (retakes > 0) {
        title = "Проект возвращен на доработку";
        text = "Команда исправляет замечания после пересдачи. Когда проект снова отправят на финальное ревью, итоговая оценка будет учитывать небольшой штраф.";
      } else {
        title = "Команда в активной работе";
        text = "Ревью-форма пока закрыта. Следующий преподавательский шаг появится после отправки проекта на финальную оценку.";
      }
    } else if (status === "GRADING") {
      title = retakes > 0 ? "Повторное финальное ревью" : "Проект открыт для финального ревью";
      text = retakes > 0
        ? "Это повторная попытка после пересдачи. Можно завершить оценивание или вернуть проект на еще одну доработку."
        : "Отметьте каждый критерий, добавьте точечные комментарии и завершите оценивание только после полного покрытия чек-листа.";
    } else if (status === "COMPLETED") {
      title = "Оценивание уже завершено";
      text = "Страница остается полезной как итоговый отчет: можно просмотреть критерии, покрытие и комментарии без редактирования.";
    }

    if (ui.gradingStageTitle) ui.gradingStageTitle.textContent = title;
    if (ui.gradingStageText) ui.gradingStageText.textContent = text;
    if (ui.gradingSignalChips) {
      ui.gradingSignalChips.innerHTML = [
        guideChip(`Статус: ${statusMeta(status).label}`, statusMeta(status).tone === "done" ? "done" : statusMeta(status).tone),
        guideChip(`Ревью: ${professorState}`, professorState === "ACCEPTED" ? "done" : professorState === "PENDING" ? "current" : "blocked"),
        guideChip(`Команда: ${membersReady}`, Number(readiness.required_members || 0) > 0 && Number(readiness.active_members || 0) >= Number(readiness.required_members || 0) ? "done" : "current"),
        guideChip(`Критерии: ${criteriaCount}`, criteriaCount > 0 ? "done" : "blocked"),
        guideChip(`Пересдачи: ${retakes}`, retakes > 0 ? "current" : "done"),
      ].join("");
    }
  }

  function renderGuidanceList() {
    if (!ui.gradingChecklistIntro) return;

    const project = state.project;
    if (!project) {
      ui.gradingChecklistIntro.innerHTML = `
        <article class="grading-guide-item grading-guide-item--empty">
          <strong>Нет выбранного проекта</strong>
          <p>Откройте проект из списка сверху, чтобы увидеть условия ревью и прогресс оценивания.</p>
        </article>
      `;
      return;
    }

    const readiness = state.readiness || {};
    const retakes = retakeCount(project);
    const items = [
      {
        label: "Команда",
        value: `${Number(readiness.active_members || 0)} из ${Number(readiness.required_members || 0)}`,
        tone: Number(readiness.required_members || 0) > 0 && Number(readiness.active_members || 0) >= Number(readiness.required_members || 0) ? "done" : "current",
      },
      {
        label: "Преподаватель",
        value: reviewCode(project),
        tone: reviewCode(project) === "ACCEPTED" ? "done" : reviewCode(project) === "PENDING" ? "current" : "blocked",
      },
      {
        label: "Критерии",
        value: `${Number(readiness.criteria_count || state.criteria.length || 0)}`,
        tone: Number(readiness.criteria_count || state.criteria.length || 0) > 0 ? "done" : "blocked",
      },
      {
        label: "Режим страницы",
        value: state.canEdit ? "Редактирование доступно" : state.canPublish ? "Готово к завершению" : "Только просмотр",
        tone: state.canEdit ? "done" : state.gradingRestricted ? "blocked" : "current",
      },
      {
        label: "Пересдачи",
        value: retakes > 0 ? `${retakes} · штраф ${retakePenaltyPercent(project)}%` : "Без штрафа",
        tone: retakes > 0 ? "current" : "done",
      },
    ];

    ui.gradingChecklistIntro.innerHTML = items.map((item) => `
      <article class="grading-guide-item grading-guide-item--${escapeHTML(item.tone)}">
        <small>${escapeHTML(item.label)}</small>
        <strong>${escapeHTML(item.value)}</strong>
      </article>
    `).join("");
  }

  function renderSummary() {
    const total = state.criteria.length;
    let answered = 0;
    let met = 0;
    let weightTotal = 0;
    let weightMet = 0;

    state.criteria.forEach((criterion) => {
      const grade = ensureGradeEntry(criterion.id);
      const weight = Number(criterion.weight || 0) > 0 ? Number(criterion.weight) : 1;
      weightTotal += weight;
      if (grade.is_met === true || grade.is_met === false) {
        answered += 1;
      }
      if (grade.is_met === true) {
        met += 1;
        weightMet += weight;
      }
    });

    const coverage = total > 0 ? Math.round((answered * 100) / total) : 0;
    const rawScore = weightTotal > 0 ? Math.round((weightMet * 100) / weightTotal) : 0;
    const penalty = retakePenaltyPercent();
    const score = Math.max(0, rawScore - penalty);
    state.isComplete = total > 0 && answered === total;

    ui.summaryCoverage.textContent = `${coverage}%`;
    ui.summaryMet.textContent = `${met}/${total}`;
    ui.summaryScore.textContent = `${score}/100`;
    if (ui.summaryPenalty) {
      if (penalty > 0) {
        const attempts = retakeCount();
        ui.summaryPenalty.hidden = false;
        ui.summaryPenalty.textContent = `Финальный результат уже учитывает ${attempts} ${retakeWord(attempts)}: -${penalty} баллов от итоговой оценки.`;
      } else {
        ui.summaryPenalty.hidden = true;
        ui.summaryPenalty.textContent = "";
      }
    }

    if (ui.gradingProgressFill) {
      ui.gradingProgressFill.style.width = `${coverage}%`;
    }
    if (ui.gradingProgressText) {
      ui.gradingProgressText.textContent = `${coverage}%`;
    }
  }

  function renderActionButtons() {
    const status = statusCode(state.project);
    const canReturn = status === "GRADING" && state.canPublish;
    if (ui.returnForRetakeBtn) {
      ui.returnForRetakeBtn.disabled = !canReturn;
    }
    if (ui.saveGradingBtn) {
      ui.saveGradingBtn.disabled = !state.canEdit;
    }
    if (ui.publishGradingBtn) {
      ui.publishGradingBtn.disabled = !(state.canPublish && state.isComplete);
    }
    if (status === "COMPLETED") {
      if (ui.returnForRetakeBtn) ui.returnForRetakeBtn.disabled = true;
      if (ui.saveGradingBtn) ui.saveGradingBtn.disabled = true;
      if (ui.publishGradingBtn) ui.publishGradingBtn.disabled = true;
    }
  }

  function buildTeamActivityRows() {
    const rows = new Map();
    const activeMembers = extractItems(state.members).filter((item) => String(item && item.status || "").toUpperCase() === "ACTIVE");

    activeMembers.forEach((member) => {
      const userID = String(member.user_id || "").trim();
      if (!userID) return;
      rows.set(userID, {
        userID,
        name: displayName(userID, member.email),
        email: String(member.email || "").trim(),
        completed: 0,
        inProgress: 0,
        events: 0,
        attachments: 0,
        score: 0,
        progress: 0,
      });
    });

    if (state.project && state.project.created_by && !rows.has(String(state.project.created_by))) {
      const ownerID = String(state.project.created_by);
      rows.set(ownerID, {
        userID: ownerID,
        name: displayName(ownerID, state.project.created_by_email),
        email: String(state.project.created_by_email || "").trim(),
        completed: 0,
        inProgress: 0,
        events: 0,
        attachments: 0,
        score: 0,
        progress: 0,
      });
    }

    extractItems(state.tasks).forEach((task) => {
      const assigneeID = String(task && task.assignee_user_id || "").trim();
      if (!assigneeID || !rows.has(assigneeID)) return;
      const entry = rows.get(assigneeID);
      const taskStatus = String(task.status || "").toUpperCase();
      if (taskStatus === "DONE") {
        entry.completed += 1;
      } else if (taskStatus === "IN_PROGRESS") {
        entry.inProgress += 1;
      }
    });

    extractItems(state.taskActivities).forEach((item) => {
      const actorID = String(item && item.actor_user_id || "").trim();
      if (!actorID || !rows.has(actorID)) return;
      const entry = rows.get(actorID);
      entry.events += 1;
      entry.attachments += Array.isArray(item.attachments) ? item.attachments.length : 0;
    });

    const values = Array.from(rows.values()).map((entry) => {
      entry.score = (entry.completed * 6) + (entry.inProgress * 2) + entry.events + Math.min(entry.attachments, 4);
      return entry;
    });

    const maxScore = values.reduce((best, entry) => Math.max(best, entry.score), 0);
    values.forEach((entry) => {
      entry.progress = maxScore > 0 ? Math.max(6, Math.round((entry.score * 100) / maxScore)) : 0;
    });

    return values.sort((a, b) => {
      if (b.score !== a.score) return b.score - a.score;
      if (b.completed !== a.completed) return b.completed - a.completed;
      return a.name.localeCompare(b.name, "ru");
    });
  }

  function renderTeamActivity() {
    if (!ui.teamActivityList) return;

    const rows = buildTeamActivityRows();
    if (!rows.length) {
      ui.teamActivityList.innerHTML = `
        <article class="team-activity-empty">
          <strong>Пока нет данных для сводки</strong>
          <p>Когда в проекте появятся задачи и события, здесь отобразится вклад участников.</p>
        </article>
      `;
      return;
    }

    ui.teamActivityList.innerHTML = rows.map((entry, index) => {
      const summary = [];
      if (entry.completed > 0) summary.push(`${entry.completed} закрыто`);
      if (entry.inProgress > 0) summary.push(`${entry.inProgress} в работе`);
      if (entry.events > 0) summary.push(`${entry.events} событий`);
      const meta = summary.length ? summary.join(" • ") : "Пока без зафиксированной активности";
      const rank = index === 0 && entry.score > 0 ? "Лидирует" : `Уровень ${entry.progress}%`;

      return `
        <article class="team-activity-item">
          <div class="team-activity-item__head">
            <div>
              <strong>${escapeHTML(entry.name)}</strong>
              <p>${escapeHTML(meta)}</p>
            </div>
            <span class="team-activity-item__rank">${escapeHTML(rank)}</span>
          </div>
          <div class="team-activity-track" aria-hidden="true">
            <span style="width:${Math.max(0, Math.min(100, entry.progress))}%"></span>
          </div>
        </article>
      `;
    }).join("");
  }

  function renderRestrictedState(title, text, tone) {
    return `
      <article class="grading-empty grading-empty--${escapeHTML(tone)}">
        <span class="material-symbols-outlined" aria-hidden="true">${tone === "blocked" ? "lock_clock" : "rate_review"}</span>
        <div>
          <h3>${escapeHTML(title)}</h3>
          <p>${escapeHTML(text)}</p>
        </div>
      </article>
    `;
  }

  function renderGradingList() {
    if (!ui.gradingList) return;

    const status = statusCode(state.project);

    if (state.gradingRestricted) {
      ui.gradingList.innerHTML = renderRestrictedState(
        "Оценивание еще не открыто",
        "Эта страница останется рабочим местом ревьюера, но сами отметки станут доступны после перехода проекта в REVIEW или GRADING.",
        "blocked"
      );
      state.isComplete = false;
      renderSummary();
      renderActionButtons();
      return;
    }

    if (!state.criteria.length) {
      ui.gradingList.innerHTML = renderRestrictedState(
        "Критерии пока не настроены",
        "Сначала откройте страницу критериев и соберите чек-лист оценки. После этого форма ревью автоматически станет осмысленной.",
        "current"
      );
      state.isComplete = false;
      renderSummary();
      renderActionButtons();
      return;
    }

    ui.gradingList.innerHTML = state.criteria.map((criterion, index) => {
      const id = String(criterion.id || "");
      const grade = ensureGradeEntry(id);
      const disabledAttr = state.canEdit ? "" : "disabled";
      const yesActive = grade.is_met === true ? "active" : "";
      const noActive = grade.is_met === false ? "active" : "";
      const answerState = grade.is_met === true ? "Критерий подтвержден" : grade.is_met === false ? "Критерий не выполнен" : "Ответ еще не выбран";

      return `
        <article class="grading-item ${grade.is_met === null ? "grading-item--pending" : grade.is_met ? "grading-item--yes" : "grading-item--no"}" data-criterion-id="${escapeHTML(id)}">
          <div class="grading-head">
            <div class="grading-head__main">
              <p class="criterion-number">Критерий ${index + 1}</p>
              <strong>${escapeHTML(criterion.title || "Без названия")}</strong>
            </div>
            <div class="grading-right">
              <span class="criterion-weight">Вес ${escapeHTML(criterion.weight || 1)}</span>
              <span class="grading-answer-state">${escapeHTML(answerState)}</span>
            </div>
          </div>

          <p class="grading-desc">${escapeHTML(criterion.description || "Описание критерия отсутствует.")}</p>

          <div class="grade-switch grade-switch--wide">
            <button class="grade-btn yes ${yesActive}" data-grade-value="yes" ${disabledAttr}>Да, выполнено</button>
            <button class="grade-btn no ${noActive}" data-grade-value="no" ${disabledAttr}>Нет, не выполнено</button>
          </div>

          <label for="comment-${escapeHTML(id)}">Комментарий преподавателя</label>
          <textarea id="comment-${escapeHTML(id)}" class="grade-comment" data-grade-comment="${escapeHTML(id)}" placeholder="Добавьте короткий комментарий только там, где он поможет команде понять решение." ${disabledAttr}>${escapeHTML(grade.comment || "")}</textarea>
        </article>
      `;
    }).join("");

    renderSummary();
    renderActionButtons();
  }

  async function loadPageData() {
    if (!state.projectID) {
      state.project = null;
      state.readiness = null;
      state.criteria = [];
      state.members = [];
      state.tasks = [];
      state.taskActivities = [];
      applyGradingPayload([]);
      renderProjectPicker();
      renderProjectHeader();
      renderStagePanel();
      renderGuidanceList();
      renderGradingList();
      renderTeamActivity();
      setStatus("Нет доступных проектов для оценивания.", true);
      return;
    }

    try {
      const [project, criteria, readiness, membersResp, tasksResp, taskActivityResp] = await Promise.all([
        request("GET", `/v2/projects/${state.projectID}`, undefined, { skipAccessAlert: true }),
        request("GET", `/v2/projects/${state.projectID}/criteria`, undefined, { skipAccessAlert: true }),
        request("GET", `/v2/projects/${state.projectID}/readiness`, undefined, { skipAccessAlert: true }).catch(() => null),
        request("GET", `/v2/projects/${state.projectID}/members`, undefined, { skipAccessAlert: true }).catch(() => []),
        request("GET", `/v2/projects/${state.projectID}/tasks`, undefined, { skipAccessAlert: true }).catch(() => []),
        request("GET", `/v2/projects/${state.projectID}/tasks/activity`, undefined, { skipAccessAlert: true }).catch(() => ({ items: [] })),
      ]);

      state.project = project;
      state.criteria = Array.isArray(criteria) ? criteria : [];
      state.readiness = readiness && typeof readiness === "object" ? readiness : null;
      state.members = extractItems(membersResp);
      state.tasks = extractItems(tasksResp);
      state.taskActivities = extractItems(taskActivityResp);
      state.gradingRestricted = false;

      try {
        const gradingResp = await request("GET", `/v2/projects/${state.projectID}/grading`, undefined, { skipAccessAlert: true });
        applyGradingPayload(gradingResp && Array.isArray(gradingResp.items) ? gradingResp.items : []);
      } catch (err) {
        if (err.status === 403) {
          state.gradingRestricted = true;
          applyGradingPayload([]);
        } else {
          throw err;
        }
      }
    } catch (err) {
      if (err.status === 404 && auth && typeof auth.redirectToNotFound === "function") {
        auth.redirectToNotFound(window.location.href);
        return;
      }
      throw err;
    }

    const status = statusCode(state.project);
    state.canEdit = EDITABLE_STATUSES.has(status);
    state.canPublish = FINALIZABLE_STATUSES.has(status);

    renderProjectPicker();
    renderProjectHeader();
    renderStagePanel();
    renderGuidanceList();
    renderGradingList();
    renderTeamActivity();

    if (state.gradingRestricted || !state.canEdit) {
      if (status === "COMPLETED") {
        setStatus("Оценивание завершено. Страница работает как итоговый отчет по проекту.", false);
      } else if (status === "ACTIVE" && retakeCount() > 0) {
        setStatus(`Проект возвращен на пересдачу. При следующем финале будет учтен штраф ${retakePenaltyPercent()}%.`, false);
      } else if (status === "ACTIVE") {
        setStatus("Команда еще работает над проектом. Форма оценки откроется после отправки на ревью.", false);
      } else {
        setStatus(`Сейчас проект находится в статусе ${statusMeta(status).label}. Редактирование оценок на этом этапе ограничено.`, status !== "COMPLETED");
      }
    } else if (state.canPublish && !state.isComplete) {
      setStatus("Чтобы завершить оценивание, отметьте каждый критерий как выполненный или невыполненный.", false);
    } else {
      setStatus("Рабочее место ревьюера готово. Можно сохранять оценку.", false);
    }
  }

  function setCriterionDecision(criterionID, value) {
    const entry = ensureGradeEntry(criterionID);
    entry.is_met = value;
    state.grading.set(String(criterionID), entry);
  }

  function setCriterionComment(criterionID, comment) {
    const entry = ensureGradeEntry(criterionID);
    entry.comment = String(comment || "").trim();
    state.grading.set(String(criterionID), entry);
  }

  async function saveGrading() {
    if (!state.projectID) {
      setStatus("Сначала выберите проект.", true);
      return;
    }
    if (!state.canEdit) {
      setStatus("Для этого статуса проекта редактирование оценки недоступно.", true);
      return;
    }

    const items = state.criteria.map((criterion) => {
      const entry = ensureGradeEntry(criterion.id);
      return {
        criterion_id: String(criterion.id || ""),
        is_met: entry.is_met,
        comment: String(entry.comment || ""),
      };
    });

    ui.saveGradingBtn.disabled = true;
    try {
      const resp = await request("PUT", `/v2/projects/${state.projectID}/grading`, { items }, { skipAccessAlert: true });
      applyGradingPayload(resp && Array.isArray(resp.items) ? resp.items : []);
      renderGradingList();
      setStatus("Оценка сохранена. Можно продолжать ревью или завершить его позже.", false);
    } catch (err) {
      setStatus(err.message || String(err), true);
    } finally {
      renderActionButtons();
    }
  }

  async function returnForRetake() {
    if (!state.projectID) {
      setStatus("Сначала выберите проект.", true);
      return;
    }
    if (!state.canPublish) {
      setStatus("Вернуть на пересдачу можно только на этапе финального оценивания.", true);
      return;
    }

    const nextRetakeCount = retakeCount() + 1;
    const nextPenalty = Math.min(RETAKE_PENALTY_CAP, nextRetakeCount * RETAKE_PENALTY_PER_ATTEMPT);
    const confirmed = await confirmAction({
      title: "Отправить на пересдачу",
      message: `Проект вернется в ACTIVE. После повторной сдачи итоговая оценка будет включать штраф ${nextPenalty}% (всего пересдач: ${nextRetakeCount}).`,
      confirmText: "Вернуть на пересдачу",
      danger: true,
    });
    if (!confirmed) return;

    if (ui.returnForRetakeBtn) {
      ui.returnForRetakeBtn.disabled = true;
    }
    try {
      await request("POST", `/v2/projects/${state.projectID}/grading/return`, {}, { skipAccessAlert: true });
      await loadPageData();
      setStatus(`Проект отправлен на пересдачу. Текущий накопленный штраф: ${retakePenaltyPercent()}%.`, false);
    } catch (err) {
      setStatus(err.message || String(err), true);
    } finally {
      renderActionButtons();
    }
  }

  async function publishGrading() {
    if (!state.projectID) {
      setStatus("Сначала выберите проект.", true);
      return;
    }
    if (!state.canPublish) {
      setStatus("Завершение доступно только на этапе финального оценивания.", true);
      return;
    }
    if (!state.isComplete) {
      setStatus("Нельзя завершить ревью, пока не отмечены все критерии.", true);
      return;
    }

    const confirmed = await confirmAction({
      title: "Завершить оценивание",
      message: retakePenaltyPercent() > 0
        ? `Проект будет переведен в завершенный статус, а итоговая оценка зафиксируется с учетом штрафа ${retakePenaltyPercent()}% за пересдачу.`
        : "Проект будет переведен в завершенный статус, а итоговые оценки зафиксируются в карточке проекта.",
      confirmText: "Завершить оценивание",
      danger: true,
    });
    if (!confirmed) return;

    ui.publishGradingBtn.disabled = true;
    try {
      await request("POST", `/v2/projects/${state.projectID}/grading/publish`, {}, { skipAccessAlert: true });
      await loadPageData();
      setStatus(
        retakePenaltyPercent() > 0
          ? `Оценивание завершено. Итог закреплен в проекте с учетом штрафа ${retakePenaltyPercent()}%.`
          : "Оценивание завершено. Итог закреплен в проекте.",
        false,
      );
    } catch (err) {
      setStatus(err.message || String(err), true);
    } finally {
      renderActionButtons();
    }
  }

  function attachEvents() {
    if (ui.projectSelect) {
      ui.projectSelect.addEventListener("change", async () => {
        state.projectID = String(ui.projectSelect.value || "").trim();
        setProjectInURL(state.projectID);
        try {
          setStatus("Переключаю проект...", false);
          await loadPageData();
        } catch (err) {
          setStatus(err.message || String(err), true);
        }
      });
    }

    if (ui.gradingList) {
      ui.gradingList.addEventListener("click", (event) => {
        const btn = event.target.closest("button[data-grade-value]");
        if (!btn || !state.canEdit) return;
        const row = btn.closest("[data-criterion-id]");
        if (!row) return;
        const criterionID = String(row.getAttribute("data-criterion-id") || "");
        if (!criterionID) return;

        const value = String(btn.getAttribute("data-grade-value") || "");
        if (value === "yes") {
          setCriterionDecision(criterionID, true);
        } else if (value === "no") {
          setCriterionDecision(criterionID, false);
        }
        renderGradingList();
      });

      ui.gradingList.addEventListener("input", (event) => {
        const input = event.target.closest("textarea[data-grade-comment]");
        if (!input) return;
        const criterionID = String(input.getAttribute("data-grade-comment") || "");
        if (!criterionID) return;
        setCriterionComment(criterionID, input.value);
        renderSummary();
      });
    }

    if (ui.saveGradingBtn) {
      ui.saveGradingBtn.addEventListener("click", () => {
        void saveGrading();
      });
    }

    if (ui.publishGradingBtn) {
      ui.publishGradingBtn.addEventListener("click", () => {
        void publishGrading();
      });
    }

    if (ui.returnForRetakeBtn) {
      ui.returnForRetakeBtn.addEventListener("click", () => {
        void returnForRetake();
      });
    }
  }

  async function bootstrap() {
    state.profile = await auth.ensureSession("professor");
    if (!state.profile) return;

    attachEvents();
    setStatus("Собираю контекст ревьюера...", false);

    try {
      await loadProjectList();
      renderProjectPicker();
      setProjectInURL(state.projectID);
      await loadPageData();
    } catch (err) {
      setStatus(err.message || String(err), true);
    }
  }

  void bootstrap();
})();
