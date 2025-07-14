document.addEventListener("DOMContentLoaded", function () {
 
  console.log("✅ JS Loaded!");
  
  // Pastikan elemen 'product-page' ada di DOM
  const productPage = document.getElementById("product-page");

  if (productPage) { // Kondisi ini sekarang akan terpenuhi
    const priceElement = document.getElementById("productPrice");
    const quantityInput = document.getElementById("qty");
    const subtotal = document.getElementById("subtotal");

    if (!priceElement || !quantityInput || !subtotal) {
      console.warn("Elemen produk tidak ditemukan:", {
        priceElement,
        quantityInput,
        subtotal,
      });
    } else {
      const price = parseFloat(priceElement.dataset.value) || 0;

      window.setMainImage = function (imageUrl) {
        const mainImg = document.getElementById("mainImage");
        if (mainImg && imageUrl) {
          mainImg.src = imageUrl;
        }
      };

      function updateSubtotal() {
        const qty = parseInt(quantityInput.value) || 1;
        subtotal.textContent = new Intl.NumberFormat("id-ID", {
          style: "currency",
          currency: "IDR",
        }).format(price * qty);
      }

      window.increaseQuantity = function () {
        quantityInput.value = parseInt(quantityInput.value || "1") + 1;
        updateSubtotal();
      };

      window.decreaseQuantity = function () {
        const currentQty = parseInt(quantityInput.value || "1");
        if (currentQty > 1) {
          quantityInput.value = currentQty - 1;
          updateSubtotal();
        }
      };

      quantityInput.addEventListener("input", updateSubtotal);
      updateSubtotal(); // Panggil sekali saat dimuat untuk inisialisasi subtotal
    }
  }


  const slider = document.querySelector("#slider");
  if (slider) {
    const slides = slider.querySelectorAll(".slides img");
    let currentSlide = 0;

    function showSlide(index) {
      slides.forEach((slide, i) => {
        slide.classList.toggle("opacity-100", i === index);
        slide.classList.toggle("opacity-0", i !== index);
      });
      currentSlide = index;
    }

    function nextSlide() {
      const nextIndex = (currentSlide + 1) % slides.length;
      showSlide(nextIndex);
    }

    if (slides.length > 0) {
      window.showSlide = showSlide;
      showSlide(currentSlide);
      setInterval(nextSlide, 5000);
    }
  }

});

