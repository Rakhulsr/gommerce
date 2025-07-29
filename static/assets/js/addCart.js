function setAction(value) {
    document.getElementById("actionInput").value = value;
  }

async function updateCartCount() {
  const res = await fetch("/carts/count");
  if (res.ok) {
    const count = await res.text();
    const badge = document.getElementById("cart-count-badge");
    badge.textContent = count;
    badge.classList.toggle("hidden", count === "0");
  }
}
