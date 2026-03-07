(() => {
  const form = document.getElementById("authorContactForm");
  const status = document.getElementById("authorFormStatus");

  if (!form || !status) return;

  form.addEventListener("submit", (event) => {
    event.preventDefault();
    status.textContent = "Спасибо! Aibolat свяжется с вами в ближайшее время.";
  });
})();
