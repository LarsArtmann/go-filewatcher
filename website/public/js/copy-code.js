const btn = document.querySelector("#copy-btn");
if (btn) {
  btn.addEventListener("click", async () => {
    const code = btn.dataset.code;
    try {
      await navigator.clipboard.writeText(code || "");
      btn.textContent = "Copied!";
      btn.classList.add("text-success");
      setTimeout(() => {
        btn.textContent = "Copy";
        btn.classList.remove("text-success");
      }, 2000);
    } catch {
      btn.textContent = "Failed";
      setTimeout(() => {
        btn.textContent = "Copy";
      }, 2000);
    }
  });
}
