(() => {
  const auth = window.IDSAIAuth;
  const i18n = window.IDSAI18n;
  if (window.__idsaiNotificationsPopupInitialized) {
    return;
  }
  window.__idsaiNotificationsPopupInitialized = true;

  const API_LIST = "/v2/notifications?limit=100&offset=0";
  const API_MARK_READ = (id) => `/v2/notifications/${id}/read`;
  const API_MARK_ALL_READ = "/v2/notifications/read-all";
  const API_DELETE = (id) => `/v2/notifications/${id}`;
  const API_CLEAR = "/v2/notifications";
  const POLL_MS = 3000;
  const TOAST_TTL_MS = 2600;
  const DEDUPE_WINDOW_MS = 15000;
  const SHOWN_TOASTS_STORAGE_KEY = "idsai_notifications_shown_toasts";
  const MAX_SHOWN_TOASTS = 200;

  const state = {
    items: [],
    shownToastIDs: loadShownToastIDs(),
    isPanelOpen: false,
    unauthorized: false,
  };

  const layer = ensureLayer();
  const bell = ensureBell();
  const panel = ensurePanel();
  if (!layer || !bell || !panel) return;

  const panelList = panel.querySelector("#idsaiNotifyList");
  let pollTimer = null;

  function confirmAction(options) {
    if (auth && typeof auth.showConfirmDialog === "function") {
      return auth.showConfirmDialog(options);
    }
    return Promise.resolve(window.confirm(String((options && options.message) || "")));
  }

  function handleUnauthorized() {
    if (state.unauthorized) return;
    state.unauthorized = true;
    if (pollTimer) {
      clearInterval(pollTimer);
      pollTimer = null;
    }
    auth.clearClientState();
    if (!window.location.pathname.startsWith("/dev/login")) {
      window.location.href = "/dev/login";
    }
  }

  async function apiNoBody(method, url) {
    return auth.requestJSON(url, {
      method,
    });
  }

  function ensureLayer() {
    let el = document.getElementById("idsaiToastLayer");
    if (el) return el;
    el = document.createElement("section");
    el.id = "idsaiToastLayer";
    el.className = "idsai-toast-layer";
    el.setAttribute("aria-live", "polite");
    document.body.appendChild(el);
    return el;
  }

  function ensureBell() {
    const topbarActions = document.querySelector(".topbar-actions");
    if (!topbarActions) return null;
    let btn = document.getElementById("idsaiToastBell");
    if (btn) return btn;

    btn = document.createElement("button");
    btn.id = "idsaiToastBell";
    btn.className = "idsai-toast-bell";
    btn.type = "button";
    btn.setAttribute("aria-label", "Открыть уведомления");
    btn.setAttribute("aria-haspopup", "dialog");
    btn.setAttribute("aria-expanded", "false");
    btn.innerHTML =
      '<span class="material-symbols-outlined" aria-hidden="true">notifications</span><span id="idsaiToastBadge" class="idsai-toast-badge" hidden>0</span>';
    btn.addEventListener("click", (event) => {
      event.stopPropagation();
      togglePanel();
    });
    topbarActions.appendChild(btn);
    return btn;
  }

  function ensurePanel() {
    let el = document.getElementById("idsaiNotifyPanel");
    if (el) return el;
    el = document.createElement("section");
    el.id = "idsaiNotifyPanel";
    el.className = "idsai-notify-panel";
    el.hidden = true;
    el.setAttribute("role", "dialog");
    el.setAttribute("aria-label", "Уведомления");
    el.innerHTML = `
      <header class="idsai-notify-head">
        <h3>Уведомления</h3>
        <div class="idsai-notify-actions">
          <button type="button" data-act="mark-all">Прочитать все</button>
          <button type="button" data-act="clear-all">Очистить</button>
        </div>
      </header>
      <div id="idsaiNotifyList" class="idsai-notify-list"></div>
    `;
    document.body.appendChild(el);
    bindPanelActions(el);
    return el;
  }

  function badgeEl() {
    return document.getElementById("idsaiToastBadge");
  }

  function setBadge(count) {
    const badge = badgeEl();
    if (!badge) return;
    const n = Number(count) || 0;
    badge.textContent = String(n);
    badge.hidden = n <= 0;
  }

  function unreadCount() {
    return state.items.reduce((acc, item) => acc + (item && item.is_read === false ? 1 : 0), 0);
  }

  function loadShownToastIDs() {
    try {
      const raw = sessionStorage.getItem(SHOWN_TOASTS_STORAGE_KEY);
      if (!raw) return new Set();
      const parsed = JSON.parse(raw);
      if (!Array.isArray(parsed)) return new Set();
      return new Set(parsed.filter((item) => typeof item === "string" && item));
    } catch (_) {
      return new Set();
    }
  }

  function persistShownToastIDs() {
    try {
      sessionStorage.setItem(
        SHOWN_TOASTS_STORAGE_KEY,
        JSON.stringify(Array.from(state.shownToastIDs).slice(-MAX_SHOWN_TOASTS)),
      );
    } catch (_) {}
  }

  function notificationFingerprint(item) {
    const payload = parsePayload(item?.payload);
    const projectID = String(payload?.project_id || "").trim().toLowerCase();
    return [
      String(item?.type || "").trim().toLowerCase(),
      String(item?.title || "").trim().toLowerCase(),
      String(item?.body || "").trim().toLowerCase(),
      projectID,
    ].join("|");
  }

  function normalizeItems(items) {
    if (!Array.isArray(items)) return [];
    const dedupeState = new Map();
    const normalized = items
      .filter((item) => item && item.id)
      .sort((a, b) => Date.parse(b.created_at || 0) - Date.parse(a.created_at || 0));
    const out = [];

    normalized.forEach((item) => {
      const candidate = { ...item, duplicate_ids: [] };
      const fingerprint = notificationFingerprint(candidate);
      const createdAt = Date.parse(candidate.created_at || 0);
      const prev = dedupeState.get(fingerprint);
      if (
        fingerprint &&
        prev &&
        Number.isFinite(prev.createdAt) &&
        Number.isFinite(createdAt) &&
        prev.createdAt-createdAt <= DEDUPE_WINDOW_MS
      ) {
        prev.item.duplicate_ids.push(candidate.id);
        return;
      }
      dedupeState.set(fingerprint, { item: candidate, createdAt });
      out.push(candidate);
    });

    return out;
  }

  function notificationIDs(id) {
    const item = state.items.find((entry) => entry && entry.id === id);
    if (!item) return id ? [id] : [];
    return [item.id, ...(Array.isArray(item.duplicate_ids) ? item.duplicate_ids : [])]
      .filter((value, index, arr) => value && arr.indexOf(value) === index);
  }

  async function loadNotifications() {
    if (state.unauthorized) return [];
    const { resp, data } = await auth.requestJSON(API_LIST, { method: "GET" });
    if (resp.status === 401) {
      handleUnauthorized();
      return [];
    }
    if (!resp.ok) return [];
    return normalizeItems(Array.isArray(data.items) ? data.items : []);
  }

  function formatDate(value) {
    if (!value) return i18n ? i18n.t("только что") : "только что";
    const ts = Date.parse(value);
    if (!Number.isFinite(ts)) return i18n ? i18n.t("только что") : "только что";
    if (i18n) {
      return i18n.formatDateTime(ts, {
        day: "2-digit",
        month: "2-digit",
        hour: "2-digit",
        minute: "2-digit",
      });
    }
    return new Date(ts).toLocaleString("ru-RU", {
      day: "2-digit",
      month: "2-digit",
      hour: "2-digit",
      minute: "2-digit",
    });
  }

  function parsePayload(raw) {
    if (!raw) return {};
    if (typeof raw === "object") return raw;
    try {
      return JSON.parse(raw);
    } catch (_) {
      return {};
    }
  }

  function openFromPayload(payload) {
    const parsed = parsePayload(payload);
    if (parsed && parsed.project_id) {
      window.location.href = `/dev/projects/${parsed.project_id}`;
      return;
    }
    window.location.href = "/dev/projects";
  }

  function updateLocalRead(id) {
    state.items = state.items.map((item) => {
      if (!item || !notificationIDs(id).includes(item.id)) return item;
      return { ...item, is_read: true, read_at: item.read_at || new Date().toISOString() };
    });
  }

  async function markRead(id) {
    if (state.unauthorized) return false;
    if (!id) return false;
    for (const currentID of notificationIDs(id)) {
      const { resp } = await apiNoBody("POST", API_MARK_READ(currentID));
      if (resp.status === 401) {
        handleUnauthorized();
        return false;
      }
      if (!resp.ok && resp.status !== 404) return false;
    }
    updateLocalRead(id);
    renderPanel();
    setBadge(unreadCount());
    return true;
  }

  async function markAllRead() {
    if (state.unauthorized) return;
    const { resp } = await apiNoBody("POST", API_MARK_ALL_READ);
    if (resp.status === 401) {
      handleUnauthorized();
      return;
    }
    if (!resp.ok) return;
    const now = new Date().toISOString();
    state.items = state.items.map((item) => ({ ...item, is_read: true, read_at: item.read_at || now }));
    renderPanel();
    setBadge(0);
  }

  async function deleteNotification(id) {
    if (state.unauthorized) return;
    if (!id) return;
    for (const currentID of notificationIDs(id)) {
      const { resp } = await apiNoBody("DELETE", API_DELETE(currentID));
      if (resp.status === 401) {
        handleUnauthorized();
        return;
      }
      if (!resp.ok && resp.status !== 404) return;
    }
    state.items = state.items.filter((item) => item && item.id !== id);
    renderPanel();
    setBadge(unreadCount());
  }

  async function clearAllNotifications() {
    if (state.unauthorized) return;
    if (!state.items.length) return;
    if (!await confirmAction({
      title: "Очистить уведомления",
      message: "Вся история уведомлений будет удалена из списка. Продолжить?",
      confirmText: "Очистить",
      danger: true,
    })) {
      return;
    }
    const { resp } = await apiNoBody("DELETE", API_CLEAR);
    if (resp.status === 401) {
      handleUnauthorized();
      return;
    }
    if (!resp.ok) return;
    state.items = [];
    renderPanel();
    setBadge(0);
  }

  function escapeHTML(value) {
    return String(value || "")
      .replaceAll("&", "&amp;")
      .replaceAll("<", "&lt;")
      .replaceAll(">", "&gt;")
      .replaceAll('"', "&quot;")
      .replaceAll("'", "&#039;");
  }

  const TYPE_REASONS = {
    "project.member.application.rejected": "Тимлид не принял заявку в команду",
    "project.member.removed": "Тимлид отозвал доступ к проекту",
    "project.submission.rejected": "Работа не прошла проверку",
    "project.grade.failed": "Оценка не зачтена",
    "project.submission.retake": "Требуется исправить и пересдать работу",
  };

  function notificationReason(item) {
    if (!item?.type) return "";
    return TYPE_REASONS[String(item.type).toLowerCase()] || "";
  }

  function notificationTone(item) {
    const type = String(item?.type || "").toLowerCase();

    if (/\.(rejected|removed|denied|failed|error)(\.|$)/.test(type)) return "error";
    if (/\.(retake|warning)(\.|$)/.test(type) || type.includes("retake")) return "warning";
    if (/\.(accepted|approved|created|updated|activated|published|finished|sent_to_grading)(\.|$)/.test(type)) return "success";

    // fallback: check only type + title, never body (avoid false positives like "отклонить")
    const safeText = `${type} ${String(item?.title || "").toLowerCase()}`;
    if (safeText.includes("retake") || safeText.includes("пересдач")) return "warning";
    if (safeText.includes("rejected") || safeText.includes("removed")) return "error";
    if (
      safeText.includes("accepted") || safeText.includes("approved") ||
      safeText.includes("created") || safeText.includes("updated") ||
      safeText.includes("activated") || safeText.includes("published") ||
      safeText.includes("finished") || safeText.includes("принят") ||
      safeText.includes("создан") || safeText.includes("заверш") ||
      safeText.includes("опублик")
    ) return "success";

    return "info";
  }

  function notificationIcon(tone) {
    if (tone === "success") return "check_circle";
    if (tone === "warning") return "warning";
    if (tone === "error") return "error";
    return "info";
  }

  function notificationToneLabel(tone) {
    if (tone === "success") return "Успех";
    if (tone === "warning") return "Внимание";
    if (tone === "error") return "Ошибка";
    return "Инфо";
  }

  function panelItemMarkup(item) {
    const title = escapeHTML(item.title || "Уведомление");
    const body = escapeHTML(item.body || "Новое событие в системе.");
    const createdAt = escapeHTML(formatDate(item.created_at));
    const tone = notificationTone(item);
    const icon = notificationIcon(tone);
    const toneLabel = notificationToneLabel(tone);
    const reason = notificationReason(item);
    const reasonHTML = reason
      ? `<span class="idsai-notify-reason">${escapeHTML(reason)}</span>`
      : "";
    const itemCls = item.is_read
      ? `idsai-notify-item idsai-notify-item--${tone} is-read`
      : `idsai-notify-item idsai-notify-item--${tone}`;
    return `
      <article class="${itemCls}" data-id="${escapeHTML(item.id)}">
        <button class="idsai-notify-delete" type="button" data-act="delete" aria-label="Удалить">×</button>
        <button class="idsai-notify-open" type="button" data-act="open">
          <span class="idsai-notify-item-icon material-symbols-outlined" aria-hidden="true">${icon}</span>
          <span class="idsai-notify-content">
            <span class="idsai-notify-pill">${toneLabel}</span>
            <h4>${title}</h4>
            <p>${body}</p>
            ${reasonHTML}
            <time>${createdAt}</time>
          </span>
        </button>
      </article>
    `;
  }

  function renderPanel() {
    if (!panelList) return;
    if (!state.items.length) {
      panelList.innerHTML = '<div class="idsai-notify-empty">Пока нет уведомлений</div>';
      return;
    }
    panelList.innerHTML = state.items.map(panelItemMarkup).join("");
  }

  function toastMarkup(item) {
    const title = escapeHTML(item.title || "Уведомление");
    const body = escapeHTML(item.body || "Новое событие в системе.");
    const tone = notificationTone(item);
    const icon = notificationIcon(tone);
    const toneLabel = notificationToneLabel(tone);
    return `
      <article class="idsai-toast idsai-toast--${tone}" data-id="${escapeHTML(item.id)}">
        <div class="idsai-toast-head">
          <div class="idsai-toast-title-wrap">
            <span class="idsai-toast-icon material-symbols-outlined" aria-hidden="true">${icon}</span>
            <div class="idsai-toast-text">
              <span class="idsai-toast-pill">${toneLabel}</span>
              <h4 class="idsai-toast-title">${title}</h4>
            </div>
          </div>
          <button type="button" class="idsai-toast-close" data-act="close" aria-label="Закрыть">×</button>
        </div>
        <p class="idsai-toast-body">${body}</p>
        <div class="idsai-toast-actions">
          <button type="button" class="idsai-toast-learn" data-act="open">Открыть</button>
        </div>
      </article>
    `;
  }

  function removeToastCard(card) {
    if (!card) return;
    card.style.opacity = "0";
    card.style.transform = "translateY(-6px)";
    setTimeout(() => card.remove(), 150);
  }

  function pushToast(item) {
    if (!item || !item.id || state.shownToastIDs.has(item.id)) return;
    state.shownToastIDs.add(item.id);
    persistShownToastIDs();
    layer.insertAdjacentHTML("beforeend", toastMarkup(item));
    const card = layer.querySelector(`.idsai-toast[data-id="${item.id}"]`);
    if (!card) return;

    const closeBtn = card.querySelector('[data-act="close"]');
    const openBtn = card.querySelector('[data-act="open"]');

    closeBtn?.addEventListener("click", async () => {
      await markRead(item.id);
      removeToastCard(card);
    });
    openBtn?.addEventListener("click", async () => {
      await markRead(item.id);
      openFromPayload(item.payload);
    });

    setTimeout(() => {
      if (card.isConnected) removeToastCard(card);
    }, TOAST_TTL_MS);
  }

  function showNewToasts() {
    const unread = state.items.filter((item) => item && item.is_read === false);
    unread
      .filter((item) => !state.shownToastIDs.has(item.id))
      .slice(0, 3)
      .forEach((item) => pushToast(item));
  }

  function setPanelOpen(open) {
    state.isPanelOpen = !!open;
    panel.hidden = !state.isPanelOpen;
    bell.setAttribute("aria-expanded", state.isPanelOpen ? "true" : "false");
    if (state.isPanelOpen) renderPanel();
  }

  function togglePanel() {
    setPanelOpen(!state.isPanelOpen);
  }

  async function sync() {
    if (state.unauthorized) return;
    const items = await loadNotifications();
    state.items = items;
    renderPanel();
    setBadge(unreadCount());
    showNewToasts();
  }

  function bindPanelActions(el) {
    el.addEventListener("click", async (event) => {
      const target = event.target;
      if (!(target instanceof Element)) return;

      const actionEl = target.closest("[data-act]");
      if (!actionEl) return;

      const action = actionEl.getAttribute("data-act");
      if (action === "mark-all") {
        await markAllRead();
        return;
      }
      if (action === "clear-all") {
        await clearAllNotifications();
        return;
      }

      const itemEl = actionEl.closest(".idsai-notify-item");
      const id = itemEl?.getAttribute("data-id") || "";
      if (!id) return;

      if (action === "delete") {
        await deleteNotification(id);
        return;
      }
      if (action === "open") {
        await markRead(id);
        const item = state.items.find((entry) => entry && entry.id === id);
        openFromPayload(item?.payload);
      }
    });
  }

  async function start() {
    const profile = await auth.ensureSession(undefined, { redirectOnMissing: false });
    if (!profile) return;
    renderPanel();
    await sync();
    pollTimer = setInterval(() => {
      void sync();
    }, POLL_MS);
  }

  document.addEventListener("click", (event) => {
    if (!state.isPanelOpen) return;
    const target = event.target;
    if (!(target instanceof Node)) return;
    if (panel.contains(target) || bell.contains(target)) return;
    setPanelOpen(false);
  });

  document.addEventListener("keydown", (event) => {
    if (event.key === "Escape" && state.isPanelOpen) {
      setPanelOpen(false);
    }
  });

  void start();
})();
