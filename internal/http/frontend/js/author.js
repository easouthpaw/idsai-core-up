(() => {
  const form = document.getElementById("authorContactForm");
  const status = document.getElementById("authorFormStatus");

  if (!form || !status) return;

  function humanizeError(message) {
    switch (String(message || "").trim()) {
      case "contact is required":
        return "Укажите ваши контактные данные.";
      case "message is required":
        return "Опишите, чем я могу помочь.";
      case "failed to deliver message":
      case "contact delivery unavailable":
        return "Сообщение пока не отправилось. Попробуйте ещё раз чуть позже.";
      default:
        return String(message || "Не удалось отправить форму.");
    }
  }

  form.addEventListener("submit", async (event) => {
    event.preventDefault();

    const submitBtn = form.querySelector("button[type='submit']");
    const defaultSubmitLabel = submitBtn ? submitBtn.textContent.trim() : "";
    const formData = new FormData(form);
    const payload = {
      contact: String(formData.get("contact") || "").trim(),
      message: String(formData.get("message") || "").trim(),
    };

    if (!payload.contact || !payload.message) {
      status.textContent = "Заполните поле с контактами и опишите ваш запрос.";
      return;
    }

    if (submitBtn) {
      submitBtn.disabled = true;
      submitBtn.textContent = "Отправляем...";
    }
    status.textContent = "";

    try {
      const resp = await fetch("/v2/contact", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "X-CSRF-Check": "1",
        },
        body: JSON.stringify(payload),
      });

      const raw = await resp.text();
      let data = null;
      try {
        data = raw ? JSON.parse(raw) : null;
      } catch (_) {}

      if (!resp.ok) {
        const message = data && typeof data === "object" && data.error
          ? String(data.error)
          : "Не удалось отправить форму.";
        throw new Error(humanizeError(message));
      }

      form.reset();
      status.textContent = "Сообщение отправлено. Я свяжусь с вами в ближайшее время.";
    } catch (err) {
      status.textContent = humanizeError(err && err.message);
    } finally {
      if (submitBtn) {
        submitBtn.disabled = false;
        submitBtn.textContent = defaultSubmitLabel || "Отправить";
      }
    }
  });
})();
