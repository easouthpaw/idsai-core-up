(() => {
  const auth = window.IDSAIAuth;
  const i18n = window.IDSAI18n;
  const GUIDE_STORAGE_KEY = "idsai_professor_reviews_guide_hidden";

  const ui = {
    reviewsBody: document.getElementById("reviewsBody"),
    refreshBtn: document.getElementById("refreshBtn"),
    pageStatus: document.getElementById("pageStatus"),
    guidePanel: document.getElementById("profGuidePanel"),
    guideSteps: document.getElementById("profGuideSteps"),
    guideToggleBtn: document.getElementById("profGuideToggleBtn"),
    guideRestoreBtn: document.getElementById("profGuideRestoreBtn"),
  };

  const state = {
    invites: [],
    loading: false,
  };

  function escapeHTML(value) {
    return String(value || "")
      .replaceAll("&", "&amp;")
      .replaceAll("<", "&lt;")
      .replaceAll(">", "&gt;")
      .replaceAll('"', "&quot;")
      .replaceAll("'", "&#039;");
  }

  async function request(method, url, body) {
    const options = { method };
    if (body !== undefined) {
      options.body = JSON.stringify(body);
      options.headers = { "Content-Type": "application/json" };
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

  function setStatus(message, isError) {
    if (!ui.pageStatus) return;
    ui.pageStatus.textContent = message || "";
    ui.pageStatus.classList.toggle("err", Boolean(isError));
  }

  function formatDate(value) {
    if (!value) return "—";
    try {
      return i18n ? i18n.formatDateTime(value) : new Date(value).toLocaleString("ru-RU");
    } catch (_) {
      return String(value);
    }
  }

  function statusMeta(status) {
    const code = String(status || "DRAFT").toUpperCase();
    if (code === "REVIEW") return { label: "Подготовка", tone: "review" };
    if (code === "RECRUITMENT") return { label: "Набор команды", tone: "recruitment" };
    if (code === "ACTIVE") return { label: "В работе", tone: "active" };
    if (code === "GRADING") return { label: "На оценке", tone: "grading" };
    if (code === "COMPLETED") return { label: "Завершен", tone: "done" };
    return { label: "Черновик", tone: "default" };
  }

  function guideHidden() {
    try {
      return localStorage.getItem(GUIDE_STORAGE_KEY) === "1";
    } catch (_) {
      return false;
    }
  }

  function applyGuideVisibility() {
    const hidden = guideHidden();
    if (ui.guidePanel) ui.guidePanel.hidden = hidden;
    if (ui.guideRestoreBtn) ui.guideRestoreBtn.hidden = !hidden;
    if (ui.guideToggleBtn) ui.guideToggleBtn.hidden = hidden;
  }

  function setGuideHidden(hidden) {
    try {
      if (hidden) {
        localStorage.setItem(GUIDE_STORAGE_KEY, "1");
      } else {
        localStorage.removeItem(GUIDE_STORAGE_KEY);
      }
    } catch (_) {}
    applyGuideVisibility();
  }

  function renderGuide() {
    if (!ui.guideSteps) return;
    const count = state.invites.length;
    const firstProjectID = count > 0 ? String(state.invites[0].id || "") : "";
    const steps = [
      {
        tone: count > 0 ? "current" : "done",
        kicker: "Шаг 1",
        title: count > 0 ? `Новых заявок: ${count}` : "Очередь пустая",
        text: count > 0
          ? "Сначала проверьте, действительно ли вы готовы сопровождать эти проекты."
          : "Сейчас новых приглашений нет. Здесь будут появляться только заявки, которые ждут вашего решения.",
        actions: [{ act: "refresh", label: "Обновить" }],
      },
      {
        tone: count > 0 ? "current" : "done",
        kicker: "Шаг 2",
        title: "Примите или отклоните приглашение",
        text: "После принятия проект исчезнет из этой очереди и появится в ваших рабочих страницах критериев и оценивания.",
        actions: firstProjectID ? [{ act: "open-project", projectID: firstProjectID, label: "Открыть первый проект" }] : [],
      },
      {
        tone: count > 0 ? "blocked" : "done",
        kicker: "Шаг 3",
        title: "Дальше переходите в критерии",
        text: "Следующий преподавательский шаг после принятия заявки: собрать критерии и затем дождаться этапа финального ревью.",
        actions: [{ act: "dashboard", label: "К дашборду" }],
      },
    ];

    ui.guideSteps.innerHTML = steps.map((step) => (
      `<article class="prof-guide-step prof-guide-step--${escapeHTML(step.tone)}">` +
        `<small>${escapeHTML(step.kicker)}</small>` +
        `<strong>${escapeHTML(step.title)}</strong>` +
        `<p>${escapeHTML(step.text)}</p>` +
        `<div class="prof-guide-step__actions">` +
          (Array.isArray(step.actions) ? step.actions.map((action) => (
            `<button class="ghost-btn" type="button" data-guide-act="${escapeHTML(action.act)}" data-project-id="${escapeHTML(action.projectID || "")}">${escapeHTML(action.label)}</button>`
          )).join("") : "") +
        `</div>` +
      `</article>`
    )).join("");
  }

  function renderTable() {
    if (!ui.reviewsBody) return;
    if (state.loading) {
      ui.reviewsBody.innerHTML = '<tr><td colspan="4">Загрузка заявок...</td></tr>';
      return;
    }
    if (!state.invites.length) {
      ui.reviewsBody.innerHTML = '<tr><td colspan="4">Новых заявок на ревью пока нет.</td></tr>';
      return;
    }

    ui.reviewsBody.innerHTML = state.invites.map((project) => {
      const status = statusMeta(project.status);
      const owner = project.created_by_name || project.created_by_email || "Команда";
      return (
        `<tr>` +
          `<td>` +
            `<strong>${escapeHTML(project.title || "Без названия")}</strong>` +
            `<div class="muted">Обновлен: ${escapeHTML(formatDate(project.updated_at || project.created_at))}</div>` +
          `</td>` +
          `<td>` +
            `${escapeHTML(owner)}` +
            `<div class="muted">${escapeHTML(project.description || "Описание пока не заполнено.")}</div>` +
          `</td>` +
          `<td>` +
            `<span class="status-pill ${escapeHTML(status.tone)}">${escapeHTML(status.label)}</span>` +
            `<div class="muted">Ждет вашего ответа</div>` +
          `</td>` +
          `<td>` +
            `<div class="actions">` +
              `<button class="action-btn" type="button" data-row-act="open" data-project-id="${escapeHTML(project.id)}">Открыть</button>` +
              `<button class="action-btn primary" type="button" data-row-act="accept" data-project-id="${escapeHTML(project.id)}">Принять</button>` +
              `<button class="action-btn" type="button" data-row-act="reject" data-project-id="${escapeHTML(project.id)}">Отклонить</button>` +
            `</div>` +
          `</td>` +
        `</tr>`
      );
    }).join("");
  }

  async function loadInvites() {
    state.loading = true;
    renderTable();
    setStatus("Загружаю новые заявки на ревью...", false);
    try {
      const items = await request("GET", "/v2/professor/review-invites?limit=100");
      state.invites = Array.isArray(items) ? items : [];
      renderGuide();
      setStatus(
        state.invites.length > 0
          ? `Новых заявок: ${state.invites.length}.`
          : "Новых заявок на ревью пока нет.",
        false,
      );
    } catch (err) {
      state.invites = [];
      renderGuide();
      setStatus(err.message || String(err), true);
    } finally {
      state.loading = false;
      renderTable();
    }
  }

  async function respond(projectID, accept) {
    const project = state.invites.find((item) => String(item.id || "") === String(projectID || ""));
    if (!projectID || !project) return;
    setStatus(accept ? "Принимаю заявку..." : "Отклоняю заявку...", false);
    try {
      await request("POST", `/v2/projects/${projectID}/professor/respond`, { accept: Boolean(accept) });
      await loadInvites();
      setStatus(
        accept
          ? `Заявка принята: ${project.title || "проект"}. Теперь можно переходить к критериям и оцениванию.`
          : `Заявка отклонена: ${project.title || "проект"}.`,
        false,
      );
    } catch (err) {
      setStatus(err.message || String(err), true);
    }
  }

  function handleGuideAction(act, projectID) {
    if (act === "refresh") {
      void loadInvites();
      return;
    }
    if (act === "dashboard") {
      window.location.href = "/dev/professor";
      return;
    }
    if (act === "open-project" && projectID) {
      window.location.href = `/dev/projects/${encodeURIComponent(projectID)}`;
    }
  }

  function attachEvents() {
    if (ui.refreshBtn) {
      ui.refreshBtn.addEventListener("click", () => {
        void loadInvites();
      });
    }
    if (ui.guideToggleBtn) {
      ui.guideToggleBtn.addEventListener("click", () => {
        setGuideHidden(true);
      });
    }
    if (ui.guideRestoreBtn) {
      ui.guideRestoreBtn.addEventListener("click", () => {
        setGuideHidden(false);
      });
    }
    if (ui.guideSteps) {
      ui.guideSteps.addEventListener("click", (event) => {
        const btn = event.target.closest("button[data-guide-act]");
        if (!btn) return;
        handleGuideAction(btn.getAttribute("data-guide-act") || "", btn.getAttribute("data-project-id") || "");
      });
    }
    if (ui.reviewsBody) {
      ui.reviewsBody.addEventListener("click", (event) => {
        const btn = event.target.closest("button[data-row-act]");
        if (!btn) return;
        const action = btn.getAttribute("data-row-act") || "";
        const projectID = btn.getAttribute("data-project-id") || "";
        if (action === "open") {
          window.location.href = `/dev/projects/${encodeURIComponent(projectID)}`;
          return;
        }
        if (action === "accept") {
          void respond(projectID, true);
          return;
        }
        if (action === "reject") {
          void respond(projectID, false);
        }
      });
    }
  }

  void (async () => {
    try {
      const profile = await auth.ensureSession("professor");
      if (!profile) return;
      applyGuideVisibility();
      renderGuide();
      attachEvents();
      await loadInvites();
      auth.setPageLoading(false);
    } catch (err) {
      auth.setPageLoading(false);
      setStatus(err.message || String(err), true);
    }
  })();
})();
