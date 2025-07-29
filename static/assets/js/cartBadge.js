document.addEventListener("DOMContentLoaded", async () => {
  const badge = document.getElementById("cart-count-badge");

  try {
    const res = await fetch("/carts/count");
    const data = await res.json();

    if (data.count > 0 && badge) {
      badge.textContent = data.count;
      badge.classList.remove("hidden");
      badge.style.opacity = count === "0" ? "0" : "1";

    }
  } catch (err) {
    console.error("Gagal mengambil jumlah cart:", err);
  }
});
