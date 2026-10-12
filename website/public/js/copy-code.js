const COPY_RESET_DELAY_MS = 2000;

const btn = document.querySelector("#copy-btn");
if (btn) {
  btn.addEventListener("click", async () => {
    const { code } = btn.dataset;
    try {
      await navigator.clipboard.writeText(code || "");
      btn.textContent = "Copied!";
      btn.classList.add("text-success");
      setTimeout(() => {
        btn.textContent = "Copy";
        btn.classList.remove("text-success");
      }, COPY_RESET_DELAY_MS);
    } catch {
      btn.textContent = "Failed";
      setTimeout(() => {
        btn.textContent = "Copy";
      }, COPY_RESET_DELAY_MS);
    }
  });
}
