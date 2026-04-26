(() => {
  const auth = window.IDSAIAuth;
  const i18n = window.IDSAI18n;

  const ui = {
    profGreeting: document.getElementById("profGreeting"),
    projectsBody: document.getElementById("projectsBody"),
    focusList: document.getElementById("focusList"),
    refreshBtn: document.getElementById("refreshBtn"),
    pageStatus: document.getElementById("pageStatus"),
    statTotal: document.getElementById("statTotal"),
    statReview: document.getElementById("statReview"),
    statActive: document.getElementById("statActive"),
    statRecruitment: document.getElementById("statRecruitment"),
  };

  const state = {
    profile: null,
    projects: [],
    reviewInvites: [],
    readiness: new Map(),
  };

  function escapeHTML(value) {
    return String(value || "")
      .replaceAll("&", "&amp;")
      .replaceAll("<", "&lt;")
      .replaceAll(">", "&gt;")
      .replaceAll('"', "&quot;")
      .replaceAll("'", "&#039;");
  }

  function capitalize(value) {
    const text = String(value || "").trim();
    if (!text) return "";
    return text.charAt(0).toUpperCase() + text.slice(1);
  }

  function greetingName(profile) {
    const fullName = String(profile?.full_name || profile?.name || "").trim();
    if (fullName) {
      return fullName.split(/\s+/).filter(Boolean)[0] || fullName;
    }
    const email = String(profile?.email || "").trim();
    const local = email.split("@")[0] || "";
    return local ? capitalize(local.split(/[._-]/)[0]) : "коллега";
  }

  function setStatus(message, isError) {
    if (!ui.pageStatus) return;
    ui.pageStatus.textContent = message || "";
    ui.pageStatus.classList.toggle("err", Boolean(isError));
  }

  function renderGreeting(profile) {
    if (!ui.profGreeting) return;
    ui.profGreeting.textContent = `Привет, ${greetingName(profile)}!`;
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

  function formatDate(value) {
    if (!value) return "—";
    try {
      return i18n ? i18n.formatDateTime(value) : new Date(value).toLocaleString("ru-RU");
    } catch (_) {
      return String(value);
    }
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

  function isInvitePending(project) {
    return isAssignedToMe(project) && reviewCode(project) === "PENDING";
  }

  function isAcceptedReviewer(project) {
    return isAssignedToMe(project) && reviewCode(project) === "ACCEPTED";
  }

  function isRelevantProject(project) {
    return isCreatedByMe(project) || isAssignedToMe(project) || reviewCode(project) === "PENDING" || statusCode(project) === "GRADING";
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

  function professorMeta(project) {
    if (isInvitePending(project)) {
      return { label: "Ждет ваш ответ", tone: "review" };
    }
    if (isAcceptedReviewer(project)) {
      return { label: "Ревью закреплено за вами", tone: "active" };
    }
    if (reviewCode(project) === "REJECTED") {
      return { label: "Приглашение отклонено", tone: "default" };
    }
    if (project?.professor_id) {
      return { label: "Назначен другой преподаватель", tone: "default" };
    }
    return { label: "Преподаватель пока не закреплен", tone: "default" };
  }

  function compactCheck(label, value, tone) {
    return `<span class="prof-check prof-check--${escapeHTML(tone)}">${escapeHTML(label)}: ${escapeHTML(value)}</span>`;
  }

  function buildReadinessBadges(project) {
    const ready = state.readiness.get(String(project.id || ""));
    const items = [];

    if (ready && typeof ready === "object") {
      const memberTone = Number(ready.active_members || 0) >= Number(ready.required_members || 0) && Number(ready.required_members || 0) > 0 ? "done" : "current";
      items.push(compactCheck("Команда", `${Number(ready.active_members || 0)}/${Number(ready.required_members || 0)}`, memberTone));

      const professorTone = String(ready.professor_status || "NONE").toUpperCase() === "ACCEPTED"
        ? "done"
        : String(ready.professor_status || "NONE").toUpperCase() === "PENDING"
          ? "current"
          : "blocked";
      items.push(compactCheck("Ревью", String(ready.professor_status || "NONE").toUpperCase(), professorTone));

      const criteriaTone = Number(ready.criteria_count || 0) > 0 ? "done" : "blocked";
      items.push(compactCheck("Критерии", String(ready.criteria_count || 0), criteriaTone));
    } else {
      const meta = professorMeta(project);
      items.push(compactCheck("Роль", meta.label, meta.tone === "active" ? "done" : meta.tone === "review" ? "current" : "blocked"));
    }

    return items.join("");
  }

  function canActivateProject(project) {
    const ready = state.readiness.get(String(project?.id || ""));
    const status = statusCode(project);
    return isAcceptedReviewer(project) &&
      Boolean(ready && ready.can_activate) &&
      (status === "RECRUITMENT" || status === "REVIEW");
  }

  function projectNarrative(project) {
    const ready = state.readiness.get(String(project.id || ""));
    const status = statusCode(project);

    if (isInvitePending(project)) {
      return "Команда ждет вашего решения. После принятия вы сможете вести критерии и финальное ревью.";
    }
    if (isAcceptedReviewer(project) && status === "GRADING") {
      return "Проект уже передан на финальную проверку. Откройте оценивание и завершите ревью по критериям.";
    }
    if (canActivateProject(project)) {
      return "Команда, ревью преподавателя и критерии готовы. Можно разрешить запуск проекта.";
    }
    if (isAcceptedReviewer(project) && (status === "DRAFT" || status === "REVIEW" || status === "RECRUITMENT")) {
      if (ready && Number(ready.criteria_count || 0) === 0) {
        return "До запуска не хватает критериев. Здесь главное действие преподавателя именно настройка чек-листа.";
      }
      return "Проект еще готовится к активной фазе. Проверьте критерии и дождитесь готовности команды.";
    }
    if (isAcceptedReviewer(project) && status === "ACTIVE") {
      return "Команда сейчас в работе. На этом этапе преподаватель сопровождает проект и ждет отправки на оценивание.";
    }
    if (status === "COMPLETED") {
      return "Ревью завершено. Можно открыть оценивание, чтобы посмотреть итоговую картину и комментарии.";
    }
    if (isCreatedByMe(project)) {
      return "Проект принадлежит вам. Управление набором, запуском и составом команды остается внутри карточки самого проекта.";
    }
    return "Проект доступен для просмотра, но преподавательские действия здесь пока не требуются.";
  }

  function projectPrimaryAction(project) {
    const status = statusCode(project);
    if (isInvitePending(project)) {
      return { label: "Принять ревью", act: "accept", primary: true };
    }
    if (canActivateProject(project)) {
      return { label: "Дать разрешение на запуск и запустить", act: "activate", primary: true };
    }
    if (isAcceptedReviewer(project) && (status === "GRADING" || status === "COMPLETED")) {
      return { label: status === "COMPLETED" ? "Открыть результат" : "Открыть оценивание", act: "grade", primary: true };
    }
    if (isAcceptedReviewer(project) && (status === "DRAFT" || status === "REVIEW" || status === "RECRUITMENT")) {
      return { label: "Настроить критерии", act: "criteria", primary: true };
    }
    return { label: "Открыть проект", act: "open", primary: false };
  }

  function projectSecondaryActions(project) {
    const status = statusCode(project);
    const actions = [{ label: "Открыть", act: "open", primary: false }];

    if (status === "COMPLETED" || status === "ARCHIVE") {
      actions.unshift({ label: "Отчёт PDF", act: "report", primary: false });
    }

    if (isInvitePending(project)) {
      actions.unshift({ label: "Отклонить", act: "reject", primary: false, danger: true });
      return actions;
    }

    if (canActivateProject(project)) {
      actions.unshift({ label: "Критерии", act: "criteria", primary: false });
      return actions;
    }

    if (isAcceptedReviewer(project) && (status === "DRAFT" || status === "REVIEW" || status === "RECRUITMENT")) {
      actions.unshift({ label: "Критерии", act: "criteria", primary: false });
      return actions;
    }

    if (isAcceptedReviewer(project) && (status === "GRADING" || status === "COMPLETED")) {
      actions.unshift({ label: "Оценивание", act: "grade", primary: false });
      return actions;
    }

    return actions;
  }

  function renderActionButton(action) {
    const classes = ["action-btn"];
    if (action.primary) classes.push("primary");
    if (action.danger) classes.push("danger");
    return `<button class="${classes.join(" ")}" data-act="${escapeHTML(action.act)}" data-id="${escapeHTML(action.id)}">${escapeHTML(action.label)}</button>`;
  }

  function relevantProjects() {
    const base = state.projects.filter((project) => statusCode(project) !== "ARCHIVE");
    return base.filter(isRelevantProject);
  }

  function sortProjects(items) {
    return [...items].sort((a, b) => {
      const scoreA = (isInvitePending(a) ? 40 : 0) + (statusCode(a) === "GRADING" ? 20 : 0) + (isRelevantProject(a) ? 10 : 0);
      const scoreB = (isInvitePending(b) ? 40 : 0) + (statusCode(b) === "GRADING" ? 20 : 0) + (isRelevantProject(b) ? 10 : 0);
      if (scoreA !== scoreB) return scoreB - scoreA;
      const dateA = new Date(a.updated_at || a.created_at || 0).getTime();
      const dateB = new Date(b.updated_at || b.created_at || 0).getTime();
      return dateB - dateA;
    });
  }

  async function loadProjects() {
    const [faculty, invites] = await Promise.all([
      request("GET", "/v2/projects/faculty"),
      request("GET", "/v2/professor/review-invites?limit=100"),
    ]);

    state.reviewInvites = Array.isArray(invites) ? invites : [];

    const merged = new Map();
    [faculty, state.reviewInvites].forEach((list) => {
      (Array.isArray(list) ? list : []).forEach((item) => {
        if (!item || !item.id) return;
        merged.set(item.id, item);
      });
    });

    state.projects = sortProjects(Array.from(merged.values()));
  }

  async function loadReadiness() {
    state.readiness.clear();
    const projects = state.projects.filter((project) => isAssignedToMe(project) || isCreatedByMe(project));
    await Promise.all(projects.map(async (project) => {
      try {
        const readiness = await request("GET", `/v2/projects/${project.id}/readiness`, undefined, { skipAccessAlert: true });
        state.readiness.set(String(project.id), readiness);
      } catch (_) {
        state.readiness.set(String(project.id), null);
      }
    }));
  }

  function updateStats(items) {
    if (!ui.statTotal) return;
    ui.statTotal.textContent = String(items.length);
    ui.statReview.textContent = String(items.filter(isInvitePending).length);
    ui.statActive.textContent = String(items.filter((project) => statusCode(project) === "ACTIVE").length);
    ui.statRecruitment.textContent = String(items.filter((project) => {
      const status = statusCode(project);
      return status === "GRADING" || status === "REVIEW";
    }).length);
  }

  function renderFocus(items) {
    if (!ui.focusList) return;

    const focus = [];
    items.forEach((project) => {
      const status = statusCode(project);
      const ready = state.readiness.get(String(project.id || ""));

      if (isInvitePending(project)) {
        focus.push({
          title: project.title || "Без названия",
          text: "Нужно ответить на приглашение преподавателя-ревьюера.",
          actions: [
            { label: "Принять", act: "accept", id: project.id, primary: true },
            { label: "Отклонить", act: "reject", id: project.id, danger: true },
          ],
        });
        return;
      }

      if (canActivateProject(project)) {
        focus.push({
          title: project.title || "Без названия",
          text: "Команда готова к старту. Можно разрешить запуск проекта.",
          actions: [
            { label: "Дать разрешение на запуск и запустить", act: "activate", id: project.id, primary: true },
            { label: "Открыть критерии", act: "criteria", id: project.id, primary: false },
          ],
        });
        return;
      }

      if (isAcceptedReviewer(project) && ready && Number(ready.criteria_count || 0) === 0) {
        focus.push({
          title: project.title || "Без названия",
          text: "У проекта еще нет критериев. Без них команда не выйдет в понятный ревью-поток.",
          actions: [
            { label: "Открыть критерии", act: "criteria", id: project.id, primary: true },
          ],
        });
        return;
      }

      if (isAcceptedReviewer(project) && status === "GRADING") {
        focus.push({
          title: project.title || "Без названия",
          text: "Команда завершила работу и отправила проект на итоговое оценивание.",
          actions: [
            { label: "Оценить сейчас", act: "grade", id: project.id, primary: true },
          ],
        });
        return;
      }

      if (isAcceptedReviewer(project) && status === "ACTIVE") {
        focus.push({
          title: project.title || "Без названия",
          text: "Проект в активной фазе. Здесь полезнее открыть карточку проекта и посмотреть динамику команды.",
          actions: [
            { label: "Открыть проект", act: "open", id: project.id, primary: false },
          ],
        });
      }
    });

    if (!focus.length) {
      ui.focusList.innerHTML = `
        <article class="focus-empty">
          <span class="material-symbols-outlined" aria-hidden="true">done_all</span>
          <strong>Срочных действий нет</strong>
          <p>Сейчас можно спокойно посмотреть каталог проектов или вернуться к заявкам на ревью.</p>
        </article>
      `;
      return;
    }

    ui.focusList.innerHTML = focus.slice(0, 4).map((item) => `
      <article class="focus-item">
        <div class="focus-item__body">
          <strong>${escapeHTML(item.title)}</strong>
          <p>${escapeHTML(item.text)}</p>
        </div>
        <div class="focus-item__actions">
          ${item.actions.map(renderActionButton).join("")}
        </div>
      </article>
    `).join("");
  }

  function renderProjects(items) {
    if (!ui.projectsBody) return;
    if (!items.length) {
      ui.projectsBody.innerHTML = `
        <article class="prof-project-card prof-project-card--empty">
          <div class="prof-project-card__meta">
            <strong>Проекты пока не найдены</strong>
            <p>Когда в учебном контуре появятся проекты, они отобразятся здесь автоматически.</p>
          </div>
        </article>
      `;
      return;
    }

    ui.projectsBody.innerHTML = items.map((project) => {
      const status = statusMeta(project.status);
      const reviewer = professorMeta(project);
      const primary = projectPrimaryAction(project);
      primary.id = project.id;
      const secondary = projectSecondaryActions(project)
        .filter((action) => action.act !== primary.act)
        .map((action) => ({ ...action, id: project.id }));

      return `
        <article class="prof-project-card">
          <div class="prof-project-card__head">
            <div class="prof-project-card__meta">
              <div class="prof-project-card__labels">
                <span class="status-pill ${escapeHTML(status.tone)}">${escapeHTML(status.label)}</span>
                <span class="status-pill status-pill--subtle ${escapeHTML(reviewer.tone)}">${escapeHTML(reviewer.label)}</span>
              </div>
              <strong class="prof-project-card__title">${escapeHTML(project.title || "Без названия")}</strong>
              <p class="prof-project-card__desc">${escapeHTML(project.description || "Описание проекта пока не заполнено.")}</p>
            </div>
            <div class="prof-project-card__aside">
              <span class="prof-mini-label">Обновлен</span>
              <strong>${escapeHTML(formatDate(project.updated_at || project.created_at))}</strong>
            </div>
          </div>

          <div class="prof-project-card__narrative">
            ${escapeHTML(projectNarrative(project))}
          </div>

          <div class="prof-project-card__checks">
            ${buildReadinessBadges(project)}
          </div>

          <div class="prof-project-card__actions">
            ${renderActionButton(primary)}
            ${secondary.map(renderActionButton).join("")}
          </div>
        </article>
      `;
    }).join("");
  }

  async function actionRespondProfessorInvite(projectID, accept) {
    await request("POST", `/v2/projects/${projectID}/professor/respond`, { accept: Boolean(accept) });
    setStatus(accept ? "Приглашение на ревью принято." : "Приглашение на ревью отклонено.", false);
  }

  async function actionActivateProject(projectID) {
    await request("POST", `/v2/projects/${projectID}/approve`, {});
    setStatus("Разрешение на запуск выдано. Проект переведен в активную фазу.", false);
  }

  function actionOpenCriteria(projectID) {
    window.location.href = `/dev/professor/criteria?project_id=${encodeURIComponent(projectID)}`;
  }

  function actionOpenGrading(projectID) {
    window.location.href = `/dev/professor/grading?project_id=${encodeURIComponent(projectID)}`;
  }

  function actionOpenReport(projectID) {
    const url = new URL(`/v2/projects/${encodeURIComponent(projectID)}/final-report.pdf`, window.location.origin);
    const lang = i18n && typeof i18n.getLanguage === "function"
      ? String(i18n.getLanguage() || "").trim().toLowerCase()
      : "";
    if (lang) {
      url.searchParams.set("lang", lang);
    }
    const popup = window.open(url.toString(), "_blank", "noopener,noreferrer");
    if (!popup) {
      window.location.href = url.toString();
    }
  }

  async function handleAction(act, projectID) {
    if (!act || !projectID) return;
    try {
      if (act === "open") {
        window.location.href = `/dev/projects/${projectID}`;
        return;
      }
      if (act === "criteria") {
        actionOpenCriteria(projectID);
        return;
      }
      if (act === "grade") {
        actionOpenGrading(projectID);
        return;
      }
      if (act === "report") {
        actionOpenReport(projectID);
        return;
      }
      if (act === "activate") {
        await actionActivateProject(projectID);
      } else if (act === "accept") {
        await actionRespondProfessorInvite(projectID, true);
      } else if (act === "reject") {
        await actionRespondProfessorInvite(projectID, false);
      }
      await refreshPage();
    } catch (err) {
      setStatus(err.message || String(err), true);
    }
  }

  function attachEvents() {
    if (ui.refreshBtn) {
      ui.refreshBtn.addEventListener("click", () => {
        void refreshPage();
      });
    }

    [ui.projectsBody, ui.focusList].forEach((host) => {
      if (!host) return;
      host.addEventListener("click", (event) => {
        const btn = event.target.closest("button[data-act][data-id]");
        if (!btn) return;
        void handleAction(btn.dataset.act, btn.dataset.id);
      });
    });
  }

  async function refreshPage() {
    try {
      setStatus("Обновляю контур преподавателя...", false);
      await loadProjects();
      await loadReadiness();

      const focusProjects = relevantProjects();
      updateStats(state.projects);
      renderFocus(focusProjects);
      renderProjects(state.projects);
      setStatus(`Вижу проектов в учебном контуре: ${state.projects.length}.`, false);
    } catch (err) {
      setStatus(err.message || String(err), true);
    }
  }

  void (async () => {
    try {
      state.profile = await auth.ensureSession("professor");
      if (!state.profile) return;

      renderGreeting(state.profile);
      attachEvents();
      await refreshPage();
      auth.setPageLoading(false);
    } catch (err) {
      setStatus(err.message || String(err), true);
      auth.setPageLoading(false);
    }
  })();
})();
