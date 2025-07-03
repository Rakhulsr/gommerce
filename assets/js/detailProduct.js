document.addEventListener("DOMContentLoaded", function () {
  const price = parseFloat(document.getElementById("productPrice").dataset.value); 
  const quantityInput = document.getElementById("quantity");
  const subtotalEl = document.getElementById("subtotal");

  window.setMainImage = function (imageUrl) {
    const mainImg = document.getElementById("mainImage");
    if (mainImg && imageUrl) {
      mainImg.src = imageUrl;
    }
  }

  function updateSubtotal() {
    const qty = parseInt(quantityInput.value) || 1;
    subtotalEl.textContent = new Intl.NumberFormat("id-ID", {
      style: "currency",
      currency: "IDR",
    }).format(price * qty);
  }

  window.increaseQuantity = function () {
    quantityInput.value = parseInt(quantityInput.value) + 1;
    updateSubtotal();
  }

  window.decreaseQuantity = function () {
    if (parseInt(quantityInput.value) > 1) {
      quantityInput.value = parseInt(quantityInput.value) - 1;
      updateSubtotal();
    }
  }

  quantityInput.addEventListener("input", updateSubtotal);

  updateSubtotal();
});
