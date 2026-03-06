(() => {
  if (window.__idsaiNotificationsPopupInitialized) {
    return;
  }
  window.__idsaiNotificationsPopupInitialized = true;

  const LS_ACCESS = "idsai_access_token";
  const LS_REFRESH = "idsai_refresh_token";
  const API_LIST = "/v2/notifications?limit=100&offset=0";
  const API_MARK_READ = (id) => `/v2/notifications/${id}/read`;
  const API_MARK_ALL_READ = "/v2/notifications/read-all";
  const API_DELETE = (id) => `/v2/notifications/${id}`;
  const API_CLEAR = "/v2/notifications";
  const POLL_MS = 3000;
  const TOAST_TTL_MS = 2600;

  const state = {
    items: [],
    shownToastIDs: new Set(),
    isPanelOpen: false,
    unauthorized: false,
  };

  const layer = ensureLayer();
  const bell = ensureBell();
  const panel = ensurePanel();
  if (!layer || !bell || !panel) return;

  const panelList = panel.querySelector("#idsaiNotifyList");
  let pollTimer = null;

  function token() {
    const access = localStorage.getItem(LS_ACCESS) || "";
    if (!access) return "";
    if (isTokenExpired(access)) {
      localStorage.removeItem(LS_ACCESS);
      localStorage.removeItem(LS_REFRESH);
      return "";
    }
    return access;
  }

  function authHeaders() {
    const access = token();
    if (!access) return {};
    return { Authorization: `Bearer ${access}` };
  }

  function handleUnauthorized() {
    if (state.unauthorized) return;
    state.unauthorized = true;
    if (pollTimer) {
      clearInterval(pollTimer);
      pollTimer = null;
    }
    localStorage.removeItem(LS_ACCESS);
    localStorage.removeItem(LS_REFRESH);
    if (!window.location.pathname.startsWith("/dev/login")) {
      window.location.href = "/dev/login";
    }
  }

  function isTokenExpired(jwt) {
    const token = String(jwt || "").trim();
    if (!token) return true;
    const parts = token.split(".");
    if (parts.length < 2) return true;
    try {
      let payload = parts[1].replace(/-/g, "+").replace(/_/g, "/");
      const mod = payload.length % 4;
      if (mod > 0) payload += "=".repeat(4 - mod);
      const decoded = JSON.parse(atob(payload));
      const exp = Number(decoded.exp);
      if (!Number.isFinite(exp) || exp <= 0) return false;
      return Date.now() >= exp * 1000;
    } catch (_) {
      return true;
    }
  }

  async function fetchJSON(url, options = {}) {
    const resp = await fetch(url, options);
    const text = await resp.text();
    let data = {};
    try {
      data = text ? JSON.parse(text) : {};
    } catch (_) {
      data = {};
    }
    return { resp, data };
  }

  async function apiNoBody(method, url) {
    return fetchJSON(url, {
      method,
      headers: authHeaders(),
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

  function normalizeItems(items) {
    if (!Array.isArray(items)) return [];
    return items
      .filter((item) => item && item.id)
      .sort((a, b) => Date.parse(b.created_at || 0) - Date.parse(a.created_at || 0));
  }

  async function loadNotifications() {
    if (state.unauthorized) return [];
    if (!token()) return [];
    const { resp, data } = await fetchJSON(API_LIST, { headers: authHeaders() });
    if (resp.status === 401) {
      handleUnauthorized();
      return [];
    }
    if (!resp.ok) return [];
    return normalizeItems(Array.isArray(data.items) ? data.items : []);
  }

  function formatDate(value) {
    if (!value) return "только что";
    const ts = Date.parse(value);
    if (!Number.isFinite(ts)) return "только что";
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
      if (!item || item.id !== id) return item;
      return { ...item, is_read: true, read_at: item.read_at || new Date().toISOString() };
    });
  }

  async function markRead(id) {
    if (state.unauthorized) return false;
    if (!id || !token()) return false;
    const { resp } = await apiNoBody("POST", API_MARK_READ(id));
    if (resp.status === 401) {
      handleUnauthorized();
      return false;
    }
    if (!resp.ok && resp.status !== 404) return false;
    updateLocalRead(id);
    renderPanel();
    setBadge(unreadCount());
    return true;
  }

  async function markAllRead() {
    if (state.unauthorized) return;
    if (!token()) return;
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
    if (!id || !token()) return;
    const { resp } = await apiNoBody("DELETE", API_DELETE(id));
    if (resp.status === 401) {
      handleUnauthorized();
      return;
    }
    if (!resp.ok && resp.status !== 404) return;
    state.items = state.items.filter((item) => item && item.id !== id);
    renderPanel();
    setBadge(unreadCount());
  }

  async function clearAllNotifications() {
    if (state.unauthorized) return;
    if (!token()) return;
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

  function panelItemMarkup(item) {
    const title = escapeHTML(item.title || "Уведомление");
    const body = escapeHTML(item.body || "Новое событие в системе.");
    const createdAt = escapeHTML(formatDate(item.created_at));
    const itemCls = item.is_read ? "idsai-notify-item is-read" : "idsai-notify-item";
    return `
      <article class="${itemCls}" data-id="${escapeHTML(item.id)}">
        <button class="idsai-notify-delete" type="button" data-act="delete" aria-label="Удалить">×</button>
        <button class="idsai-notify-open" type="button" data-act="open">
          <h4>${title}</h4>
          <p>${body}</p>
          <time>${createdAt}</time>
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
    return `
      <article class="idsai-toast idsai-toast--light" data-id="${escapeHTML(item.id)}">
        <div class="idsai-toast-head">
          <div class="idsai-toast-title-wrap">
            <span class="idsai-toast-icon material-symbols-outlined" aria-hidden="true">info</span>
            <h4 class="idsai-toast-title">${title}</h4>
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
    if (!token()) return;
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
    if (!token()) return;
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
